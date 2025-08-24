package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/utils"
)

// GenerateDescriptions processes clients to generate session descriptions using git analysis
func (s *TimesheetService) GenerateDescriptions(ctx context.Context, clientName, sessionID string, update bool) error {
	if sessionID != "" {
		return s.processSession(ctx, sessionID, update)
	}

	clients, err := s.getTargetClients(ctx, clientName)
	if err != nil {
		return err
	}

	if len(clients) == 0 {
		fmt.Println("No clients with directories found.")
		return nil
	}

	var wg sync.WaitGroup
	for _, client := range clients {
		sessions, err := s.db.GetSessionsWithoutDescription(ctx, &client.Name, nil)
		if err != nil {
			fmt.Printf("Error getting sessions for client %s: %v\n", client.Name, err)
			continue
		}

		if len(sessions) == 0 {
			fmt.Printf("No sessions missing descriptions for client: %s\n", client.Name)
			continue
		}

		for _, session := range sessions {
			wg.Add(1)
			go func(sess *models.WorkSession) {
				defer wg.Done()
				s.processSessionWithClient(ctx, sess, client, update)
			}(session)
		}
	}

	wg.Wait()
	return nil
}

// DescriptionResult contains both the final summary and full work details
type DescriptionResult struct {
	FinalSummary    string
	FullWorkSummary string
}

var (
	ErrSessionNotFinished       = errors.New("session is not finished")
	ErrConfiguredClientRequired = errors.New("client with a configured dir is required")
)

// RepositoryResult holds the result of analyzing a single git repository
type RepositoryResult struct {
	RepoPath string
	Output   string
	Error    error
}

func (s *TimesheetService) processSession(ctx context.Context, sessionID string, update bool) error {
	session, err := s.db.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session '%s': %w", sessionID, err)
	}
	if session == nil {
		return fmt.Errorf("session '%s' does not exist", sessionID)
	}

	client, err := s.db.GetClientByID(ctx, session.ClientID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	return s.processSessionWithClient(ctx, session, client, update)
}

func (s *TimesheetService) processSessionWithClient(ctx context.Context, session *models.WorkSession, client *models.Client, update bool) error {
	if session.EndTime == nil {
		fmt.Printf("  Skipping active session %s (not ended)\n", session.ID)
		return nil
	}

	fmt.Printf("  Processing session %s (%s to %s)\n",
		session.ID,
		session.StartTime.Format("2006-01-02 15:04"),
		session.EndTime.Format("2006-01-02 15:04"))

	result, err := s.analyzeSession(ctx, client, session)
	if err != nil {
		fmt.Printf("    Error analyzing session: %v\n", err)
		return err
	}

	if update {
		_, err = s.db.UpdateSessionDescription(ctx, session.ID, result.FinalSummary, &result.FullWorkSummary)
		if err != nil {
			return fmt.Errorf("failed to update session description: %w", err)
		}
	}

	return nil
}

func (s *TimesheetService) getTargetClients(ctx context.Context, clientName string) ([]*models.Client, error) {
	if clientName != "" {
		client, err := s.db.GetClientByName(ctx, clientName)
		if err != nil {
			return nil, fmt.Errorf("failed to get client '%s': %w", clientName, err)
		}

		if client.Dir == nil || *client.Dir == "" {
			return nil, fmt.Errorf("client '%s' does not have a directory configured", clientName)
		}

		return []*models.Client{client}, nil
	}

	clients, err := s.db.GetClientsWithDirectories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get clients with directories: %w", err)
	}

	return clients, nil
}

