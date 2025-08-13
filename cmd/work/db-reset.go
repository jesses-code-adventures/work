package main

import (
	"fmt"
	"os"

	"github.com/jessewilliams/work/internal/config"
	"github.com/jessewilliams/work/internal/database"
	"github.com/jessewilliams/work/internal/service"
	"github.com/spf13/cobra"
)

func newDbResetCmd(timesheetService *service.TimesheetService) *cobra.Command {
	return &cobra.Command{
		Use:   "db-reset",
		Short: "Delete and recreate the SQLite database",
		Long: `Delete the existing SQLite database file and recreate it with a fresh schema.
This will permanently delete all existing work sessions and clients.

WARNING: This operation cannot be undone!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("WARNING: This will permanently delete all work sessions and clients!")
			fmt.Print("Are you sure you want to continue? (y/N): ")

			var response string
			fmt.Scanln(&response)

			if response != "y" && response != "Y" {
				fmt.Println("Database reset cancelled.")
				return nil
			}

			dbPath := cfg.DatabaseURL
			if dbPath == "" {
				dbPath = "work.db"
			}

			// Remove the existing database file
			if _, err := os.Stat(dbPath); err == nil {
				if err := os.Remove(dbPath); err != nil {
					return fmt.Errorf("failed to delete database file: %w", err)
				}
				fmt.Printf("Deleted existing database: %s\n", dbPath)
			}

			// Create a new database connection to initialize the file
			db, err := database.NewSQLiteDB(dbPath)
			if err != nil {
				return fmt.Errorf("failed to create new database: %w", err)
			}
			defer db.Close()

			// Execute the schema creation SQL
			schemaSQL := `
-- Create clients table
CREATE TABLE clients (
    id TEXT PRIMARY KEY NOT NULL,
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    hourly_rate DECIMAL(10,2) DEFAULT 0.00
);

-- Create sessions table
CREATE TABLE sessions (
    id TEXT PRIMARY KEY NOT NULL,
    client_id TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    hourly_rate DECIMAL(10,2),
    FOREIGN KEY (client_id) REFERENCES clients(id)
);

-- Create indexes for performance
CREATE INDEX idx_sessions_client_id ON sessions(client_id);
CREATE INDEX idx_sessions_start_time ON sessions(start_time);
CREATE INDEX idx_sessions_end_time ON sessions(end_time);
CREATE INDEX idx_clients_name ON clients(name);

-- Create trigger to update updated_at on clients
CREATE TRIGGER clients_updated_at 
    AFTER UPDATE ON clients 
    BEGIN
        UPDATE clients SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Create trigger to update updated_at on sessions
CREATE TRIGGER sessions_updated_at 
    AFTER UPDATE ON sessions 
    BEGIN
        UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
`

			// Get the underlying SQL connection to execute raw SQL
			conn := db.GetConnection()
			if _, err := conn.Exec(schemaSQL); err != nil {
				return fmt.Errorf("failed to create database schema: %w", err)
			}

			fmt.Printf("Successfully recreated database: %s\n", dbPath)
			fmt.Println("Database is ready for use with a fresh schema.")

			return nil
		},
	}
}
