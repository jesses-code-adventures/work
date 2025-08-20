package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newListCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var limit int32
	var fromDate, toDate string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent work sessions",
		Long:  "Show a list of recent work sessions with durations and billable amounts. Filter by date range using -f and -t flags. Use -v for verbose output including full work summaries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var sessions []*models.WorkSession
			var err error

			if fromDate != "" || toDate != "" {
				if fromDate == "" {
					fromDate = "1900-01-01"
				}
				if toDate == "" {
					toDate = "2099-12-31"
				}
				sessions, err = timesheetService.ListSessionsWithDateRange(ctx, fromDate, toDate, limit)
			} else {
				sessions, err = timesheetService.ListRecentSessions(ctx, limit)
			}

			if err != nil {
				return err
			}

			if len(sessions) == 0 {
				fmt.Println("No work sessions found.")
				return nil
			}

			for _, session := range sessions {
				displaySession(session, timesheetService, verbose)
			}

			return nil
		},
	}

	cmd.Flags().Int32VarP(&limit, "limit", "l", 10, "Number of sessions to show")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Show sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Show sessions to this date (YYYY-MM-DD)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full work summaries")

	return cmd
}

// displaySession formats and displays a single work session
func displaySession(session *models.WorkSession, timesheetService *service.TimesheetService, verbose bool) {
	duration := timesheetService.CalculateDuration(session)
	billable := timesheetService.CalculateBillableAmount(session)
	status := "Active"
	endTime := "now"

	if session.EndTime != nil {
		status = "Completed"
		endTime = session.EndTime.Format("15:04:05")
	}

	billableStr := ""
	if billable > 0 {
		billableStr = fmt.Sprintf(" | %s", timesheetService.FormatBillableAmount(billable))
	}

	// Main session info
	fmt.Printf("%s | %s | %s - %s (%s)%s | %s\n",
		session.ClientName,
		session.StartTime.Format("2006-01-02"),
		session.StartTime.Format("15:04:05"),
		endTime,
		timesheetService.FormatDuration(duration),
		billableStr,
		status)

	// Description (always shown if present)
	if session.Description != nil && *session.Description != "" {
		fmt.Printf("  → %s\n", *session.Description)
	}

	// Full work summary (only in verbose mode)
	if verbose && session.FullWorkSummary != nil && *session.FullWorkSummary != "" {
		fmt.Printf("\n  ┌─ Full Work Summary ─────────────────────────────────────────────────\n")

		// Format the summary with strategic linebreaks for better readability
		summary := formatSummaryWithBreaks(*session.FullWorkSummary)
		lines := wrapText(summary, 68) // Leave room for indentation

		for _, line := range lines {
			fmt.Printf("  │ %s\n", line)
		}
		fmt.Printf("  └─────────────────────────────────────────────────────────────────────\n")
	}

	fmt.Println() // Add spacing between sessions
}

// wrapText wraps text to the specified width
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := []string{}

	// Split by words, preserving spaces
	current := ""
	for _, char := range text {
		if char == ' ' || char == '\n' || char == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
			if char == '\n' {
				words = append(words, "\n") // Preserve line breaks
			} else if char != '\t' { // Skip tabs
				words = append(words, string(char))
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		words = append(words, current)
	}

	currentLine := ""
	for _, word := range words {
		if word == "\n" {
			lines = append(lines, currentLine)
			currentLine = ""
			continue
		}

		testLine := currentLine + word
		if len(testLine) <= width {
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

// formatSummaryWithBreaks adds strategic linebreaks to improve readability
func formatSummaryWithBreaks(text string) string {
	// Add linebreaks after common markdown patterns and structural elements
	result := text

	// Add linebreaks before markdown headers (## and ###)
	result = strings.ReplaceAll(result, " ### ", "\n\n### ")
	result = strings.ReplaceAll(result, " ## ", "\n\n## ")

	// Add linebreaks after sentences that end with repository/section indicators
	result = strings.ReplaceAll(result, " Repository - ", " Repository\n\n- ")
	result = strings.ReplaceAll(result, " Project - ", " Project\n\n- ")

	// Add linebreaks before bullet point patterns
	result = strings.ReplaceAll(result, " - **", "\n- **")
	result = strings.ReplaceAll(result, " • ", "\n• ")

	// Add breaks after definitions (sentences ending with colon)
	result = strings.ReplaceAll(result, "definitions ", "definitions\n\n")

	// Add linebreaks after long sentences ending with common patterns
	result = strings.ReplaceAll(result, "capabilities. ", "capabilities.\n\n")
	result = strings.ReplaceAll(result, "functionality. ", "functionality.\n\n")
	result = strings.ReplaceAll(result, "improvements. ", "improvements.\n\n")
	result = strings.ReplaceAll(result, "workflow. ", "workflow.\n\n")
	result = strings.ReplaceAll(result, "experience. ", "experience.\n\n")
	result = strings.ReplaceAll(result, "integrity ", "integrity\n\n")

	// Clean up excessive linebreaks (more than 2 consecutive)
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	// Trim leading/trailing whitespace
	result = strings.TrimSpace(result)

	return result
}
