package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jessewilliams/work/internal/config"
	"github.com/jessewilliams/work/internal/database"
	"github.com/jessewilliams/work/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewSQLiteDB(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	timesheetService := service.NewTimesheetService(db)

	rootCmd := newRootCmd(timesheetService)
	return rootCmd.ExecuteContext(context.Background())
}
