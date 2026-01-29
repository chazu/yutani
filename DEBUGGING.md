# Debugging Yutani Applications

This guide covers all debugging tools and techniques for Yutani applications, including the DebugService for TUI inspection, using grpcurl for ad-hoc gRPC interaction, server and client logging, and common troubleshooting scenarios.

## Overview

Yutani provides multiple layers of debugging support:

- **DebugService** -- A gRPC service for inspecting screen state, widget properties, and widget bounds. Especially useful for LLM agents that cannot see the terminal directly.
- **grpcurl** -- A command-line gRPC client that lets you interact with any Yutani service without writing code.
- **Server logging** -- Structured log output for tracking session lifecycle, widget operations, and event dispatch.
- **Client logging** -- Log output from demo and client applications showing connection status and received events.

## DebugService

The DebugService exposes three RPCs for inspecting the state of a running Yutani application:

```protobuf
service DebugService {
  rpc GetScreenDump(GetScreenDumpRequest) returns (GetScreenDumpResponse);
  rpc GetWidgetState(GetWidgetStateRequest) returns (GetWidgetStateResponse);
  rpc GetAllWidgetBounds(GetAllWidgetBoundsRequest) returns (GetAllWidgetBoundsResponse);
}
```

### Finding the Session ID

When a Yutani application starts, it logs the session ID:

```
YutaniSession: connected, creating session
YutaniSession: session created with ID: f1e6eb39-a02f-43cf-8db3-01cc27925db1
```

You can also list active sessions:

```bash
yutani debug sessions
```

### Dump Current Screen

See exactly what is rendered on screen:

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
+--------------------------+
| MaggieIDE              X |
+--------------------------+
| [A]Class Browser         |
| [A]Inspector             |
| [A]REPL                  |
| [B]                      |
+--------------------------+

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
---------------------------------------------------------
abc123          List        (2, 3)      23x5      *
def456          TextView    (2, 9)      23x1
ghi789          Flex        (0, 0)      80x24
```

### DebugService with grpcurl

```bash
# Screen dump
grpcurl -plaintext -d '{"session_id": {"id": "SESSION_ID"}, "include_widget_bounds": true}' \
  localhost:7755 industries.loosh.yutani.v1.DebugService/GetScreenDump

# Widget state
grpcurl -plaintext -d '{"session_id": {"id": "SESSION_ID"}, "widget_id": {"id": "WIDGET_ID"}}' \
  localhost:7755 industries.loosh.yutani.v1.DebugService/GetWidgetState
```

### Common Debugging Scenarios

#### "My button click isn't working"

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

#### "My list selection isn't updating"

1. Check the list's current state:
   ```bash
   yutani debug widget -s <session-id> -w <list-id>
   ```
   Look for `selectedIndex` in properties.

2. Verify the list has focus (keyboard events go to focused widget):
   ```bash
   yutani debug bounds -s <session-id> | grep -E "Focused|<list-id>"
   ```

#### "I can't see what's on screen" (LLM agents)

For LLM agents that cannot see the terminal:

```bash
# Get a complete picture
yutani debug screen -s <session-id> --bounds --legend --format json
```

This returns JSON with:
- `ascii_art`: The screen as text
- `widgets`: Array of widget bounds with IDs and types
- `width`, `height`: Screen dimensions

### Using with LLM Agents

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
- What is visible on screen
- Which widget has focus
- What events have been received
- Why an interaction might not be working

### DebugService Tips

1. **Keep session IDs handy**: Log them at startup for easy debugging
2. **Use JSON format**: Easier to parse and search through
3. **Check focus first**: Most keyboard interactions require the widget to be focused
4. **Compare bounds**: If clicks are not working, verify click coordinates vs widget bounds
5. **Watch recent events**: They show what the widget actually received

## Using grpcurl

grpcurl is a command-line tool for interacting with gRPC services. Since Yutani has gRPC reflection enabled, grpcurl can discover and call any Yutani RPC without needing proto files.

### Prerequisites

Install grpcurl:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Make sure the Yutani server is running:
```bash
./bin/yutani server
```

The server runs on `localhost:7755` by default.

### Basic Commands

#### List Available Services

```bash
grpcurl -plaintext localhost:7755 list
```

Output:
```
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
industries.loosh.yutani.v1.EventService
industries.loosh.yutani.v1.FormService
industries.loosh.yutani.v1.LayoutService
industries.loosh.yutani.v1.ListService
industries.loosh.yutani.v1.ScreenService
industries.loosh.yutani.v1.SessionService
industries.loosh.yutani.v1.TableService
industries.loosh.yutani.v1.TreeService
industries.loosh.yutani.v1.WidgetService
```

#### Describe a Service

```bash
grpcurl -plaintext localhost:7755 describe industries.loosh.yutani.v1.SessionService
```

#### Describe a Method

```bash
grpcurl -plaintext localhost:7755 describe industries.loosh.yutani.v1.SessionService.CreateSession
```

### Session Management

#### Ping the Server

```bash
grpcurl -plaintext -d '{"timestamp": 1234567890}' \
  localhost:7755 industries.loosh.yutani.v1.SessionService/Ping
