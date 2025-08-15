# Work CLI

Who needs interns when we have clankers to do our bidding?

Simple CLI time tracker for freelancers. Track work sessions across multiple clients with automatic billing calculations, PDF invoice generation, and AI-powered work summaries.

**Dependencies:** Requires [OpenCode](https://github.com/sst/opencode) for `summarize` and `descriptions populate` commands that analyze local git repositories to generate work summaries per-client. Requires your own [Turso](https://turso.tech/) sqlite database if you want to share sessions between machines.

## Installation

### Local Use Only (Single Machine)
```bash
# Clone and build
git clone <repository-url>
cd work
make install
```

### Remote Database Usage (If you need to track work across multiple machines)
```bash
# Clone and configure
git clone <repository-url>
cd work

# Configure .env with your Turso database
echo "TURSO_TOKEN=your_turso_token_here" >> .env.mine
echo "PROD_DATABASE_URL=libsql://your-database.turso.io" >> .env.mine

# Build and install
make prod-install
```

## Features

- **Time Tracking**: Start/stop sessions with automatic client switching
- **Session Management**: Create sessions with custom start/end times
- **Billing**: Set hourly rates per client with automatic calculations  
- **AI Work Summaries**: Analyze git repositories to generate work descriptions
- **Invoicing**: Generate professional PDF invoices with client details
- **Export**: CSV export for accounting software integration
- **Client Management**: Store billing details and directory paths
- **Database Querying**: Run arbitrary SQL queries against your timesheet data

## Commands

### Time Tracking

#### `work start`
Start a new work session for a client. Automatically stops any active session.
```bash
work start -c clientname                    # Start session for client
work start -c clientname -d "Bug fixes"     # With description
```

#### `work stop`
Stop the currently active work session.
```bash
work stop
```

#### `work status`
Show the currently active work session.
```bash
work status
```

### Session Management

#### `work list`
Show recent work sessions with durations and billable amounts.
```bash
work list                                   # Show last 10 sessions
work list -l 20                            # Show last 20 sessions  
work list -f 2025-08-01 -t 2025-08-31     # Filter by date range
work list -v                               # Verbose: include full work summaries
```

#### `work session create`
Create work sessions with custom start and end times.
```bash
work session create -c client -f "09:00" -t "17:00"                    # Today 9am-5pm
work session create -c client -f "2025-08-15 09:00" -t "2025-08-15 17:00"  # Specific date
work session create -c client -f "09:00" -t "17:00" -d "Development work"   # With description
```

#### `work clear`
Delete work sessions (use with caution).
```bash
work clear --force                         # Delete all sessions
work clear -f 2025-08-01 -t 2025-08-31    # Delete date range
```

### Client Management

#### `work create`
Create new clients.
```bash
work create -c clientname                  # Create new client
```

#### `work clients list`
Display clients with hourly rates.
```bash
work clients list                          # Basic list
work clients list -v                       # Detailed billing info
```

#### `work clients update`
Update client details and billing information.
```bash
work clients update -c clientname -r 150                    # Set hourly rate
work clients update -c clientname --company "ACME Corp"     # Company name
work clients update -c clientname --contact "John Doe"      # Contact person
work clients update -c clientname --email "john@acme.com"   # Email
work clients update -c clientname --phone "+1234567890"     # Phone
work clients update -c clientname --address1 "123 Main St"  # Address line 1
work clients update -c clientname --address2 "Suite 100"    # Address line 2
work clients update -c clientname --city "New York"         # City
work clients update -c clientname --state "NY"              # State/Province
work clients update -c clientname --postcode "10001"        # Postal code
work clients update -c clientname --country "USA"           # Country
work clients update -c clientname --tax "12-3456789"        # Tax/VAT number
work clients update -c clientname --dir "~/code/client"         # Git directory for analysis
```

### Reporting

#### `work export`
Export sessions to CSV format.
```bash
work export                                 # Export to stdout
work export -o timesheet.csv              # Export to file
work export -l 500                        # Limit to 500 sessions
work export -f 2025-08-01 -t 2025-08-31   # Export date range
```

#### `work invoices`
Generate PDF invoices for clients with billable hours.
```bash
work invoices                              # Week ending today
work invoices -p day -d 2025-08-23        # Single day
work invoices -p week -d 2025-08-23       # Week containing date
work invoices -p fortnight -d 2025-08-23  # Fortnight containing date
work invoices -p month -d 2025-08-23      # Month containing date
```

### AI-Powered Work Analysis

#### `work summarize descriptions`
Analyze git repositories to generate work summaries for specified time periods.
```bash
work summarize descriptions                              # Current week, all clients
work summarize descriptions -p day -d 2025-08-15       # Specific day
work summarize descriptions -c clientname              # Single client only
work summarize descriptions -p month -d 2025-08-01     # Entire month
```

#### `work descriptions populate`
Automatically populate missing session descriptions using git analysis. Requires clients to have directory paths configured.
```bash
work descriptions populate                    # All clients with directories  
work descriptions populate -c clientname     # Single client only
```

### Database Operations

#### `make db-query`
Execute arbitrary SQL queries against your timesheet database (development).
```bash
make db-query QUERY="SELECT * FROM clients"
make db-query QUERY="SELECT * FROM sessions WHERE start_time > '2025-08-01'" FORMAT=column
make db-query QUERY="SELECT client_name, COUNT(*) FROM sessions GROUP BY client_name" FORMAT=csv
```

## Usage Examples

### Complete Workflow
```bash
# Set up a client
work create -c acme
work clients update -c acme -r 150 --company "ACME Corp" --dir "~/code/acme"

# Track work
work start -c acme -d "Feature development"
# ... do work ...
work stop

# Generate descriptions from git activity
work descriptions populate -c acme

# View work with summaries
work list -v

# Generate invoice
work invoices -p week
```

### Custom Session Entry
```bash
# Add work done yesterday
work session create -c acme -f "2025-08-14 09:00" -t "2025-08-14 17:00" -d "Bug fixes"

# Analyze that period
work summarize descriptions -p day -d 2025-08-14 -c acme
```
