package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseName   string
	DatabasePath   string
	DatabaseURL    string
	DatabaseDriver string
	TempDir        string
}

func Load(dbConn, dbDriver string) (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	if dbConn == "" {
		dbConn = getEnv("DATABASE_URL", "./work.db")
	}

	if dbDriver == "" {
		dbDriver = getEnv("DATABASE_DRIVER", "sqlite3")
	}

	cfg := &Config{
		DatabaseName:   getEnv("DATABASE_NAME", "work"),
		DatabaseURL:    dbConn,
		DatabaseDriver: dbDriver,
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
