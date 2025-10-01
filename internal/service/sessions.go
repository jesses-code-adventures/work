package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/shopspring/decimal"
)

// ParseTimeString parses time strings in various formats
func (s *TimesheetService) ParseTimeString(timeStr string) (time.Time, error) {
	now := time.Now()

	// Try full date-time format first
	if t, err := time.ParseInLocation("2006-01-02 15:04", timeStr, now.Location()); err == nil {
		return t, nil
	}

	// Try time-only format (assume today)
	if t, err := time.ParseInLocation("15:04", timeStr, now.Location()); err == nil {
		// Combine with today's date
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
	}

	return time.Time{}, fmt.Errorf("time must be in format 'YYYY-MM-DD HH:MM' or 'HH:MM'")
}

// DisplaySession formats and displays a single work session
func (s *TimesheetService) DisplaySession(session *models.WorkSession, verbose bool) {
	duration := s.CalculateDuration(session)
	billable := s.CalculateBillableAmount(session)
	status := "Active"
	endTime := "now"

	if session.EndTime != nil {
		status = "Completed"
		endTime = session.EndTime.Format("15:04:05")
	}

	billableStr := ""
	if billable.GreaterThan(decimal.Zero) {
		billableStr = fmt.Sprintf(" | %s", s.FormatBillableAmount(billable))
	}

	// Main session info
	fmt.Printf("%s | %s | %s - %s (%s)%s | %s\n",
		session.ClientName,
		session.StartTime.Format("2006-01-02"),
		session.StartTime.Format("15:04:05"),
		endTime,
		s.FormatDuration(duration),
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
		summary := s.formatSummaryWithBreaks(*session.FullWorkSummary)
		lines := s.wrapText(summary, 68) // Leave room for indentation

		for _, line := range lines {
			fmt.Printf("  │ %s\n", line)
		}
		fmt.Printf("  └─────────────────────────────────────────────────────────────────────\n")
	}

	fmt.Println() // Add spacing between sessions
}

// ExportSessionsCSV exports work sessions to CSV format
func (s *TimesheetService) ExportSessionsCSV(ctx context.Context, fromDate, toDate string, limit int32, output string) error {
	var sessions []*models.WorkSession
	var err error

	if fromDate != "" || toDate != "" {
		if fromDate == "" {
			fromDate = "1900-01-01"
		}
		if toDate == "" {
			toDate = "2099-12-31"
		}
		fmt.Printf("Exporting with date range %s to %s\n with limit %d\n", fromDate, toDate, limit)
		sessions, err = s.ListSessionsWithDateRange(ctx, fromDate, toDate, limit)
	} else {
		fmt.Printf("Exporting recent sessions with limit %d\n", limit)
		sessions, err = s.ListRecentSessions(ctx, limit)
	}
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found to export.")
		return nil
	}

	var file *os.File
	if output == "" || output == "-" {
		file = os.Stdout
	} else {
		file, err = os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	if err := writer.Write([]string{
		"ID", "Client", "Start Time", "End Time", "Duration (minutes)", "Hourly Rate", "Billable Amount", "Description", "Outside Git Notes", "Date",
	}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write session data
	for _, session := range sessions {
		duration := s.CalculateDuration(session)
		durationMinutes := strconv.FormatFloat(duration.Minutes(), 'f', 0, 64)
		billable := s.CalculateBillableAmount(session)

		endTimeStr := ""
		if session.EndTime != nil {
			endTimeStr = session.EndTime.Format("15:04:05")
		}

		description := ""
		if session.Description != nil {
			description = *session.Description
		}

		outsideGitNotes := ""
		if session.OutsideGit != nil {
			outsideGitNotes = *session.OutsideGit
		}

		hourlyRate := "0.00"
		if session.HourlyRate != nil && session.HourlyRate.GreaterThan(decimal.Zero) {
			hourlyRate = session.HourlyRate.StringFixed(2)
		}

		billableAmount := billable.StringFixed(2)

		record := []string{
			session.ID,
			session.ClientName,
			session.StartTime.Format("15:04:05"),
			endTimeStr,
			durationMinutes,
			hourlyRate,
			billableAmount,
			description,
			outsideGitNotes,
			session.StartTime.Format("2006-01-02"),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	if output != "" && output != "-" {
		fmt.Printf("Exported %d sessions to %s\n", len(sessions), output)
	}

	return nil
}

// wrapText wraps text to the specified width
func (s *TimesheetService) wrapText(text string, width int) []string {
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
func (s *TimesheetService) formatSummaryWithBreaks(text string) string {
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
