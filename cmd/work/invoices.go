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
	cmd.AddCommand(newInvoicesRegenerateCmd(timesheetService))
	cmd.AddCommand(newInvoicesListCmd(timesheetService))
	return cmd
}

func newInvoicesGenerateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var period string
	var date string
	var client string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate PDF invoices for clients",
		Long:  "Generate PDF invoices for each client with billable hours > 0 in the specified period",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return timesheetService.GenerateInvoices(ctx, period, date, client)
		},
	}

	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in the period (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Generate invoice for specific client only")
	cmd.MarkFlagRequired("date")

	return cmd
}

func newInvoicesRegenerateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var period string
	var date string
	var client string

	cmd := &cobra.Command{
		Use:   "regenerate",
		Short: "Regenerate invoices for a period (clears existing invoices for that period)",
		Long:  "Regenerate invoices for each client with billable hours > 0 in the specified period. This will clear existing invoices for the period and regenerate them.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return timesheetService.RegenerateInvoices(ctx, period, date, client)
		},
	}

	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in the period (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Regenerate invoice for specific client only")
	cmd.MarkFlagRequired("date")

	return cmd
}

func newInvoicesListCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var limit int32
	var client string
	var unpaidOnly bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices",
		Long:  "List invoices showing client, period, dates, amounts and payment status",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return timesheetService.ListInvoices(ctx, limit, client, unpaidOnly)
		},
	}

	cmd.Flags().Int32VarP(&limit, "limit", "l", 20, "Number of invoices to show")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Filter by specific client")
	cmd.Flags().BoolVarP(&unpaidOnly, "unpaid", "u", false, "Show only unpaid invoices")

	return cmd
}