```

Response:
```json
{
  "timestamp": "1234567890",
  "serverTimestamp": "1704123456789"
}
```

#### Get Server Info

```bash
grpcurl -plaintext localhost:7755 \
  industries.loosh.yutani.v1.SessionService/GetServerInfo
```

Response:
```json
{
  "version": "0.1.0",
  "activeSessions": 0,
  "maxSessions": 10,
  "screenSize": {
    "width": 120,
    "height": 40
  },
  "capabilities": {
    "mouseSupport": true,
    "pasteSupport": true,
    "trueColor": true,
    "supportedWidgets": ["BOX", "TEXT_VIEW", "LIST", "TABLE", ...]
  }
}
```

#### Create a Session

```bash
grpcurl -plaintext -d '{
  "client_name": "grpcurl-client",
  "preferences": {
    "receive_key_events": true,
    "receive_mouse_events": true,
    "receive_resize_events": true
  }
}' localhost:7755 industries.loosh.yutani.v1.SessionService/CreateSession
```

Response:
```json
{
  "sessionId": {
    "id": "abc-123-def-456"
  },
  "serverVersion": "0.1.0"
}
```

Save the session ID for subsequent commands:
```bash
export SESSION_ID="abc-123-def-456"
```

#### Destroy a Session

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.SessionService/DestroySession
```

### Screen Operations

#### Get Screen Size

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.ScreenService/GetSize
```

#### Clear Screen

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.ScreenService/Clear
```

#### Draw Text

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"position\": {\"x\": 10, \"y\": 5},
  \"text\": \"Hello from grpcurl!\",
  \"style\": {
    \"foreground\": {\"name\": \"yellow\"},
    \"attributes\": [\"ATTR_BOLD\"]
  }
}" localhost:7755 industries.loosh.yutani.v1.ScreenService/DrawText
```

#### Draw Box

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"top_left\": {\"x\": 5, \"y\": 3},
  \"bottom_right\": {\"x\": 50, \"y\": 15},
  \"style\": {
    \"foreground\": {\"name\": \"cyan\"}
  }
}" localhost:7755 industries.loosh.yutani.v1.ScreenService/DrawBox
```

#### Sync Screen

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.ScreenService/Sync
```

### Widget Operations

#### Create a TextView Widget

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"type\": \"WIDGET_TEXT_VIEW\",
  \"properties\": {
    \"border\": true,
    \"title\": \"My Text View\",
    \"type_properties\": {
      \"text_view\": {
        \"text\": \"Welcome to Yutani!\\n\\nThis is a text view widget.\",
        \"dynamic_colors\": true,
        \"word_wrap\": true
      }
    }
  }
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/CreateWidget
```

