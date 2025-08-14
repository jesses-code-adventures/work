package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newSummarizeCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize and analyze client data",
		Long:  "Commands for summarizing and analyzing client data, including directory-based analysis.",
	}

	cmd.AddCommand(newSummarizeDescriptionsCmd(timesheetService))

	return cmd
}

func newSummarizeDescriptionsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var period string
	var date string
	var client string

	cmd := &cobra.Command{
		Use:   "descriptions",
		Short: "Summarize descriptions from client directories",
		Long:  "Analyze client directories and summarize descriptions for the specified time period.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return summarizeDescriptions(ctx, timesheetService, period, date, client)
		},
	}

	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in the period (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Process only the specified client (optional)")

	return cmd
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
		tempDir, err = ioutil.TempDir("", "work-summarize-*")
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
		ioutil.WriteFile(outputFile, []byte("NO GIT ACTIVITY"), 0644)
		return
	}

	// Find all git repositories in subdirectories
	gitRepos := findGitRepositories(dir)

	if len(gitRepos) == 0 {
		fmt.Printf("  No git repositories found in %s\n", dir)
		outputFile := filepath.Join(tempDir, sanitizeClientName(clientName)+".txt")
		ioutil.WriteFile(outputFile, []byte("NO COMMITS"), 0644)
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
	err := ioutil.WriteFile(outputFile, []byte(combinedOutput), 0644)
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
func findGitRepositories(rootDir string) []string {
	var gitRepos []string

	// Walk through all subdirectories
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't access
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
