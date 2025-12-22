# Display Issue Fix - Yutani Server

## Problem Description

When running the Yutani server and demo client, the server's graphical TUI display was not rendering properly. Instead of showing the expected visual interface with widgets (List, Table, Form, Tree, etc.), only log messages were visible in the terminal.

## Root Causes

The display issues were caused by **multiple problems**:

### Problem 1: Logging Interference
1. **tcell/tview** uses the terminal for rendering the TUI
2. **slog** was writing log messages that interfered with the display
3. Log messages were appearing over the TUI, corrupting the display

### Problem 2: Ctrl+C Not Working
1. **tview** was capturing all keyboard input including Ctrl+C
2. The signal handler in main.go never received the interrupt signal
3. Server couldn't be stopped gracefully

### Problem 3: TUI Running in Wrong Goroutine
1. **tview app.Run()** was running in a background goroutine
2. TUI applications MUST run in the main goroutine to work properly
3. This caused the display to not appear and Ctrl+C to not work

### Problem 4: No Initial Display
1. Server started with no root widget set
2. TUI showed a blank screen until a client connected
3. Confusing user experience

## Solutions Applied

### Solution 1: Disable Logging by Default

**File:** `cmd/yutani-server/main.go`

Logs are now sent to `/dev/null` by default to avoid interfering with the TUI.
To enable logging, set the `YUTANI_LOG_FILE` environment variable:

```go
logFile := os.Getenv("YUTANI_LOG_FILE")
if logFile != "" {
    logWriter, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
} else {
    // Discard logs when no log file is specified (TUI mode)
    logWriter, _ = os.Open(os.DevNull)
}
```

### Solution 2: Handle Ctrl+C in TUI

**File:** `pkg/server/server.go`

Added Ctrl+C handling in the input capture function:

```go
s.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    // Handle Ctrl+C to allow graceful shutdown
    if event.Key() == tcell.KeyCtrlC {
        s.app.Stop()
        return nil
    }
    // Dispatch other key events
    return s.handleInputEvent(event)
})
```

### Solution 3: Run TUI in Main Goroutine

**Files:** `pkg/server/server.go` and `cmd/yutani-server/main.go`

**Critical Fix:** TUI applications MUST run in the main goroutine.

**In pkg/server/server.go** - Added a Run() method:
```go
// Run starts the TUI application (blocking call)
// This should be called from the main goroutine
func (s *Server) Run() error {
    // Run the application (blocks until app.Stop() is called)
    return s.app.Run()
}
```

**In cmd/yutani-server/main.go** - Restructured to run TUI in main goroutine:
```go
// Start gRPC server in a goroutine (background)
go func() {
    if err := grpcServer.Serve(listener); err != nil {
        // Server stopped
    }
}()

// Run the TUI application (blocks until Ctrl+C or app.Stop())
// This MUST run in the main goroutine
if err := yutaniServer.Run(); err != nil {
    slog.Error("TUI error", "error", err)
}
```

### Solution 4: Show Placeholder Widget

**File:** `pkg/server/server.go`

Added a welcome screen that displays when no client is connected:

```go
placeholder := tview.NewTextView().
    SetText("[yellow::b]Yutani Display Server[-::-]\n\n" +
        "Waiting for client connections...\n\n" +
        "Server is running on localhost:7755\n" +
        "Press [red::b]Ctrl+C[-::-] to exit").
    SetDynamicColors(true).
    SetTextAlign(tview.AlignCenter)
s.app.SetRoot(placeholder, true)
```

## Solution

Changed the logging output from `os.Stdout` to `os.Stderr`. This allows:
- **TUI display** → renders to `stdout` (the terminal screen)
- **Log messages** → output to `stderr` (can be redirected separately)

## How to Use

### Running the Server and Demo

**Build everything (one time):**
```bash
make build
```
This builds the server, test-client, and phase4-demo all at once.

**Terminal 1 - Start the server:**
```bash
make run
# or directly: ./bin/yutani-server
```

The TUI display will now appear correctly in this terminal, showing the graphical interface.

**Terminal 2 - Run the demo client:**
```bash
make demo
# or directly: ./bin/phase4-demo
```

The demo will create widgets that appear in the server's terminal (Terminal 1).

### Enabling Logs

By default, logs are disabled to avoid interfering with the TUI.
To enable logging to a file:

```bash
# Set log file environment variable
export YUTANI_LOG_FILE=server.log
./bin/yutani-server

# Or inline
YUTANI_LOG_FILE=server.log ./bin/yutani-server

# View logs in real-time in another terminal
tail -f server.log
```

## Expected Behavior

After the fix:
- ✅ Server terminal displays a clean TUI (no log messages)
- ✅ Shows a welcome screen when no client is connected
- ✅ **Ctrl+C works** to exit the server gracefully
- ✅ Widgets appear when client connects (List, Table, Form, Tree in a grid layout)
- ✅ Widgets update in real-time as the client sends commands
- ✅ Mouse and keyboard input work in the server terminal
- ✅ Logs are disabled by default (enable with YUTANI_LOG_FILE)

## Architecture Notes

**Yutani is a display server**, similar to X11 or Wayland but for terminal UIs:
- **Server** (`yutani-server`): Manages the display and renders widgets
- **Clients** (`phase4-demo`, `test-client`): Send gRPC commands to create/modify widgets
- **Display location**: Always in the server's terminal, not the client's terminal

This is the intended architecture - clients control the display remotely via gRPC.

## Testing

To verify the fix works:

1. Start the server: `./bin/yutani-server`
2. Run the demo: `./bin/phase4-demo`
3. Check the server terminal - you should see:
   - Yellow title: "Yutani Phase 4 Demo - Complex Widgets"
   - Top row: List widget (left) and Table widget (right)
   - Bottom row: Form widget (left) and Tree widget (right)
   - All widgets with borders and content

## Related Files

- `cmd/yutani-server/main.go` - Server entry point (logging configuration)
- `pkg/server/server.go` - Server initialization and TUI setup
- `pkg/services/widget.go` - Widget rendering logic
- `cmd/phase4-demo/main.go` - Demo client that creates complex widgets

## Makefile Targets

The Makefile has been updated with convenient targets:

- `make build` - Build everything (server + all clients)
- `make build-server` - Build only the server
- `make build-test-client` - Build only the test client
- `make build-phase4-demo` - Build only the phase4 demo
- `make run` - Build and run the server
- `make demo` - Build and run the phase4 demo (server must be running)
- `make clean` - Clean all generated files
- `make dev` - Clean, build server, and run (development workflow)

## Future Improvements

Consider adding:
1. Optional log file configuration via config file
2. Structured logging with JSON output for production
3. Log level filtering for different components
4. Visual client that renders the display locally (more complex)