Response:
```json
{
  "widgetId": {
    "id": "widget-xyz-789"
  }
}
```

Save the widget ID:
```bash
export WIDGET_ID="widget-xyz-789"
```

#### Display a Widget (Set as Root)

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"widget_id\": {\"id\": \"$WIDGET_ID\"}
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/SetRoot
```

#### Create a List Widget

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"type\": \"WIDGET_LIST\",
  \"properties\": {
    \"border\": true,
    \"title\": \"Menu\"
  }
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/CreateWidget
```

#### Add Items to a List

```bash
export LIST_ID="widget-list-123"

grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"widget_id\": {\"id\": \"$LIST_ID\"},
  \"main_text\": \"New File\",
  \"secondary_text\": \"Create a new file\",
  \"shortcut\": \"n\"
}" localhost:7755 industries.loosh.yutani.v1.ListService/AddItem
```

#### List All Widgets

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.WidgetService/ListWidgets
```

#### Delete a Widget

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"widget_id\": {\"id\": \"$WIDGET_ID\"}
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/DeleteWidget
```

### Event Streaming

The EventService provides server-streaming RPC for receiving events from the server. This is useful for monitoring keyboard input, mouse events, and widget interactions.

#### Subscribe to Events (Streaming)

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"filter\": {
    \"receive_key_events\": true,
    \"receive_mouse_events\": true,
    \"receive_resize_events\": true,
    \"receive_focus_events\": true,
    \"receive_widget_events\": true
  }
}" localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe
```

This command will stream events continuously. Press `Ctrl+C` to stop.

Example output:
```json
{
  "sessionId": {"id": "abc-123-def-456"},
  "timestamp": "1704123456789",
  "key": {
    "key": "KEY_ENTER",
    "rune": 13,
    "modifiers": []
  }
}
{
  "sessionId": {"id": "abc-123-def-456"},
  "timestamp": "1704123456790",
  "mouse": {
    "x": 25,
    "y": 10,
    "button": "MOUSE_LEFT",
    "action": "MOUSE_CLICK"
  }
}
```

#### Subscribe to Specific Event Types

```bash
# Only keyboard events
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"filter\": {
    \"receive_key_events\": true,
    \"receive_mouse_events\": false,
    \"receive_resize_events\": false,
    \"receive_focus_events\": false,
    \"receive_widget_events\": false
  }
}" localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe
```

#### Inject a Test Event

In another terminal, you can inject synthetic events for testing:

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"event\": {
    \"session_id\": {\"id\": \"$SESSION_ID\"},
    \"timestamp\": 1234567890,
    \"key\": {
      \"key\": \"KEY_ENTER\",
      \"rune\": 13
    }
  }
}" localhost:7755 industries.loosh.yutani.v1.EventService/InjectEvent
```

### Advanced grpcurl Workflows

#### Complete Workflow: Create and Display a Widget

```bash
# 1. Create session
SESSION_RESP=$(grpcurl -plaintext -d '{"client_name": "demo"}' \
  localhost:7755 industries.loosh.yutani.v1.SessionService/CreateSession)
SESSION_ID=$(echo $SESSION_RESP | jq -r '.sessionId.id')
echo "Session ID: $SESSION_ID"

# 2. Create a text view widget
WIDGET_RESP=$(grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"type\": \"WIDGET_TEXT_VIEW\",
  \"properties\": {
    \"border\": true,
    \"title\": \"Demo\",
    \"type_properties\": {
      \"text_view\": {
        \"text\": \"Hello, World!\",
        \"word_wrap\": true
      }
    }
  }
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/CreateWidget)
WIDGET_ID=$(echo $WIDGET_RESP | jq -r '.widgetId.id')
echo "Widget ID: $WIDGET_ID"

# 3. Display the widget
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"widget_id\": {\"id\": \"$WIDGET_ID\"}
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/SetRoot

echo "Widget displayed! Check the server terminal."
```

