package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jesses-code-adventures/work/internal/models"
)

// ShowTotalHours displays total worked hours with optional filtering
func (s *TimesheetService) ShowTotalHours(ctx context.Context, client, period, periodDate, fromDate, toDate string) error {
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

		fromDateTime, toDateTime := s.CalculatePeriodRange(period, targetDate)
		fromDate = fromDateTime.Format("2006-01-02")
		toDate = toDateTime.Format("2006-01-02")
	}

	// Get sessions based on filters
	var sessions []*models.WorkSession
	var err error

	if client != "" {
		if fromDate != "" || toDate != "" {
			// Get all sessions for client, then filter by date range
			allSessions, err := s.ListSessionsByClient(ctx, client, 10000)
			if err != nil {
				return fmt.Errorf("failed to get sessions for client: %w", err)
			}
			sessions = s.filterSessionsByDateRange(allSessions, fromDate, toDate)
		} else {
			sessions, err = s.ListSessionsByClient(ctx, client, 10000)
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
		sessions, err = s.ListSessionsWithDateRange(ctx, fromDate, toDate, 10000)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
	} else {
		sessions, err = s.ListRecentSessions(ctx, 10000)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
	}

	if len(sessions) == 0 {
		fmt.Println("0.0")
		return nil
	}

	// Calculate total hours and billable amount
	totalDuration := time.Duration(0)
	totalBillable := 0.0
	for _, session := range sessions {
		duration := s.CalculateDuration(session)
		totalDuration += duration
		totalBillable += s.CalculateBillableAmount(session)
	}

	totalHours := totalDuration.Hours()
	fmt.Printf("%.1f hours", totalHours)

	if totalBillable > 0 {
		fmt.Printf(" | %s", s.FormatBillableAmountWithGST(totalBillable))
	}
	fmt.Println()

	return nil
}

func (s *TimesheetService) filterSessionsByDateRange(sessions []*models.WorkSession, fromDate, toDate string) []*models.WorkSession {
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

func (s *TimesheetService) CalculatePeriodRange(period string, targetDate time.Time) (time.Time, time.Time) {
	switch period {
	case "day":
		start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		end := start.Add(24*time.Hour - time.Nanosecond)
		return start, end
	case "week":
		// Find Monday of this week
		weekday := targetDate.Weekday()
		if weekday == 0 { // Sunday
			weekday = 7
		}
		start := targetDate.AddDate(0, 0, -int(weekday-1))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
		return start, end
	case "fortnight":
		// Find Monday of this week
		weekday := targetDate.Weekday()
		if weekday == 0 { // Sunday
			weekday = 7
		}
		start := targetDate.AddDate(0, 0, -int(weekday-1))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end := start.AddDate(0, 0, 14).Add(-time.Nanosecond)
		return start, end
	case "month":
		start := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		return start, end
	default:
		// Default to day if unknown period
		start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		end := start.Add(24*time.Hour - time.Nanosecond)
		return start, end
	}
}
