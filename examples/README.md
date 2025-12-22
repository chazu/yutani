# Yutani Examples

This directory contains example applications demonstrating various features of the Yutani TUI framework.

## Overview

| Example | Difficulty | Widgets Used | Features Demonstrated |
|---------|-----------|--------------|----------------------|
| [Hello World](hello-world/) | Beginner | Box, TextView | Basic connection, simple widgets |
| [List Example](list-example/) | Beginner | List | List widget, item selection, events |
| [Table Example](table-example/) | Beginner | Table | Table widget, cell updates, selection |
| [Form Example](form-example/) | Intermediate | Form | Form widget, input handling, validation |
| [File Browser](file-browser/) | Intermediate | TreeView | Tree navigation, directory traversal |
| [Dashboard](dashboard/) | Intermediate | Grid, TextView | Grid layout, real-time updates |
| [Process Monitor](process-monitor/) | Intermediate | Table | Real-time data, process listing |
| [Chat App](chat-app/) | Advanced | Pages, InputField, List, Flex | Multi-page UI, user input, messaging |
| [Text Editor](text-editor/) | Advanced | TextView, Flex | Syntax highlighting, file I/O, keyboard shortcuts |

## Quick Start

### Prerequisites

1. **Start the Yutani server**:
   ```bash
   cd /path/to/yutani
   go run cmd/server/main.go
   ```

2. **In a new terminal, run an example**:
   ```bash
   cd examples/hello-world
   go run main.go
   ```

## Example Descriptions

### 1. Hello World
**File**: `hello-world/main.go`  
**Difficulty**: ⭐ Beginner

The simplest possible Yutani application. Creates a box with a text view displaying "Hello, World!".

**What you'll learn**:
- Connecting to the Yutani server
- Creating basic widgets (Box, TextView)
- Starting the event stream
- Handling graceful shutdown

**Run**:
```bash
cd examples/hello-world
go run main.go
```

---

### 2. List Example
**File**: `list-example/main.go`  
**Difficulty**: ⭐ Beginner

Demonstrates the List widget with selectable items and keyboard shortcuts.

**What you'll learn**:
- Creating and populating lists
- Handling selection events
- Adding keyboard shortcuts
- Dynamic list updates

**Run**:
```bash
cd examples/list-example
go run main.go
```

**Controls**:
- Arrow keys: Navigate
- Enter: Select item
- `a`: Add item
- `d`: Delete selected item
- `q`: Quit

---

### 3. Table Example
**File**: `table-example/main.go`  
**Difficulty**: ⭐ Beginner

Shows how to use the Table widget with headers, cell colors, and selection.

**What you'll learn**:
- Creating tables with headers
- Setting cell values and colors
- Handling cell selection
- Batch cell updates

**Run**:
```bash
cd examples/table-example
go run main.go
```

**Controls**:
- Arrow keys: Navigate cells
- `r`: Refresh data
- `q`: Quit

---

### 4. Form Example
**File**: `form-example/main.go`  
**Difficulty**: ⭐⭐ Intermediate

Demonstrates form creation with various input types and validation.

**What you'll learn**:
- Creating forms with multiple fields
- Input validation
- Form submission handling
- Field types (text, password, checkbox, dropdown)

**Run**:
```bash
cd examples/form-example
go run main.go
```

**Controls**:
- Tab: Next field
- Shift+Tab: Previous field
- Enter: Submit form
- `q`: Quit

---

### 5. File Browser
**File**: `file-browser/main.go`  
**Difficulty**: ⭐⭐ Intermediate

A file system browser using the TreeView widget.

**What you'll learn**:
- TreeView widget usage
- Hierarchical data structures
- Lazy loading of tree nodes
- File system operations

**Features**:
- 📁 Directory navigation
- 📄 File size display
- Expandable/collapsible folders
- Keyboard navigation

**Run**:
```bash
cd examples/file-browser
go run main.go
```

**Controls**:
- Arrow keys: Navigate tree
- Enter: Expand/collapse directory
- `r`: Refresh current directory
- `h`: Show help
- `q`: Quit

---

### 6. Dashboard
**File**: `dashboard/main.go`  
**Difficulty**: ⭐⭐ Intermediate

A system monitoring dashboard using Grid layout.

**What you'll learn**:
- Grid layout for complex UIs
- Real-time data updates
- Multiple widgets in a grid
- Periodic refresh with timers

**Features**:
- CPU usage monitoring
- Memory statistics
- Goroutine tracking
- System information
- Auto-refresh every 2 seconds

**Run**:
```bash
cd examples/dashboard
go run main.go
```

**Controls**:
- `r`: Force refresh
- `h`: Show help
- `q`: Quit

---

