package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jessewilliams/work/internal/service"
)

func newStartCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var clientName string
	var description string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a work session",
		Long:  "Start a new work session for a client. This will automatically stop any active session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientName == "" {
				return fmt.Errorf("client name is required (use -c flag)")
			}

			ctx := cmd.Context()

			var desc *string
			if description != "" {
				desc = &description
			}

			session, err := timesheetService.StartWork(ctx, clientName, desc)
			if err != nil {
				return err
			}

			fmt.Printf("Started work session for %s at %s\n",
				clientName,
				session.StartTime.Format("15:04:05"))

			if desc != nil {
				fmt.Printf("Description: %s\n", *desc)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&clientName, "client", "c", "", "Client name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Optional description of the work")
	cmd.MarkFlagRequired("client")

	return cmd
}
