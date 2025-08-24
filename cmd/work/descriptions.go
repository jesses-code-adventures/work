package main

import (
	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newDescriptionsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "descriptions",
		Short: "Manage session descriptions using git and AI summarization",
		Long:  "Commands for managing and populating session descriptions using git analysis.",
	}

	cmd.AddCommand(newDescriptionsGenerateCmd(timesheetService))

	return cmd
}

func newDescriptionsGenerateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var client string
	var period string
	var date string
	var session string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate missing session descriptions using git analysis",
		Long:  "Gets all sessions missing descriptions and runs summarize analysis using the session start/end times to populate descriptions and full work summaries.",
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Process only the specified client (optional)")
	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in the period (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&session, "session", "s", "", "The ID of the session to analyze")
	update := cmd.Flags().BoolP("update", "u", false, "Update the session descriptions in the database")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return timesheetService.GenerateDescriptions(ctx, client, session, *update)
	}

	return cmd
}
