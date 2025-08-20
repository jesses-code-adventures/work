package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newExportCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var fromDate, toDate string
	var output string
	var limit int32

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export work sessions to CSV",
		Long:  "Export work sessions to CSV format with hourly rates and billable amounts. Supports optional date filtering.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			return exportSessions(ctx, timesheetService, fromDate, toDate, limit, output)
		},
	}

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
