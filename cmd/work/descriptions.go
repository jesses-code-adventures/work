package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
	"github.com/jesses-code-adventures/work/internal/utils"
)

func newDescriptionsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "descriptions",
		Short: "Manage session descriptions using git and AI summarization",
		Long:  "Commands for managing and populating session descriptions using git analysis.",
	}

	cmd.AddCommand(newDescriptionsGenerateCmd(timesheetService))

	return cmd
}

func newDescriptionsGenerateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var client string
	var period string
	var date string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate missing session descriptions using git analysis",
		Long:  "Gets all sessions missing descriptions and runs summarize analysis using the session start/end times to populate descriptions and full work summaries.",
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Process only the specified client (optional)")
	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in the period (YYYY-MM-DD)")
	update := cmd.Flags().BoolP("update", "u", false, "Update the session descriptions in the database")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if utils.FromPtr(update) {
			return populateDescriptions(ctx, timesheetService, client)
		}
		return summarizeDescriptions(ctx, timesheetService, period, date, client)
	}

	return cmd
}

func populateDescriptions(ctx context.Context, timesheetService *service.TimesheetService, clientName string) error {
	// Get clients with directories
	var clients []*models.Client
	var err error

	if clientName != "" {
		// Get specific client by name
		client, err := timesheetService.GetClientByName(ctx, clientName)
		if err != nil {
			return fmt.Errorf("failed to get client '%s': %w", clientName, err)
		}

		// Check if client has a directory
		if client.Dir == nil || *client.Dir == "" {
			return fmt.Errorf("client '%s' does not have a directory configured", clientName)
		}

		clients = []*models.Client{client}
		fmt.Printf("Processing single client: %s\n", clientName)
	} else {
		// Get all clients that have directories
		clients, err = timesheetService.GetClientsWithDirectories(ctx)
		if err != nil {
			return fmt.Errorf("failed to get clients with directories: %w", err)
		}

		if len(clients) == 0 {
			fmt.Println("No clients with directories found.")
			return nil
		}
	}

	fmt.Printf("Found %d clients with directories\n", len(clients))

	// For each client, get sessions without descriptions
	totalProcessed := 0
	for _, client := range clients {
		sessions, err := timesheetService.GetSessionsWithoutDescription(ctx, &client.Name)
		if err != nil {
			fmt.Printf("Error getting sessions for client %s: %v\n", client.Name, err)
			continue
		}

		if len(sessions) == 0 {
			fmt.Printf("No sessions missing descriptions for client: %s\n", client.Name)
			continue
		}

		fmt.Printf("Processing %d sessions for client: %s\n", len(sessions), client.Name)

		for _, session := range sessions {
			if session.EndTime == nil {
				fmt.Printf("  Skipping active session %s (not ended)\n", session.ID)
				continue
			}

			fmt.Printf("  Processing session %s (%s to %s)\n",
				session.ID,
				session.StartTime.Format("2006-01-02 15:04"),
				session.EndTime.Format("2006-01-02 15:04"))

			// Run summarize analysis for this session's time period
			err := analyzeAndUpdateSession(ctx, timesheetService, client, session)
			if err != nil {
				fmt.Printf("    Error analyzing session: %v\n", err)
				continue
			}

			totalProcessed++
			fmt.Printf("    Successfully updated session description\n")
		}
	}

	fmt.Printf("\nCompleted! Processed %d sessions total.\n", totalProcessed)
	return nil
}

