# Yutani Go Client Library

A high-level, idiomatic Go client library for the Yutani Terminal Display Server.

## Overview

The Yutani client library provides a fluent, easy-to-use API for building terminal UIs through the Yutani server. It abstracts away the low-level gRPC details and provides a clean, builder-pattern interface for creating and managing widgets.

## Features

- **Fluent Builder Pattern** - Chainable methods for easy widget configuration
- **Type-Safe API** - Strongly typed widget interfaces
- **Event Handling** - Simple callback-based event system
- **Automatic Resource Management** - Widgets are automatically registered and cleaned up
- **Comprehensive Widget Support** - All Yutani widget types supported

## Installation

```bash
go get industries/loosh/yutani/pkg/client
```

## Quick Start

```go
package main

import (
    "log"
    "industries/loosh/yutani/pkg/client"
)

func main() {
    // Connect to server
    c, err := client.Connect("localhost:50051")
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Create a list widget
    list, err := c.NewList().
        Title("Menu").
        Border(true).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Add items
    list.AddItem("New File", "Create a new file", strPtr("n"))
    list.AddItem("Open File", "Open an existing file", strPtr("o"))

    // Handle events
    c.OnEvent(func(event *client.Event) {
        if event.IsWidget() {
            log.Printf("Widget event: %s", event.Widget.Type)
        }
    })

    // Start event stream
    c.StartEventStream()

    // Keep running...
}

func strPtr(s string) *string { return &s }
```

## Supported Widgets

### Basic Widgets

- **Box** - Simple container with border and title
- **TextView** - Display text with word wrap and dynamic colors
- **Button** - Clickable button with label and colors
- **Checkbox** - Boolean toggle with label
- **InputField** - Single-line text input with label and placeholder

### Complex Widgets

- **List** - Scrollable list with items and shortcuts
- **Table** - Grid of cells with headers and selection
- **Form** - Form with input fields, checkboxes, dropdowns, and buttons
- **TreeView** - Hierarchical tree structure with expandable nodes

### Layout Widgets

- **Flex** - Flexible box layout (row or column)
- **Grid** - Grid layout with cells and spans
- **Pages** - Multi-page container with page switching

## Widget Examples

### List Widget

```go
list, _ := c.NewList().
    Title("File Menu").
    Border(true).
    BorderColor(client.Color("blue")).
    Build()

list.AddItem("New File", "Create a new file", strPtr("n"))
list.AddItem("Open File", "Open an existing file", strPtr("o"))
list.SetSelected(0)

selected, _ := list.GetSelected()
count, _ := list.GetItemCount()
```

### Table Widget

```go
table, _ := c.NewTable().
    Title("Data Table").
    Border(true).
    Build()

// Set header
table.SetCell(0, 0, client.NewTableCellWithColor("Name", client.Color("yellow")))
table.SetCell(0, 1, client.NewTableCellWithColor("Age", client.Color("yellow")))

// Set data
table.SetCells([]*pb.TableCellUpdate{
    {Row: 1, Column: 0, Cell: client.NewTableCell("Alice")},
    {Row: 1, Column: 1, Cell: client.NewTableCell("30")},
})

table.SetFixed(1, 0) // Fix header row
```

### Form Widget

```go
form, _ := c.NewForm().
    Title("Login").
    Border(true).
    Build()

usernameIdx, _ := form.AddInputField("Username", 30, "")
passwordIdx, _ := form.AddPasswordField("Password", 30)
rememberIdx, _ := form.AddCheckbox("Remember me", false)

form.AddButton("Login")

// Get values
username, _ := form.GetFieldValue(usernameIdx)
password, _ := form.GetFieldValue(passwordIdx)
```

### TextView Widget

```go
textView, _ := c.NewTextView().
    Title("Output").
    Border(true).
    Text("Hello, World!").
    WordWrap(true).
    DynamicColors(true).
    Build()

textView.SetText("Updated text")
```

### Button Widget

```go
button, _ := c.NewButton().
    Title("Action").
    Border(true).
    Label("Click Me!").
    LabelColor(client.Color("white")).
    BackgroundColor(client.Color("blue")).
    ActivatedColor(client.Color("green")).
    Build()

button.SetLabel("Clicked!")
```

### Checkbox Widget

