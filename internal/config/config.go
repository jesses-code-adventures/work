package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseName         string
	DatabasePath         string
	DatabaseURL          string
	DatabaseDriver       string
	TempDir              string
	GitAnalysisPrompt    string
	DevMode              bool
	BillingBank          string
	BillingAccountName   string
	BillingAccountNumber string
	BillingBSB           string
}

func Load(dbConn, dbDriver, gitPrompt, devMode, billingBank, billingAccountName, billingAccountNumber, billingBSB string) (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	if dbConn == "" {
		dbConn = getEnv("DATABASE_URL", "./work.db")
	}

	if dbDriver == "" {
		dbDriver = getEnv("DATABASE_DRIVER", "sqlite3")
	}

	if gitPrompt == "" {
		gitPrompt = getEnv("GIT_ANALYSIS_PROMPT", "use git log --since=\"{from_date}\" --until=\"{to_date}\" to review the commits between date {from_date} and date {to_date}. create a curt list of dot points explaining what has been done in the commits. feel free to look at the diffs in the commits themselves if needed for clarification. if there are no commits, say NO COMMITS and nothing else.")
	}

	if billingBank == "" {
		billingBank = getEnv("BILLING_BANK", "bank")
	}

	if billingAccountName == "" {
		billingAccountName = getEnv("BILLING_ACCOUNT_NAME", "account name")
	}

	if billingAccountNumber == "" {
		billingAccountNumber = getEnv("BILLING_ACCOUNT_NUMBER", "account number")
	}

	if billingBSB == "" {
		billingBSB = getEnv("BILLING_BSB", "bsb")
	}

	// Dev mode defaults to true for local builds, false for prod
	isDevMode := devMode == "true" || (devMode == "" && getEnv("DEV_MODE", "true") == "true")

	cfg := &Config{
		DatabaseName:         getEnv("DATABASE_NAME", "work"),
		DatabaseURL:          dbConn,
		DatabaseDriver:       dbDriver,
		GitAnalysisPrompt:    gitPrompt,
		DevMode:              isDevMode,
		BillingBank:          billingBank,
		BillingAccountName:   billingAccountName,
		BillingAccountNumber: billingAccountNumber,
		BillingBSB:           billingBSB,
	}

	return cfg, nil
}

func (c *Config) Dump() {
	fmt.Printf("Database Name: %s\n", c.DatabaseName)
	fmt.Printf("Database URL: %s\n", c.DatabaseURL)
	fmt.Printf("Database Driver: %s\n", c.DatabaseDriver)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	value := getEnv(key, "")
	if value == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return value
}
