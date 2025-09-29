package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"github.com/jesses-code-adventures/work/internal/db"
	"github.com/jesses-code-adventures/work/internal/models"
)

// GenerateInvoices generates PDF invoices for clients with billable hours
func (s *TimesheetService) GenerateInvoices(ctx context.Context, period, date, clientName string) error {
	// Parse the date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Calculate date range based on period
	fromDate, toDate := s.CalculatePeriodRange(period, targetDate)

	// Get sessions for the period that haven't been invoiced yet
	var sessions []*models.WorkSession

	if clientName != "" {
		sessions, err = s.db.GetSessionsForPeriodWithoutInvoiceByClient(ctx, fromDate, toDate, clientName)
		if err != nil {
			return fmt.Errorf("failed to get uninvoiced sessions for client %s: %w", clientName, err)
		}
	} else {
		sessions, err = s.db.GetSessionsForPeriodWithoutInvoice(ctx, fromDate, toDate)
		if err != nil {
			return fmt.Errorf("failed to get uninvoiced sessions: %w", err)
		}
	}

	// Group sessions by client and calculate totals
	clientSessions := s.groupSessionsByClient(sessions)

	invoiceCount := 0
	for clientName, clientSessionList := range clientSessions {
		subtotal := s.calculateClientTotal(clientSessionList)
		if subtotal <= 0 {
			continue // Skip clients with no billable hours
		}

		// Calculate GST and total
		var gstAmount float64
		var total float64
		if s.cfg.GSTRegistered {
			gstAmount = subtotal * 0.1 // 10% GST
			total = subtotal + gstAmount
		} else {
			total = subtotal
		}

		// Get client details for billing information
		client, err := s.GetClientByName(ctx, clientName)
		if err != nil {
			return fmt.Errorf("failed to get client details for %s: %w", clientName, err)
		}

		// Generate invoice number
		invoiceNumber := fmt.Sprintf("INV-%s-%s-%s", clientName, period, date)
		invoiceNumber = s.sanitizeFileName(invoiceNumber)

		// Create invoice record in database
		invoice, err := s.db.CreateInvoice(ctx, client.ID, invoiceNumber, period, fromDate, toDate, subtotal, gstAmount, total, 0.00)
		if err != nil {
			return fmt.Errorf("failed to create invoice record for %s: %w", clientName, err)
		}

		// Update sessions with invoice ID
		for _, session := range clientSessionList {
			err = s.db.UpdateSessionInvoiceID(ctx, session.ID, invoice.ID)
			if err != nil {
				return fmt.Errorf("failed to update session %s with invoice ID: %w", session.ID, err)
			}
		}

		// Generate PDF invoice
		fileName := fmt.Sprintf("invoice_%s_%s_%s.pdf", clientName, period, date)
		fileName = s.sanitizeFileName(fileName)

		err = s.generateInvoicePDF(fileName, client, clientSessionList, period, fromDate, toDate)
		if err != nil {
			return fmt.Errorf("failed to generate invoice for %s: %w", clientName, err)
		}

		var totalDisplay string
		if s.cfg.GSTRegistered {
			totalDisplay = fmt.Sprintf("$%.2f ($%.2f inc. GST)", subtotal, total)
		} else {
			totalDisplay = fmt.Sprintf("$%.2f", total)
		}

		fmt.Printf("Generated invoice: %s (Total: %s)\n", fileName, totalDisplay)
		invoiceCount++
	}

	if invoiceCount == 0 {
		fmt.Println("No invoices generated - no clients with billable hours > 0 for the specified period")
	}

	return nil
}

