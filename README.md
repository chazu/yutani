# Yutani - Terminal Display Server

Yutani is a Go-based terminal display server that provides networked, widget-based windowing capabilities for text-mode applications. Inspired by TWIN, Yutani uses gRPC for client-server communication and leverages tcell/tview for the underlying TUI rendering.

## Features

- **Networked Terminal UI**: Control terminal UI remotely via gRPC
- **Widget-Based**: High-level widget abstractions (forms, lists, tables, etc.)
- **Low-Level Access**: Direct cell manipulation for custom rendering
- **Multiple Clients**: Support for concurrent client sessions
- **gRPC Reflection**: Easy introspection with tools like `grpcurl` and `grpcui`

## Current Status: Phase 5 Complete ✅

**Phase 1** - Foundation ✅
- ✅ Project structure and build system
- ✅ Proto definitions for core types
- ✅ SessionService implementation
- ✅ Basic ScreenService (GetSize, Clear, Sync)
- ✅ gRPC server with reflection
- ✅ Configuration system (file → env → flags)

**Phase 2** - Low-Level API ✅
- ✅ Extended ScreenService (SetCell, SetCells, Fill, DrawText, DrawBox, GetCell)
- ✅ EventService with streaming support
- ✅ Event types (Key, Mouse, Resize, Focus, Widget)
- ✅ Event dispatcher with filtering
- ✅ Event capture from tview/tcell
- ✅ Unit tests for business logic

**Phase 3** - Widget System ✅
- ✅ WidgetService with 10 RPCs (Create, Delete, SetProperties, GetProperties, etc.)
- ✅ 5 widget types: Box, TextView, InputField, Button, Checkbox
- ✅ Comprehensive property system (common + type-specific)
- ✅ Widget hierarchy infrastructure (parent-child relationships)
- ✅ Focus management (SetFocus, GetFocus)
- ✅ Automatic widget event emission (CHANGED, SUBMITTED, DONE)
- ✅ Enhanced widget registry with metadata

**Phase 4** - Complex Widgets ✅ **COMPLETE**
- ✅ Protocol buffer definitions for all complex widgets
- ✅ Widget factory support for List, Table, TreeView, Form, Flex, Grid, Pages
- ✅ ListService fully implemented (7 RPCs) with unit tests
- ✅ TableService fully implemented (8 RPCs) with unit tests
- ✅ FormService fully implemented (6 RPCs) with unit tests
- ✅ TreeService fully implemented (7 RPCs) with unit tests
- ✅ LayoutService fully implemented (11 RPCs) with unit tests
- ✅ Comprehensive end-to-end tests for all complex widgets

**Phase 5** - Client Library, Documentation, and Examples ✅ **COMPLETE**
- ✅ Go client library with fluent builder pattern
- ✅ Widget builders for Box, TextView, List, Table, Form
- ✅ Event handling with callbacks and type-safe events
- ✅ 3 complete example applications (list, table, form)
- ✅ Comprehensive tutorial with 5 lessons
- ✅ Full API documentation for client library
- ✅ Helper functions for colors and common operations

**Phase 6.1** - Additional Widget Builders ✅ **COMPLETE**
- ✅ TreeView widget builder with fluent API
- ✅ Flex layout widget builder
- ✅ Grid layout widget builder
- ✅ Pages layout widget builder
- ✅ InputField widget builder
- ✅ Button widget builder
- ✅ Checkbox widget builder
- ✅ Comprehensive tests for all new builders
- ✅ Updated documentation with examples

**Phase 6.2** - Advanced Event Handling ✅ **COMPLETE**
- ✅ Event filtering by widget ID
- ✅ Event filtering by event type
- ✅ Custom event filters with predicates
- ✅ Event middleware/interceptor pipeline
- ✅ Event batching for high-frequency events
- ✅ Event recording and replay for debugging
- ✅ Server-side event filtering
- ✅ Comprehensive tests (10 test cases)
- ✅ Updated documentation with examples

**Phase 6.5** - Additional Examples ✅ **COMPLETE**
- ✅ File Browser - TreeView navigation example
- ✅ Dashboard - Grid layout with system stats
- ✅ Process Monitor - Real-time table updates
- ✅ Chat Application - Multi-page messaging app
- ✅ Text Editor - Syntax highlighting and file I/O
- ✅ Comprehensive examples README
- ✅ All examples compile and run
- ✅ Common patterns documented