func analyzeAndUpdateSession(ctx context.Context, timesheetService *service.TimesheetService, client *models.Client, session *models.WorkSession) error {
	// Calculate the period as "day" and use the session start date
	period := "day"
	date := session.StartTime.Format("2006-01-02")

	// Run the summarize analysis for this specific client and time period
	result, err := performSummarizeAnalysis(ctx, timesheetService, period, date, client.Name)
	if err != nil {
		return fmt.Errorf("failed to perform analysis: %w", err)
	}

	// Update the session with the results
	_, err = timesheetService.UpdateSessionDescription(ctx, session.ID, result.FinalSummary, &result.FullWorkSummary)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func summarizeDescriptions(ctx context.Context, timesheetService *service.TimesheetService, period string, date string, clientName string) error {
	// Default to today if no date specified
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Parse the date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Calculate date range based on period
	fromDate, toDate := calculatePeriodRange(period, targetDate)

	// Get clients that have directories
	var clients []*models.Client

	if clientName != "" {
		// Get specific client by name
		client, err := timesheetService.GetClientByName(ctx, clientName)
		if err != nil {
			return fmt.Errorf("failed to get client '%s': %w", clientName, err)
		}

		// Check if client has a directory
		if client.Dir == nil || *client.Dir == "" {
			return fmt.Errorf("client '%s' does not have a directory configured", clientName)
		}

		clients = []*models.Client{client}
		fmt.Printf("Processing single client: %s\n", clientName)
	} else {
		// Get all clients that have directories
		var err error
		clients, err = timesheetService.GetClientsWithDirectories(ctx)
		if err != nil {
			return fmt.Errorf("failed to get clients with directories: %w", err)
		}

		if len(clients) == 0 {
			fmt.Println("No clients with directories found.")
			return nil
		}
	}

	fmt.Printf("Found %d clients with directories for period: %s %s\n", len(clients), period, date)
	fmt.Printf("Date range: %s to %s\n", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Create temp directory for storing outputs
	var tempDir string

	if timesheetService.Config().DevMode {
		// In dev mode, create a local directory that persists
		tempDir = "./work-summarize-temp"
		err := os.MkdirAll(tempDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create dev temp directory: %w", err)
		}
		// Clean up existing files but keep directory
		files, _ := filepath.Glob(filepath.Join(tempDir, "*.txt"))
		for _, file := range files {
			os.Remove(file)
		}
		fmt.Printf("Using persistent temp directory (dev mode): %s\n", tempDir)
	} else {
		// In prod mode, use system temp directory
		var err error
		tempDir, err = os.MkdirTemp("", "work-summarize-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tempDir)
		fmt.Printf("Using temp directory: %s\n", tempDir)
	}

	// Process directories concurrently
	var wg sync.WaitGroup
	for _, client := range clients {
		if client.Dir != nil {
			wg.Add(1)
			go func(clientName, dir string) {
				defer wg.Done()
				processDirectory(clientName, dir, fromDate, toDate, tempDir, timesheetService)
			}(client.Name, *client.Dir)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()
	fmt.Println("All directories processed.")

	// Final summarization step
	err = generateFinalSummary(tempDir)
	if err != nil {
		return fmt.Errorf("failed to generate final summary: %w", err)
	}

	return nil
}

// processDirectory finds git repositories in the client directory and analyzes each one
func processDirectory(clientName, dir string, fromDate, toDate time.Time, tempDir string, timesheetService *service.TimesheetService) {
	fmt.Printf("Processing directory for client '%s': %s\n", clientName, dir)
	fmt.Printf("  Date range: %s to %s\n", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Expand tilde in directory path
	if strings.HasPrefix(dir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("  Error getting home directory: %v\n", err)
			return
		}
		dir = filepath.Join(homeDir, dir[2:])
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("  Directory does not exist: %s\n", dir)
		// Write "NO GIT ACTIVITY" to output file for missing directories
		outputFile := filepath.Join(tempDir, sanitizeClientName(clientName)+".txt")
		os.WriteFile(outputFile, []byte("NO GIT ACTIVITY"), 0644)
		return
	}

	// Find all git repositories in subdirectories
	gitRepos := findGitRepositories(dir)

	if len(gitRepos) == 0 {
		fmt.Printf("  No git repositories found in %s\n", dir)
		outputFile := filepath.Join(tempDir, sanitizeClientName(clientName)+".txt")
		os.WriteFile(outputFile, []byte("NO COMMITS"), 0644)
		return
	}

	fmt.Printf("  Found %d git repositories: %v\n", len(gitRepos), gitRepos)

	// Process each git repository in parallel
	var wg sync.WaitGroup
	results := make(chan RepositoryResult, len(gitRepos))

	for _, repoDir := range gitRepos {
		wg.Add(1)
		go func(repoPath string) {
			defer wg.Done()
			result := analyzeGitRepository(repoPath, fromDate, toDate, timesheetService)
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
	combinedOutput := combineRepositoryResults(clientName, allResults)

	// Write combined output to file
	outputFile := filepath.Join(tempDir, sanitizeClientName(clientName)+".txt")
	err := os.WriteFile(outputFile, []byte(combinedOutput), 0644)
	if err != nil {
		fmt.Printf("  Error writing output file for %s: %v\n", clientName, err)
		return
	}

	fmt.Printf("  Analysis complete for %s, output written to %s\n", clientName, outputFile)
}

// generateFinalSummary processes all individual client analyses and generates a final summary
func generateFinalSummary(tempDir string) error {
	fmt.Println("Generating final summary...")

	// Create a more specific prompt that focuses on the content of the files
	finalPrompt := "I have individual client analysis files in this directory. Each file contains git activity analysis for a specific client and time period. Please read all the .txt files in this directory and create a single invoice description summarizing what work was done across all clients. Focus on the actual work described in the files, not on analyzing git repositories. If all files indicate no commits or no git activity, return 'NO GIT ACTIVITY'."

	// Create the shell command to cd into temp directory and run opencode
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && echo %s | opencode run",
		shellescape(tempDir),
		shellescape(finalPrompt)))

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate final summary: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("\n=== FINAL SUMMARY ===")
	fmt.Printf("%s\n", string(output))
	fmt.Println("===================")

	return nil
}

// sanitizeClientName creates a safe filename from client name
func sanitizeClientName(clientName string) string {
	// Replace spaces and special characters with underscores
	result := strings.ReplaceAll(clientName, " ", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "\\", "_")
	result = strings.ReplaceAll(result, ":", "_")
	return result
}

// RepositoryResult holds the result of analyzing a single git repository
type RepositoryResult struct {
	RepoPath string
	Output   string
	Error    error
}

// findGitRepositories searches for .git directories in the given directory and its subdirectories
// Uses find command with time-based filtering to only check recently modified repositories
func findGitRepositories(root string) []string {
	var gitRepos []string

	// Use find command to locate .git directories modified in the last 7 days
	// This is much faster than walking through all directories
	cmd := exec.Command("find", root, "-type", "d", "-name", ".git", "-mtime", "-7", "-maxdepth", "3")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  Warning: find command failed, falling back to directory walk: %v\n", err)
		return findGitRepositoriesWalk(root)
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
		gitRepos = findGitRepositoriesWithRecentCommits(root)
	}

	return gitRepos
}

// findGitRepositoriesWalk is the original implementation as fallback
func findGitRepositoriesWalk(root string) []string {
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

// findGitRepositoriesWithRecentCommits finds git repos that have commits in the last week
func findGitRepositoriesWithRecentCommits(root string) []string {
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

			// Check if this repo has commits in the last week
			cmd := exec.Command("git", "-C", repoDir, "log", "--since=1 week ago", "--oneline", "-n", "1")
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
func analyzeGitRepository(repoDir string, fromDate, toDate time.Time, timesheetService *service.TimesheetService) RepositoryResult {
	// Create prompt with actual dates
	prompt := strings.ReplaceAll(timesheetService.Config().GitAnalysisPrompt, "{from_date}", fromDate.Format("2006-01-02"))
	prompt = strings.ReplaceAll(prompt, "{to_date}", toDate.Format("2006-01-02"))

	// Create the shell command to cd into repository directory and run opencode
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && echo %s | opencode run",
		shellescape(repoDir),
		shellescape(prompt)))

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()

	return RepositoryResult{
		RepoPath: repoDir,
		Output:   string(output),
		Error:    err,
	}
}

// combineRepositoryResults combines results from multiple repositories into a single output
func combineRepositoryResults(clientName string, results []RepositoryResult) string {
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
			combinedOutput.WriteString(result.Output)
			combinedOutput.WriteString("\n\n")
			hasContent = true
		}
	}

	if !hasContent {
		return "NO COMMITS"
	}

	return combinedOutput.String()
}

// shellescape escapes a string for safe use in shell commands
func shellescape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// SummarizeResult contains both the final summary and full work details
type SummarizeResult struct {
	FinalSummary    string
	FullWorkSummary string
}

// performSummarizeAnalysis runs the summarize analysis and returns structured results for a single client
func performSummarizeAnalysis(ctx context.Context, timesheetService *service.TimesheetService, period string, date string, clientName string) (*SummarizeResult, error) {
	// Default to today if no date specified
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Parse the date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Calculate date range based on period
	fromDate, toDate := calculatePeriodRange(period, targetDate)

	// Get specific client by name
	client, err := timesheetService.GetClientByName(ctx, clientName)
	if err != nil {
		return nil, fmt.Errorf("failed to get client '%s': %w", clientName, err)
	}

	// Check if client has a directory
	if client.Dir == nil || *client.Dir == "" {
		return nil, fmt.Errorf("client '%s' does not have a directory configured", clientName)
	}

	// Create temp directory for storing outputs
	tempDir, err := os.MkdirTemp("", "work-analyze-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Process the single client directory
	var wg sync.WaitGroup
	wg.Add(1)
	go func(clientName, dir string) {
		defer wg.Done()
		processDirectory(clientName, dir, fromDate, toDate, tempDir, timesheetService)
	}(client.Name, *client.Dir)

	// Wait for processing to complete
	wg.Wait()

	// Generate brief description for the session
	briefDescription, err := generateBriefDescription(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate brief description: %w", err)
	}

	// Generate detailed full work summary
	fullWorkSummary, err := generateDetailedSummary(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate detailed summary: %w", err)
	}

	return &SummarizeResult{
		FinalSummary:    briefDescription,
		FullWorkSummary: fullWorkSummary,
	}, nil
}

// generateFinalSummaryString generates final summary and returns it as string (without printing)
func generateFinalSummaryString(tempDir string) (string, error) {
	// Create a more specific prompt that focuses on the content of the files
	finalPrompt := "I have individual client analysis files in this directory. Each file contains git activity analysis for a specific client and time period. Please read all the .txt files in this directory and create a single invoice description summarizing what work was done across all clients. Focus on the actual work described in the files, not on analyzing git repositories. If all files indicate no commits or no git activity, return 'NO GIT ACTIVITY'."

	// Create the shell command to cd into temp directory and run opencode
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && echo %s | opencode run",
		shellescape(tempDir),
		shellescape(finalPrompt)))

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate final summary: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// generateBriefDescription creates a concise 1-2 sentence description suitable for a line item
func generateBriefDescription(tempDir string) (string, error) {
	briefPrompt := "Read all .txt files in this directory and provide ONLY a single, concise line item description (maximum 1-2 sentences) of the work done. Focus on business value, not technical details. Do not show your thinking or tool usage. Output only the final description. If no work was done, respond 'No development activity'."

	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && echo %s | opencode run",
		shellescape(tempDir),
		shellescape(briefPrompt)))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate brief description: %v\nOutput: %s", err, string(output))
	}

	return cleanOpenCodeOutput(string(output)), nil
}