// Replace the placeholder analyzeSession method with the real implementation
func (s *TimesheetService) analyzeSession(ctx context.Context, client *models.Client, session *models.WorkSession) (*DescriptionResult, error) {
	if session.EndTime == nil {
		return nil, ErrSessionNotFinished
	}

	// Create temp directory for this session analysis
	tempDir, err := os.MkdirTemp("", "work-analyze-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clean up temp directory after processing
	defer os.RemoveAll(tempDir)

	// Run the analysis for this specific client and time period
	result, err := s.performAnalysis(session.StartTime, *session.EndTime, client, tempDir)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// performAnalysis runs the git analysis and returns structured results for a single client
func (s *TimesheetService) performAnalysis(fromDate, toDate time.Time, client *models.Client, tempDir string) (*DescriptionResult, error) {
	if client == nil || utils.FromPtr(client.Dir) == "" {
		return nil, ErrConfiguredClientRequired
	}

	// Process the client directory
	err := s.processDirectory(client.Name, *client.Dir, fromDate, toDate, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to process directory: %w", err)
	}

	// Generate brief description for the session
	briefDescription, err := s.generateBriefDescription(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate brief description: %w", err)
	}

	// Generate detailed full work summary
	fullWorkSummary, err := s.generateDetailedSummary(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate detailed summary: %w", err)
	}

	return &DescriptionResult{
		FinalSummary:    briefDescription,
		FullWorkSummary: fullWorkSummary,
	}, nil
}

// processDirectory finds git repositories in the client directory and analyzes each one
func (s *TimesheetService) processDirectory(clientName, dir string, fromDate, toDate time.Time, tempDir string) error {
	// Trim whitespace from the directory path
	dir = strings.TrimSpace(dir)
	if strings.HasPrefix(dir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %v", err)
		}
		dir = filepath.Join(homeDir, dir[2:])
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// Find all git repositories in subdirectories
	gitRepos := s.findGitRepositories(dir)

	if len(gitRepos) == 0 {
		return fmt.Errorf("no git repositories found in %s", dir)
	}

	// Process each git repository in parallel
	var wg sync.WaitGroup
	results := make(chan RepositoryResult, len(gitRepos))

	for _, repoDir := range gitRepos {
		wg.Add(1)
		go func(repoPath string) {
			defer wg.Done()
			result := s.analyzeGitRepository(repoPath, fromDate, toDate)
			results <- result
		}(repoDir)
	}

	// Wait for all repositories to be processed
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	var allResults []RepositoryResult
	for result := range results {
		allResults = append(allResults, result)
	}

	// Combine results into a single output
	combinedOutput := s.combineRepositoryResults(clientName, allResults)

	// Write combined output to file
	outputFile := filepath.Join(tempDir, s.sanitizeClientName(clientName, fromDate, toDate)+".txt")
	err := os.WriteFile(outputFile, []byte(combinedOutput), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file for %s: %v", clientName, err)
	}
	return nil
}

// sanitizeClientName creates a safe filename from client name
func (s *TimesheetService) sanitizeClientName(clientName string, fromDate, toDate time.Time) string {
	// Replace spaces and special characters with underscores
	result := strings.ReplaceAll(clientName, " ", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "\\", "_")
	result = strings.ReplaceAll(result, ":", "_")
	result = fmt.Sprintf("%s_%s_%s", result, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
	return result
}

// findGitRepositories searches for .git directories in the given directory and its subdirectories
// Uses find command with time-based filtering to only check recently modified repositories
func (s *TimesheetService) findGitRepositories(root string) []string {
	var gitRepos []string

	// Use find command to locate .git directories modified in the last 30 days
	// This is much faster than walking through all directories
	cmd := exec.Command("find", root, "-type", "d", "-name", ".git", "-mtime", "-30", "-maxdepth", "3")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  Warning: find command failed, falling back to directory walk: %v\n", err)
		return s.findGitRepositoriesWalk(root)
	}

	// Parse find output to get repository directories
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" {
			// Get the parent directory (the actual repository directory)
			repoDir := filepath.Dir(line)
			gitRepos = append(gitRepos, repoDir)
		}
	}

	// If no recently modified repos found, also check for repos with recent commits
	if len(gitRepos) == 0 {
		fmt.Printf("  No recently modified .git directories found, checking for repos with recent commits...\n")
		gitRepos = s.findGitRepositoriesWithRecentCommits(root)
	}

	return gitRepos
}

