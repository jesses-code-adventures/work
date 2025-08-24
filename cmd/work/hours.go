package main

import (
	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newHoursCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var client string
	var period string
	var periodDate string
	var fromDate string
	var toDate string

	cmd := &cobra.Command{
		Use:   "hours",
		Short: "Display total worked hours",
		Long:  "Display total worked hours with optional filtering by client, period, or date range.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return timesheetService.ShowTotalHours(ctx, client, period, periodDate, fromDate, toDate)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Filter by client name")
	cmd.Flags().StringVarP(&period, "period", "p", "", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&periodDate, "date", "d", "", "Date in the period (YYYY-MM-DD), defaults to today when using -p")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Show hours from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Show hours to this date (YYYY-MM-DD)")

	return cmd
}
