package main

import (
	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newGitCheckCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-check <session-id>",
		Short: "Debug git commands for a specific session",
		Long:  "Shows exactly what git commands are executed for a session's time period and their outputs.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			return timesheetService.GitCheckSession(sessionID)
		},
	}

	return cmd
}
