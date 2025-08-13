# Work CLI

Simple CLI time tracker for freelancers. Track work sessions across multiple clients with automatic billing calculations and PDF invoice generation.

## Installation

```bash
# Auto-detect platform and install
curl -sSL https://raw.githubusercontent.com/jesses-code-adventures/work/main/install.sh | bash

# Or manually:
# Linux/WSL
curl -L https://github.com/jesses-code-adventures/work/releases/download/latest/work-linux-amd64 -o work && chmod +x work && sudo mv work /usr/local/bin/

# macOS (Intel)  
curl -L https://github.com/jesses-code-adventures/work/releases/download/latest/work-darwin-amd64 -o work && chmod +x work && sudo mv work /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/jesses-code-adventures/work/releases/download/latest/work-darwin-arm64 -o work && chmod +x work && sudo mv work /usr/local/bin/
```

## Basic Usage

```bash
work start -c clientname                    # Start tracking time
work stop                                   # Stop current session
work status                                 # Show active session
work list                                   # View recent sessions
work export -o timesheet.csv              # Export to CSV
work invoices -p week -d 2025-08-23       # Generate PDF invoices
```

## Features

- **Time Tracking**: Start/stop sessions with automatic client switching
- **Billing**: Set hourly rates per client with automatic calculations  
- **Invoicing**: Generate professional PDF invoices with client details
- **Export**: CSV export for accounting software integration
- **Client Management**: Store billing details (company, address, tax numbers)

## Client Setup

```bash
work clients update -c clientname -r 150                           # Set hourly rate
work clients update -c clientname --company "ACME Corp"           # Add billing details
work clients list -v                                              # View all client details
```