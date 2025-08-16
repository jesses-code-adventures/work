package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newStartCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var clientName string
	var description string
	var fromTime string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a work session",
		Long:  "Start a new work session for a client. This will automatically stop any active session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientName == "" {
				return fmt.Errorf("client name is required (use -c flag)")
			}

			ctx := cmd.Context()

			var desc *string
			if description != "" {
				desc = &description
			}

			var session *models.WorkSession
			var err error

			if fromTime != "" {
				// Parse the custom start time
				startTime, parseErr := parseStartTime(fromTime)
				if parseErr != nil {
					return fmt.Errorf("invalid time format: %w", parseErr)
				}
				session, err = timesheetService.StartWorkWithTime(ctx, clientName, startTime, desc)
			} else {
				session, err = timesheetService.StartWork(ctx, clientName, desc)
			}

			if err != nil {
				return err
			}

			fmt.Printf("Started work session for %s at %s\n",
				clientName,
				session.StartTime.Format("15:04:05"))

			if desc != nil {
				fmt.Printf("Description: %s\n", *desc)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&clientName, "client", "c", "", "Client name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Optional description of the work")
	cmd.Flags().StringVarP(&fromTime, "from", "f", "", "Start time (YYYY-MM-DD HH:MM or HH:MM)")
	cmd.MarkFlagRequired("client")

	return cmd
}

func parseStartTime(timeStr string) (time.Time, error) {
	now := time.Now()

	// Try HH:MM format first (apply to current date)
	if len(timeStr) == 5 && strings.Contains(timeStr, ":") {
		parsed, err := time.Parse("15:04", timeStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format, expected HH:MM: %w", err)
		}
		// Apply the time to today's date
		return time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location()), nil
	}

	// Try YYYY-MM-DD HH:MM format
	if len(timeStr) == 16 && strings.Count(timeStr, "-") == 2 && strings.Contains(timeStr, ":") {
		parsed, err := time.Parse("2006-01-02 15:04", timeStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid datetime format, expected YYYY-MM-DD HH:MM: %w", err)
		}
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("invalid time format, expected HH:MM or YYYY-MM-DD HH:MM")
}
