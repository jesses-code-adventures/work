package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jessewilliams/work/internal/service"
)

func newStopCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the current work session",
		Long:  "Stop the currently active work session and record the end time.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			session, err := timesheetService.StopWork(ctx)
			if err != nil {
				return err
			}

			duration := timesheetService.CalculateDuration(session)

			fmt.Printf("Stopped work session for %s\n", session.ClientName)
			fmt.Printf("Duration: %s\n", timesheetService.FormatDuration(duration))
			fmt.Printf("Started: %s, Ended: %s\n",
				session.StartTime.Format("15:04:05"),
				session.EndTime.Format("15:04:05"))

			return nil
		},
	}

	return cmd
}