#### Monitor Events While Interacting

```bash
# Terminal 1: Subscribe to events
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"filter\": {\"receive_key_events\": true}
}" localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe

# Terminal 2: Interact with the server terminal
# Press keys and watch them appear in Terminal 1
```

### grpcurl Tips

#### Using jq for JSON Processing

Install `jq` for easier JSON handling:
```bash
brew install jq  # macOS
apt-get install jq  # Linux
```

Extract specific fields:
```bash
grpcurl -plaintext localhost:7755 \
  industries.loosh.yutani.v1.SessionService/GetServerInfo | jq '.version'
```

#### Save Session ID Automatically

```bash
SESSION_ID=$(grpcurl -plaintext -d '{"client_name": "auto"}' \
  localhost:7755 industries.loosh.yutani.v1.SessionService/CreateSession | \
  jq -r '.sessionId.id')
```

#### Timeout for Streaming

Use shell timeout to limit streaming duration:
```bash
timeout 10s grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe
```

### grpcurl Troubleshooting

**Connection Refused**: Make sure the Yutani server is running (`./bin/yutani server`).

**Invalid Session ID**: Create a new session or verify your session ID is correct.

**Method Not Found**: List available services with `grpcurl -plaintext localhost:7755 list` and verify the method name with `describe`.

**Streaming Hangs**: This is expected behavior for streaming RPCs. Press `Ctrl+C` to stop, or use `timeout`.

## Quick Start: Debug Mode

The easiest way to debug the phase4-demo is to use the debug script:

```bash
./run-phase4-demo-debug.sh
```

This will:
1. Start the server with logging to `yutani-server.log`
2. Run the phase4-demo with logging to `phase4-demo.log`
3. Capture all events and operations

## Server-Side Logging

### Enable Server Logging

The Yutani server supports logging via the `--log-file` flag (or the `YUTANI_LOG_FILE` environment variable):

```bash
# Start server with logging
./bin/yutani server --log-file yutani-server.log

# With debug level logging
./bin/yutani server --log-file yutani-server.log --log-level debug

# Headless mode logs to stderr by default
./bin/yutani server --headless
```

### Log Levels

- `debug` - All events, including event dispatch details
- `info` - Important operations (session creation, widget creation, etc.)
- `warn` - Warnings (dropped events, etc.)
- `error` - Errors only

### What Gets Logged (Server)

Server logs include:
- Session creation/destruction
- Widget creation/deletion
- Event subscriptions
- Event dispatch (with debug level)
- Service operations (List, Table, Form, etc.)
- Errors and warnings

### Viewing Server Logs

```bash
# Follow logs in real-time
tail -f yutani-server.log

# View all event dispatches
grep "Event dispatch" yutani-server.log

# View warnings (like dropped events)
grep WARN yutani-server.log

# View errors
grep ERROR yutani-server.log
```

## Client-Side Logging

### phase4-demo Logging

The phase4-demo automatically logs to `phase4-demo.log` with detailed event information.

### What Gets Logged (Client)

Client logs include:
- Connection status
- Session creation
- Widget creation
- All received events with timestamps
- Event details (key, mouse, widget, etc.)

### Viewing Client Logs

```bash
# Follow logs in real-time
tail -f phase4-demo.log

# View all events
grep EVENT phase4-demo.log

# View only key events
grep "EVENT.*KEY" phase4-demo.log

# View only widget events
grep "EVENT.*WIDGET" phase4-demo.log

# Count events by type
grep EVENT phase4-demo.log | cut -d']' -f3 | cut -d':' -f1 | sort | uniq -c
```

## Common Issues and Solutions

### Issue: No Events Received

**Symptoms:**
- Client logs show "Event stream started" but no events
- Server logs show no "Event dispatch" messages

**Debugging:**
1. Check if event stream is connected:
   ```bash
   grep "Event stream started" phase4-demo.log
   ```

