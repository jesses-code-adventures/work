package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newHoursCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var client string
	var period string
	var periodDate string
	var fromDate string
	var toDate string

	cmd := &cobra.Command{
		Use:   "hours",
		Short: "Display total worked hours",
		Long:  "Display total worked hours with optional filtering by client, period, or date range.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return showTotalHours(ctx, timesheetService, client, period, periodDate, fromDate, toDate)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Filter by client name")
	cmd.Flags().StringVarP(&period, "period", "p", "", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&periodDate, "date", "d", "", "Date in the period (YYYY-MM-DD), defaults to today when using -p")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Show hours from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Show hours to this date (YYYY-MM-DD)")

	return cmd
}

func showTotalHours(ctx context.Context, timesheetService *service.TimesheetService, client, period, periodDate, fromDate, toDate string) error {
	// Handle period filtering
	if period != "" {
		var targetDate time.Time
		var err error

		if periodDate == "" {
			// Default to today's date
			targetDate = time.Now()
		} else {
			targetDate, err = time.Parse("2006-01-02", periodDate)
			if err != nil {
				return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
			}
		}

		fromDateTime, toDateTime := calculatePeriodRange(period, targetDate)
		fromDate = fromDateTime.Format("2006-01-02")
		toDate = toDateTime.Format("2006-01-02")
	}

	// Get sessions based on filters
	var sessions []*models.WorkSession
	var err error

	if client != "" {
		if fromDate != "" || toDate != "" {
			// Get all sessions for client, then filter by date range
			allSessions, err := timesheetService.ListSessionsByClient(ctx, client, 10000)
			if err != nil {
				return fmt.Errorf("failed to get sessions for client: %w", err)
			}
			sessions = filterSessionsByDateRange(allSessions, fromDate, toDate)
		} else {
			sessions, err = timesheetService.ListSessionsByClient(ctx, client, 10000)
			if err != nil {
				return fmt.Errorf("failed to get sessions for client: %w", err)
			}
		}
	} else if fromDate != "" || toDate != "" {
		if fromDate == "" {
			fromDate = "1900-01-01"
		}
		if toDate == "" {
			toDate = "2099-12-31"
		}
		sessions, err = timesheetService.ListSessionsWithDateRange(ctx, fromDate, toDate, 10000)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
	} else {
		sessions, err = timesheetService.ListRecentSessions(ctx, 10000)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
	}

	if len(sessions) == 0 {
		fmt.Println("0.0")
		return nil
	}

	// Calculate total hours
	totalDuration := time.Duration(0)
	for _, session := range sessions {
		duration := timesheetService.CalculateDuration(session)
		totalDuration += duration
	}

	totalHours := totalDuration.Hours()
	fmt.Printf("%.1f\n", totalHours)

	return nil
}

func filterSessionsByDateRange(sessions []*models.WorkSession, fromDate, toDate string) []*models.WorkSession {
	if fromDate == "" && toDate == "" {
		return sessions
	}

	var filtered []*models.WorkSession

	var from, to time.Time
	var err error

	if fromDate != "" {
		from, err = time.Parse("2006-01-02", fromDate)
		if err != nil {
			return sessions // If parsing fails, return all sessions
		}
	}

	if toDate != "" {
		to, err = time.Parse("2006-01-02", toDate)
		if err != nil {
			return sessions // If parsing fails, return all sessions
		}
		// Set to end of day
		to = to.Add(24*time.Hour - time.Nanosecond)
	}

	for _, session := range sessions {
		sessionDate := session.StartTime

		if fromDate != "" && sessionDate.Before(from) {
			continue
		}

		if toDate != "" && sessionDate.After(to) {
			continue
		}

		filtered = append(filtered, session)
	}

	return filtered
}
