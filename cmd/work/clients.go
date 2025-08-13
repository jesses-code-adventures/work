package main

import (
	"context"
	"fmt"

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
	return &cobra.Command{
		Use:   "list",
		Short: "List all clients with their hourly rates",
		Long:  "Display a list of all clients along with their configured hourly rates for billing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return listClients(ctx, timesheetService)
		},
	}
}

func newClientsUpdateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var hourlyRate float32
	var client string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update details about a client",
		Long:  "Update attributes of the client, such as the hourly rate.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if client == "" {
				return fmt.Errorf("client name is required")
			}
			if hourlyRate == 0 {
				return fmt.Errorf("hourly rate is required")
			}
			return updateClient(ctx, timesheetService, client, float64(hourlyRate))
		},
	}
	cmd.Flags().StringVarP(&client, "client", "c", "", "Name of the client to update")
	cmd.Flags().Float32VarP(&hourlyRate, "rate", "r", 0.0, "Hourly rate for the client")
	return cmd
}

func listClients(ctx context.Context, timesheetService *service.TimesheetService) error {
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
		fmt.Printf("%s - %s - %s\n", client.ID, client.Name, rateStr)
	}

	return nil
}

func updateClient(ctx context.Context, timesheetService *service.TimesheetService, client string, rate float64) error {
	clients, err := timesheetService.UpdateClient(ctx, client, rate)
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	fmt.Printf("Updated client '%s' to $%v\n", clients.Name, clients.HourlyRate)
	return nil
}
