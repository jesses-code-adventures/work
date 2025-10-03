package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

func newSessionsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage sessions",
		Long:  "Commands for managing sessions, including listing sessions and their hourly rates.",
	}

	cmd.AddCommand(newSessionsCreateCmd(timesheetService))
	cmd.AddCommand(newSessionsListCmd(timesheetService))
	cmd.AddCommand(newSessionsUpdateCmd(timesheetService))
	cmd.AddCommand(newSessionsDeleteCmd(timesheetService))
	cmd.AddCommand(newSessionsCsvCmd(timesheetService))

	return cmd
}

func newSessionsCreateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var client string
	var fromTime string
	var toTime string
	var description string
	var includesGst bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a work session with custom start and end times",
		Long:  "Create a work session for a client with specified start and end times. Times should be in format 'YYYY-MM-DD HH:MM' or 'HH:MM' for today.",
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client name (required)")
	cmd.Flags().StringVarP(&fromTime, "from", "f", "", "Start time (required, format: 'YYYY-MM-DD HH:MM' or 'HH:MM')")
	cmd.Flags().StringVarP(&toTime, "to", "t", "", "End time (required, format: 'YYYY-MM-DD HH:MM' or 'HH:MM')")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Session description (optional)")
	cmd.Flags().BoolVar(&includesGst, "includes-gst", false, "Session amount includes GST (default: false)")

	cmd.MarkFlagRequired("client")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		startTime, err := timesheetService.ParseTimeString(fromTime)
		if err != nil {
			return fmt.Errorf("invalid start time format: %w", err)
		}

		endTime, err := timesheetService.ParseTimeString(toTime)
		if err != nil {
			return fmt.Errorf("invalid end time format: %w", err)
		}

		if endTime.Before(startTime) || endTime.Equal(startTime) {
			return fmt.Errorf("end time must be after start time")
		}

		var desc *string
		if description != "" {
			desc = &description
		}

		session, err := timesheetService.CreateSessionWithTimes(ctx, client, startTime, endTime, desc, includesGst)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		duration := timesheetService.CalculateDuration(session)
		billableAmount := timesheetService.CalculateBillableAmount(session)

		// Display session details
		fmt.Printf("Created session for %s:\n", client)
		fmt.Printf("  Start: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  End: %s\n", session.EndTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Duration: %s\n", timesheetService.FormatDuration(duration))
		if description != "" {
			fmt.Printf("  Description: %s\n", description)
		}
		if billableAmount.GreaterThan(decimal.Zero) {
			fmt.Printf("  Billable: %s\n", timesheetService.FormatSessionBillableAmount(session))
		}
		return nil
	}

	return cmd
}

func newSessionsListCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var limit int32
	var fromDate, toDate string
	var verbose bool
	var client string
	var period string
	var periodDate string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work sessions",
		Long:  "Show a list of work sessions with durations and billable amounts. Filter by date range using -f and -t flags, by period using -p flag, or by client using -c flag. Use -v for verbose output including full work summaries.",
	}

	cmd.Flags().Int32VarP(&limit, "limit", "l", 10, "Number of sessions to show")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Show sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Show sessions to this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&period, "period", "p", "", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&periodDate, "date", "d", "", "Date in the period (YYYY-MM-DD), defaults to today when using -p")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full work summaries")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Filter sessions by client name")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Handle period filtering (same logic as hours command)
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

			fromDateTime, toDateTime := timesheetService.CalculatePeriodRange(period, targetDate)
			fromDate = fromDateTime.Format("2006-01-02")
			toDate = toDateTime.Format("2006-01-02")
		}

		var sessions, err = func() ([]*models.WorkSession, error) {
			if client != "" {
				if fromDate != "" || toDate != "" {
					// Get all sessions for client, then filter by date range
					allSessions, err := timesheetService.ListSessionsByClient(ctx, client, 10000)
					if err != nil {
						return nil, fmt.Errorf("failed to get sessions for client: %w", err)
					}
					filtered := timesheetService.FilterSessionsByDateRange(allSessions, fromDate, toDate)
					// Apply limit after filtering
					if int32(len(filtered)) > limit {
						filtered = filtered[:limit]
					}
					return filtered, nil
				} else {
					return timesheetService.ListSessionsByClient(ctx, client, limit)
				}
			}
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
			if client != "" {
				fmt.Printf("No work sessions found for client '%s'.\n", client)
			} else {
				fmt.Println("No work sessions found.")
			}
			return nil
		}

		for _, session := range sessions {
			timesheetService.DisplaySession(session, verbose)
		}

		return nil
	}

	return cmd
}

func newSessionsDeleteCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var fromDate, toDate string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete work sessions",
		Long:  "Delete work sessions. Use with caution - this action cannot be undone.",
	}

	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Delete sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Delete sessions to this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if !force {
			var rangeStr string
			if fromDate != "" && toDate != "" {
				rangeStr = fmt.Sprintf("work sessions from %s to %s", fromDate, toDate)
			} else if fromDate != "" {
				rangeStr = fmt.Sprintf("work sessions since %s", fromDate)
			} else if toDate != "" {
				rangeStr = fmt.Sprintf("work sessions before %s", toDate)
			} else {
				rangeStr = "all work sessions"
			}

			fmt.Printf("This will permanently delete %s. Are you sure? (y/N): ", rangeStr)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		if fromDate != "" || toDate != "" {
			if fromDate == "" {
				fromDate = "1900-01-01"
			}
			if toDate == "" {
				toDate = "2099-12-31"
			}

			err := timesheetService.DeleteSessionsByDateRange(ctx, fromDate, toDate)
			if err != nil {
				return err
			}

			fmt.Printf("Deleted work sessions from %s to %s\n", fromDate, toDate)
		} else {
			err := timesheetService.DeleteAllSessions(ctx)
			if err != nil {
				return err
			}

			fmt.Println("Deleted all work sessions")
		}

		return nil
	}

	return cmd
}

func newSessionsUpdateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var hourlyRate float64
	var companyName, contactName, email, phone string
	var addressLine1, addressLine2, city, state, postalCode, country, taxNumber, dir string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update details about a session",
		Long:  "Update attributes of the session, such as timeframe and hourly rate.",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.Flags().Float64VarP(&hourlyRate, "rate", "r", 0.0, "Hourly rate for the session")

	// Billing detail flags
	cmd.Flags().StringVar(&companyName, "company", "", "Company name")
	cmd.Flags().StringVar(&contactName, "contact", "", "Contact person name")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&addressLine1, "address1", "", "Address line 1")
	cmd.Flags().StringVar(&addressLine2, "address2", "", "Address line 2")
	cmd.Flags().StringVar(&city, "city", "", "City")
	cmd.Flags().StringVar(&state, "state", "", "State/Province")
	cmd.Flags().StringVar(&postalCode, "postcode", "", "Postal/ZIP code")
	cmd.Flags().StringVar(&country, "country", "", "Country")
	cmd.Flags().StringVar(&taxNumber, "tax", "", "Tax/VAT number")
	cmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory path for the session")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		print("not implemented")
		// sessionID := args[0]
		// if sessionID == "" {
		// 	return fmt.Errorf("session name is required")
		// }
		// ctx := cmd.Context()
		// if session == "" {
		// 	return fmt.Errorf("session name is required")
		// }
		//
		// updatedSession, err := timesheetService.UpdateSession(ctx, session, &database.SessionUpdateDetails{
		// 	HourlyRate:   &hourlyRate,
		// 	CompanyName:  &companyName,
		// 	ContactName:  &contactName,
		// 	Email:        &email,
		// 	Phone:        &phone,
		// 	AddressLine1: &addressLine1,
		// 	AddressLine2: &addressLine2,
		// 	City:         &city,
		// 	State:        &state,
		// 	PostalCode:   &postalCode,
		// 	Country:      &country,
		// 	TaxNumber:    &taxNumber,
		// 	Dir:          &dir,
		// })
		// if err != nil {
		// 	return fmt.Errorf("failed to update session billing: %w", err)
		// }
		//
		// fmt.Printf("Updated session '%s'\nNew state: \n", updatedSession.Name)
		// timesheetService.DisplaySession(ctx, updatedSession)
		// return nil
		return nil
	}

	return cmd
}

func newSessionsCsvCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var fromDate, toDate string
	var output string
	var limit int32
	var period string
	var date string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export work sessions to CSV",
		Long:  "Export work sessions to CSV format with hourly rates and billable amounts. Supports optional date filtering.",
	}

	cmd.Flags().StringVarP(&period, "period", "p", "", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "If using period, the date in the period (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Export sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Export sessions to this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().Int32VarP(&limit, "limit", "l", 1000, "Maximum number of sessions to export")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if fromDate == "" && toDate == "" && period != "" && date != "" {
			fmt.Printf("Period: %s, date: %s\n", period, date)
			var d time.Time
			if date != "" {
				d, _ = time.Parse("2006-01-02", date)
			}
			fromDateTime, toDateTime := timesheetService.CalculatePeriodRange(period, d)
			fromDate = fromDateTime.Format("2006-01-02")
			toDate = toDateTime.Format("2006-01-02")
		}

		fmt.Printf("Flags: period: %s, date: %s, from: %s, to: %s, output: %s, limit: %d\n", period, date, fromDate, toDate, output, limit)

		return timesheetService.ExportSessionsCSV(ctx, fromDate, toDate, limit, output)
	}

	return cmd
}
