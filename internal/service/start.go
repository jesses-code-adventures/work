package service

import (
	"fmt"
	"strings"
	"time"
)

// ParseStartTime parses time strings in various formats for start times
func (s *TimesheetService) ParseStartTime(timeStr string) (time.Time, error) {
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
