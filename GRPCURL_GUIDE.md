# Using grpcurl with Yutani

This guide shows how to interact with the Yutani Terminal Display Server using `grpcurl`, a command-line tool for gRPC services.

## Prerequisites

1. **Install grpcurl**:
   ```bash
   go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
   ```

2. **Start the Yutani server**:
   ```bash
   ./bin/yutani-server
   ```
   The server runs on `localhost:7755` by default with gRPC reflection enabled.

## Basic Commands

### List Available Services

```bash
grpcurl -plaintext localhost:7755 list
```

**Output**:
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

### Describe a Service

```bash
grpcurl -plaintext localhost:7755 describe industries.loosh.yutani.v1.SessionService
```

### Describe a Method

```bash
grpcurl -plaintext localhost:7755 describe industries.loosh.yutani.v1.SessionService.CreateSession
```

## Session Management

### Ping the Server

```bash
grpcurl -plaintext -d '{"timestamp": 1234567890}' \
  localhost:7755 industries.loosh.yutani.v1.SessionService/Ping
```

**Response**:
```json
{
  "timestamp": "1234567890",
  "serverTimestamp": "1704123456789"
}
```

### Get Server Info

```bash
grpcurl -plaintext localhost:7755 \
  industries.loosh.yutani.v1.SessionService/GetServerInfo
```

**Response**:
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

### Create a Session

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

**Response**:
```json
{
  "sessionId": {
    "id": "abc-123-def-456"
  },
  "serverVersion": "0.1.0"
}
```

**Save the session ID** - you'll need it for subsequent commands. For convenience:
```bash
export SESSION_ID="abc-123-def-456"
```

### Destroy a Session

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.SessionService/DestroySession
```

## Screen Operations

### Get Screen Size

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.ScreenService/GetSize
```

### Clear Screen

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.ScreenService/Clear
```

### Draw Text

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

### Draw Box

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

### Sync Screen

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.ScreenService/Sync
```

## Widget Operations

### Create a TextView Widget

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

**Response**:
```json
{
  "widgetId": {
    "id": "widget-xyz-789"
  }
}
```

**Save the widget ID**:
```bash
export WIDGET_ID="widget-xyz-789"
```

### Display a Widget (Set as Root)

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"widget_id\": {\"id\": \"$WIDGET_ID\"}
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/SetRoot
```

### Create a List Widget

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

### Add Items to List

```bash
# Save the list widget ID first
export LIST_ID="widget-list-123"

grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"widget_id\": {\"id\": \"$LIST_ID\"},
  \"main_text\": \"New File\",
  \"secondary_text\": \"Create a new file\",
  \"shortcut\": \"n\"
}" localhost:7755 industries.loosh.yutani.v1.ListService/AddItem
```

### List All Widgets

```bash
grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.WidgetService/ListWidgets
```

### Delete a Widget

```bash
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"widget_id\": {\"id\": \"$WIDGET_ID\"}
}" localhost:7755 industries.loosh.yutani.v1.WidgetService/DeleteWidget
```

## Event Streaming

The EventService provides server-streaming RPC for receiving events from the server. This is useful for monitoring keyboard input, mouse events, and widget interactions.

### Subscribe to Events (Streaming)

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

**This command will stream events continuously**. Press `Ctrl+C` to stop.

**Example output**:
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

### Subscribe to Specific Event Types

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

### Inject a Test Event

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


## Advanced Examples

### Complete Workflow: Create and Display a Widget

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

### Monitor Events While Interacting

```bash
# Terminal 1: Subscribe to events
grpcurl -plaintext -d "{
  \"session_id\": {\"id\": \"$SESSION_ID\"},
  \"filter\": {\"receive_key_events\": true}
}" localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe

# Terminal 2: Interact with the server terminal
# Press keys and watch them appear in Terminal 1
```

## Tips and Tricks

### Using jq for JSON Processing

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

### Save Session ID Automatically

```bash
SESSION_ID=$(grpcurl -plaintext -d '{"client_name": "auto"}' \
  localhost:7755 industries.loosh.yutani.v1.SessionService/CreateSession | \
  jq -r '.sessionId.id')
```

### Pretty Print Responses

grpcurl automatically pretty-prints JSON, but you can pipe through jq for more control:
```bash
grpcurl -plaintext localhost:7755 \
  industries.loosh.yutani.v1.SessionService/GetServerInfo | jq '.'
```

### Timeout for Streaming

Use shell timeout to limit streaming duration:
```bash
timeout 10s grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe
```

## Troubleshooting

### Connection Refused

**Problem**: `Failed to dial target host "localhost:7755": dial tcp [::1]:7755: connect: connection refused`

**Solution**: Make sure the Yutani server is running:
```bash
./bin/yutani-server
```

### Invalid Session ID

**Problem**: `Session not found` or `Invalid session ID`

**Solution**: Create a new session or verify your session ID is correct:
```bash
grpcurl -plaintext -d '{"client_name": "test"}' \
  localhost:7755 industries.loosh.yutani.v1.SessionService/CreateSession
```

### Method Not Found

**Problem**: `unknown service` or `unknown method`

**Solution**: List available services and verify the method name:
```bash
grpcurl -plaintext localhost:7755 list
grpcurl -plaintext localhost:7755 describe industries.loosh.yutani.v1.SessionService
```

### Streaming Hangs

**Problem**: Event streaming command doesn't return

**Solution**: This is expected behavior - streaming RPCs stay open. Press `Ctrl+C` to stop, or use `timeout`:
```bash
timeout 5s grpcurl -plaintext -d "{\"session_id\": {\"id\": \"$SESSION_ID\"}}" \
  localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe
```

## See Also

- [README.md](README.md) - Project overview
- [QUICKSTART.md](QUICKSTART.md) - Getting started guide
- [TUTORIAL.md](TUTORIAL.md) - Client library tutorial
- [pkg/client/README.md](pkg/client/README.md) - Go client library documentation
- [PRD.md](PRD.md) - Complete API reference
- [grpcurl documentation](https://github.com/fullstorydev/grpcurl) - Official grpcurl docs


