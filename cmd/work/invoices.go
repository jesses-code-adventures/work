package main

import (
	"time"

	"github.com/shopspring/decimal"
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
	cmd.AddCommand(newInvoicesPayCmd(timesheetService))
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

func newInvoicesPayCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var amount float64
	var dateStr string

	cmd := &cobra.Command{
		Use:   "pay",
		Short: "Pay an invoice",
		Long:  "Pay an invoice with the specified invoice number and amount, amount defaults to the total amount of the invoice",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().Float64VarP(&amount, "amount", "a", 0.0, "Amount being paid, defaulting to the total amount of the invoice")
	cmd.Flags().StringVarP(&dateStr, "date", "d", "", "Date the payment was made (YYYY-MM-DD)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		id := args[0]
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil && dateStr != "" {
			return err
		}
		return timesheetService.PayInvoice(ctx, id, decimal.NewFromFloat(amount), date)
	}

	return cmd
}
