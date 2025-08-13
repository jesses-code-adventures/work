package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newStatusCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current work status",
		Long:  "Display the currently active work session, if any.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			session, err := timesheetService.GetActiveSession(ctx)
			if err != nil {
				return err
			}

			if session == nil {
				fmt.Println("No active work session.")
				return nil
			}

			duration := timesheetService.CalculateDuration(session)
			billableAmount := timesheetService.CalculateBillableAmount(session)

			fmt.Printf("Active work session:\n")
			fmt.Printf("Client: %s\n", session.ClientName)
			fmt.Printf("Started: %s (%s)\n",
				session.StartTime.Format("15:04:05"),
				session.StartTime.Format("2006-01-02"))
			fmt.Printf("Duration: %s\n", timesheetService.FormatDuration(duration))
			fmt.Printf("Billable amount: %s\n", timesheetService.FormatBillableAmount(billableAmount))

			if session.Description != nil && *session.Description != "" {
				fmt.Printf("Description: %s\n", *session.Description)
			}

			return nil
		},
	}

	return cmd
}
