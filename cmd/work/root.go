package main

import (
	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newRootCmd(timesheetService *service.TimesheetService) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "work",
		Short: "CLI work time tracker for freelance work",
		Long: `Track your work sessions across multiple clients with simple start/stop commands.
Supports hourly rate tracking and automatic billable amount calculations for freelance work.`,
	}

	rootCmd.AddCommand(
		newStartCmd(timesheetService),
		newStopCmd(timesheetService),
		newCreateCmd(timesheetService),
		newClientsCmd(timesheetService),
		newListCmd(timesheetService),
		newStatusCmd(timesheetService),
		newClearCmd(timesheetService),
		newExportCmd(timesheetService),
		newInvoicesCmd(timesheetService),
		newSummarizeCmd(timesheetService),
		newDescriptionsCmd(timesheetService),
		newSessionCmd(timesheetService),
	)

	return rootCmd
}