// findGitRepositoriesWalk is the original implementation as fallback
func (s *TimesheetService) findGitRepositoriesWalk(root string) []string {
	var gitRepos []string
	maxDepth := 2

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		depth := len(strings.Split(rel, string(filepath.Separator)))

		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			// Add the parent directory (the actual repository directory)
			repoDir := filepath.Dir(path)
			gitRepos = append(gitRepos, repoDir)
			return filepath.SkipDir // Don't traverse into .git directory
		}

		return nil
	})

	return gitRepos
}

// findGitRepositoriesWithRecentCommits finds git repos that have commits in the last month
func (s *TimesheetService) findGitRepositoriesWithRecentCommits(root string) []string {
	var gitRepos []string
	maxDepth := 2

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		depth := len(strings.Split(rel, string(filepath.Separator)))

		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoDir := filepath.Dir(path)

			// Check if this repo has commits in the last month
			cmd := exec.Command("git", "-C", repoDir, "log", "--since=1 month ago", "--oneline", "-n", "1")
			output, err := cmd.Output()
			if err == nil && len(strings.TrimSpace(string(output))) > 0 {
				gitRepos = append(gitRepos, repoDir)
			}

			return filepath.SkipDir // Don't traverse into .git directory
		}

		return nil
	})

	return gitRepos
}

// analyzeGitRepository runs git analysis on a single repository
func (s *TimesheetService) analyzeGitRepository(repoDir string, fromDate, toDate time.Time) RepositoryResult {
	// Create prompt with actual dates
	prompt := strings.ReplaceAll(s.cfg.GitAnalysisPrompt, "{from_date}", fromDate.Format("2006-01-02 15:04"))
	prompt = strings.ReplaceAll(prompt, "{to_date}", toDate.Format("2006-01-02 15:04"))

	// Create the shell command to cd into repository directory and run opencode
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && echo %s | opencode run",
		s.shellescape(repoDir),
		s.shellescape(prompt)))

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()

	return RepositoryResult{
		RepoPath: repoDir,
		Output:   string(output),
		Error:    err,
	}
}

// combineRepositoryResults combines results from multiple repositories into a single output
func (s *TimesheetService) combineRepositoryResults(clientName string, results []RepositoryResult) string {
	if len(results) == 0 {
		return "NO COMMITS"
	}

	var combinedOutput strings.Builder
	hasContent := false

	for _, result := range results {
		repoName := filepath.Base(result.RepoPath)

		if result.Error != nil {
			combinedOutput.WriteString(fmt.Sprintf("ERROR analyzing %s: %v\n", repoName, result.Error))
			continue
		}

		// Check if this repository had any commits (not just "NO COMMITS")
		if strings.TrimSpace(result.Output) != "" && !strings.Contains(strings.ToUpper(result.Output), "NO COMMITS") {
			combinedOutput.WriteString(fmt.Sprintf("=== %s ===\n", repoName))
			// Clean the output before adding it to remove ANSI codes and tool invocations
			cleanedOutput := s.cleanRepositoryOutput(result.Output)
			combinedOutput.WriteString(cleanedOutput)
			combinedOutput.WriteString("\n\n")
			hasContent = true
		}
	}

	if !hasContent {
		return "NO COMMITS"
	}

	return combinedOutput.String()
}

// cleanRepositoryOutput removes ANSI codes and tool invocations from repository analysis output
func (s *TimesheetService) cleanRepositoryOutput(output string) string {
	lines := strings.Split(output, "\n")
	var cleanLines []string
	inToolOutput := false
	seenLines := make(map[string]bool) // Track seen lines to remove duplicates

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if trimmedLine == "" {
			continue
		}

		// Detect tool output lines (they contain ANSI codes and tool indicators)
		if strings.Contains(line, "[0m") ||
			strings.Contains(line, "[90m") ||
			strings.Contains(line, "[94m") ||
			strings.Contains(line, "[96m") ||
			strings.Contains(line, "[91m") ||
			strings.Contains(line, "[1m|") ||
			strings.Contains(line, "Glob") ||
			strings.Contains(line, "Read") ||
			strings.Contains(line, "Bash") ||
			strings.Contains(line, "{\"pattern\":") {
			inToolOutput = true
			continue
		}

		// If we were in tool output and now we're not, we've found the actual content
		if inToolOutput && !strings.Contains(line, "\033") {
			inToolOutput = false
		}

		// Skip tool output lines
		if inToolOutput {
			continue
		}

		// Clean any remaining ANSI codes from the actual content
		cleaned := strings.ReplaceAll(line, "\033[0m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[90m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[94m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[96m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[91m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[1m", "")

		// Add the cleaned line if it has content and hasn't been seen before
		if strings.TrimSpace(cleaned) != "" {
			if !seenLines[cleaned] {
				cleanLines = append(cleanLines, strings.TrimSpace(cleaned))
				seenLines[cleaned] = true
			}
		}
	}

	// Join and return the clean content
	return strings.Join(cleanLines, "\n")
}

