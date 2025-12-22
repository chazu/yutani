# Yutani Testing Utilities

Comprehensive testing utilities for Yutani applications, including mock clients, test helpers, event simulation, and integration test support.

## Overview

The `testing` package provides everything you need to test Yutani applications:

- **Mock Client** - In-memory client that doesn't require a server
- **Test Helpers** - Convenience functions for common test scenarios
- **Event Simulation** - Simulate keyboard, mouse, and widget events
- **Widget Assertions** - Assert widget state and properties
- **Integration Test Helpers** - Start test servers and manage lifecycle

## Installation

```bash
import testing "github.com/chazu/yutani/pkg/client/testing"
```

## Quick Start

### Unit Testing with Mock Client

```go
func TestMyApp(t *testing.T) {
    // Create mock client
    mc := testing.NewMockClient()
    
    // Create widgets
    widgetID := testing.CreateMockList(mc, "Test List")
    
    // Simulate user interaction
    mc.SimulateKeyPress('q')
    
    // Verify behavior
    if mc.GetCallCount("SimulateKeyPress") != 1 {
        t.Error("Expected key press to be simulated")
    }
}
```

### Integration Testing with Test Server

```go
func TestWithRealServer(t *testing.T) {
    testing.WithTestClient(t, func(t *testing.T, c *client.Client) {
        // Create real widgets
        list, err := c.NewList().Title("Test").Build()
        if err != nil {
            t.Fatal(err)
        }
        
        // Test real interactions
        list.AddItem("Item 1", "", nil)
    })
}
```

## Mock Client

### Creating a Mock Client

```go
mc := testing.NewMockClient()
```

### Recording Method Calls

```go
mc.RecordCall("MyMethod", arg1, arg2)

// Check call count
count := mc.GetCallCount("MyMethod")

// Get all calls
calls := mc.GetCalls("MyMethod")

// Clear recorded calls
mc.ClearCalls()
```

### Configuring Errors

```go
// Make a method return an error
mc.SetError("CreateWidget", fmt.Errorf("test error"))

// Clear error
mc.ClearError("CreateWidget")
```

### Registering Mock Widgets

```go
mc.RegisterWidget("widget-1", pb.WidgetType_WIDGET_LIST)

widget, ok := mc.GetWidget("widget-1")
if ok {
    widget.Properties["title"] = "Test Title"
}
```

## Event Simulation

### Simulating Keyboard Events

```go
// Simple key press
mc.SimulateKeyPress('q')

// Full key event
mc.SimulateKeyEvent("KEY_ENTER", '\n')
```

### Simulating Mouse Events

```go
// Mouse click at position (10, 20) with left button (1)
mc.SimulateMouseClick(10, 20, 1)
```

### Simulating Resize Events

```go
// Resize to 80x24
mc.SimulateResize(80, 24)
```

### Simulating Widget Events

```go
mc.SimulateWidgetEvent("widget-1", "WIDGET_SELECTED")
```

### Handling Simulated Events

```go
var receivedEvent *client.Event
mc.OnEvent(func(event *client.Event) {
    receivedEvent = event
})

mc.SimulateKeyPress('q')

// receivedEvent now contains the simulated event
```

## Test Helpers

### Creating a Test Helper

```go
h := testing.NewTestHelper(t)
mc := h.Client() // Get the mock client
```

### Assertions

```go
// Call count assertions
h.AssertCallCount("MyMethod", 2)
h.AssertCalled("MyMethod")
h.AssertNotCalled("MyMethod")

// Widget assertions
h.AssertWidgetExists("widget-1")
h.AssertWidgetType("widget-1", pb.WidgetType_WIDGET_LIST)
h.AssertWidgetProperty("widget-1", "title", "Test Title")

// General assertions
h.AssertEqual(actual, expected, "values should match")
h.AssertTrue(condition, "condition should be true")
h.AssertFalse(condition, "condition should be false")
h.AssertNoError(err)
h.AssertError(err)
```

### Convenience Functions

