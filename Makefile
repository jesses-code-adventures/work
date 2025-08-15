.PHONY: build install migrate migrate-down sqlc-gen dev test clean deps db-schema db-inspect db-stats db-query

BIN_NAME := work
DB_FILE := $(BIN_NAME).db
MIGRATIONS := ./migrations

-include .env

PROD_DATABASE := $(PROD_DATABASE_URL)?authToken=$(TURSO_TOKEN)

# Build binary
build:
	go build -o bin/$(BIN_NAME) \
		-ldflags "-X 'main.GitPrompt=$(GIT_ANALYSIS_PROMPT)'" \
		./cmd/$(BIN_NAME)

build-prod:
	go build -ldflags "\
		-X 'main.DBConn=$(PROD_DATABASE)' \
		-X 'main.DBDriver=$(PROD_DATABASE_DRIVER)' \
		-X 'main.GitPrompt=$(GIT_ANALYSIS_PROMPT)' \
		-X 'main.DevMode=false'" \
		-o bin/$(BIN_NAME) \
		./cmd/$(BIN_NAME)

# Install to system
install: build
	cp bin/$(BIN_NAME) ~/.local/bin/$(BIN_NAME)

prod-install: build-prod
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

dump:
	@echo "Database Name: $(DATABASE_NAME)"
	@echo "Database Driver: $(DATABASE_DRIVER)"
	@echo "Local Database URL: $(DATABASE_URL)"
	@echo "Prod Database: $(PROD_DATABASE)"

prod-db-shell:
	turso db shell $(PROD_DATABASE)

db-reset:
	rm -f $(DB_FILE)
	for f in $(MIGRATIONS)/*.sql; do \
		echo "Executing migration: $$f"; \
		sqlite3 $(DB_FILE) ".read $$f"; \
	done
	$(MAKE) db-schema-dump

# Execute arbitrary SQL query
# Usage: make db-query QUERY="SELECT * FROM clients"
# Usage: make db-query QUERY="SELECT * FROM clients" FORMAT=column  (for column format)
# Usage: make db-query QUERY="SELECT * FROM clients" FORMAT=csv     (for CSV output)
db-query:
ifndef QUERY
	@echo "Usage: make db-query QUERY=\"<your SQL query>\""
	@echo "Examples:"
	@echo "  make db-query QUERY=\"SELECT * FROM clients\""
	@echo "  make db-query QUERY=\"SELECT * FROM clients\" FORMAT=column"
	@echo "  make db-query QUERY=\"SELECT * FROM clients\" FORMAT=csv"
	@echo "  make db-query QUERY=\"SELECT name FROM clients WHERE hourly_rate > 50\""
else
	@if [ -f "./$(DB_FILE)" ]; then \
		if [ "$(FORMAT)" = "column" ]; then \
			sqlite3 $(DB_FILE) -header -column "$(QUERY)"; \
		elif [ "$(FORMAT)" = "csv" ]; then \
			sqlite3 $(DB_FILE) -header -csv "$(QUERY)"; \
		else \
			sqlite3 $(DB_FILE) "$(QUERY)"; \
		fi \
	else \
		echo "Database file not found: $(DB_FILE)"; \
	fi
endif

e2e: db-reset install
	work create -c givetel
	work clients update -c givetel -r 100 -d ~/coding/givetel
	work session create -c givetel -f "2025-08-14 16:30" -t "2025-08-15 02:30"
	work descriptions populate
	$(MAKE) db-query QUERY="SELECT * FROM sessions WHERE client = 'givetel' ORDER BY start_time ASC"
