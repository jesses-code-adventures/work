package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
	"github.com/spf13/cobra"
)

func newSessionsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage sessions",
		Long:  "Commands for managing sessions, including listing sessions and their hourly rates.",
	}

	cmd.AddCommand(newSessionsCreateCmd(timesheetService))
	cmd.AddCommand(newSessionsListCmd(timesheetService))
	cmd.AddCommand(newSessionsUpdateCmd(timesheetService))
	cmd.AddCommand(newSessionsDeleteCmd(timesheetService))
	cmd.AddCommand(newSessionsCsvCmd(timesheetService))

	return cmd
}

func newSessionsCreateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var client string
	var fromTime string
	var toTime string
	var description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a work session with custom start and end times",
		Long:  "Create a work session for a client with specified start and end times. Times should be in format 'YYYY-MM-DD HH:MM' or 'HH:MM' for today.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Parse the times
			startTime, err := parseTimeString(fromTime)
			if err != nil {
				return fmt.Errorf("invalid start time format: %w", err)
			}

			endTime, err := parseTimeString(toTime)
			if err != nil {
				return fmt.Errorf("invalid end time format: %w", err)
			}

			// Validate that start time is before end time
			if endTime.Before(startTime) || endTime.Equal(startTime) {
				return fmt.Errorf("end time must be after start time")
			}

			// Create the session
			var desc *string
			if description != "" {
				desc = &description
			}

			session, err := timesheetService.CreateSessionWithTimes(ctx, client, startTime, endTime, desc)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			// Calculate duration and billable amount
			duration := timesheetService.CalculateDuration(session)
			billableAmount := timesheetService.CalculateBillableAmount(session)

			// Display session details
			fmt.Printf("Created session for %s:\n", client)
			fmt.Printf("  Start: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("  End: %s\n", session.EndTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Duration: %s\n", timesheetService.FormatDuration(duration))
			if description != "" {
				fmt.Printf("  Description: %s\n", description)
			}
			if billableAmount > 0 {
				fmt.Printf("  Billable: %s\n", timesheetService.FormatBillableAmount(billableAmount))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client name (required)")
	cmd.Flags().StringVarP(&fromTime, "from", "f", "", "Start time (required, format: 'YYYY-MM-DD HH:MM' or 'HH:MM')")
	cmd.Flags().StringVarP(&toTime, "to", "t", "", "End time (required, format: 'YYYY-MM-DD HH:MM' or 'HH:MM')")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Session description (optional)")

	cmd.MarkFlagRequired("client")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

func parseTimeString(timeStr string) (time.Time, error) {
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

func newSessionsListCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var limit int32
	var fromDate, toDate string
	var verbose bool
	var client string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work sessions",
		Long:  "Show a list of work sessions with durations and billable amounts. Filter by date range using -f and -t flags, or by client using -c flag. Use -v for verbose output including full work summaries.",
	}

	cmd.Flags().Int32VarP(&limit, "limit", "l", 10, "Number of sessions to show")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Show sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Show sessions to this date (YYYY-MM-DD)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full work summaries")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Filter sessions by client name")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		var sessions, err = func() ([]*models.WorkSession, error) {
			if client != "" {
				return timesheetService.ListSessionsByClient(ctx, client, limit)
			}
			if fromDate != "" || toDate != "" {
				if fromDate == "" {
					fromDate = "1900-01-01"
				}
				if toDate == "" {
					toDate = "2099-12-31"
				}
				return timesheetService.ListSessionsWithDateRange(ctx, fromDate, toDate, limit)
			}
			return timesheetService.ListRecentSessions(ctx, limit)
		}()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			if client != "" {
				fmt.Printf("No work sessions found for client '%s'.\n", client)
			} else {
				fmt.Println("No work sessions found.")
			}
			return nil
		}

		for _, session := range sessions {
			displaySession(session, timesheetService, verbose)
		}

		return nil
	}

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

func newSessionsDeleteCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var fromDate, toDate string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete work sessions",
		Long:  "Delete work sessions. Use with caution - this action cannot be undone.",
	}

	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Delete sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Delete sessions to this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if !force {
			var rangeStr string
			if fromDate != "" && toDate != "" {
				rangeStr = fmt.Sprintf("work sessions from %s to %s", fromDate, toDate)
			} else if fromDate != "" {
				rangeStr = fmt.Sprintf("work sessions since %s", fromDate)
			} else if toDate != "" {
				rangeStr = fmt.Sprintf("work sessions before %s", toDate)
			} else {
				rangeStr = "all work sessions"
			}

			fmt.Printf("This will permanently delete %s. Are you sure? (y/N): ", rangeStr)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		if fromDate != "" || toDate != "" {
			if fromDate == "" {
				fromDate = "1900-01-01"
			}
			if toDate == "" {
				toDate = "2099-12-31"
			}

			err := timesheetService.DeleteSessionsByDateRange(ctx, fromDate, toDate)
			if err != nil {
				return err
			}

			fmt.Printf("Deleted work sessions from %s to %s\n", fromDate, toDate)
		} else {
			err := timesheetService.DeleteAllSessions(ctx)
			if err != nil {
				return err
			}

			fmt.Println("Deleted all work sessions")
		}

		return nil
	}

	return cmd
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

func newSessionsUpdateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var hourlyRate float64
	var session string
	var companyName, contactName, email, phone string
	var addressLine1, addressLine2, city, state, postalCode, country, taxNumber, dir string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update details about a session",
		Long:  "Update attributes of the session, such as timeframe and hourly rate.",
		RunE: func(cmd *cobra.Command, args []string) error {
			print("not implemented")
			// ctx := cmd.Context()
			// if session == "" {
			// 	return fmt.Errorf("session name is required")
			// }
			//
			// updatedSession, err := timesheetService.UpdateSession(ctx, session, &database.SessionUpdateDetails{
			// 	HourlyRate:   &hourlyRate,
			// 	CompanyName:  &companyName,
			// 	ContactName:  &contactName,
			// 	Email:        &email,
			// 	Phone:        &phone,
			// 	AddressLine1: &addressLine1,
			// 	AddressLine2: &addressLine2,
			// 	City:         &city,
			// 	State:        &state,
			// 	PostalCode:   &postalCode,
			// 	Country:      &country,
			// 	TaxNumber:    &taxNumber,
			// 	Dir:          &dir,
			// })
			// if err != nil {
			// 	return fmt.Errorf("failed to update session billing: %w", err)
			// }
			//
			// fmt.Printf("Updated session '%s'\nNew state: \n", updatedSession.Name)
			// timesheetService.DisplaySession(ctx, updatedSession)
			return nil
		},
	}

	cmd.Flags().StringVarP(&session, "session", "c", "", "Name of the session to update")
	cmd.Flags().Float64VarP(&hourlyRate, "rate", "r", 0.0, "Hourly rate for the session")

	// Billing detail flags
	cmd.Flags().StringVar(&companyName, "company", "", "Company name")
	cmd.Flags().StringVar(&contactName, "contact", "", "Contact person name")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&addressLine1, "address1", "", "Address line 1")
	cmd.Flags().StringVar(&addressLine2, "address2", "", "Address line 2")
	cmd.Flags().StringVar(&city, "city", "", "City")
	cmd.Flags().StringVar(&state, "state", "", "State/Province")
	cmd.Flags().StringVar(&postalCode, "postcode", "", "Postal/ZIP code")
	cmd.Flags().StringVar(&country, "country", "", "Country")
	cmd.Flags().StringVar(&taxNumber, "tax", "", "Tax/VAT number")
	cmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory path for the session")

	return cmd
}

func newSessionsCsvCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var fromDate, toDate string
	var output string
	var limit int32
	var period string
	var date string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export work sessions to CSV",
		Long:  "Export work sessions to CSV format with hourly rates and billable amounts. Supports optional date filtering.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if fromDate == "" && toDate == "" && period != "" && date != "" {
				var d time.Time
				if date != "" {
					d, _ = time.Parse("2006-01-02", date)
				}
				fromDateTime, toDateTime := calculatePeriodRange(period, d)
				fromDate = fromDateTime.Format("2006-01-02")
				toDate = toDateTime.Format("2006-01-02")
			}
			return exportSessions(ctx, timesheetService, fromDate, toDate, limit, output)
		},
	}

	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "date", "If using period, the date in the period (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Export sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Export sessions to this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().Int32VarP(&limit, "limit", "l", 1000, "Maximum number of sessions to export")

	return cmd
}

func exportSessions(ctx context.Context, timesheetService *service.TimesheetService, fromDate, toDate string, limit int32, output string) error {
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
		duration := timesheetService.CalculateDuration(session)
		durationMinutes := strconv.FormatFloat(duration.Minutes(), 'f', 0, 64)
		billable := timesheetService.CalculateBillableAmount(session)

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
		if session.HourlyRate != nil && *session.HourlyRate > 0 {
			hourlyRate = fmt.Sprintf("%.2f", *session.HourlyRate)
		}

		billableAmount := fmt.Sprintf("%.2f", billable)

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