2. Check server-side subscription:
   ```bash
   grep "subscribing to events" yutani-server.log
   ```

3. Verify events are being captured:
   ```bash
   # Try pressing keys and check server logs
   grep "Event dispatch" yutani-server.log
   ```

**Solutions:**
- Ensure server is running before starting client
- Check that Subscribe RPC is being called (not StreamEvents)
- Verify event filter allows all event types

### Issue: Events Filtered Out

**Symptoms:**
- Some events received but not others
- Server logs show "Event dispatch: filtered out"

**Debugging:**
```bash
# Check what's being filtered
grep "filtered out" yutani-server.log
```

**Solutions:**
- Check event filter in Subscribe request
- Ensure all event types are enabled:
  ```go
  Filter: &pb.EventFilter{
      ReceiveKeyEvents:    true,
      ReceiveMouseEvents:  true,
      ReceiveResizeEvents: true,
      ReceiveFocusEvents:  true,
      ReceiveWidgetEvents: true,
  }
  ```

### Issue: Events Dropped

**Symptoms:**
- Server logs show "Event dropped: channel full"
- Some events missing from client

**Debugging:**
```bash
# Count dropped events
grep "Event dropped" yutani-server.log | wc -l
```

**Solutions:**
- Process events faster in client
- Reduce event generation rate
- Increase buffer size in EventDispatcher (default: 100)

### Issue: Focus Not Working

**Symptoms:**
- Cannot change focus between widgets
- No focus events received

**Debugging:**
1. Check if focus events are enabled:
   ```bash
   grep "FOCUS" phase4-demo.log
   ```

2. Check if SetFocus is being called:
   ```bash
   grep "SetFocus" yutani-server.log
   ```

**Solutions:**
- Ensure widgets are focusable
- Call SetFocus explicitly
- Check widget hierarchy (parent-child relationships)

## Event Flow Diagram

```
User Input (keyboard/mouse)
    |
    v
tview captures event
    |
    v
Server converts to protobuf Event
    |
    v
EventDispatcher.Dispatch()
    |
    v
Check filter
    |
    v
Send to subscriber channel
    |
    v
EventService.Subscribe stream
    |
    v
Client receives event
    |
    v
Client event handlers called
```

## Debugging Checklist

When events are not working:

- [ ] Server is running with logging enabled
- [ ] Client successfully connects and creates session
- [ ] Event stream is started (Subscribe called)
- [ ] Event filter allows all event types
- [ ] Server logs show "Event dispatch" messages
- [ ] Client logs show "EVENT" messages
- [ ] No "Event dropped" warnings in server logs
- [ ] No "filtered out" messages for expected events

## Advanced Debugging

### Enable Debug Logging in Code

Add temporary debug logging:

```go
// In client code
log.Printf("DEBUG: Received event type: %T", event.Event)

// In server code
slog.Debug("Custom debug message", "key", value)
```

### Monitor Event Rate

```bash
# Count events per second
watch -n 1 'grep EVENT phase4-demo.log | tail -100 | wc -l'
```

### Analyze Event Timing

```bash
# Show event timestamps
grep EVENT phase4-demo.log | cut -d'@' -f2 | cut -d']' -f1
```

## Getting Help

If you are still having issues:

1. Collect logs:
   ```bash
   tar czf yutani-debug.tar.gz yutani-server.log phase4-demo.log
   ```

2. Include:
   - What you are trying to do
   - What is happening vs. what you expect
   - Relevant log excerpts
   - Steps to reproduce

3. Check existing issues or create a new one with the debug information

## See Also

- [README.md](README.md) - Project overview
- [QUICKSTART.md](QUICKSTART.md) - Getting started guide
- [TUTORIAL.md](TUTORIAL.md) - Client library tutorial
- [examples/](examples/) - Example applications
- [grpcurl documentation](https://github.com/fullstorydev/grpcurl) - Official grpcurl docs