**Phase 6.6** - Testing Utilities ✅ **COMPLETE**
- ✅ Mock client for unit testing
- ✅ Event simulation (keyboard, mouse, resize)
- ✅ Test helper utilities and assertions
- ✅ Widget-specific assertions (List, Table)
- ✅ Integration test helpers with test server
- ✅ Server options pattern
- ✅ Comprehensive tests (13 test cases)
- ✅ Complete documentation with examples

See [PHASE3_COMPLETE.md](PHASE3_COMPLETE.md) for Phase 3 documentation.
See [PHASE4_COMPLETE.md](PHASE4_COMPLETE.md) for Phase 4 documentation and usage examples.
See [PHASE5_COMPLETE.md](PHASE5_COMPLETE.md) for Phase 5 client library documentation.

## Quick Start

**New to Yutani?** See [QUICKSTART.md](QUICKSTART.md) for a step-by-step guide!

### Prerequisites

- Go 1.24.5 or later
- Protocol Buffers compiler (`protoc`)

On macOS:
```bash
brew install protobuf
```

On Linux:
```bash
apt-get install protobuf-compiler
```

### Build

```bash
# Install protoc plugins
make install-tools

# Build everything (server + clients + examples)
make build
```

This creates:
- `bin/yutani-server` - The display server
- `bin/test-client` - Basic test client
- `bin/phase4-demo` - Phase 4 demo with complex widgets
- `bin/examples/*` - All example applications (8 examples)

To build only examples:
```bash
make build-examples
```

To list available examples:
```bash
make list-examples
```

### Run

**Terminal 1 - Start the server:**
```bash
make run
```

**Terminal 2 - Run a client or example:**
```bash
# Run the demo
make demo

# Or run an example
make run-example EXAMPLE=text-editor

# Or run directly
./bin/examples/text-editor
./bin/test-client
```

**Available Examples**:
- `simple-list` - Basic list widget
- `data-table` - Table widget demo
- `login-form` - Form widget demo
- `file-browser` - TreeView file browser
- `dashboard` - Grid layout system monitor
- `process-monitor` - Real-time process table
- `chat-app` - Multi-room chat application
- `text-editor` - Text editor with syntax highlighting

See [examples/README.md](examples/README.md) for detailed documentation.

The graphical UI will appear in the server terminal (Terminal 1).

**Features:**
- ✅ Clean TUI display (no log messages)
- ✅ Welcome screen when no client is connected
- ✅ Press **Ctrl+C** to exit gracefully
- ✅ Enable logging with `YUTANI_LOG_FILE=server.log ./bin/yutani-server`

**Note:** If you have display issues, see [DISPLAY_FIX.md](DISPLAY_FIX.md) for troubleshooting.

## Configuration

Yutani uses a three-tier configuration system with the following priority:

1. `.yutani.conf` file (lowest priority)
2. Environment variables (middle priority)
3. Command-line flags (highest priority)

### Configuration File

Create a `.yutani.conf` file in the current directory:

```conf
YUTANI_ADDRESS=:7755
YUTANI_MAX_SESSIONS=100
YUTANI_MOUSE=true
YUTANI_PASTE=true
YUTANI_LOG_LEVEL=info
```

See `.yutani.conf.example` for a template.

### Environment Variables

All configuration options can be set via environment variables with the `YUTANI_` prefix:

```bash
export YUTANI_ADDRESS=:7755
export YUTANI_LOG_LEVEL=debug
```

### Command-Line Flags

```bash
./bin/yutani-server \
  --address=:7755 \
  --max-sessions=100 \
  --mouse=true \
  --paste=true \
  --log-level=info
```

## Running the Server

```bash
# Using make
make run

# Or directly
./bin/yutani-server
```

The server will start and listen on the configured address (default: `:7755`).

## Testing with grpcurl

Since gRPC reflection is enabled, you can easily test the server with `grpcurl`:

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:7755 list

# Get server info
grpcurl -plaintext localhost:7755 industries.loosh.yutani.v1.SessionService/GetServerInfo

# Ping the server
grpcurl -plaintext -d '{"timestamp": 123456}' localhost:7755 industries.loosh.yutani.v1.SessionService/Ping

# Create a session
grpcurl -plaintext -d '{"client_name": "test-client"}' localhost:7755 industries.loosh.yutani.v1.SessionService/CreateSession

# Get screen size (replace SESSION_ID with actual session ID from CreateSession)
grpcurl -plaintext -d '{"session_id": {"id": "SESSION_ID"}}' localhost:7755 industries.loosh.yutani.v1.ScreenService/GetSize

