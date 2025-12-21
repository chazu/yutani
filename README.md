# Yutani - Terminal Display Server

Yutani is a Go-based terminal display server that provides networked, widget-based windowing capabilities for text-mode applications. Inspired by TWIN, Yutani uses gRPC for client-server communication and leverages tcell/tview for the underlying TUI rendering.

## Features

- **Networked Terminal UI**: Control terminal UI remotely via gRPC
- **Widget-Based**: High-level widget abstractions (forms, lists, tables, etc.)
- **Low-Level Access**: Direct cell manipulation for custom rendering
- **Multiple Clients**: Support for concurrent client sessions
- **gRPC Reflection**: Easy introspection with tools like `grpcurl` and `grpcui`

## Current Status: Phase 4 In Progress 🚧

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

**Phase 4** - Complex Widgets 🚧 (In Progress)
- ✅ Protocol buffer definitions for all complex widgets
- ✅ Widget factory support for List, Table, TreeView, Form, Flex, Grid, Pages
- ✅ ListService fully implemented with 7 RPCs and unit tests
- ⏳ TableService (8 RPCs) - pending implementation
- ⏳ FormService (6 RPCs) - pending implementation
- ⏳ TreeService (7 RPCs) - pending implementation
- ⏳ LayoutService (11 RPCs) - pending implementation
- ⏳ End-to-end tests for complex widgets

See [PHASE3_COMPLETE.md](PHASE3_COMPLETE.md) for Phase 3 documentation.
See [PHASE4_PROGRESS.md](PHASE4_PROGRESS.md) for Phase 4 progress tracking.

## Installation

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

# Build the server
make build
```

The binary will be created at `bin/yutani-server`.

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
- **Unit Tests**: 46 tests covering server, services, widget factory, and ListService
  - 17 server tests (registry, events, lifecycle)
  - 20 widget factory tests (basic + complex widgets)
  - 5 ListService tests
  - 4 other service tests
- **E2E Tests**: 4 comprehensive tests using in-memory gRPC (bufconn)
- **All tests pass** in under 1 second

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

## Roadmap

- ✅ **Phase 1**: Foundation (SessionService, basic ScreenService, configuration)
- ✅ **Phase 2**: Low-Level API (complete ScreenService, EventService, event streaming)
- **Phase 3**: Widget system with basic widgets (Box, TextView, InputField, Button)
- **Phase 4**: Complex widgets (List, Table, TreeView, Form, Layouts)
- **Phase 5**: Client library, documentation, and examples

## License

TBD

## Contributing

TBD

