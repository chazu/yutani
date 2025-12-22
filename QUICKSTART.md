# Yutani Quick Start Guide

## 🚀 Getting Started in 3 Steps

### Step 1: Build Everything

```bash
make build
```

This compiles protobuf definitions and builds:
- `bin/yutani-server` - The display server
- `bin/test-client` - Basic test client
- `bin/phase4-demo` - Phase 4 complex widgets demo

### Step 2: Start the Server

Open a terminal and run:

```bash
make run
```

Or directly:

```bash
./bin/yutani-server
```

**Important:** Keep this terminal open - the graphical UI will appear here!

### Step 3: Run the Demo

Open a **second terminal** and run:

```bash
make demo
```

Or directly:

```bash
./bin/phase4-demo
```

## 🎨 What You'll See

In the **server terminal** (Terminal 1), you'll see a beautiful TUI layout:

```
┌─────────────────────────────────────────────────────────────┐
│ Yutani Phase 4 Demo - Complex Widgets                      │
├──────────────────────┬──────────────────────────────────────┤
│ ┌─ Task List ──────┐ │ ┌─ Service Stats ───────────────┐  │
│ │ ✓ Phase 1: Core  │ │ │ Service    │ RPCs │ Status   │  │
│ │ ✓ Phase 2: Widge │ │ │ ListServ.. │  5   │    ✓     │  │
│ │ ✓ Phase 3: Basic │ │ │ TableSer.. │  8   │    ✓     │  │
│ │ ✓ Phase 4: Compl │ │ │ FormServ.. │  6   │    ✓     │  │
│ │ ⏳ Phase 5: Adva │ │ │ TreeServ.. │  7   │    ✓     │  │
│ └──────────────────┘ │ │ LayoutSe.. │ 11   │    ✓     │  │
│                      │ └───────────────────────────────┘  │
├──────────────────────┼──────────────────────────────────────┤
│ ┌─ User Settings ─┐ │ ┌─ Project Structure ────────────┐  │
│ │ Username: ______ │ │ │ ▼ yutani/                     │  │
│ │ Password: ______ │ │ │   ▼ pkg/                      │  │
│ │ ☐ Remember me   │ │ │     • server/                 │  │
│ │ Theme: [Dark ▼] │ │ │     • services/               │  │
│ │ [Submit][Cancel]│ │ │     • proto/                  │  │
│ └─────────────────┘ │ │   • cmd/                      │  │
│                      │ └───────────────────────────────┘  │
└──────────────────────┴──────────────────────────────────────┘
```

In the **demo terminal** (Terminal 2), you'll see log output:

```
Creating session...
✓ Session created: abc-123-def
✓ Main flex container created
✓ Title created
✓ List widget created
✓ Table widget created
...
```

## 🔧 Troubleshooting

### Problem: Only seeing log messages in server terminal

**Solution:** This was fixed! Make sure you've rebuilt after the fix:

```bash
make clean
make build
```

Logs are now disabled by default. The TUI should display cleanly.

### Problem: Can't exit server with Ctrl+C

**Solution:** This was fixed! Ctrl+C now works properly to exit the server.

If you're still having issues, make sure you've rebuilt:

```bash
make clean
make build-server
```

### Problem: "Failed to connect" error

**Solution:** Make sure the server is running first (Terminal 1) before running the demo (Terminal 2).

### Problem: Port already in use

**Solution:** Kill any existing server process:

```bash
pkill yutani-server
```

Then restart the server.

## 📚 Next Steps

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

- `cmd/yutani-server/main.go` - Server entry point
- `cmd/phase4-demo/main.go` - Demo client showing complex widgets
- `pkg/server/` - Core server implementation
- `pkg/services/` - gRPC service implementations
- `api/proto/` - Protocol buffer definitions

### Read the Documentation

- `README.md` - Project overview and architecture
- `DISPLAY_FIX.md` - Details about the display fix
- `PHASE4_PROGRESS.md` - Phase 4 implementation status
- `PRD.md` - Product requirements and design

## 🎯 Common Use Cases

### Development Workflow

```bash
# Clean and rebuild everything
make clean build

# Terminal 1: Run server with logs enabled
YUTANI_LOG_FILE=server.log ./bin/yutani-server

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

## 💡 Tips

1. **Logs are disabled by default** - Enable with `YUTANI_LOG_FILE=server.log ./bin/yutani-server`
2. **Server must run first** - Always start the server before clients
3. **UI appears in server terminal** - Not in the client terminal
4. **Multiple clients supported** - You can run multiple demo clients simultaneously
5. **Ctrl+C to exit** - Gracefully shuts down the server
6. **Welcome screen** - Server shows a placeholder until a client connects

## 🐛 Known Issues

None currently! The display issue has been fixed.

## 📞 Getting Help

- Check `DISPLAY_FIX.md` for display-related issues
- Review `README.md` for architecture details
- Run `make help` to see all available targets (coming soon)

