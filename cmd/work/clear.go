package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jessewilliams/work/internal/service"
)

func newClearCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var fromDate, toDate string
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete work sessions",
		Long:  "Delete work sessions. Use with caution - this action cannot be undone.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if !force {
				fmt.Print("This will permanently delete work sessions. Are you sure? (y/N): ")
				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if fromDate != "" || toDate != "" {
				if fromDate == "" {
					fromDate = "1900-01-01"
				}
				if toDate == "" {
					toDate = "2099-12-31"
				}

				err := timesheetService.DeleteSessionsByDateRange(ctx, fromDate, toDate)
				if err != nil {
					return err
				}

				fmt.Printf("Deleted work sessions from %s to %s\n", fromDate, toDate)
			} else {
				err := timesheetService.DeleteAllSessions(ctx)
				if err != nil {
					return err
				}

				fmt.Println("Deleted all work sessions")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "Delete sessions from this date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&toDate, "to", "t", "", "Delete sessions to this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
