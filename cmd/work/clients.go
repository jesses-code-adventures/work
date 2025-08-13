package main

import (
	"context"
	"fmt"

	"github.com/jessewilliams/work/internal/database"
	"github.com/jessewilliams/work/internal/service"
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
			return listClients(ctx, timesheetService, verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed billing information")
	return cmd
}

func newClientsUpdateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var hourlyRate float32
	var client string
	var companyName, contactName, email, phone string
	var addressLine1, addressLine2, city, state, postalCode, country, taxNumber string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update details about a client",
		Long:  "Update attributes of the client, such as hourly rate and billing details.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if client == "" {
				return fmt.Errorf("client name is required")
			}

			// If only hourly rate is provided, use the old function
			if hourlyRate > 0 && !hasBillingFlags(cmd) {
				return updateClient(ctx, timesheetService, client, float64(hourlyRate))
			}

			// If billing details are provided, update billing
			if hasBillingFlags(cmd) {
				return updateClientBilling(ctx, timesheetService, client, &database.ClientBillingDetails{
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
				})
			}

			return fmt.Errorf("either hourly rate or billing details must be provided")
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Name of the client to update")
	cmd.Flags().Float32VarP(&hourlyRate, "rate", "r", 0.0, "Hourly rate for the client")

	// Billing detail flags
	cmd.Flags().StringVar(&companyName, "company", "", "Company name")
	cmd.Flags().StringVar(&contactName, "contact", "", "Contact person name")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&addressLine1, "address1", "", "Address line 1")
	cmd.Flags().StringVar(&addressLine2, "address2", "", "Address line 2")
	cmd.Flags().StringVar(&city, "city", "", "City")
	cmd.Flags().StringVar(&state, "state", "", "State/Province")
	cmd.Flags().StringVar(&postalCode, "postal", "", "Postal/ZIP code")
	cmd.Flags().StringVar(&country, "country", "", "Country")
	cmd.Flags().StringVar(&taxNumber, "tax", "", "Tax/VAT number")

	return cmd
}

func listClients(ctx context.Context, timesheetService *service.TimesheetService, verbose bool) error {
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

			if client.CompanyName != nil {
				fmt.Printf("  Company: %s\n", *client.CompanyName)
			}
			if client.ContactName != nil {
				fmt.Printf("  Contact: %s\n", *client.ContactName)
			}
			if client.Email != nil {
				fmt.Printf("  Email: %s\n", *client.Email)
			}
			if client.Phone != nil {
				fmt.Printf("  Phone: %s\n", *client.Phone)
			}
			if client.AddressLine1 != nil {
				fmt.Printf("  Address: %s", *client.AddressLine1)
				if client.AddressLine2 != nil {
					fmt.Printf(", %s", *client.AddressLine2)
				}
				fmt.Printf("\n")
			}
			if client.City != nil || client.State != nil || client.PostalCode != nil {
				fmt.Printf("  Location: ")
				if client.City != nil {
					fmt.Printf("%s", *client.City)
				}
				if client.State != nil {
					fmt.Printf(", %s", *client.State)
				}
				if client.PostalCode != nil {
					fmt.Printf(" %s", *client.PostalCode)
				}
				fmt.Printf("\n")
			}
			if client.Country != nil {
				fmt.Printf("  Country: %s\n", *client.Country)
			}
			if client.TaxNumber != nil {
				fmt.Printf("  Tax Number: %s\n", *client.TaxNumber)
			}
		} else {
			fmt.Printf("%s - %s - %s\n", client.ID, client.Name, rateStr)
		}
	}

	return nil
}

func updateClient(ctx context.Context, timesheetService *service.TimesheetService, client string, rate float64) error {
	clients, err := timesheetService.UpdateClient(ctx, client, rate)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	fmt.Printf("Updated client '%s' to $%v\n", clients.Name, clients.HourlyRate)
	return nil
}

func updateClientBilling(ctx context.Context, timesheetService *service.TimesheetService, clientName string, billing *database.ClientBillingDetails) error {
	client, err := timesheetService.UpdateClientBilling(ctx, clientName, billing)
	if err != nil {
		return fmt.Errorf("failed to update client billing: %w", err)
	}

	fmt.Printf("Updated billing details for client '%s'\n", client.Name)
	return nil
}

func hasBillingFlags(cmd *cobra.Command) bool {
	flags := []string{"company", "contact", "email", "phone", "address1", "address2", "city", "state", "postal", "country", "tax"}
	for _, flag := range flags {
		if cmd.Flags().Changed(flag) {
			return true
		}
	}
	return false
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
