# Work CLI

Simple CLI time tracker for freelancers. Track work sessions across multiple clients with automatic billing calculations and PDF invoice generation.

**Platform Support:** Apple Silicon (macOS ARM64) only

## Installation

```bash
# Auto-detect platform and install
curl -sSL https://raw.githubusercontent.com/jesses-code-adventures/work/main/install.sh | bash

# Or manually for macOS Apple Silicon:
curl -L https://github.com/jesses-code-adventures/work/releases/download/latest/work -o work && chmod +x work && sudo mv work /usr/local/bin/
```

## Features

- **Time Tracking**: Start/stop sessions with automatic client switching
- **Billing**: Set hourly rates per client with automatic calculations  
- **Invoicing**: Generate professional PDF invoices with client details
- **Export**: CSV export for accounting software integration
- **Client Management**: Store billing details (company, address, tax numbers)

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