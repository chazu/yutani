# Yutani Tutorial: Building Terminal UIs

This tutorial will guide you through building terminal user interfaces with the Yutani Terminal Display Server and its Go client library.

## Prerequisites

- Go 1.21 or later
- Yutani server running (see [QUICKSTART.md](QUICKSTART.md))
- Basic understanding of Go programming

## Tutorial 1: Your First Widget

Let's create a simple "Hello World" application with a text view widget.

### Step 1: Import the Client Library

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "industries/loosh/yutani/pkg/client"
)
```

### Step 2: Connect to the Server

```go
func main() {
    // Connect to the Yutani server
    c, err := client.Connect("localhost:50051")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer c.Close()
    
    log.Println("Connected to Yutani server")
```

### Step 3: Create a Widget

```go
    // Create a text view widget
    textView, err := c.NewTextView().
        Title("Hello World").
        Border(true).
        Text("Welcome to Yutani!").
        WordWrap(true).
        Build()
    if err != nil {
        log.Fatalf("Failed to create widget: %v", err)
    }
    
    log.Printf("Created widget: %s", textView.ID())
```

### Step 4: Keep the Application Running

```go
    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    log.Println("Shutting down...")
}
```

### Run It!

```bash
# Terminal 1: Start the server
./bin/yutani server

# Terminal 2: Run your application
go run your-app.go
```

You should see a bordered text view with "Welcome to Yutani!" displayed.

## Tutorial 2: Interactive List

Let's build an interactive menu using a list widget.

### Step 1: Create the List

```go
list, err := c.NewList().
    Title("Main Menu").
    Border(true).
    BorderColor(client.Color("cyan")).
    Build()
if err != nil {
    log.Fatal(err)
}
```

### Step 2: Add Menu Items

```go
items := []struct {
    main      string
    secondary string
    shortcut  string
}{
    {"New Project", "Create a new project", "n"},
    {"Open Project", "Open an existing project", "o"},
    {"Settings", "Configure application settings", "s"},
    {"Help", "Show help documentation", "h"},
    {"Exit", "Exit the application", "q"},
}

for _, item := range items {
    shortcut := item.shortcut
    list.AddItem(item.main, item.secondary, &shortcut)
}

list.SetSelected(0) // Select first item
```

### Step 3: Handle Events

```go
// Start event stream
if err := c.StartEventStream(); err != nil {
    log.Fatal(err)
}

// Handle events
c.OnEvent(func(event *client.Event) {
    if event.IsWidget() && event.Widget.WidgetID == list.ID() {
        // Get selected item
        if selected, err := list.GetSelected(); err == nil {
            if main, _, _, err := list.GetItem(selected); err == nil {
                log.Printf("Selected: %s", main)
                
                // Handle selection
                if main == "Exit" {
                    os.Exit(0)
                }
            }
        }
    }
    
    if event.IsKey() && event.Key.Rune == 'q' {
        os.Exit(0)
    }
})
```

## Tutorial 3: Data Table

Let's create a data table to display structured information.

### Step 1: Create the Table

```go
table, err := c.NewTable().
    Title("Employee Directory").
    Border(true).
    Build()
if err != nil {
    log.Fatal(err)
}
```

### Step 2: Set Headers

```go
headers := []string{"Name", "Department", "Email"}
for i, header := range headers {
    table.SetCell(0, i, &pb.TableCell{
        Text:  header,
        Color: client.Color("yellow"),
    })
}

// Fix the header row
table.SetFixed(1, 0)
```

### Step 3: Add Data

```go
employees := [][]string{
    {"Alice Johnson", "Engineering", "alice@example.com"},
    {"Bob Smith", "Marketing", "bob@example.com"},
    {"Carol White", "Sales", "carol@example.com"},
}

// Batch set cells for better performance
var cells []*pb.TableCellUpdate
for row, employee := range employees {
    for col, value := range employee {
        cells = append(cells, &pb.TableCellUpdate{
            Row:    int32(row + 1), // +1 for header
            Column: int32(col),
            Cell:   client.NewTableCell(value),
        })
    }
}

