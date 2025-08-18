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

# If you want to be able to share sessions between machines, sign up to Turso and configure .env.mine with your database URL & turso token
# In this case, make sure your database name is the same as the BIN_NAME in .env
# After creating your db, run `make prod-turso-setup`

# Install with sqlite on your machine
make install

# Or install with turso, using `make prod-init && make prod-install`. You should only ever need to run `make prod-init` once.
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