// cleanOpenCodeOutput removes OpenCode tool invocations and ANSI codes, returning only the final content
func cleanOpenCodeOutput(output string) string {
	lines := strings.Split(output, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Skip lines with ANSI color codes and tool indicators
		if strings.Contains(line, "[0m") ||
			strings.Contains(line, "[90m") ||
			strings.Contains(line, "[94m") ||
			strings.Contains(line, "[96m") ||
			strings.Contains(line, "[91m") ||
			strings.Contains(line, "[1m|") ||
			strings.Contains(line, "Glob") ||
			strings.Contains(line, "Read") ||
			strings.Contains(line, "Bash") ||
			strings.Contains(line, "{\"pattern\":") ||
			strings.TrimSpace(line) == "" {
			continue
		}

		// Clean any remaining ANSI codes
		cleaned := strings.ReplaceAll(line, "\033[0m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[90m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[94m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[96m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[91m", "")
		cleaned = strings.ReplaceAll(cleaned, "\033[1m", "")

		if strings.TrimSpace(cleaned) != "" {
			cleanLines = append(cleanLines, strings.TrimSpace(cleaned)+"\n")
		}
	}

	// Join and return the clean content, preserving line structure
	result := strings.Join(cleanLines, "\n")
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
func generateDetailedSummary(tempDir string) (string, error) {
	detailedPrompt := "Read all .txt files in this directory and provide ONLY a comprehensive summary of all work performed. Include technical details, specific changes made, and context. Organize by repository/area if multiple areas were worked on. Do not show your thinking or tool usage. Output only the final detailed summary. If no work was done, respond 'No development activity during this period'."

	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && echo %s | opencode run",
		shellescape(tempDir),
		shellescape(detailedPrompt)))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate detailed summary: %v\nOutput: %s", err, string(output))
	}

	return cleanOpenCodeOutput(string(output)), nil
}

