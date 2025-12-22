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

### Complex Widgets

- **List** - Scrollable list with items and shortcuts
- **Table** - Grid of cells with headers and selection
- **Form** - Form with input fields, checkboxes, dropdowns, and buttons

### Layout Widgets

- **Flex** - Flexible box layout (coming soon)
- **Grid** - Grid layout with cells (coming soon)
- **Pages** - Multi-page container (coming soon)

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

## Event Handling

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

