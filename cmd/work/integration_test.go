package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jesses-code-adventures/work/internal/config"
	"github.com/jesses-code-adventures/work/internal/database"
	"github.com/jesses-code-adventures/work/internal/service"
)

func TestIntegrationWorkCommands(t *testing.T) {
	// Create a temporary directory for test database
	tempDir, err := os.MkdirTemp("", "work-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup test database
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := &config.Config{
		DatabaseURL:    dbPath,
		DatabaseDriver: "sqlite3",
		DatabaseName:   "test",
		DevMode:        true,
	}

	// Initialize database
	db, err := database.NewDB(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Run migrations using sqlite3 command directly
	if err := runMigrationsWithSQLite(cfg); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create service
	timesheetService := service.NewTimesheetService(db, cfg)

	// Create root command
	rootCmd := newRootCmd(timesheetService)

	// Test context
	ctx := context.Background()

	t.Run("Work Start", func(t *testing.T) {
		// First create a client
		_, err := timesheetService.CreateClient(ctx, "test-client", 50.0, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create test client: %v", err)
		}

		// Test work start command
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"start", "-c", "test-client", "-d", "Test session"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work start command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Started work session") {
			t.Errorf("Expected 'Started work session' in output, got: %s", output)
		}
	})

	t.Run("Work Status", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"status"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work status command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Active work session") {
			t.Errorf("Expected 'Active work session' in output, got: %s", output)
		}
	})

	t.Run("Work Stop", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"stop"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work stop command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Stopped work session") {
			t.Errorf("Expected 'Stopped work session' in output, got: %s", output)
		}
	})

	t.Run("Work Git-Check", func(t *testing.T) {
		// Skip git-check test as it requires a git repository setup
		// and specific database configuration that's complex to set up in tests
		t.Skip("Skipping git-check test - requires git repository setup")
	})

	t.Run("Work Note", func(t *testing.T) {
		// Start a new session first
		_, err := timesheetService.StartWork(ctx, "test-client", nil)
		if err != nil {
			t.Fatalf("Failed to start work session: %v", err)
		}

		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"note", "Test note for session"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work note command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Added note to session") {
			t.Errorf("Expected 'Added note to session' in output, got: %s", output)
		}

		// Stop the session
		_, err = timesheetService.StopWork(ctx)
		if err != nil {
			t.Errorf("Failed to stop session: %v", err)
		}
	})

	t.Run("Work Sessions Create", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"sessions", "create", "-c", "test-client", "-f", "2025-08-20 10:00", "-t", "2025-08-20 12:00", "-d", "Test session"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work sessions create command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Created session for") {
			t.Errorf("Expected 'Created session for' in output, got: %s", output)
		}
	})

	t.Run("Work Sessions List", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"sessions", "list"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work sessions list command failed: %v", err)
			}
		})

		if !strings.Contains(output, "test-client") {
			t.Errorf("Expected client name in output, got: %s", output)
		}
	})

	t.Run("Work Sessions CSV", func(t *testing.T) {
		// Create a temporary CSV file
		csvFile := filepath.Join(tempDir, "test_export.csv")

		// Clear all sessions
		err := timesheetService.DeleteAllSessions(ctx)
		if err != nil {
			t.Fatalf("Failed to delete all sessions: %v", err)
		}

		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"sessions", "export", "-o", csvFile})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work sessions csv command failed: %v", err)
			}
		})

		if !strings.Contains(output, "No sessions found to export") {
			t.Errorf("Expected 'No sessions found to export' in output, got: %s", output)
		}

		if _, err := os.Stat(csvFile); err == nil {
			t.Errorf("CSV file was created and should not have been")
		}

		// Create a new session
		_, err = timesheetService.CreateSessionWithTimes(ctx, "test-client", time.Now(), time.Now(), nil)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		output = captureOutput(func() {
			rootCmd.SetArgs([]string{"sessions", "export", "-o", csvFile})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work sessions csv command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Exported") {
			t.Errorf("Expected 'Exported' in output, got: %s", output)
		}

		// Check if CSV file was created
		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			t.Errorf("CSV file was not created")
		}
	})

	t.Run("Work Clients Create", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"clients", "create", "new-client", "-r", "75.0"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work clients create command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Created client") {
			t.Errorf("Expected 'Created client' in output, got: %s", output)
		}
	})

	t.Run("Work Clients Update", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"clients", "update", "new-client", "-r", "80.0"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work clients update command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Updated client") {
			t.Errorf("Expected 'Updated client' in output, got: %s", output)
		}
	})

	t.Run("Work Clients List", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"clients", "list"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work clients list command failed: %v", err)
			}
		})

		if !strings.Contains(output, "test-client") && !strings.Contains(output, "new-client") {
			t.Errorf("Expected client names in output, got: %s", output)
		}
	})

	t.Run("Work Sessions Delete", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"sessions", "delete", "--force"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work sessions delete command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Deleted") {
			t.Errorf("Expected 'Deleted' in output, got: %s", output)
		}
	})

	t.Run("Work Invoices Generate", func(t *testing.T) {
		output := captureOutput(func() {
			rootCmd.SetArgs([]string{"invoices", "generate", "-d", "2025-08-20"})
			err := rootCmd.ExecuteContext(ctx)
			if err != nil {
				t.Errorf("Work invoices generate command failed: %v", err)
			}
		})

		if !strings.Contains(output, "Generated invoice") && !strings.Contains(output, "No invoices generated") {
			t.Errorf("Expected 'Generated invoice' or 'No invoices generated' in output, got: %s", output)
		}
	})
}

// Helper function to capture stdout output
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	io.Copy(&buf, r)
	return buf.String()
}

// Helper function to run database migrations using sqlite3 command
func runMigrationsWithSQLite(cfg *config.Config) error {
	// Read migration files
	migrationFiles := []string{
		"001_initial_schema.sql",
		"002_add_rates.sql",
		"003_add_billing_details.sql",
		"004_add_dir.sql",
		"005_add_full_work_summary.sql",
		"006_add_outside_git.sql",
	}

	for _, file := range migrationFiles {
		content, err := os.ReadFile(filepath.Join("../../migrations", file))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration using sqlite3 command
		cmd := exec.Command("sqlite3", cfg.DatabaseURL)
		cmd.Stdin = strings.NewReader(string(content))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}
