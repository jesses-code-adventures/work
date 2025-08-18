.PHONY: build install sqlc-gen dev test clean deps db-schema db-inspect db-stats db-query

-include .env .env.mine

DATABASE_NAME := $(BIN_NAME)
DB_FILE := $(DATABASE_NAME).db
MIGRATIONS := ./migrations
DATABASE_URL := ./$(DB_FILE)


PROD_DATABASE := $(PROD_DATABASE_URL)?authToken=$(TURSO_TOKEN)

# Build binary
build:
	go build -o bin/$(BIN_NAME) \
		-ldflags "-X 'main.GitPrompt=$(GIT_ANALYSIS_PROMPT)'" \
		./cmd/$(BIN_NAME)

prod-build:
	go build -ldflags "\
		-X 'main.DBConn=$(PROD_DATABASE)' \
		-X 'main.DBDriver=$(PROD_DATABASE_DRIVER)' \
		-X 'main.GitPrompt=$(GIT_ANALYSIS_PROMPT)' \
		-X 'main.DevMode=false'" \
		-o bin/$(BIN_NAME) \
		./cmd/$(BIN_NAME)

check-turso-creds:
	@if ! grep -q '^TURSO_TOKEN=.*[^[:space:]]' .env.mine; then \
		echo "ERROR: TURSO_TOKEN is missing or empty in .env.mine" >&2; \
		exit 1; \
	fi
	@if ! grep -q '^PROD_DATABASE_URL=.*[^[:space:]]' .env.mine; then \
		echo "ERROR: PROD_DATABASE_URL is missing or empty in .env.mine" >&2; \
		exit 1; \
	fi

prod-turso-setup:
	@if ! command -v turso >/dev/null 2>&1; then \
		echo "Installing turso..."; \
		curl -sSfL https://get.tur.so/install.sh | bash; \
	else \
		echo "turso is already installed"; \
	fi
	@turso db create $(DATABASE_NAME) 2>&1 | grep -vq "already exists" || true

prod-init: check-turso-creds
	@for f in $(MIGRATIONS)/*.sql; do \
		echo "Executing migration: $$f"; \
		turso db shell "$(PROD_DATABASE)" < "$$f" || { echo "Migration failed: $$f" >&2; exit 1; }; \
	done
	@turso db shell "$(PROD_DATABASE)" < "scripts/tables.sql" || { echo "Migration failed: $$f" >&2; exit 1; }; \

prod-tables:
	@turso db shell "$(PROD_DATABASE)" < scripts/tables.sql

# Install to system
install: build
	cp bin/$(BIN_NAME) ~/.local/bin/$(BIN_NAME)

prod-install: prod-build
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
	work clients update -c givetel -r 100 -d "~/coding/givetel"
	work session create -c givetel -f "2025-08-14 16:30" -t "2025-08-15 02:30"
	work create -c personal
	work clients update -c personal -d "~/coding/personal"
	work session create -c personal -f "2025-08-18 18:30" -t "2025-08-19 01:30"
	work descriptions populate
	work list -v
	work export -d 2025-08-15 -o givetel.csv
	work invoices -p fortnight -d 2025-08-15
