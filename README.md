# Work CLI

Who needs interns when we have clankers to do our bidding?

Simple CLI time tracker for freelancers. Track work sessions across multiple clients with automatic billing calculations, PDF invoice generation, and AI-powered work summaries.

**Dependencies:** Requires [OpenCode](https://github.com/sst/opencode) for `descriptions generate` that analyzes local git repositories to generate session invoice descriptions per-client. Requires your own [Turso](https://turso.tech/) sqlite database if you want to share sessions between machines.

## Installation

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

```bash
➜ work help
Track your work sessions across multiple clients with simple start/stop commands.
Supports hourly rate tracking and automatic billable amount calculations for freelance work.

Usage:
  work [command]

Available Commands:
  clients      Create, update and list clients
  descriptions Manage session descriptions using git and AI summarization
  help         Help about any command
  hours        Display total worked hours
  invoices     Manage invoices for clients
  note         Add a note to the active session
  sessions     Manage sessions
  start        Start a work session
  status       Show current work status
  stop         Stop the current work session
```

### Example

In the below, we create a client with an hourly rate of $100, start a session, leave a note about something we did outside git, and stop the session.

```bash
➜ work clients create "My Client" -r 100
Created client: My Client (ID: 0198c7b0-96e8-7320-848a-29b07f2d717f, Rate: $100.00/hr)
➜ work start -c "My Client"
Started work session for My Client at 23:36:01
➜ work note -c "My Client" -m "Worked on architectural design for a new website"
Added note to session for My Client:
- Worked on architectural design for a new website

All notes for this session:
- Worked on architectural design for a new website
➜ work stop
Stopped work session for My Client
Duration: 0h 2m
Started: 23:36:01, Ended: 23:38:11
```