table.SetCells(cells)
```

### Step 4: Handle Selection

```go
c.OnEvent(func(event *client.Event) {
    if event.IsWidget() && event.Widget.WidgetID == table.ID() {
        if row, col, err := table.GetSelection(); err == nil {
            if cell, err := table.GetCell(row, col); err == nil {
                log.Printf("Selected: [%d,%d] = %s", row, col, cell.Text)
            }
        }
    }
})
```

## Tutorial 4: Login Form

Let's build a complete login form with validation.

### Step 1: Create the Form

```go
form, err := c.NewForm().
    Title("Login").
    Border(true).
    Build()
if err != nil {
    log.Fatal(err)
}
```

### Step 2: Add Form Fields

```go
usernameIdx, _ := form.AddInputField("Username", 30, "")
passwordIdx, _ := form.AddPasswordField("Password", 30)
rememberIdx, _ := form.AddCheckbox("Remember me", false)

form.AddButton("Login")
form.AddButton("Cancel")
```

### Step 3: Handle Form Submission

```go
c.OnEvent(func(event *client.Event) {
    if event.IsWidget() && event.Widget.WidgetID == form.ID() {
        if event.Widget.Type == "SUBMITTED" {
            // Get form values
            username, _ := form.GetFieldValue(usernameIdx)
            password, _ := form.GetFieldValue(passwordIdx)
            remember, _ := form.GetFieldValue(rememberIdx)
            
            // Validate
            if username == "" || password == "" {
                log.Println("Please fill in all fields")
                return
            }
            
            // Process login
            log.Printf("Login: %s (remember: %s)", username, remember)
            
            // Clear form after submission
            form.Clear()
        }
    }
})
```

## Tutorial 5: Window Management

Yutani provides a floating window management system. Windows can be dragged, resized, minimized, maximized, and closed — all rendered in the terminal.

### Step 1: Create a Window Manager

The window manager is a container that holds windows and handles their positioning, z-ordering, and mouse interactions. Set it as the root widget.

```go
wm, err := c.NewWindowManager().
    Title("Desktop").
    Build()
if err != nil {
    log.Fatal(err)
}

if err := c.SetRoot(wm); err != nil {
    log.Fatal(err)
}
```

### Step 2: Create Windows

Each window wraps a child content widget. Specify position, size, and behavior constraints.

```go
win1, err := c.NewWindow().
    Title("Editor").
    Rect(2, 2, 40, 15).       // x, y, width, height
    Resizable(true).
    Movable(true).
    MinSize(20, 8).            // minimum width, height
    MaxSize(80, 30).           // maximum width, height
    Build()
if err != nil {
    log.Fatal(err)
}

win2, err := c.NewWindow().
    Title("Output").
    Rect(25, 8, 35, 12).
    Resizable(true).
    Movable(true).
    Build()
if err != nil {
    log.Fatal(err)
}
```

### Step 3: Add Windows to the Manager

```go
if err := wm.AddWindow(win1); err != nil {
    log.Fatal(err)
}
if err := wm.AddWindow(win2); err != nil {
    log.Fatal(err)
}
```

Windows are stacked in the order they are added — the last window added is on top.

### Step 4: Control Windows Programmatically

```go
// Move a window
win1.Move(10, 5)

// Resize a window
win1.Resize(50, 20)

// Minimize, maximize, restore
win1.Minimize()
win1.Maximize()
win1.Restore()

// Query current state and geometry
state, err := win1.GetState()
if err == nil {
    fmt.Printf("State: %v, Rect: (%d,%d,%d,%d)\n",
        state.State, state.Rect.X, state.Rect.Y,
        state.Rect.Width, state.Rect.Height)
}

// Set constraints after creation
win1.SetConstraints(&pb.WindowConstraints{
    Resizable: testutil.BoolPtr(false),
    Movable:   testutil.BoolPtr(true),
    MinWidth:  testutil.Int32Ptr(15),
    MinHeight: testutil.Int32Ptr(8),
})
```

### Step 5: Z-Order Control

```go
// Bring a window to the front
wm.BringToFront(win1)

