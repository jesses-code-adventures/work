package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jesses-code-adventures/work/internal/config"
	"github.com/jesses-code-adventures/work/internal/database"
	"github.com/jesses-code-adventures/work/internal/service"
)

var DBConn string
var DBDriver string
var GitPrompt string
var DevMode string
var BillingBank string
var BillingAccountName string
var BillingAccountNumber string
var BillingBSB string

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// TODO: test and use this
// func runWithEmbeddedReplica() error {
// 	cfg, err := config.Load(DBConn, DBDriver, GitPrompt, DevMode, BillingBank, BillingAccountName, BillingAccountNumber, BillingBSB)
// 	if err != nil {
// 		return fmt.Errorf("failed to load config: %w", err)
// 	}
//
// 	dir, err := os.MkdirTemp("", "libsql-*")
// 	if err != nil {
// 		fmt.Println("Error creating temporary directory:", err)
// 		os.Exit(1)
// 	}
// 	cfg.TempDir = dir
// 	cfg.DatabasePath = filepath.Join(dir, cfg.DatabaseName)
// 	defer os.RemoveAll(dir)
//
// 	db, err := database.NewTursoDBWithEmbeddedReplica(cfg)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to database: %w", err)
// 	}
// 	defer db.Close()
//
// 	timesheetService := service.NewTimesheetService(db, cfg)
//
// 	rootCmd := newRootCmd(timesheetService)
// 	return rootCmd.ExecuteContext(context.Background())
// }

func run() error {
	cfg, err := config.Load(DBConn, DBDriver, GitPrompt, DevMode, BillingBank, BillingAccountName, BillingAccountNumber, BillingBSB)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	timesheetService := service.NewTimesheetService(db, cfg)

	rootCmd := newRootCmd(timesheetService)
	return rootCmd.ExecuteContext(context.Background())
}