```go
// Create mock widgets
listID := testing.CreateMockList(mc, "Test List")
tableID := testing.CreateMockTable(mc, "Test Table")
```

## Widget Assertions

### Basic Widget Assertions

```go
wa := testing.NewWidgetAssertion(t, mc, "widget-1")

wa.Exists().
   HasProperty("title", "Test Title").
   HasBorder(true).
   HasText("Hello, World!")
```

### List Assertions

```go
la := testing.NewListAssertion(t, mc, "list-1")

la.AddItem("Item 1", "Description", "")
la.AddItem("Item 2", "Description", "")

la.HasItemCount(2).
   HasItem(0, "Item 1").
   HasItem(1, "Item 2")
```

### Table Assertions

```go
ta := testing.NewTableAssertion(t, mc, "table-1")

ta.SetCell(0, 0, "Header 1")
ta.SetCell(0, 1, "Header 2")
ta.SetCell(1, 0, "Value 1")
ta.SetCell(1, 1, "Value 2")

ta.HasDimensions(2, 2).
   HasCell(0, 0, "Header 1").
   HasCell(1, 0, "Value 1")
```

## Integration Testing

### Starting a Test Server

```go
ts := testing.StartTestServer(t)
defer ts.Stop()

// Get server address
address := ts.Address()
```

### Connecting to Test Server

```go
c := testing.ConnectToTestServer(t, ts)
defer c.Close()

// Use client normally
list, _ := c.NewList().Title("Test").Build()
```

### Using Helper Functions

```go
// Automatically manages server lifecycle
testing.WithTestServer(t, func(t *testing.T, ts *testing.TestServer) {
    // Test with server
})

// Automatically manages server and client lifecycle
testing.WithTestClient(t, func(t *testing.T, c *client.Client) {
    // Test with client
})
```

### Creating Test Widgets

```go
testing.WithTestClient(t, func(t *testing.T, c *client.Client) {
    widgetID := testing.CreateTestWidget(t, c, "list")
    // widgetID is the ID of the created list widget
})
```

## Complete Examples

### Example 1: Testing Event Handling

```go
func TestEventHandling(t *testing.T) {
    mc := testing.NewMockClient()
    
    var quitCalled bool
    mc.OnEvent(func(event *client.Event) {
        if event.Type == client.EventTypeKey && event.Key.Rune == 'q' {
            quitCalled = true
        }
    })
    
    mc.SimulateKeyPress('q')
    
    if !quitCalled {
        t.Error("Expected quit to be called")
    }
}
```

### Example 2: Testing Widget Creation

```go
func TestWidgetCreation(t *testing.T) {
    h := testing.NewTestHelper(t)
    mc := h.Client()
    
    widgetID := testing.CreateMockList(mc, "My List")
    
    h.AssertWidgetExists(widgetID)
    h.AssertWidgetType(widgetID, pb.WidgetType_WIDGET_LIST)
    h.AssertWidgetProperty(widgetID, "title", "My List")
}
```

### Example 3: Integration Test

```go
func TestRealListWidget(t *testing.T) {
    testing.WithTestClient(t, func(t *testing.T, c *client.Client) {
        list, err := c.NewList().
            Title("Test List").
            Border(true).
            Build()
        
        if err != nil {
            t.Fatalf("Failed to create list: %v", err)
        }
        
        err = list.AddItem("Item 1", "Description", nil)
        if err != nil {
            t.Errorf("Failed to add item: %v", err)
        }
    })
}
```

## Best Practices

1. **Use Mock Client for Unit Tests** - Fast, no server required
2. **Use Test Server for Integration Tests** - Tests real server behavior
3. **Use Test Helpers** - Reduces boilerplate code
4. **Chain Assertions** - More readable tests
5. **Clean Up Resources** - Use `defer` to close clients and stop servers
6. **Test Event Handling** - Simulate user interactions
7. **Verify Call Counts** - Ensure methods are called expected number of times

## See Also

- [Client Library Documentation](../README.md)
- [Examples](../../../examples/README.md)
- [API Reference](../../../docs/API.md)

