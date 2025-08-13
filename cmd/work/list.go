package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jessewilliams/work/internal/models"
	"github.com/jessewilliams/work/internal/service"
)

func newListCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var limit int32
	var fromDate, toDate string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent work sessions",
		Long:  "Show a list of recent work sessions with durations and billable amounts. Filter by date range using -f and -t flags.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var sessions, err = func() ([]*models.WorkSession, error) {
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
				fmt.Println("No work sessions found.")
				return nil
			}

			for _, session := range sessions {
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

				fmt.Printf("%s | %s | %s - %s (%s)%s | %s\n",
					session.ClientName,
					session.StartTime.Format("2006-01-02"),
					session.StartTime.Format("15:04:05"),
					endTime,
					timesheetService.FormatDuration(duration),
					billableStr,
					status)

				if session.Description != nil && *session.Description != "" {
					fmt.Printf("- %s\n", *session.Description)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().Int32VarP(&limit, "limit", "l", 10, "Number of sessions to show")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Show sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Show sessions to this date (YYYY-MM-DD)")

	return cmd
}