// shellescape escapes a string for safe use in shell commands
func (s *TimesheetService) shellescape(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "'\"'\"'") + "'"
}

// generateBriefDescription creates a concise 1-2 sentence description suitable for a line item
func (s *TimesheetService) generateBriefDescription(tempDir string) (string, error) {
	briefPrompt := "Read all .txt files in this directory and provide ONLY a single, concise line item description (maximum 1-2 sentences) of the work done. Focus on business value, not technical details. Do not show your thinking or tool usage. Output only the final description. If no work was done, respond 'No development activity'."

	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && echo %s | opencode run",
		s.shellescape(tempDir),
		s.shellescape(briefPrompt)))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate brief description: %v\nOutput: %s", err, string(output))
	}

	return s.cleanOpenCodeOutput(string(output)), nil
}

// cleanOpenCodeOutput removes OpenCode tool invocations and ANSI codes, returning only the final content
func (s *TimesheetService) cleanOpenCodeOutput(output string) string {
	lines := strings.Split(output, "\n")
	var cleanLines []string
	inToolOutput := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if trimmedLine == "" {
			continue
		}

		// Detect tool output lines (they contain ANSI codes and tool indicators)
		if strings.Contains(line, "[0m") ||
			strings.Contains(line, "[90m") ||
			strings.Contains(line, "[94m") ||
			strings.Contains(line, "[96m") ||
			strings.Contains(line, "[91m") ||
			strings.Contains(line, "[1m|") ||
			strings.Contains(line, "Glob") ||
			strings.Contains(line, "Read") ||
			strings.Contains(line, "Bash") ||
			strings.Contains(line, "{\"pattern\":") {
			inToolOutput = true
			continue
		}

		// If we were in tool output and now we're not, we've found the actual content
		if inToolOutput && !strings.Contains(line, "\033") {
			inToolOutput = false
		}

		// Skip tool output lines
		if inToolOutput {
			continue
		}

		// Clean any remaining ANSI codes from the actual content
		cleaned := strings.ReplaceAll(line, "\033[0m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[90m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[94m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[96m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[91m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[1m", "")

		// Add the cleaned line if it has content
		if strings.TrimSpace(cleaned) != "" {
			cleanLines = append(cleanLines, strings.TrimSpace(cleaned))
		}
	}

	// Join and return the clean content
	result := strings.Join(cleanLines, " ")
	result = strings.TrimSpace(result)

	// Remove duplicate phrases (simple deduplication)
	words := strings.Fields(result)
	if len(words) > 0 {
		// Check for repeated phrases
		half := len(words) / 2
		if half > 0 {
			firstHalf := strings.Join(words[:half], " ")
			secondHalf := strings.Join(words[half:], " ")
			if firstHalf == secondHalf {
				return firstHalf
			}
		}
	}

	return result
}

// generateDetailedSummary creates a comprehensive summary for the full work summary field
func (s *TimesheetService) generateDetailedSummary(tempDir string) (string, error) {
	contents, err := os.ReadDir(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory contents: %w", err)
	}

	var builder strings.Builder
	for _, file := range contents {
		fileContents, err := os.ReadFile(filepath.Join(tempDir, file.Name()))
		if err != nil {
			return "", fmt.Errorf("failed to read file contents: %w", err)
		}
		builder.WriteString(string(file.Name()))
		builder.WriteString("\n")
		builder.WriteString(string(fileContents))
	}
	return builder.String(), nil
}
