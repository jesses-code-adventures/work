package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"github.com/jesses-code-adventures/work/internal/models"
)

// GenerateInvoices generates PDF invoices for clients with billable hours
func (s *TimesheetService) GenerateInvoices(ctx context.Context, period, date string) error {
	// Parse the date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Calculate date range based on period
	fromDate, toDate := s.CalculatePeriodRange(period, targetDate)

	// Get sessions for the period
	sessions, err := s.ListSessionsWithDateRange(ctx, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"), 10000)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	// Group sessions by client and calculate totals
	clientSessions := s.groupSessionsByClient(sessions)

	invoiceCount := 0
	for clientName, clientSessionList := range clientSessions {
		total := s.calculateClientTotal(clientSessionList)
		if total <= 0 {
			continue // Skip clients with no billable hours
		}

		// Get client details for billing information
		client, err := s.GetClientByName(ctx, clientName)
		if err != nil {
			return fmt.Errorf("failed to get client details for %s: %w", clientName, err)
		}

		// Generate PDF invoice
		fileName := fmt.Sprintf("invoice_%s_%s_%s.pdf", clientName, period, date)
		fileName = s.sanitizeFileName(fileName)

		err = s.generateInvoicePDF(fileName, client, clientSessionList, period, fromDate, toDate)
		if err != nil {
			return fmt.Errorf("failed to generate invoice for %s: %w", clientName, err)
		}

		fmt.Printf("Generated invoice: %s (Total: $%.2f)\n", fileName, total)
		invoiceCount++
	}

	if invoiceCount == 0 {
		fmt.Println("No invoices generated - no clients with billable hours > 0 for the specified period")
	}

	return nil
}

func (s *TimesheetService) sanitizeFileName(fileName string) string {
	// Replace spaces and special characters
	result := ""
	for _, r := range fileName {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			result += string(r)
		} else if r == ' ' {
			result += "_"
		}
	}
	return result
}

