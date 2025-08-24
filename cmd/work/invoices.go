package main

import (
	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newInvoicesCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "Manage invoices for clients",
		Long:  "Manage invoices for clients.",
	}

	cmd.AddCommand(newInvoicesGenerateCmd(timesheetService))
	return cmd
}

func newInvoicesGenerateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var period string
	var date string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate PDF invoices for clients",
		Long:  "Generate PDF invoices for each client with billable hours > 0 in the specified period",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return timesheetService.GenerateInvoices(ctx, period, date)
		},
	}

	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in the period (YYYY-MM-DD)")
	cmd.MarkFlagRequired("date")

	return cmd
}
