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

# Download turso, authenticate, create a database with the name $(BIN_NAME) from the .env file
make prod-turso-setup

# Configure .env with your Turso database - you'll need to create a token and get your database URL for these steps
echo "TURSO_TOKEN=your_turso_token_here" >> .env.mine
echo "PROD_DATABASE_URL=libsql://your-database.turso.io" >> .env.mine

# Run initial migrations - should only ever need to do this once
make prod-init

# Build and install
make prod-install
```

## Usage

After installation, create a client using `work create -c clientname -r 100`, to create a client with a $100 hourly rate.

The main use case is starting and stopping work sessions, using `work start -c clientname` and `work stop`. You may then use `work list` to see your logged sessions.

To generate invoices for a period of time, use `work invoices -p fortnight -d 2025-08-23`, which generates an invoice per client for the specified period.

For more information, use `work help` or `work help <command>`. Current output is below.

```bash
➜ work help
Track your work sessions across multiple clients with simple start/stop commands.
Supports hourly rate tracking and automatic billable amount calculations for freelance work.

Usage:
  work [command]

Available Commands:
  clear        Delete work sessions
  clients      Manage clients
  create       Create various entities
  descriptions Manage session descriptions
  export       Export work sessions to CSV
  help         Help about any command
  invoices     Generate PDF invoices for clients
  list         List recent work sessions
  session      Manage work sessions
  start        Start a work session
  status       Show current work status
  stop         Stop the current work session
  summarize    Summarize and analyze client data
```
