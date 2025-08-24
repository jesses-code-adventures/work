package main

import (
	"strings"
	"time"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

// shellescape escapes a string for safe use in shell commands
func shellescape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func calculatePeriodRange(period string, targetDate time.Time) (time.Time, time.Time) {
	switch period {
	case "day":
		start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		end := start.Add(24*time.Hour - time.Nanosecond)
		return start, end

	case "week":
		// Find Monday of the week containing targetDate
		daysFromMonday := int(targetDate.Weekday()-time.Monday+7) % 7
		monday := targetDate.AddDate(0, 0, -daysFromMonday)
		start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
		return start, end

	case "fortnight":
		// Find Monday of the week containing targetDate, then determine if it's first or second week
		daysFromMonday := int(targetDate.Weekday()-time.Monday+7) % 7
		monday := targetDate.AddDate(0, 0, -daysFromMonday)

		// Find the first Monday of the month
		firstOfMonth := time.Date(monday.Year(), monday.Month(), 1, 0, 0, 0, 0, monday.Location())
		daysToFirstMonday := int(time.Monday-firstOfMonth.Weekday()+7) % 7
		firstMonday := firstOfMonth.AddDate(0, 0, daysToFirstMonday)

		// Determine which fortnight we're in
		daysSinceFirstMonday := int(monday.Sub(firstMonday).Hours() / 24)
		fortnightNumber := daysSinceFirstMonday / 14

		start := firstMonday.AddDate(0, 0, fortnightNumber*14)
		end := start.AddDate(0, 0, 14).Add(-time.Nanosecond)
		return start, end

	case "month":
		start := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		return start, end

	default:
		// Default to week
		daysFromMonday := int(targetDate.Weekday()-time.Monday+7) % 7
		monday := targetDate.AddDate(0, 0, -daysFromMonday)
		start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
		return start, end
	}
}

func groupSessionsByClient(sessions []*models.WorkSession) map[string][]*models.WorkSession {
	clientSessions := make(map[string][]*models.WorkSession)
	for _, session := range sessions {
		if session.EndTime != nil { // Only include completed sessions
			clientSessions[session.ClientName] = append(clientSessions[session.ClientName], session)
		}
	}
	return clientSessions
}

func calculateClientTotal(timesheetService *service.TimesheetService, sessions []*models.WorkSession) float64 {
	total := 0.0
	for _, session := range sessions {
		total += timesheetService.CalculateBillableAmount(session)
	}
	return total
}

func formatClientName(name string) string {
	// Convert snake_case to Capitalized Case With Spaces
	words := strings.Split(name, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func wrapDescriptionText(text string, maxChars int) []string {
	if len(text) <= maxChars {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= maxChars {
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
