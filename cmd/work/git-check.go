package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newGitCheckCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-check <session-id>",
		Short: "Debug git commands for a specific session",
		Long:  "Shows exactly what git commands are executed for a session's time period and their outputs.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			return gitCheckSession(timesheetService, sessionID)
		},
	}

	return cmd
}

func gitCheckSession(timesheetService *service.TimesheetService, sessionID string) error {
	// Use SQLite command to get session and client info together
	sqlCmd := fmt.Sprintf(`sqlite3 work.db "SELECT s.id, c.name, s.start_time, s.end_time, c.dir FROM sessions s JOIN clients c ON s.client_id = c.id WHERE s.id = '%s';"`, sessionID)

	cmd := exec.Command("sh", "-c", sqlCmd)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to query session '%s': %w", sessionID, err)
	}

	if len(output) == 0 {
		return fmt.Errorf("session '%s' not found", sessionID)
	}

	// Parse the output: id|client_name|start_time|end_time|dir
	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) < 5 {
		return fmt.Errorf("invalid session data returned: %s", string(output))
	}

	sessionIDResult := parts[0]
	clientName := parts[1]
	startTime := parts[2]
	endTime := parts[3]
	clientDir := parts[4]

	if endTime == "" {
		return fmt.Errorf("session '%s' is still active (no end time)", sessionID)
	}

	fmt.Printf("=== GIT CHECK FOR SESSION ===\n")
	fmt.Printf("Session ID: %s\n", sessionIDResult)
	fmt.Printf("Client: %s\n", clientName)
	fmt.Printf("Session Time: %s to %s\n", startTime, endTime)

	// Parse start time to get the date (handle multiple formats)
	var startTimeParsed time.Time
	var parseErr error

	// Try different time formats
	formats := []string{
		"2006-01-02 15:04:05.000000-07:00",
		"2006-01-02 15:04:05.000000+07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		startTimeParsed, parseErr = time.Parse(format, startTime)
		if parseErr == nil {
			break
		}
	}

	if parseErr != nil {
		// If all parsing fails, extract just the date part
		if len(startTime) >= 10 {
			dateOnly := startTime[:10]
			startTimeParsed, parseErr = time.Parse("2006-01-02", dateOnly)
			if parseErr != nil {
				return fmt.Errorf("failed to parse start time '%s': %w", startTime, parseErr)
			}
		} else {
			return fmt.Errorf("failed to parse start time '%s': %w", startTime, parseErr)
		}
	}

	// Parse end time as well
	var endTimeParsed time.Time
	for _, format := range formats {
		endTimeParsed, parseErr = time.Parse(format, endTime)
		if parseErr == nil {
			break
		}
	}

	if parseErr != nil {
		// If all parsing fails, extract just the date part
		if len(endTime) >= 10 {
			dateOnly := endTime[:10] + " 23:59:59"
			endTimeParsed, parseErr = time.Parse("2006-01-02 15:04:05", dateOnly)
			if parseErr != nil {
				return fmt.Errorf("failed to parse end time '%s': %w", endTime, parseErr)
			}
		} else {
			return fmt.Errorf("failed to parse end time '%s': %w", endTime, parseErr)
		}
	}

	// Use session start and end times for precise git analysis
	fromDateTime := startTimeParsed.Format("2006-01-02 15:04")
	toDateTime := endTimeParsed.Format("2006-01-02 15:04")

	fmt.Printf("Git Time Range: %s to %s\n", fromDateTime, toDateTime)

	// Process the directory
	dir := strings.TrimSpace(clientDir)
	fmt.Printf("Client Directory (raw): '%s'\n", clientDir)
	fmt.Printf("Client Directory (trimmed): '%s'\n", dir)

	// Expand tilde
	if strings.HasPrefix(dir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %w", err)
		}
		expandedDir := filepath.Join(homeDir, dir[2:])
		fmt.Printf("Directory (expanded): %s\n", expandedDir)
		dir = expandedDir
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	fmt.Printf("Directory exists: ✓\n")

	// Find git repositories
	fmt.Printf("\n=== FINDING GIT REPOSITORIES ===\n")
	gitRepos := findGitRepositoriesDebug(dir)

	if len(gitRepos) == 0 {
		fmt.Printf("No git repositories found in %s\n", dir)
		return nil
	}

	fmt.Printf("Found %d git repositories:\n", len(gitRepos))
	for i, repo := range gitRepos {
		fmt.Printf("  %d. %s\n", i+1, repo)
	}

	// Get the git analysis prompt
	gitPrompt := timesheetService.Config().GitAnalysisPrompt
	actualPrompt := strings.ReplaceAll(gitPrompt, "{from_date}", fromDateTime)
	actualPrompt = strings.ReplaceAll(actualPrompt, "{to_date}", toDateTime)

	fmt.Printf("\n=== GIT ANALYSIS PROMPT ===\n")
	fmt.Printf("%s\n", actualPrompt)

	// Process each repository
	for i, repoDir := range gitRepos {
		fmt.Printf("\n=== REPOSITORY %d: %s ===\n", i+1, filepath.Base(repoDir))
		fmt.Printf("Full path: %s\n", repoDir)

		// Check if it's actually a git repository
		gitDir := filepath.Join(repoDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			fmt.Printf("❌ Not a git repository (no .git directory)\n")
			continue
		}
		fmt.Printf("✓ Valid git repository\n")

		// Run basic git commands to show repository state
		fmt.Printf("\n--- Git Status ---\n")
		runGitCommand(repoDir, "git", "status", "--porcelain")

		fmt.Printf("\n--- Git Log for Time Range ---\n")
		logCmd := fmt.Sprintf("git log --since=\"%s\" --until=\"%s\" --oneline", fromDateTime, toDateTime)
		fmt.Printf("Command: %s\n", logCmd)
		runGitCommand(repoDir, "git", "log", fmt.Sprintf("--since=%s", fromDateTime), fmt.Sprintf("--until=%s", toDateTime), "--oneline")

		fmt.Printf("\n--- Git Log with Details ---\n")
		runGitCommand(repoDir, "git", "log", fmt.Sprintf("--since=%s", fromDateTime), fmt.Sprintf("--until=%s", toDateTime), "--stat")

		fmt.Printf("\n--- Recent Git Log (last 5 commits) ---\n")
		runGitCommand(repoDir, "git", "log", "--oneline", "-5")

		fmt.Printf("\n--- Recent Git Log with Timestamps ---\n")
		runGitCommand(repoDir, "git", "log", "--pretty=format:%h %cd %s", "--date=iso", "-5")

		// Test the actual opencode command that would be run
		fmt.Printf("\n--- Testing OpenCode Command ---\n")
		fmt.Printf("Would run in directory: %s\n", repoDir)
		fmt.Printf("Command: cd %s && echo '%s' | opencode run\n", repoDir, actualPrompt)

		// Actually run the opencode command to see what happens
		fmt.Printf("\n--- OpenCode Output ---\n")
		runOpenCodeCommand(repoDir, actualPrompt)
	}

	return nil
}