// Send a window to the back
wm.SendToBack(win2)

// Get the current z-order (front to back)
order, err := wm.GetZOrder()
if err == nil {
    fmt.Println("Z-order (front to back):", order)
}
```

### Step 6: Handle Window Events

Windows emit events when moved, resized, or state-changed.

```go
c.OnEventType(client.EventTypeWidget, func(event *client.Event) {
    if event.Widget == nil {
        return
    }
    switch event.Widget.Type {
    case client.WidgetEventWindowMoved:
        fmt.Printf("Window %s moved\n", event.Widget.WidgetID)
    case client.WidgetEventWindowResized:
        fmt.Printf("Window %s resized\n", event.Widget.WidgetID)
    case client.WidgetEventWindowStateChanged:
        fmt.Printf("Window %s state changed\n", event.Widget.WidgetID)
    case client.WidgetEventWindowClosed:
        fmt.Printf("Window %s closed\n", event.Widget.WidgetID)
    case client.WidgetEventWindowActivated:
        fmt.Printf("Window %s activated\n", event.Widget.WidgetID)
    }
})
```

### Mouse Interactions

The window manager handles all mouse interactions automatically:

- **Title bar**: Drag to move the window
- **`[_]` button**: Minimize (collapse to title bar) or restore
- **`[^]` button**: Maximize (fill manager area) or restore
- **`[X]` button**: Close and remove the window
- **Edges/corners**: Drag to resize
- **Click a window**: Brings it to front

### Run the Demo

A complete working example is in `examples/window-demo/`:

```bash
# Terminal 1: Start server
./bin/yutani server

# Terminal 2: Run demo
go run ./examples/window-demo
```

## Tutorial 6: Best Practices

### 1. Always Close the Client

```go
c, err := client.Connect("localhost:50051")
if err != nil {
    log.Fatal(err)
}
defer c.Close() // Always defer Close()
```

### 2. Check Errors

```go
// Don't ignore errors
list, err := c.NewList().Build()
if err != nil {
    log.Printf("Failed to create list: %v", err)
    return
}

_, err = list.AddItem("Item", "Description", nil)
if err != nil {
    log.Printf("Failed to add item: %v", err)
}
```

### 3. Use Batch Operations for Performance

```go
// Instead of multiple SetCell calls (slow):
for row := 0; row < 100; row++ {
    for col := 0; col < 10; col++ {
        table.SetCell(row, col, cell)
    }
}

// Use SetCells for better performance:
var cells []*pb.TableCellUpdate
for row := 0; row < 100; row++ {
    for col := 0; col < 10; col++ {
        cells = append(cells, &pb.TableCellUpdate{
            Row: int32(row), Column: int32(col), Cell: cell,
        })
    }
}
table.SetCells(cells) // Much faster!
```

### 4. Handle Signals Gracefully

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Println("Shutting down gracefully...")
    c.Close()
    os.Exit(0)
}()
```

## Next Steps

- Explore the [examples/](examples/) directory for complete applications
- Read the [client library documentation](pkg/client/README.md)
- Check out [PHASE4_COMPLETE.md](PHASE4_COMPLETE.md) for advanced features
- Build your own terminal UI application!

## Troubleshooting

### Connection Refused

Make sure the Yutani server is running:
```bash
./bin/yutani server
```

### Widget Not Appearing

1. Check that the widget was created successfully (no error)
2. Ensure the server is running
3. Verify the widget has `Visible(true)` set (default is true)

### Events Not Firing

1. Make sure you called `c.StartEventStream()`
2. Check that your event handler is registered before starting the stream
3. Verify the server is sending events (check server logs)

## Getting Help

- **GitHub Issues**: Report bugs or request features
- **Documentation**: See [README.md](README.md) for project overview
- **Examples**: Check [examples/](examples/) for working code
- **API Reference**: See [pkg/client/README.md](pkg/client/README.md)


