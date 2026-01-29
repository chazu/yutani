# Yutani Debug Guide

This guide explains how to use Yutani's debugging tools to inspect and troubleshoot TUI applications. These tools are especially useful for LLM agents debugging text-based user interfaces.

## Overview

Yutani provides a `DebugService` that exposes:
- **Screen dumps**: ASCII representation of what's currently rendered
- **Widget state**: Properties, bounds, focus, and recent events for any widget
- **Widget bounds**: Position and size of all widgets in a session

## CLI Commands

### Dump Current Screen

See exactly what's rendered on screen:

```bash
# Basic screen dump
yutani debug screen -s <session-id>

# With widget bounds overlay (shows widget positions as markers)
yutani debug screen -s <session-id> --bounds

# With legend explaining widget markers
yutani debug screen -s <session-id> --bounds --legend

# JSON output for programmatic use
yutani debug screen -s <session-id> --format json
```

Example output with `--bounds --legend`:
```
┌─────────────────────────┐
│ MaggieIDE             X │
├─────────────────────────┤
│ [A]Class Browser        │
│ [A]Inspector            │
│ [A]REPL                 │
│ [B]                     │
└─────────────────────────┘

Legend:
  A: List (abc123) - focused
  B: TextView (def456)
```

### Inspect Widget State

Get detailed information about a specific widget:

```bash
yutani debug widget -s <session-id> -w <widget-id>

# JSON output
yutani debug widget -s <session-id> -w <widget-id> --format json
```

Example output:
```
Widget: abc123
  Type: List
  Bounds: (2, 3) - (25, 8) [23x5]
  Focused: true
  Visible: true
  Properties:
    title: "Tools"
    border: true
    selectedIndex: 2
  Parent: root_flex_456
  Children: (none)
  Recent Events:
    [12:34:56.123] KEY_DOWN
    [12:34:56.456] KEY_ENTER
```

### List All Widget Bounds

See positions of all widgets:

```bash
yutani debug bounds -s <session-id>

# JSON output
yutani debug bounds -s <session-id> --format json
```

Example output:
```
Widget ID       Type        Position    Size      Focused
─────────────────────────────────────────────────────────
abc123          List        (2, 3)      23x5      *
def456          TextView    (2, 9)      23x1
ghi789          Flex        (0, 0)      80x24
```

## Finding the Session ID

When a Maggie/Yutani application starts, it logs the session ID:

```
YutaniSession: connected, creating session
YutaniSession: session created with ID: f1e6eb39-a02f-43cf-8db3-01cc27925db1
```

You can also list active sessions:

```bash
yutani debug sessions
```

## Common Debugging Scenarios

### "My button click isn't working"

1. Dump the screen to see if the button is visible:
   ```bash
   yutani debug screen -s <session-id> --bounds --legend
   ```

2. Check if the button has focus:
   ```bash
   yutani debug widget -s <session-id> -w <button-id>
   ```

3. Look at recent events to see if clicks are being received:
   ```bash
   yutani debug widget -s <session-id> -w <button-id> --format json | jq '.recent_events'
   ```

### "My list selection isn't updating"

1. Check the list's current state:
   ```bash
   yutani debug widget -s <session-id> -w <list-id>
   ```
   Look for `selectedIndex` in properties.

2. Verify the list has focus (keyboard events go to focused widget):
   ```bash
   yutani debug bounds -s <session-id> | grep -E "Focused|<list-id>"
   ```

### "I can't see what's on screen"

For LLM agents that can't see the terminal:

```bash
# Get a complete picture
yutani debug screen -s <session-id> --bounds --legend --format json
```

This returns JSON with:
- `ascii_art`: The screen as text
- `widgets`: Array of widget bounds with IDs and types
- `width`, `height`: Screen dimensions

## Using with LLM Agents

When debugging with an LLM agent (like Claude), provide this context:

```bash
# Capture current state
echo "=== SCREEN ==="
yutani debug screen -s <session-id> --bounds --legend
echo ""
echo "=== FOCUSED WIDGET ==="
yutani debug widget -s <session-id> -w $(yutani debug bounds -s <session-id> --format json | jq -r '.widgets[] | select(.has_focus) | .widget_id')
```

The LLM can then reason about:
- What's visible on screen
- Which widget has focus
- What events have been received
- Why an interaction might not be working

## gRPC API

For programmatic access, use the DebugService directly:

```protobuf
service DebugService {
  rpc GetScreenDump(GetScreenDumpRequest) returns (GetScreenDumpResponse);
  rpc GetWidgetState(GetWidgetStateRequest) returns (GetWidgetStateResponse);
  rpc GetAllWidgetBounds(GetAllWidgetBoundsRequest) returns (GetAllWidgetBoundsResponse);
}
```

Example with grpcurl:

```bash
# Screen dump
grpcurl -plaintext -d '{"session_id": {"id": "SESSION_ID"}, "include_widget_bounds": true}' \
  localhost:7755 industries.loosh.yutani.v1.DebugService/GetScreenDump

# Widget state
grpcurl -plaintext -d '{"session_id": {"id": "SESSION_ID"}, "widget_id": {"id": "WIDGET_ID"}}' \
  localhost:7755 industries.loosh.yutani.v1.DebugService/GetWidgetState
```

## Tips

1. **Keep session IDs handy**: Log them at startup for easy debugging
2. **Use JSON format**: Easier to parse and search through
3. **Check focus first**: Most keyboard interactions require the widget to be focused
4. **Compare bounds**: If clicks aren't working, verify click coordinates vs widget bounds
5. **Watch recent events**: They show what the widget actually received

## See Also

- [QUICKSTART.md](QUICKSTART.md) - Getting started with Yutani
- [examples/](examples/) - Example applications
- [TUTORIAL.md](TUTORIAL.md) - Client library tutorial
