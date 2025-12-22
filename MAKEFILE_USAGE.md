# Makefile Usage Guide

This document describes all available Makefile targets for the Yutani project.

## Quick Reference

```bash
# Build everything (server, clients, examples)
make build

# Build and run server
make run

# Build only examples
make build-examples

# List available examples
make list-examples

# Run a specific example (server must be running)
make run-example EXAMPLE=text-editor

# Clean all generated files
make clean

# Development workflow
make dev
```

---

## Build Targets

### `make build`
Builds everything: server, test client, demo, and all examples.

**Output**:
- `bin/yutani-server` - Main server
- `bin/test-client` - Test client
- `bin/phase4-demo` - Phase 4 demo
- `bin/examples/*` - All example applications

**Example**:
```bash
make build
```

### `make build-server`
Builds only the Yutani server.

**Output**: `bin/yutani-server`

**Example**:
```bash
make build-server
```

### `make build-examples`
Builds all example applications.

**Output**: `bin/examples/*`

**Examples built**:
- simple-list
- data-table
- login-form
- file-browser
- dashboard
- process-monitor
- chat-app
- text-editor

**Example**:
```bash
make build-examples
```

---

## Run Targets

### `make run`
Builds and runs the Yutani server.

**Example**:
```bash
make run
```

### `make run-example EXAMPLE=<name>`
Runs a specific example application. The server must be running in another terminal.

**Available examples**:
- `simple-list` - Basic list widget demonstration
- `data-table` - Table widget with data
- `login-form` - Form widget example
- `file-browser` - TreeView file system browser
- `dashboard` - Grid layout system monitor
- `process-monitor` - Real-time process table
- `chat-app` - Multi-room chat application
- `text-editor` - Text editor with syntax highlighting

**Example**:
```bash
# Terminal 1: Start server
make run

# Terminal 2: Run example
make run-example EXAMPLE=text-editor
```

### `make demo`
Builds and runs the Phase 4 demo client. Server must be running separately.

**Example**:
```bash
# Terminal 1
make run

# Terminal 2
make demo
```

---

## Utility Targets

### `make list-examples`
Lists all available example applications with descriptions.

**Example**:
```bash
make list-examples
```

**Output**:
```
Available examples:
  simple-list       - Basic list widget demonstration
  data-table        - Table widget with data
  login-form        - Form widget example
  file-browser      - TreeView file system browser
  dashboard         - Grid layout system monitor
  process-monitor   - Real-time process table
  chat-app          - Multi-room chat application
  text-editor       - Text editor with syntax highlighting
```

### `make clean`
Removes all generated files (protobuf files and binaries).

**Example**:
```bash
make clean
```

### `make proto`
Compiles protobuf files. This is automatically run by build targets.

**Example**:
```bash
make proto
```

### `make fmt`
Formats all Go code using `go fmt`.

**Example**:
```bash
make fmt
```

### `make tidy`
Tidies Go module dependencies.

**Example**:
```bash
make tidy
```

---

## Development Targets

### `make dev`
Development workflow: cleans, builds server, and runs it.

**Example**:
```bash
make dev
```

### `make install-tools`
Installs required protobuf tools.

**Example**:
```bash
make install-tools
```

---

## Common Workflows

### First Time Setup

```bash
# Install tools
make install-tools

# Build everything
make build
```

### Running the Text Editor Example

```bash
# Terminal 1: Start server
make run

# Terminal 2: Run text editor
make run-example EXAMPLE=text-editor

# Or run the binary directly
./bin/examples/text-editor myfile.go
```

### Development Cycle

```bash
# Make code changes...

# Format code
make fmt

# Rebuild
make build

# Run server
make run
```

### Clean Build

```bash
# Clean everything
make clean

# Rebuild from scratch
make build
```

---

## Tips

1. **Always start the server first** before running examples or clients
2. **Use `make list-examples`** to see what's available
3. **Run examples directly** from `bin/examples/` if you prefer
4. **Use `make dev`** for quick server restart during development
5. **Run `make clean`** if you encounter build issues

---

## Troubleshooting

### "protoc: command not found"
Install protobuf compiler:
- macOS: `brew install protobuf`
- Linux: `apt-get install protobuf-compiler`

### "Error: EXAMPLE not specified"
You forgot to specify which example to run:
```bash
make run-example EXAMPLE=text-editor
```

### Examples won't connect
Make sure the server is running in another terminal:
```bash
make run
```

---

## See Also

- [Examples README](examples/README.md) - Detailed example documentation
- [Client Library README](pkg/client/README.md) - Client API documentation
- [Tutorial](docs/TUTORIAL.md) - Step-by-step tutorial

