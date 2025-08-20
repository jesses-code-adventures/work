package main

import (
	"fmt"

	"github.com/jesses-code-adventures/work/internal/database"
	"github.com/jesses-code-adventures/work/internal/service"
	"github.com/spf13/cobra"
)

func newClientsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clients",
		Short: "Manage clients",
		Long:  "Commands for managing clients, including listing clients and their hourly rates.",
	}

	cmd.AddCommand(newClientsListCmd(timesheetService))
	cmd.AddCommand(newClientsUpdateCmd(timesheetService))

	return cmd
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
				rateStr := fmt.Sprintf("$%.2f/hr", client.HourlyRate)
				if client.HourlyRate == 0.0 {
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
	var client string
	var companyName, contactName, email, phone string
	var addressLine1, addressLine2, city, state, postalCode, country, taxNumber, dir string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update details about a client",
		Long:  "Update attributes of the client, such as hourly rate and billing details.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if client == "" {
				return fmt.Errorf("client name is required")
			}

			updatedClient, err := timesheetService.UpdateClient(ctx, client, &database.ClientUpdateDetails{
				HourlyRate:   &hourlyRate,
				CompanyName:  stringPtr(companyName),
				ContactName:  stringPtr(contactName),
				Email:        stringPtr(email),
				Phone:        stringPtr(phone),
				AddressLine1: stringPtr(addressLine1),
				AddressLine2: stringPtr(addressLine2),
				City:         stringPtr(city),
				State:        stringPtr(state),
				PostalCode:   stringPtr(postalCode),
				Country:      stringPtr(country),
				TaxNumber:    stringPtr(taxNumber),
				Dir:          stringPtr(dir),
			})
			if err != nil {
				return fmt.Errorf("failed to update client billing: %w", err)
			}

			fmt.Printf("Updated client '%s'\nNew state: \n", updatedClient.Name)
			timesheetService.DisplayClient(ctx, updatedClient)
			return nil
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Name of the client to update")
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
	cmd.Flags().StringVar(&taxNumber, "tax", "", "Tax/VAT number")
	cmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory path for the client")

	return cmd
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
