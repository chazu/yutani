# Yutani Quick Start Guide

## Getting Started in 3 Steps

### Step 1: Build Everything

```bash
make build
```

This compiles protobuf definitions and builds:
- `bin/yutani` - The Yutani CLI (includes `server` subcommand)
- `bin/test-client` - Basic test client
- `bin/phase4-demo` - Phase 4 complex widgets demo

### Step 2: Start the Server

Open a terminal and run:

```bash
make run
```

Or directly:

```bash
./bin/yutani server
```

**Important:** Keep this terminal open -- the graphical UI will appear here!

When the server starts, you should see a TUI welcome screen:

```
+----------------------------------------------------------+
|                   Yutani v0.1.0-alpha                    |
+----------------------------------------------------------+
|                                                          |
|               Yutani Display Server                      |
|                                                          |
|           Waiting for client connections...               |
|                                                          |
|         Server is running on localhost:7755               |
|               Press Ctrl+C to exit                       |
|                                                          |
+----------------------------------------------------------+
```

The title text is displayed in yellow/bold and the server address information is centered and bordered.

### Step 3: Run the Demo

Open a **second terminal** and run:

```bash
make demo
```

Or directly:

```bash
./bin/phase4-demo
```

## What You Will See

In the **server terminal** (Terminal 1), you will see a TUI layout with widgets:

```
+--------------------------------------------------------------+
| Yutani Phase 4 Demo - Complex Widgets                        |
+----------------------+---------------------------------------+
| +-- Task List -----+ | +-- Service Stats ----------------+  |
| | Phase 1: Core    | | | Service    | RPCs | Status       |  |
| | Phase 2: Widge   | | | ListServ.. |  5   |    ok        |  |
| | Phase 3: Basic   | | | TableSer.. |  8   |    ok        |  |
| | Phase 4: Compl   | | | FormServ.. |  6   |    ok        |  |
| | Phase 5: Adva    | | | TreeServ.. |  7   |    ok        |  |
| +------------------+ | | LayoutSe.. | 11   |    ok        |  |
|                      | +--------------------------------+   |
+----------------------+---------------------------------------+
| +-- User Settings -+ | +-- Project Structure -------------+  |
| | Username: ______ | | | > yutani/                        |  |
| | Password: ______ | | |   > pkg/                         |  |
| | [ ] Remember me  | | |     server/                     |  |
| | Theme: [Dark]    | | |     services/                   |  |
| | [Submit][Cancel] | | |     proto/                      |  |
| +-----------------+  | |   cmd/                           |  |
|                      | +----------------------------------+  |
+----------------------+---------------------------------------+
```

In the **demo terminal** (Terminal 2), you will see log output:

```
Creating session...
Session created: abc-123-def
Main flex container created
Title created
List widget created
Table widget created
...
```

## Troubleshooting

### Problem: Only seeing log messages in server terminal

**Solution:** Make sure you have rebuilt after the latest changes:

```bash
make clean
make build
```

Logs are disabled by default. The TUI should display cleanly.

### Problem: Nothing appears, terminal is blank

**Possible causes:**
1. The terminal does not support the required features
2. The TERM environment variable is not set correctly
3. Running in a non-interactive environment

**Solutions:**
```bash
# Check your TERM variable
echo $TERM

# Try setting it explicitly
export TERM=xterm-256color
./bin/yutani server

# Or try screen-256color
export TERM=screen-256color
./bin/yutani server
```

The server must run in an interactive terminal to display the TUI. It will not work when run in the background with `&`.

### Problem: Can't exit server with Ctrl+C

**Solution:** Make sure you have rebuilt:

```bash
make clean
make build-yutani
```

If it still does not work, try pressing Ctrl+C multiple times, and verify you are running the server (`./bin/yutani server`).

### Problem: "Failed to connect" error

**Solution:** Make sure the server is running first (Terminal 1) before running the demo (Terminal 2).

### Problem: Port already in use

**Solution:** Kill any existing server process:

```bash
pkill yutani
```

Then restart the server.

### Problem: Want to see logs

```bash
# Enable logging to a file
./bin/yutani server --log-file server.log

# In another terminal, watch the logs
tail -f server.log
```

## Next Steps

### Run the Basic Test Client

```bash
./bin/test-client
```

This tests all Phase 1-3 features including:
- Session management
- Screen operations (Clear, DrawText, DrawBox, Fill)
- Event streaming
- Basic widgets (TextView, InputField, Button, Checkbox)

### Explore the Code

- `cmd/yutani/main.go` - CLI entry point (includes `server` subcommand)
- `cmd/phase4-demo/main.go` - Demo client showing complex widgets
- `pkg/server/` - Core server implementation
- `pkg/services/` - gRPC service implementations
- `api/proto/` - Protocol buffer definitions

### Read the Documentation

- [README.md](README.md) - Project overview and architecture
- [DEBUGGING.md](DEBUGGING.md) - Debugging tools and troubleshooting
- [TUTORIAL.md](TUTORIAL.md) - Client library tutorial
- [PRD.md](PRD.md) - Product requirements and design

## Common Use Cases

### Development Workflow

```bash
# Clean and rebuild everything
make clean build

# Terminal 1: Run server with logs enabled
./bin/yutani server --log-file server.log

# Terminal 2: Run demo
make demo

# Terminal 3: Watch logs
tail -f server.log
```

### Testing Changes

```bash
# Format code
make fmt

# Tidy dependencies
make tidy

# Run tests
go test ./...

# Rebuild and run
make dev
```

## Tips

1. **Logs are disabled by default** - Enable with `./bin/yutani server --log-file server.log`
2. **Server must run first** - Always start the server before clients
3. **UI appears in server terminal** - Not in the client terminal
4. **Multiple clients supported** - You can run multiple demo clients simultaneously
5. **Ctrl+C to exit** - Gracefully shuts down the server
6. **Welcome screen** - Server shows a placeholder until a client connects
7. **Check your TERM variable** - If the TUI is blank, try `export TERM=xterm-256color`

## Getting Help

- See [DEBUGGING.md](DEBUGGING.md) for debugging tools and common issues
- Review [README.md](README.md) for architecture details
- Run `make list-examples` to see all available example applications
