.PHONY: build install migrate migrate-down sqlc-gen dev test clean deps db-schema db-inspect db-stats

BIN_NAME := work
DB_FILE := $(BIN_NAME).db

# Build binary
build:
	go build -o bin/$(BIN_NAME) ./cmd/$(BIN_NAME)

# Install to system
install: build
	cp bin/$(BIN_NAME) ~/.local/bin/$(BIN_NAME)

# Database schema dump
db-schema:
	@echo "=== Database Schema ==="
	@if [ -f "./$(DB_FILE)" ]; then \
		sqlite3 $(DB_FILE) ".schema"; \
	else \
		echo "Database file not found."; \
	fi

db-schema-dump:
	@if [ -f "./$(DB_FILE)" ]; then \
		sqlite3 $(DB_FILE) ".schema" > $(BIN_NAME).sql; \
	else \
		echo "Database file not found."; \
	fi

# Database inspection (schema + data overview)
db-inspect:
	@echo "=== Database Schema ==="
	@if [ -f "./$(DB_FILE)" ]; then \
		sqlite3 $(DB_FILE) ".schema" && \
		echo "" && \
		echo "=== Table Information ===" && \
		sqlite3 $(DB_FILE) ".tables" && \
		echo "" && \
		echo "=== Data Overview ===" && \
		sqlite3 $(DB_FILE) "SELECT 'Clients: ' || COUNT(*) FROM clients; SELECT 'Sessions: ' || COUNT(*) FROM $(BIN_NAME)_sessions; SELECT 'Active Sessions: ' || COUNT(*) FROM $(BIN_NAME)_sessions WHERE end_time IS NULL;"; \
	else \
		echo "Database file not found"; \
	fi

# Database statistics
db-stats:
	@echo "=== Database Statistics ==="
	@if [ -f "./$(DB_FILE)" ]; then \
		sqlite3 $(DB_FILE) -header -column \
		"SELECT \
			COUNT(DISTINCT c.name) as 'Total Clients', \
			COUNT(ws.id) as 'Total Sessions', \
			COUNT(CASE WHEN ws.end_time IS NULL THEN 1 END) as 'Active Sessions', \
			COUNT(CASE WHEN ws.end_time IS NOT NULL THEN 1 END) as 'Completed Sessions', \
			ROUND(AVG(CASE WHEN ws.end_time IS NOT NULL THEN \
				(julianday(ws.end_time) - julianday(ws.start_time)) * 24 * 60 END), 2) as 'Avg Duration (min)' \
		FROM clients c \
		LEFT JOIN $(BIN_NAME)_sessions ws ON c.id = ws.client_id"; \
	else \
		echo "Database file not found."; \
	fi

# Code generation
sqlc-gen:
	sqlc generate

# Development
dev: sqlc-gen
	go run ./cmd/$(BIN_NAME)

# Testing
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go mod download
