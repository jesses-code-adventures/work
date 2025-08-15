package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newSessionCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage work sessions",
		Long:  "Commands for managing work sessions, including creating sessions with custom times.",
	}

	cmd.AddCommand(newSessionCreateCmd(timesheetService))

	return cmd
}

func newSessionCreateCmd(timesheetService *service.TimesheetService) *cobra.Command {
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
			return createSession(ctx, timesheetService, client, fromTime, toTime, description)
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

func createSession(ctx context.Context, timesheetService *service.TimesheetService, clientName, fromTime, toTime, description string) error {
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

	session, err := timesheetService.CreateSessionWithTimes(ctx, clientName, startTime, endTime, desc)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Calculate duration and billable amount
	duration := timesheetService.CalculateDuration(session)
	billableAmount := timesheetService.CalculateBillableAmount(session)

	// Display session details
	fmt.Printf("Created session for %s:\n", clientName)
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
}

// parseTimeString parses time strings in format "YYYY-MM-DD HH:MM" or "HH:MM" (for today)
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
