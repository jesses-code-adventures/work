package main

import (
	"fmt"
	"github.com/jesses-code-adventures/work/internal/service"
	"github.com/spf13/cobra"
)

func newNoteCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note <text>",
		Short: "Add a note to the active session",
		Long:  "Add a note to the currently active work session. Notes are stored as bullet points and included in invoices and exports.",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		note := args[0]

		activeSession, err := timesheetService.GetActiveSession(ctx)
		if err != nil {
			return fmt.Errorf("failed to get active session: %w", err)
		}

		if activeSession == nil {
			return fmt.Errorf("no active session found. Start a session first with 'work start <client>'")
		}

		updatedSession, err := timesheetService.AddSessionNote(ctx, activeSession.ID, note)
		if err != nil {
			return fmt.Errorf("failed to add note to session: %w", err)
		}

		fmt.Printf("Added note to session for %s:\n", activeSession.ClientName)
		fmt.Printf("- %s\n", note)

		if updatedSession.OutsideGit != nil {
			fmt.Printf("\nAll notes for this session:\n%s\n", *updatedSession.OutsideGit)
		}
		return nil
	}

	return cmd
}
