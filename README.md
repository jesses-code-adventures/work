# Work CLI

A simple command-line work time tracker for freelance software engineers working with multiple clients.

## Features

- **Start/Stop work sessions**: Track time for different clients
- **Auto client creation**: Clients are created automatically when starting work
- **Session switching**: Starting work for a new client automatically stops the current session  
- **Time tracking**: View durations and session history
- **Hourly rate tracking**: Set rates per client for billing calculations
- **Billable amount calculation**: Automatic calculation of earnings (time × rate)
- **Enhanced exports**: CSV includes rates and billable amounts for invoicing
- **SQLite storage**: Local database with migration path to PostgreSQL
- **UUID v7 primary keys**: Future-proof identifiers

## Installation

```bash
# Clone and build
git clone <repo>
cd work
make setup
make build

# Install globally (optional)
make install
```

## Quick Start

```bash
# Create a client (optional - auto-created when starting work)
work create -c givetel

# Start working
work start -c givetel

# Check status
work status

# Switch clients (auto-stops current session)
work start -c another-client

# Stop working
work stop

# View recent sessions
work list
```

## Commands

### `work start -c <client>`
Start a work session for a client. If another session is active, it will be stopped automatically.

```bash
work start -c givetel
work start -c givetel -d "Working on authentication feature"
```

### `work stop`
Stop the current active work session.

### `work status`
Show the currently active work session and its duration.

### `work create -c <client>`
Explicitly create a client (extensible for future entity types).

### `work list`
Show recent work sessions with durations, billable amounts, and status.

```bash
work list -l 20  # Show last 20 sessions with billing info
work list -f 2025-08-01 -t 2025-08-31  # Show sessions for August 2025
```

### `work export`
Export work sessions to CSV format with hourly rates and billable amounts.

```bash
work export  # Export to stdout with billing data
work export -o billing.csv  # Export to file with rates and amounts
work export -f 2025-08-01 -t 2025-08-31 -o august-billing.csv  # Export date range
```

### `work clear`
Delete work sessions (use with caution).

```bash
work clear  # Delete all sessions (with confirmation)
work clear --force  # Delete all sessions (no confirmation)
work clear -f 2025-08-01 -t 2025-08-31  # Delete sessions in date range
```

## Database

- **Default**: SQLite database stored in `./work.db`
- **Configuration**: Edit `.env` file to change database settings
- **Migration ready**: Interface-based design allows easy switch to PostgreSQL

### Schema
- `clients`: Client information with UUID v7 primary keys and hourly rates
- `work_sessions`: Time tracking with start/end times, hourly rates at time of work, linked to clients

## Development

```bash
# Generate database queries
make sqlc-gen

# Run migrations (when using PostgreSQL)
make migrate

# Build
make build

# Run tests
make test

# Database inspection
make db-schema    # Show database schema
make db-inspect   # Show schema + data overview
make db-stats     # Show database statistics
```

## Configuration

Edit `.env` file:
```
DATABASE_URL=./work.db
DATABASE_DRIVER=sqlite3

# For PostgreSQL:
# DATABASE_URL=postgres://user:password@localhost/work?sslmode=disable
# DATABASE_DRIVER=postgres
```

## Architecture

- **CLI**: Cobra-based command interface
- **Service Layer**: Business logic with session management
- **Database Interface**: Pluggable database adapters
- **Type Safety**: sqlc-generated queries
- **UUID v7**: Time-ordered identifiers for better performance

## Billing Features

The app now includes comprehensive billing support:

- **Hourly Rates**: Set rates per client (stored as DECIMAL(10,2))
- **Rate Preservation**: Session records capture client rate at time of work
- **Automatic Calculations**: Duration × Rate = Billable Amount
- **Enhanced Output**: Both `list` and `export` show billing information

### CSV Export Format
When using `work export`, the CSV includes these columns:
- ID, Client, Start Time, End Time, Duration (minutes)
- **Hourly Rate**: Rate active during the session
- **Billable Amount**: Calculated earnings for the session
- Description, Date

This format is ready for importing into invoicing systems or accounting software.