package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newDescriptionsCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "descriptions",
		Short: "Manage session descriptions",
		Long:  "Commands for managing and populating session descriptions using git analysis.",
	}

	cmd.AddCommand(newDescriptionsPopulateCmd(timesheetService))

	return cmd
}

func newDescriptionsPopulateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var client string

	cmd := &cobra.Command{
		Use:   "populate",
		Short: "Populate missing session descriptions using git analysis",
		Long:  "Gets all sessions missing descriptions and runs summarize analysis using the session start/end times to populate descriptions and full work summaries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return populateDescriptions(ctx, timesheetService, client)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Process only the specified client (optional)")

	return cmd
}

func populateDescriptions(ctx context.Context, timesheetService *service.TimesheetService, clientName string) error {
	// Get clients with directories
	var clients []*models.Client
	var err error

	if clientName != "" {
		// Get specific client by name
		client, err := timesheetService.GetClientByName(ctx, clientName)
		if err != nil {
			return fmt.Errorf("failed to get client '%s': %w", clientName, err)
		}

		// Check if client has a directory
		if client.Dir == nil || *client.Dir == "" {
			return fmt.Errorf("client '%s' does not have a directory configured", clientName)
		}

		clients = []*models.Client{client}
		fmt.Printf("Processing single client: %s\n", clientName)
	} else {
		// Get all clients that have directories
		clients, err = timesheetService.GetClientsWithDirectories(ctx)
		if err != nil {
			return fmt.Errorf("failed to get clients with directories: %w", err)
		}

		if len(clients) == 0 {
			fmt.Println("No clients with directories found.")
			return nil
		}
	}

	fmt.Printf("Found %d clients with directories\n", len(clients))

	// For each client, get sessions without descriptions
	totalProcessed := 0
	for _, client := range clients {
		sessions, err := timesheetService.GetSessionsWithoutDescription(ctx, &client.Name)
		if err != nil {
			fmt.Printf("Error getting sessions for client %s: %v\n", client.Name, err)
			continue
		}

		if len(sessions) == 0 {
			fmt.Printf("No sessions missing descriptions for client: %s\n", client.Name)
			continue
		}

		fmt.Printf("Processing %d sessions for client: %s\n", len(sessions), client.Name)

		for _, session := range sessions {
			if session.EndTime == nil {
				fmt.Printf("  Skipping active session %s (not ended)\n", session.ID)
				continue
			}

			fmt.Printf("  Processing session %s (%s to %s)\n",
				session.ID,
				session.StartTime.Format("2006-01-02 15:04"),
				session.EndTime.Format("2006-01-02 15:04"))

			// Run summarize analysis for this session's time period
			err := analyzeAndUpdateSession(ctx, timesheetService, client, session)
			if err != nil {
				fmt.Printf("    Error analyzing session: %v\n", err)
				continue
			}

			totalProcessed++
			fmt.Printf("    Successfully updated session description\n")
		}
	}

	fmt.Printf("\nCompleted! Processed %d sessions total.\n", totalProcessed)
	return nil
}

func analyzeAndUpdateSession(ctx context.Context, timesheetService *service.TimesheetService, client *models.Client, session *models.WorkSession) error {
	// Calculate the period as "day" and use the session start date
	period := "day"
	date := session.StartTime.Format("2006-01-02")

	// Run the summarize analysis for this specific client and time period
	result, err := performSummarizeAnalysis(ctx, timesheetService, period, date, client.Name)
	if err != nil {
		return fmt.Errorf("failed to perform analysis: %w", err)
	}

	// Update the session with the results
	_, err = timesheetService.UpdateSessionDescription(ctx, session.ID, result.FinalSummary, &result.FullWorkSummary)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}