# Draw text
grpcurl -plaintext -d '{"session_id": {"id": "SESSION_ID"}, "position": {"x": 10, "y": 5}, "text": "Hello!"}' localhost:7755 industries.loosh.yutani.v1.ScreenService/DrawText

# Subscribe to events (streaming)
grpcurl -plaintext -d '{"session_id": {"id": "SESSION_ID"}}' localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe
```

## Testing

### Automated Tests

The project includes comprehensive unit and end-to-end tests:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run only E2E tests
go test ./test/e2e/... -v

# Run specific test
go test ./pkg/server/... -v -run TestServerLifecycle
```

**Test Coverage:**
- **Unit Tests**: 74 tests covering all services and components
  - 17 server tests (registry, events, lifecycle)
  - 20 widget factory tests (basic + complex widgets)
  - 33 service tests (List, Table, Form, Tree, Layout)
  - 4 other service tests
- **E2E Tests**: 11 comprehensive tests using in-memory gRPC (bufconn)
  - 4 core tests (session, screen, events, widgets)
  - 7 complex widget tests (List, Table, Form, Tree, Flex, Grid, Pages)
- **All tests pass** in under 2 seconds
- **100% coverage** of Phase 4 RPCs

See `E2E_TESTS_SUMMARY.md` and `UNIT_TESTS_SUMMARY.md` for details.

### Manual Testing with Test Client

A comprehensive test client is included that demonstrates all functionality:

```bash
# Build the test client
make build-test-client

# Run it (server must be running in another terminal)
./bin/test-client
```

The test client will:
- Create a session
- Test all screen operations (Clear, SetCell, DrawText, DrawBox, Fill, GetCell)
- Subscribe to events and test event injection
- Test widget creation and management
- Clean up the session

## Development

```bash
# Clean generated files
make clean

# Rebuild everything
make build

# Run unit tests
go test ./pkg/server/... ./pkg/services/... -v

# Format code
make fmt

# Tidy dependencies
make tidy
```

## Architecture

See [PRD.md](./PRD.md) for detailed architecture and design documentation.

## Quick Start with Client Library

### Installation

```bash
go get industries/loosh/yutani/pkg/client
```

### Creating a List Widget

```go
import "industries/loosh/yutani/pkg/client"

// Connect to server
c, _ := client.Connect("localhost:50051")
defer c.Close()

// Create a list with fluent API
list, _ := c.NewList().
    Title("Menu").
    Border(true).
    BorderColor(client.Color("blue")).
    Build()

// Add items
list.AddItem("New File", "Create a new file", strPtr("n"))
list.AddItem("Open File", "Open an existing file", strPtr("o"))
list.SetSelected(0)
```

### Creating a Table Widget

```go
// Create a table
table, _ := c.NewTable().
    Title("Data Table").
    Border(true).
    Build()

// Set headers with colors
table.SetCell(0, 0, client.NewTableCellWithColor("Name", client.Color("yellow")))
table.SetCell(0, 1, client.NewTableCellWithColor("Age", client.Color("yellow")))
table.SetFixed(1, 0) // Fix header row

// Batch set data
table.SetCells([]*pb.TableCellUpdate{
    {Row: 1, Column: 0, Cell: client.NewTableCell("Alice")},
    {Row: 1, Column: 1, Cell: client.NewTableCell("30")},
})
```

### Creating a Form Widget

```go
// Create a form
form, _ := c.NewForm().
    Title("Login").
    Border(true).
    Build()

// Add fields
usernameIdx, _ := form.AddInputField("Username", 30, "")
passwordIdx, _ := form.AddPasswordField("Password", 30)
form.AddButton("Login")

// Handle events
c.OnEvent(func(event *client.Event) {
    if event.IsWidget() && event.Widget.Type == "SUBMITTED" {
        username, _ := form.GetFieldValue(usernameIdx)
        password, _ := form.GetFieldValue(passwordIdx)
        // Process login...
    }
})
c.StartEventStream()
```

For more examples, see [TUTORIAL.md](TUTORIAL.md) and [examples/](examples/).

## Roadmap

- ✅ **Phase 1**: Foundation (SessionService, basic ScreenService, configuration)
- ✅ **Phase 2**: Low-Level API (complete ScreenService, EventService, event streaming)
- ✅ **Phase 3**: Widget system with basic widgets (Box, TextView, InputField, Button, Checkbox)
- ✅ **Phase 4**: Complex widgets (List, Table, TreeView, Form, Flex, Grid, Pages)
- ✅ **Phase 5**: Client library, documentation, and examples
- **Phase 6** (Future): Advanced features, performance optimization, additional widgets

## License

TBD

## Contributing

TBD