```go
checkbox, _ := c.NewCheckbox().
    Title("Options").
    Border(true).
    Label("Enable feature").
    Checked(false).
    LabelColor(client.Color("yellow")).
    CheckedColor(client.Color("green")).
    Build()

checkbox.SetChecked(true)
checkbox.SetLabel("Feature enabled")
```

### InputField Widget

```go
input, _ := c.NewInputField().
    Title("User Input").
    Border(true).
    Label("Name: ").
    Placeholder("Enter your name").
    FieldWidth(30).
    LabelColor(client.Color("cyan")).
    Build()

input.SetText("John Doe")
input.SetLabel("Full Name: ")
```

### TreeView Widget

```go
tree, _ := c.NewTreeView().
    Title("File Browser").
    Border(true).
    NodeTextColor(client.Color("white")).
    SelectedTextColor(client.Color("black")).
    SelectedBackgroundColor(client.Color("blue")).
    ShowGraphics(true).
    Build()

// Create root node
rootNode := client.NewTreeNode("Root")
rootID, _ := tree.SetRoot(rootNode)

// Add children
child1 := client.TreeNodeWithColor("Documents", client.Color("yellow"))
child1ID, _ := tree.AddChild(rootID, child1)

child2 := client.NewTreeNode("Pictures")
tree.AddChild(rootID, child2)

// Expand/collapse nodes
tree.SetExpanded(rootID, true)

// Get selected node
nodeID, text, ref, _ := tree.GetSelected()
```

### Flex Layout Widget

```go
flex, _ := c.NewFlex().
    Title("Layout").
    Border(true).
    Direction(pb.FlexDirection_FLEX_COLUMN).
    Build()

// Create child widgets
header, _ := c.NewTextView().Title("Header").Build()
content, _ := c.NewTextView().Title("Content").Build()
footer, _ := c.NewTextView().Title("Footer").Build()

// Add items with proportions
flex.AddItem(header, 0, 3, false)  // Fixed 3 lines
flex.AddItem(content, 1, 0, true)  // Proportional, takes remaining space
flex.AddItem(footer, 0, 1, false)  // Fixed 1 line
```

### Grid Layout Widget

```go
grid, _ := c.NewGrid().
    Title("Dashboard").
    Border(true).
    Rows(2).
    Columns(2).
    Build()

// Create widgets for grid cells
topLeft, _ := c.NewTextView().Title("CPU").Build()
topRight, _ := c.NewTextView().Title("Memory").Build()
bottomLeft, _ := c.NewTextView().Title("Disk").Build()
bottomRight, _ := c.NewTextView().Title("Network").Build()

// Add items to grid (row, column, rowSpan, columnSpan, minWidth, minHeight, focus)
grid.AddItem(topLeft, 0, 0, 1, 1, 0, 0, false)
grid.AddItem(topRight, 0, 1, 1, 1, 0, 0, false)
grid.AddItem(bottomLeft, 1, 0, 1, 1, 0, 0, false)
grid.AddItem(bottomRight, 1, 1, 1, 1, 0, 0, false)
```

### Pages Layout Widget

```go
pages, _ := c.NewPages().
    Title("Multi-Page App").
    Border(true).
    ShowPageNames(true).
    PageNameColor(client.Color("cyan")).
    Build()

// Create pages
page1, _ := c.NewTextView().Text("Page 1 content").Build()
page2, _ := c.NewTextView().Text("Page 2 content").Build()
page3, _ := c.NewTextView().Text("Page 3 content").Build()

// Add pages
pages.AddPage("home", page1, true, true)
pages.AddPage("settings", page2, true, false)
pages.AddPage("about", page3, true, false)

// Switch pages
pages.ShowPage("settings")

// Get current page
currentPage, _ := pages.GetCurrentPage()
```

## Event Handling

### Basic Event Handling

```go
c.OnEvent(func(event *client.Event) {
    switch {
    case event.IsKey():
        log.Printf("Key: %s (rune: %c)", event.Key.Key, event.Key.Rune)

    case event.IsMouse():
        log.Printf("Mouse: (%d,%d) button: %d", event.Mouse.X, event.Mouse.Y, event.Mouse.Button)

    case event.IsWidget():
        log.Printf("Widget %s: %s", event.Widget.WidgetID, event.Widget.Type)

    case event.IsResize():
        log.Printf("Resize: %dx%d", event.Resize.Width, event.Resize.Height)
    }
})

c.StartEventStream()
```

