package main

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/database"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newClientsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clients",
		Short: "Create, update and list clients",
		Long:  "Commands for managing clients, including listing clients and their hourly rates.",
	}

	cmd.AddCommand(newClientsCreateCmd(timesheetService))
	cmd.AddCommand(newClientsListCmd(timesheetService))
	cmd.AddCommand(newClientsUpdateCmd(timesheetService))

	return cmd
}

func newClientsCreateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var rate float64
	var retainerAmount, retainerHours float64
	var retainerBasis, dir string

	cmd := &cobra.Command{
		Use:   "create <client-name>",
		Short: "Create a new client",
		Long:  "Create a client with a given hourly rate, optional retainer, and directory",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.Flags().Float64VarP(&rate, "rate", "r", 0.0, "Hourly rate for the client")
	cmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory path for the client")

	// Retainer flags
	cmd.Flags().Float64Var(&retainerAmount, "retainer-amount", 0.0, "Retainer amount (e.g., 5000.00)")
	cmd.Flags().Float64Var(&retainerHours, "retainer-hours", 0.0, "Hours covered by retainer (e.g., 40.0)")
	cmd.Flags().StringVar(&retainerBasis, "retainer-basis", "", "Retainer billing basis: day, week, month, quarter, year")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		clientName := args[0]

		switch {
		case clientName != "":
			return createClient(ctx, timesheetService, clientName, rate, retainerAmount, retainerHours, retainerBasis, dir)
		default:
			return fmt.Errorf("must supply a client name (usage: work clients create <client-name>)")
		}
	}

	return cmd
}

func createClient(ctx context.Context, timesheetService *service.TimesheetService, name string, rate float64, retainerAmount, retainerHours float64, retainerBasis, dir string) error {
	// Convert fields to pointers (nil if zero/empty)
	var retainerAmountPtr *decimal.Decimal
	var retainerHoursPtr *float64
	var retainerBasisPtr, dirPtr *string

	if retainerAmount > 0 {
		amt := decimal.NewFromFloat(retainerAmount)
		retainerAmountPtr = &amt
	}
	if retainerHours > 0 {
		retainerHoursPtr = &retainerHours
	}
	if retainerBasis != "" {
		retainerBasisPtr = &retainerBasis
	}
	if dir != "" {
		dirPtr = &dir
	}

	client, err := timesheetService.CreateClient(ctx, name, decimal.NewFromFloat(rate), retainerAmountPtr, retainerHoursPtr, retainerBasisPtr, dirPtr)
	if err != nil {
		return err
	}

	fmt.Printf("Created client: %s (ID: %s, Rate: $%s/hr)\n", client.Name, client.ID, client.HourlyRate.StringFixed(2))

	// Show retainer info if set
	if client.RetainerAmount != nil && client.RetainerAmount.GreaterThan(decimal.Zero) {
		fmt.Printf("Retainer: $%s for %.1f hours per %s\n", client.RetainerAmount.StringFixed(2), *client.RetainerHours, *client.RetainerBasis)
	}

	// Show directory if set
	if client.Dir != nil && *client.Dir != "" {
		fmt.Printf("Directory: %s\n", *client.Dir)
	}

	return nil
}

func newClientsListCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all clients with their hourly rates",
		Long:  "Display a list of all clients along with their configured hourly rates for billing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients, err := timesheetService.ListClients(ctx)
			if err != nil {
				return fmt.Errorf("failed to list clients: %w", err)
			}

			if len(clients) == 0 {
				fmt.Println("No clients found.")
				return nil
			}

			fmt.Println("Clients:")
			for _, client := range clients {
				rateStr := fmt.Sprintf("$%s/hr", client.HourlyRate.StringFixed(2))
				if client.HourlyRate.Equal(decimal.Zero) {
					rateStr = "No rate set"
				}

				if verbose {
					fmt.Printf("\nClient: %s (ID: %s)\n", client.Name, client.ID)
					fmt.Printf("  Rate: %s\n", rateStr)
					timesheetService.DisplayClient(ctx, client)
				} else {
					fmt.Printf("%s - %s - %s\n", client.ID, client.Name, rateStr)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed billing information")
	return cmd
}

func newClientsUpdateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var hourlyRate float64
	var companyName, contactName, email, phone string
	var addressLine1, addressLine2, city, state, postalCode, country, abn, dir string
	var retainerAmount, retainerHours float64
	var retainerBasis string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update details about a client",
		Long:  "Update attributes of the client, such as hourly rate and billing details.",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.Flags().Float64VarP(&hourlyRate, "rate", "r", 0.0, "Hourly rate for the client")

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
	cmd.Flags().StringVar(&abn, "abn", "", "Australian Business Number (ABN)")
	cmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory path for the client")

	// Retainer flags
	cmd.Flags().Float64Var(&retainerAmount, "retainer-amount", 0.0, "Retainer amount (e.g., 5000.00)")
	cmd.Flags().Float64Var(&retainerHours, "retainer-hours", 0.0, "Hours covered by retainer (e.g., 40.0)")
	cmd.Flags().StringVar(&retainerBasis, "retainer-basis", "", "Retainer billing basis: day, week, month, quarter, year")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client := args[0]
		if client == "" {
			return fmt.Errorf("client name is required")
		}

		var hourlyRateDecimal *decimal.Decimal
		var retainerAmountDecimal *decimal.Decimal
		var retainerHoursPtr *float64

		// Helper function to convert empty strings to nil pointers
		stringPtr := func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}

		if hourlyRate > 0 {
			rate := decimal.NewFromFloat(hourlyRate)
			hourlyRateDecimal = &rate
		}
		if retainerAmount > 0 {
			amount := decimal.NewFromFloat(retainerAmount)
			retainerAmountDecimal = &amount
		}
		if retainerHours > 0 {
			retainerHoursPtr = &retainerHours
		}

		updatedClient, err := timesheetService.UpdateClient(ctx, client, &database.ClientUpdateDetails{
			HourlyRate:     hourlyRateDecimal,
			CompanyName:    stringPtr(companyName),
			ContactName:    stringPtr(contactName),
			Email:          stringPtr(email),
			Phone:          stringPtr(phone),
			AddressLine1:   stringPtr(addressLine1),
			AddressLine2:   stringPtr(addressLine2),
			City:           stringPtr(city),
			State:          stringPtr(state),
			PostalCode:     stringPtr(postalCode),
			Country:        stringPtr(country),
			Abn:            stringPtr(abn),
			Dir:            stringPtr(dir),
			RetainerAmount: retainerAmountDecimal,
			RetainerHours:  retainerHoursPtr,
			RetainerBasis:  stringPtr(retainerBasis),
		})
		if err != nil {
			return fmt.Errorf("failed to update client billing: %w", err)
		}

		fmt.Printf("Updated client '%s'\nNew state: \n", updatedClient.Name)
		timesheetService.DisplayClient(ctx, updatedClient)
		return nil
	}

	return cmd
}
