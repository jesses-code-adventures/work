package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jessewilliams/work/internal/service"
)

func newCreateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var clientName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create various entities",
		Long:  "Create clients and other entities. Clients are created with hourly rates for billing calculations. Use flags to specify what to create.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			switch {
			case clientName != "":
				return createClient(ctx, timesheetService, clientName)
			default:
				return fmt.Errorf("must specify what to create (use -c for client)")
			}
		},
	}

	cmd.Flags().StringVarP(&clientName, "client", "c", "", "Create a new client")

	return cmd
}

func createClient(ctx context.Context, timesheetService *service.TimesheetService, name string) error {
	client, err := timesheetService.CreateClient(ctx, name, 0.0)
	if err != nil {
		return err
	}

	fmt.Printf("Created client: %s (ID: %s, Rate: $%.2f/hr)\n", client.Name, client.ID, client.HourlyRate)
	return nil
}