func calculatePeriodRange(period string, targetDate time.Time) (time.Time, time.Time) {
	switch period {
	case "day":
		start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		end := start.Add(24*time.Hour - time.Nanosecond)
		return start, end

	case "week":
		// Find Monday of the week containing targetDate
		daysFromMonday := int(targetDate.Weekday()-time.Monday+7) % 7
		monday := targetDate.AddDate(0, 0, -daysFromMonday)
		start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
		return start, end

	case "fortnight":
		// Find Monday of the week containing targetDate, then determine if it's first or second week
		daysFromMonday := int(targetDate.Weekday()-time.Monday+7) % 7
		monday := targetDate.AddDate(0, 0, -daysFromMonday)

		// Find the first Monday of the month
		firstOfMonth := time.Date(monday.Year(), monday.Month(), 1, 0, 0, 0, 0, monday.Location())
		daysToFirstMonday := int(time.Monday-firstOfMonth.Weekday()+7) % 7
		firstMonday := firstOfMonth.AddDate(0, 0, daysToFirstMonday)

		// Determine which fortnight we're in
		daysSinceFirstMonday := int(monday.Sub(firstMonday).Hours() / 24)
		fortnightNumber := daysSinceFirstMonday / 14

		start := firstMonday.AddDate(0, 0, fortnightNumber*14)
		end := start.AddDate(0, 0, 14).Add(-time.Nanosecond)
		return start, end

	case "month":
		start := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		return start, end

	default:
		// Default to week
		daysFromMonday := int(targetDate.Weekday()-time.Monday+7) % 7
		monday := targetDate.AddDate(0, 0, -daysFromMonday)
		start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
		return start, end
	}
}

func groupSessionsByClient(sessions []*models.WorkSession) map[string][]*models.WorkSession {
	clientSessions := make(map[string][]*models.WorkSession)
	for _, session := range sessions {
		if session.EndTime != nil { // Only include completed sessions
			clientSessions[session.ClientName] = append(clientSessions[session.ClientName], session)
		}
	}
	return clientSessions
}

func calculateClientTotal(timesheetService *service.TimesheetService, sessions []*models.WorkSession) float64 {
	total := 0.0
	for _, session := range sessions {
		total += timesheetService.CalculateBillableAmount(session)
	}
	return total
}

func formatClientName(name string) string {
	// Convert snake_case to Capitalized Case With Spaces
	words := strings.Split(name, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func wrapDescriptionText(text string, maxChars int) []string {
	if len(text) <= maxChars {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= maxChars {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