### Advanced Event Handling

#### Event Filtering by Type

```go
// Only handle key events
c.OnEventType(client.EventTypeKey, func(event *client.Event) {
    log.Printf("Key pressed: %c", event.Key.Rune)
})

// Only handle widget events
c.OnEventType(client.EventTypeWidget, func(event *client.Event) {
    log.Printf("Widget event: %s", event.Widget.Type)
})
```

#### Event Filtering by Widget

```go
// Only handle events from a specific widget
list, _ := c.NewList().Title("Menu").Build()

c.OnWidgetEvent(list.ID(), func(event *client.Event) {
    log.Printf("List event: %s", event.Widget.Type)
})
```

#### Custom Event Filters

```go
// Filter with custom logic
filter := &client.EventFilter{
    Types: []client.EventType{client.EventTypeKey},
    CustomFilter: func(e *client.Event) bool {
        // Only handle 'Enter' key
        return e.Key != nil && e.Key.Key == "KEY_ENTER"
    },
}

c.OnEventFiltered(func(event *client.Event) {
    log.Println("Enter key pressed!")
}, filter)
```

#### Event Middleware

```go
// Add middleware to log all events
c.AddEventMiddleware(func(event *client.Event) (*client.Event, bool) {
    log.Printf("Event: %v", event.Type)
    return event, true // Continue processing
})

// Add middleware to block certain events
c.AddEventMiddleware(func(event *client.Event) (*client.Event, bool) {
    if event.Type == client.EventTypeMouse {
        return event, false // Block mouse events
    }
    return event, true
})

// Add middleware to modify events
c.AddEventMiddleware(func(event *client.Event) (*client.Event, bool) {
    if event.Type == client.EventTypeKey && event.Key != nil {
        // Convert to uppercase
        if event.Key.Rune >= 'a' && event.Key.Rune <= 'z' {
            event.Key.Rune = event.Key.Rune - 32
        }
    }
    return event, true
})
```

#### Event Batching

```go
// Batch high-frequency events
batcher := client.NewEventBatcher(100*time.Millisecond, func(event *client.Event) {
    // This handler receives batched events every 100ms
    log.Printf("Batched event: %v", event.Type)
})
defer batcher.Close()

// Add events to the batcher
c.OnEvent(func(event *client.Event) {
    if event.Type == client.EventTypeMouse {
        batcher.Add(event)
    }
})
```

#### Event Recording and Replay

```go
// Enable event recording
c.EnableEventRecording(1000) // Keep last 1000 events

// Get the recorder
recorder := c.GetEventRecorder()

// Get all recorded events
events := recorder.GetEvents()
log.Printf("Recorded %d events", len(events))

// Get events since a specific time
since := time.Now().Add(-5 * time.Minute)
recentEvents := recorder.GetEventsSince(since)

// Replay events
recorder.Replay(func(event *client.Event) {
    log.Printf("Replaying: %v", event.Type)
}, false) // false = replay immediately, true = replay with original timing

// Control recording
recorder.Stop()  // Pause recording
recorder.Start() // Resume recording
recorder.Clear() // Clear all recorded events
```

#### Server-Side Event Filtering

```go
// Reduce network traffic by filtering at the server
c.SetServerEventFilterSimple(
    true,  // key events
    false, // mouse events (disabled)
    true,  // resize events
    true,  // focus events
    true,  // widget events
)

// Filter by specific widgets at the server
c.SetServerWidgetFilter([]string{
    widget1.ID(),
    widget2.ID(),
})
```

## Helper Functions

```go
// Color helpers
client.Color("red")                    // Named color
client.ColorRGB(255, 0, 0)            // RGB color
client.ColorHex("#ff0000")            // Hex color

// Table cell helpers
client.NewTableCell("text")                           // Simple cell
client.NewTableCellWithColor("text", client.Color("yellow"))  // Colored cell
```

## Examples

See the `examples/` directory for complete working examples:

- `examples/simple-list/` - Simple list widget demo
- `examples/data-table/` - Data table with employee directory
- `examples/login-form/` - Login form with multiple field types

## API Reference

See the [GoDoc](https://pkg.go.dev/industries/loosh/yutani/pkg/client) for complete API documentation.

