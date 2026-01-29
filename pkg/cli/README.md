# Yutani CLI

Command-line tools for the Yutani TUI framework.

## Installation

```bash
go install github.com/chazu/yutani/cmd/yutani@latest
```

Or build from source:

```bash
cd /path/to/yutani
go build -o bin/yutani ./cmd/yutani
```

## Commands

### `yutani server`

Start a Yutani server.

```bash
# Start server on default port (7755)
yutani server

# Start server on custom port
yutani server --address :8080

# Start server with debug logging
yutani server --debug

# Start headless server for testing/CI
yutani server --headless

# Start with log file
yutani server --log-file server.log --debug

# Start server with custom settings
yutani server --address :7755 --max-sessions 20 --mouse --paste
```

**Flags**:
- `-a, --address` - Server address to listen on (from config/env, default ":7755")
- `-m, --max-sessions` - Maximum concurrent sessions (from config/env, default 100)
- `--mouse` - Enable mouse support (default true)
- `--paste` - Enable paste support (default true)
- `-l, --log-level` - Log level: debug, info, warn, error (default "info")
- `-d, --debug` - Enable debug mode
- `--headless` - Run in headless mode (no TUI, simulated screen)
- `--log-file` - Log to file instead of stderr (TUI mode discards logs by default)

### `yutani new`

Create a new Yutani project with scaffolding.

```bash
# Create basic project
yutani new my-app

# Create project with specific template
yutani new my-app --template dashboard

# Create project with custom module name
yutani new my-app --module github.com/user/my-app
```

**Templates**:
- `basic` - Basic application with a single widget
- `list` - List-based application
- `table` - Table-based application
- `dashboard` - Dashboard with multiple widgets
- `full` - Full-featured application

**Flags**:
- `-t, --template` - Project template (default "basic")
- `-m, --module` - Go module name (default: project-name)

### `yutani inspect`

Inspect a running Yutani server.

```bash
# Inspect server on default port
yutani inspect

# Inspect server on custom port
yutani inspect --address localhost:8080

# Watch for changes
yutani inspect --watch

# Output as JSON
yutani inspect --format json
```

**Flags**:
- `-a, --address` - Server address (default "localhost:7755")
- `-f, --format` - Output format: text, json (default "text")
- `-w, --watch` - Watch for changes

### `yutani profile`

Profile a Yutani application for performance analysis.

```bash
# CPU profile for 30 seconds
yutani profile --type cpu --duration 30

# Memory profile
yutani profile --type mem --output mem.prof

# Goroutine profile
yutani profile --type goroutine

# Block profile
yutani profile --type block --duration 60

# Mutex profile
yutani profile --type mutex --duration 60
```

**Profile Types**:
- `cpu` - CPU profiling
- `mem` - Memory profiling
- `goroutine` - Goroutine profiling
- `block` - Block profiling
- `mutex` - Mutex profiling

**Flags**:
- `-t, --type` - Profile type (default "cpu")
- `-d, --duration` - Profile duration in seconds (default 30)
- `-o, --output` - Output file (default "<type>.prof")

**Analyzing Profiles**:
```bash
# After profiling
go tool pprof cpu.prof

# Interactive mode
go tool pprof -http=:8080 cpu.prof
```

### `yutani debug`

Debug utilities for Yutani applications.

```bash
# Display widget tree
yutani debug tree

# Display tree with properties
yutani debug tree --verbose

# Monitor event stream
yutani debug events

# Monitor specific event types
yutani debug events --type key,mouse
```

**Subcommands**:
- `tree` - Display widget hierarchy tree
- `events` - Monitor event stream

**Flags**:
- `-a, --address` - Server address (default "localhost:7755")
- `-v, --verbose` - Show detailed information
- `-t, --type` - Event types to monitor (for events command)

## Examples

### Starting a Development Server

```bash
# Start server with debug logging
yutani server --debug

# In another terminal, run your app
go run main.go
```

### Creating a New Project

```bash
# Create a dashboard project
yutani new my-dashboard --template dashboard

# Navigate and run
cd my-dashboard
go mod tidy
go run main.go
```

### Profiling an Application

```bash
# Terminal 1: Start server
yutani server

# Terminal 2: Run your app
go run main.go

# Terminal 3: Profile CPU usage
yutani profile --type cpu --duration 60

# Analyze results
go tool pprof -http=:8080 cpu.prof
```

### Inspecting a Running Server

```bash
# Watch server in real-time
yutani inspect --watch

# Get JSON output for scripting
yutani inspect --format json > server-state.json
```

## See Also

- [Client Library Documentation](../client/README.md)
- [Server Documentation](../server/README.md)
- [Examples](../../examples/README.md)