// RegenerateInvoices deletes existing invoices for a period and regenerates them
func (s *TimesheetService) RegenerateInvoices(ctx context.Context, period, date, clientName string) error {
	// Parse the date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Calculate date range based on period
	fromDate, toDate := s.CalculatePeriodRange(period, targetDate)

	// Get existing invoices for this period
	var existingInvoices []*models.Invoice

	if clientName != "" {
		existingInvoices, err = s.db.GetInvoicesByPeriodAndClient(ctx, fromDate, toDate, period, clientName)
		if err != nil {
			return fmt.Errorf("failed to get existing invoices for period and client %s: %w", clientName, err)
		}
	} else {
		existingInvoices, err = s.db.GetInvoicesByPeriod(ctx, fromDate, toDate, period)
		if err != nil {
			return fmt.Errorf("failed to get existing invoices for period: %w", err)
		}
	}

	// Clear sessions' invoice_id for existing invoices and delete the invoices
	for _, invoice := range existingInvoices {
		// Clear session invoice IDs
		err = s.db.ClearSessionInvoiceIDs(ctx, invoice.ID)
		if err != nil {
			return fmt.Errorf("failed to clear session invoice IDs for invoice %s: %w", invoice.ID, err)
		}

		// Delete the invoice
		err = s.db.DeleteInvoice(ctx, invoice.ID)
		if err != nil {
			return fmt.Errorf("failed to delete invoice %s: %w", invoice.ID, err)
		}

		fmt.Printf("Deleted existing invoice: %s\n", invoice.InvoiceNumber)
	}

	// Now generate new invoices
	return s.GenerateInvoices(ctx, period, date, clientName)
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

	// Header with company name
	pdf.Cell(40, 10, fmt.Sprintf("Invoice - %s", s.formatClientName(client.Name)))
	pdf.Ln(8)

	// Billing company name and ABN/ACN
	if s.cfg.BillingCompanyName != "" {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(40, 6, s.cfg.BillingCompanyName)
		pdf.Ln(6)
	}

	if s.cfg.BillingABN != "" {
		pdf.SetFont("Arial", "", 10)
		abnText := fmt.Sprintf("ABN %s", s.cfg.BillingABN)
		if s.cfg.BillingACN != "" {
			abnText = fmt.Sprintf("ABN %s (includes ACN %s)", s.cfg.BillingABN, s.cfg.BillingACN)
		}
		pdf.Cell(40, 6, abnText)
		pdf.Ln(12)
	}

	pdf.SetFont("Arial", "B", 16)

	// Client billing details in two columns
	if client.CompanyName != nil || client.ContactName != nil {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(40, 8, "Bill To:")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)

		// Left column items
		leftColY := pdf.GetY()
		leftEndY := leftColY

		// Contact name first (person above company)
		if client.ContactName != nil {
			pdf.Cell(95, 6, *client.ContactName)
			pdf.Ln(6)
			leftEndY = pdf.GetY()
		}

		// Then company name
		if client.CompanyName != nil {
			pdf.Cell(95, 6, *client.CompanyName)
			pdf.Ln(6)
			leftEndY = pdf.GetY()
		}

		// Address as single line
		address := s.formatClientAddress(client)
		if address != "" {
			pdf.Cell(95, 6, address)
			pdf.Ln(6)
			leftEndY = pdf.GetY()
		}

		// Right column items
		rightColY := leftColY
		rightEndY := rightColY
		pdf.SetXY(105, rightColY)

		if client.Email != nil {
			pdf.Cell(85, 6, fmt.Sprintf("Email: %s", *client.Email))
			rightEndY = pdf.GetY() + 6
			pdf.SetXY(105, rightEndY)
		}

		if client.Phone != nil {
			pdf.Cell(85, 6, fmt.Sprintf("Phone: %s", *client.Phone))
			rightEndY = pdf.GetY() + 6
			pdf.SetXY(105, rightEndY)
		}

		if client.Abn != nil {
			pdf.Cell(85, 6, fmt.Sprintf("ABN: %s", *client.Abn))
			rightEndY = pdf.GetY() + 6
			pdf.SetXY(105, rightEndY)
		}

		// Set Y position to the maximum of both columns
		maxY := leftEndY
		if rightEndY > maxY {
			maxY = rightEndY
		}

		// Reset to left margin and position after both columns
		pdf.SetXY(10, maxY)
		pdf.Ln(12) // Add proper spacing after Bill To section
	}

	// Payment Details (moved before totals)
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
	pdf.Ln(12) // Add space before totals

	// Calculate totals first
	subtotal := 0.0
	for _, session := range sessions {
		amount := s.CalculateBillableAmount(session)
		subtotal += amount
	}

	// Totals section on first page
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(168, 8, "Subtotal:")
	pdf.CellFormat(22, 8, fmt.Sprintf("$%.2f", subtotal), "", 1, "R", false, 0, "")

	// GST (10%) - only if GST registered
	var total float64
	if s.cfg.GSTRegistered {
		gst := subtotal * 0.1
		pdf.Cell(168, 8, "GST (10%):")
		pdf.CellFormat(22, 8, fmt.Sprintf("$%.2f", gst), "", 1, "R", false, 0, "")
		total = subtotal + gst
	} else {
		total = subtotal
	}

	// Total
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(168, 10, "Total:")
	pdf.CellFormat(22, 10, fmt.Sprintf("$%.2f", total), "", 1, "R", false, 0, "")

	// Start new page for the session details table
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, fmt.Sprintf("Session Details (%s to %s)", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02")))
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

	for _, session := range sessions {
		duration := s.CalculateDuration(session)
		amount := s.CalculateBillableAmount(session)

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

func (s *TimesheetService) formatClientAddress(client *models.Client) string {
	var addressParts []string

	if client.AddressLine1 != nil && *client.AddressLine1 != "" {
		addressParts = append(addressParts, *client.AddressLine1)
	}

	if client.AddressLine2 != nil && *client.AddressLine2 != "" {
		addressParts = append(addressParts, *client.AddressLine2)
	}

	if client.City != nil && *client.City != "" {
		addressParts = append(addressParts, *client.City)
	}

	if client.State != nil && *client.State != "" {
		addressParts = append(addressParts, *client.State)
	}

	if client.PostalCode != nil && *client.PostalCode != "" {
		addressParts = append(addressParts, *client.PostalCode)
	}

	if client.Country != nil && *client.Country != "" {
		addressParts = append(addressParts, *client.Country)
	}

	return strings.Join(addressParts, ", ")
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

// ListInvoices displays a list of invoices with client, period, amounts and payment status
func (s *TimesheetService) ListInvoices(ctx context.Context, limit int32, clientName string, unpaidOnly bool) error {
	var invoices []*models.Invoice
	var err error

	if clientName != "" {
		invoices, err = s.db.GetInvoicesByClient(ctx, clientName)
		if err != nil {
			return fmt.Errorf("failed to get invoices for client %s: %w", clientName, err)
		}
	} else {
		invoices, err = s.db.ListInvoices(ctx, limit)
		if err != nil {
			return fmt.Errorf("failed to list invoices: %w", err)
		}
	}

	// Filter for unpaid invoices if requested
	if unpaidOnly {
		var unpaidInvoices []*models.Invoice
		for _, invoice := range invoices {
			if invoice.AmountPaid < invoice.TotalAmount {
				unpaidInvoices = append(unpaidInvoices, invoice)
			}
		}
		invoices = unpaidInvoices
	}

	if len(invoices) == 0 {
		if unpaidOnly {
			fmt.Println("No unpaid invoices found.")
		} else {
			fmt.Println("No invoices found.")
		}
		return nil
	}

	// Print header
	if unpaidOnly {
		fmt.Println("Unpaid Invoices:")
	}
	fmt.Printf("%-38s %-15s %-10s %-12s %-12s %-12s %-12s %-12s\n",
		"ID", "CLIENT", "PERIOD", "FROM", "TO", "SUBTOTAL", "TOTAL", "PAID")
	fmt.Println(strings.Repeat("-", 131))

	// Print each invoice
	for _, invoice := range invoices {
		paidStatus := fmt.Sprintf("$%.2f", invoice.AmountPaid)
		if invoice.AmountPaid >= invoice.TotalAmount {
			paidStatus = "PAID"
		} else if invoice.AmountPaid > 0 {
			paidStatus = fmt.Sprintf("$%.2f", invoice.AmountPaid)
		} else {
			paidStatus = "UNPAID"
		}

		fmt.Printf("%-38s %-15s %-10s %-12s %-12s $%-11.2f $%-11.2f %-12s\n",
			invoice.ID,
			truncateString(invoice.ClientName, 14),
			invoice.PeriodType,
			invoice.PeriodStartDate.Format("2006-01-02"),
			invoice.PeriodEndDate.Format("2006-01-02"),
			invoice.SubtotalAmount,
			invoice.TotalAmount,
			paidStatus,
		)
	}

	return nil
}

func (s *TimesheetService) PayInvoice(ctx context.Context, id string, amount float64) error {
	invoice, err := s.db.GetInvoiceByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	remainingAmount := invoice.TotalAmount - invoice.AmountPaid
	if remainingAmount <= 0 {
		return fmt.Errorf("invoice already fully paid")
	}

	if amount < 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if amount == 0 {
		amount = remainingAmount
	}

	if amount > remainingAmount {
		return fmt.Errorf("payment amount ($%.2f) exceeds remaining balance ($%.2f)", amount, remainingAmount)
	}

	err = s.db.PayInvoice(ctx, db.PayInvoiceParams{
		ID:     invoice.ID,
		Amount: amount,
	})
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	newAmountPaid := invoice.AmountPaid + amount
	status := "partially paid"
	if newAmountPaid >= invoice.TotalAmount {
		status = "fully paid"
	}

	fmt.Printf("Invoice %s paid $%.2f (now %s: $%.2f/$%.2f)\n",
		invoice.InvoiceNumber, amount, status, newAmountPaid, invoice.TotalAmount)
	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