### 7. Process Monitor
**File**: `process-monitor/main.go`  
**Difficulty**: ⭐⭐ Intermediate

A process monitoring tool using the Table widget with real-time updates.

**What you'll learn**:
- Real-time table updates
- Process listing (platform-specific)
- Colored table cells
- Batch cell updates for performance

**Features**:
- Live process list
- CPU and memory usage
- Process status indicators
- Auto-refresh every 3 seconds
- Color-coded status

**Run**:
```bash
cd examples/process-monitor
go run main.go
```

**Controls**:
- Arrow keys: Navigate processes
- `r`: Refresh process list
- `k`: Kill selected process (demo only)
- `h`: Show help
- `q`: Quit

---

### 8. Chat Application
**File**: `chat-app/main.go`  
**Difficulty**: ⭐⭐⭐ Advanced

A multi-room chat application using Pages, InputField, and List widgets.

**What you'll learn**:
- Pages widget for multi-page UIs
- InputField for user input
- Flex layout for complex arrangements
- Message history management

**Features**:
- Multiple chat rooms
- Message input field
- Message history
- Room switching
- Status bar

**Run**:
```bash
cd examples/chat-app
go run main.go
```

**Controls**:
- Type message and press Enter to send
- Tab: Switch between rooms
- `q`: Quit

---

### 9. Text Editor
**File**: `text-editor/main.go`  
**Difficulty**: ⭐⭐⭐ Advanced

A text editor with syntax highlighting and file operations.

**What you'll learn**:
- TextView for text editing
- File I/O operations
- Syntax highlighting
- Keyboard shortcuts
- Multi-line text handling

**Features**:
- Open and save files
- Basic syntax highlighting (Go, Python, JavaScript, etc.)
- Find text
- Line and column tracking
- Modified indicator

**Run**:
```bash
cd examples/text-editor
go run main.go [filename]
```

**Controls**:
- Ctrl+O: Open file
- Ctrl+S: Save file
- Ctrl+F: Find text
- Ctrl+H: Show help
- Ctrl+Q: Quit

---

## Building All Examples

To build all examples at once:

```bash
cd /path/to/yutani
go build ./examples/...
```

## Common Patterns

### Connecting to the Server

```go
c, err := client.Connect("localhost:7755")
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer c.Close()
```

### Setting the Root Widget

**IMPORTANT**: You must set a root widget to make it visible on the server:

```go
// After creating your widget
if err := c.SetRoot(widget); err != nil {
    log.Fatalf("Failed to set root widget: %v", err)
}
```

Without calling `SetRoot()`, your widget will be created but won't appear on the server terminal.

### Starting the Event Stream

```go
if err := c.StartEventStream(); err != nil {
    log.Fatalf("Failed to start event stream: %v", err)
}
```

### Handling Events

```go
// Handle all events
c.OnEvent(func(event *client.Event) {
    // Process event
})

// Handle specific widget events
c.OnWidgetEvent(widget.ID(), func(event *client.Event) {
    // Process widget-specific event
})

// Handle specific event types
c.OnEventType(client.EventTypeKey, func(event *client.Event) {
    // Process key events
})
```

### Graceful Shutdown

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan
fmt.Println("Shutting down...")
```

## Tips and Best Practices

1. **Always close the client**: Use `defer c.Close()` after connecting
2. **Start event stream**: Call `c.StartEventStream()` before handling events
3. **Handle errors**: Check errors from widget creation and operations
4. **Use event filters**: Use `OnWidgetEvent()` or `OnEventType()` for specific events
5. **Graceful shutdown**: Handle interrupt signals properly
6. **Update in batches**: Use batch operations for better performance
7. **Test incrementally**: Start simple and add features gradually

## Troubleshooting

### "Failed to connect"
- Make sure the Yutani server is running
- Check that the server is listening on `localhost:7755` (default port)
- Verify no firewall is blocking the connection
- You can change the port with: `YUTANI_ADDRESS=:7755 go run cmd/yutani-server/main.go`

### "Widget not displaying"
- Ensure you called `Build()` on the widget builder
- Check that the event stream is started
- Verify the widget was added to a layout or set as root

### "Events not firing"
- Make sure `StartEventStream()` was called
- Check that event handlers are registered before events occur
- Verify the event type matches what you're listening for

## Next Steps

- Read the [Client Library Documentation](../pkg/client/README.md)
- Explore the [API Reference](../docs/API.md)
- Check out the [Tutorial](../docs/TUTORIAL.md)
- Build your own application!

## Contributing

Have an idea for a new example? Contributions are welcome! Please:

1. Follow the existing example structure
2. Include clear comments and documentation
3. Add your example to this README
4. Test thoroughly before submitting

## License

These examples are part of the Yutani project and are licensed under the same terms.

