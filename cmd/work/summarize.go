package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/service"
)

func newSummarizeCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize and analyze client data",
		Long:  "Commands for summarizing and analyzing client data, including directory-based analysis.",
	}

	cmd.AddCommand(newSummarizeDescriptionsCmd(timesheetService))

	return cmd
}

func newSummarizeDescriptionsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var period string
	var date string

	cmd := &cobra.Command{
		Use:   "descriptions",
		Short: "Summarize descriptions from client directories",
		Long:  "Analyze client directories and summarize descriptions for the specified time period.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return summarizeDescriptions(ctx, timesheetService, period, date)
		},
	}

	cmd.Flags().StringVarP(&period, "period", "p", "week", "Period type: day, week, fortnight, month")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in the period (YYYY-MM-DD)")

	return cmd
}

func summarizeDescriptions(ctx context.Context, timesheetService *service.TimesheetService, period string, date string) error {
	// Default to today if no date specified
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Parse the date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Calculate date range based on period
	fromDate, toDate := calculatePeriodRange(period, targetDate)

	// Get all clients that have directories
	clients, err := timesheetService.GetClientsWithDirectories(ctx)
	if err != nil {
		return fmt.Errorf("failed to get clients with directories: %w", err)
	}

	if len(clients) == 0 {
		fmt.Println("No clients with directories found.")
		return nil
	}

	fmt.Printf("Found %d clients with directories for period: %s %s\n", len(clients), period, date)
	fmt.Printf("Date range: %s to %s\n", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Process directories concurrently
	var wg sync.WaitGroup
	for _, client := range clients {
		if client.Dir != nil {
			wg.Add(1)
			go func(clientName, dir string) {
				defer wg.Done()
				processDirectory(clientName, dir, fromDate, toDate)
			}(client.Name, *client.Dir)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()
	fmt.Println("All directories processed.")

	return nil
}

// processDirectory is a placeholder function that will eventually analyze git repositories
func processDirectory(clientName, dir string, fromDate, toDate time.Time) {
	fmt.Printf("Processing directory for client '%s': %s\n", clientName, dir)
	fmt.Printf("  Date range: %s to %s\n", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
	// TODO: Implement git analysis logic here
	// This should:
	// 1. Check if dir is a git repository or contains git repositories
	// 2. Get git log for the specified period using fromDate and toDate
	// 3. Extract and summarize commit messages
	// 4. Generate descriptions based on the changes
}