func findGitRepositoriesDebug(root string) []string {
	var gitRepos []string

	fmt.Printf("Searching for git repositories in: %s\n", root)

	// Try the find command first (like the original code does)
	fmt.Printf("Running find command...\n")
	cmd := exec.Command("find", root, "-type", "d", "-name", ".git", "-mtime", "-30", "-maxdepth", "3")
	fmt.Printf("Command: %s\n", strings.Join(cmd.Args, " "))

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Find command failed: %v\n", err)
		fmt.Printf("Falling back to directory walk...\n")
		return findGitRepositoriesWalkDebug(root)
	}

	fmt.Printf("Find command output:\n%s\n", string(output))

	// Parse find output to get repository directories
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" {
			// Get the parent directory (the actual repository directory)
			repoDir := filepath.Dir(line)
			gitRepos = append(gitRepos, repoDir)
			fmt.Printf("Found git repo: %s\n", repoDir)
		}
	}

	// If no recently modified repos found, check for repos with recent commits
	if len(gitRepos) == 0 {
		fmt.Printf("No recently modified .git directories found, checking for repos with recent commits...\n")
		return findGitRepositoriesWithRecentCommitsDebug(root)
	}

	return gitRepos
}

func findGitRepositoriesWalkDebug(root string) []string {
	fmt.Printf("Walking directory tree...\n")
	var gitRepos []string
	maxDepth := 2

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Walk error at %s: %v\n", path, err)
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		depth := len(strings.Split(rel, string(filepath.Separator)))

		if depth > maxDepth {
			if info.IsDir() {
				fmt.Printf("Skipping deep directory: %s (depth %d)\n", path, depth)
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoDir := filepath.Dir(path)
			gitRepos = append(gitRepos, repoDir)
			fmt.Printf("Found git repo (walk): %s\n", repoDir)
			return filepath.SkipDir
		}

		return nil
	})

	return gitRepos
}

func findGitRepositoriesWithRecentCommitsDebug(root string) []string {
	fmt.Printf("Checking for repositories with recent commits...\n")
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
			fmt.Printf("Checking recent commits in: %s\n", repoDir)
			cmd := exec.Command("git", "-C", repoDir, "log", "--since=1 week ago", "--oneline", "-n", "1")
			output, err := cmd.Output()
			if err == nil && len(strings.TrimSpace(string(output))) > 0 {
				gitRepos = append(gitRepos, repoDir)
				fmt.Printf("Found repo with recent commits: %s\n", repoDir)
				fmt.Printf("Recent commit: %s\n", strings.TrimSpace(string(output)))
			} else {
				fmt.Printf("No recent commits in: %s\n", repoDir)
			}

			return filepath.SkipDir
		}

		return nil
	})

	return gitRepos
}

func runGitCommand(repoDir string, args ...string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = repoDir

	fmt.Printf("Running: %s (in %s)\n", strings.Join(args, " "), repoDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Command failed: %v\n", err)
		if len(output) > 0 {
			fmt.Printf("Output: %s\n", string(output))
		}
	} else {
		if len(output) > 0 {
			fmt.Printf("Output:\n%s\n", string(output))
		} else {
			fmt.Printf("(no output)\n")
		}
	}
}

func runOpenCodeCommand(repoDir, prompt string) {
	// Create the shell command to cd into repository directory and run opencode
	shellCmd := fmt.Sprintf("cd %s && echo %s | opencode run",
		shellescape(repoDir),
		shellescape(prompt))

	fmt.Printf("Shell command: %s\n", shellCmd)

	cmd := exec.Command("sh", "-c", shellCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ OpenCode command failed: %v\n", err)
	}

	if len(output) > 0 {
		fmt.Printf("OpenCode output:\n%s\n", string(output))
	} else {
		fmt.Printf("(no opencode output)\n")
	}
}