func (s *TimesheetService) generateInvoicePDF(fileName string, client *models.Client, sessions []*models.WorkSession, period string, fromDate, toDate time.Time) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Header
	pdf.Cell(40, 10, fmt.Sprintf("Invoice - %s", s.formatClientName(client.Name)))
	pdf.Ln(12)

	// Client billing details in two columns
	if client.CompanyName != nil || client.ContactName != nil {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(40, 8, "Bill To:")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)

		// Left column items
		leftColY := pdf.GetY()
		if client.CompanyName != nil {
			pdf.Cell(95, 6, *client.CompanyName)
			pdf.Ln(6)
		}

		if client.ContactName != nil {
			pdf.Cell(95, 6, *client.ContactName)
			pdf.Ln(6)
		}

		if client.AddressLine1 != nil {
			pdf.Cell(95, 6, *client.AddressLine1)
			pdf.Ln(6)
		}

		if client.AddressLine2 != nil {
			pdf.Cell(95, 6, *client.AddressLine2)
			pdf.Ln(6)
		}

		// City, State, Postal Code on one line
		addressLine := ""
		if client.City != nil {
			addressLine += *client.City
		}
		if client.State != nil {
			if addressLine != "" {
				addressLine += ", "
			}
			addressLine += *client.State
		}
		if client.PostalCode != nil {
			if addressLine != "" {
				addressLine += " "
			}
			addressLine += *client.PostalCode
		}
		if addressLine != "" {
			pdf.Cell(95, 6, addressLine)
			pdf.Ln(6)
		}

		if client.Country != nil {
			pdf.Cell(95, 6, *client.Country)
			pdf.Ln(6)
		}

		// Right column items
		rightColY := leftColY
		pdf.SetXY(105, rightColY)

		if client.Email != nil {
			pdf.Cell(85, 6, fmt.Sprintf("Email: %s", *client.Email))
			pdf.SetXY(105, pdf.GetY()+6)
		}

		if client.Phone != nil {
			pdf.Cell(85, 6, fmt.Sprintf("Phone: %s", *client.Phone))
			pdf.SetXY(105, pdf.GetY()+6)
		}

		if client.Abn != nil {
			pdf.Cell(85, 6, fmt.Sprintf("ABN: %s", *client.Abn))
			pdf.SetXY(105, pdf.GetY()+6)
		}

		// Reset to left margin and add space
		pdf.SetX(10)
		pdf.Ln(8)
	}

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Date Range: %s to %s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02")))
	pdf.Ln(12)

	// Table headers - adjusted widths to fit A4 (total ~190mm)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(35, 8, "Start", "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 8, "End", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 8, "Duration", "1", 0, "C", false, 0, "")
	pdf.CellFormat(18, 8, "Rate", "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 8, "Description", "1", 0, "C", false, 0, "")
	pdf.CellFormat(22, 8, "Amount", "1", 1, "C", false, 0, "")

	// Table rows
	pdf.SetFont("Arial", "", 8)
	subtotal := 0.0

	for _, session := range sessions {
		duration := s.CalculateDuration(session)
		amount := s.CalculateBillableAmount(session)
		subtotal += amount

		// Prepare description lines with text wrapping
		description := ""
		if session.Description != nil {
			description = *session.Description
		}

		// Add outside_git notes to description
		if session.OutsideGit != nil && *session.OutsideGit != "" {
			if description != "" {
				description += "\n"
			}
			description += *session.OutsideGit
		}

		descriptionLines := s.wrapDescriptionText(description, 28)

		// Calculate row height based on number of description lines
		rowHeight := float64(len(descriptionLines)) * 6
		if rowHeight < 6 {
			rowHeight = 6
		}

		// Start datetime with minute precision
		startDateTime := session.StartTime.Format("2006-01-02 15:04")
		pdf.CellFormat(35, rowHeight, startDateTime, "1", 0, "L", false, 0, "")

		// End datetime with minute precision
		endDateTime := ""
		if session.EndTime != nil {
			endDateTime = session.EndTime.Format("2006-01-02 15:04")
		}
		pdf.CellFormat(35, rowHeight, endDateTime, "1", 0, "L", false, 0, "")

		pdf.CellFormat(20, rowHeight, fmt.Sprintf("%.1fh", duration.Hours()), "1", 0, "C", false, 0, "")

		rate := ""
		if session.HourlyRate != nil {
			rate = fmt.Sprintf("$%.0f", *session.HourlyRate)
		}
		pdf.CellFormat(18, rowHeight, rate, "1", 0, "C", false, 0, "")

		// Handle multi-line description
		currentX := pdf.GetX()
		currentY := pdf.GetY()

		// Draw description cell border
		pdf.Rect(currentX, currentY, 60, rowHeight, "D")

		// Write each line of description
		for i, line := range descriptionLines {
			pdf.SetXY(currentX+1, currentY+float64(i)*6+1)
			pdf.Cell(58, 6, line)
		}

		// Move to amount column
		pdf.SetXY(currentX+60, currentY)
		pdf.CellFormat(22, rowHeight, fmt.Sprintf("$%.2f", amount), "1", 1, "R", false, 0, "")
	}

	// Totals
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 11)

	// Subtotal
	pdf.Cell(168, 8, "Subtotal:")
	pdf.CellFormat(22, 8, fmt.Sprintf("$%.2f", subtotal), "", 1, "R", false, 0, "")

	// GST (10%)
	gst := subtotal * 0.1
	pdf.Cell(168, 8, "GST (10%):")
	pdf.CellFormat(22, 8, fmt.Sprintf("$%.2f", gst), "", 1, "R", false, 0, "")

	// Total
	total := subtotal + gst
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(168, 10, "Total:")
	pdf.CellFormat(22, 10, fmt.Sprintf("$%.2f", total), "", 1, "R", false, 0, "")

	// Payment Details
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Payment Details:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(40, 6, fmt.Sprintf("Bank: %s", s.cfg.BillingBank))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Account Name: %s", s.cfg.BillingAccountName))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Account Number: %s", s.cfg.BillingAccountNumber))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("BSB: %s", s.cfg.BillingBSB))
	pdf.Ln(6)

	return pdf.OutputFileAndClose(fileName)
}

func (s *TimesheetService) groupSessionsByClient(sessions []*models.WorkSession) map[string][]*models.WorkSession {
	clientSessions := make(map[string][]*models.WorkSession)
	for _, session := range sessions {
		if session.EndTime != nil { // Only include completed sessions
			clientSessions[session.ClientName] = append(clientSessions[session.ClientName], session)
		}
	}
	return clientSessions
}

func (s *TimesheetService) calculateClientTotal(sessions []*models.WorkSession) float64 {
	total := 0.0
	for _, session := range sessions {
		total += s.CalculateBillableAmount(session)
	}
	return total
}

func (s *TimesheetService) formatClientName(name string) string {
	// Convert snake_case to Capitalized Case With Spaces
	words := strings.Split(name, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func (s *TimesheetService) wrapDescriptionText(text string, maxChars int) []string {
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
