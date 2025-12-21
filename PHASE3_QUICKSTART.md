# Phase 3 Quick Start Guide

## Running the Server

```bash
# Start the Yutani server
./bin/yutani-server

# Or with debug logging
./bin/yutani-server --log-level=debug
```

The server will start on port 7755 by default.

## Running the Test Client

In a separate terminal:

```bash
./bin/test-client
```

This will run through all Phase 1, 2, and 3 tests, demonstrating:
- Session management
- Screen operations
- Event streaming
- Widget creation and management

## Using grpcurl

### List Available Services

```bash
grpcurl -plaintext localhost:7755 list
```

### Create a Session

```bash
grpcurl -plaintext -d '{
  "client_name": "my-client",
  "preferences": {
    "receive_key_events": true,
    "receive_mouse_events": true
  }
}' localhost:7755 industries.loosh.yutani.v1.SessionService/CreateSession
```

### Create a TextView Widget

```bash
grpcurl -plaintext -d '{
  "session_id": {"id": "YOUR_SESSION_ID"},
  "type": "WIDGET_TEXT_VIEW",
  "properties": {
    "border": true,
    "title": "Hello World",
    "type_properties": {
      "text_view": {
        "text": "Welcome to Yutani!",
        "dynamic_colors": true,
        "word_wrap": true
      }
    }
  }
}' localhost:7755 industries.loosh.yutani.v1.WidgetService/CreateWidget
```

### Display a Widget

```bash
grpcurl -plaintext -d '{
  "session_id": {"id": "YOUR_SESSION_ID"},
  "widget_id": {"id": "YOUR_WIDGET_ID"}
}' localhost:7755 industries.loosh.yutani.v1.WidgetService/SetRoot
```

### List All Widgets

```bash
grpcurl -plaintext -d '{
  "session_id": {"id": "YOUR_SESSION_ID"}
}' localhost:7755 industries.loosh.yutani.v1.WidgetService/ListWidgets
```

## Widget Types

### Box
Basic container with border and title:
```json
{
  "type": "WIDGET_BOX",
  "properties": {
    "border": true,
    "title": "My Box",
    "background_color": {"r": 0, "g": 0, "b": 0}
  }
}
```

### TextView
Display text content:
```json
{
  "type": "WIDGET_TEXT_VIEW",
  "properties": {
    "border": true,
    "title": "Text Display",
    "type_properties": {
      "text_view": {
        "text": "Hello, World!",
        "dynamic_colors": true,
        "word_wrap": true,
        "scrollable": true
      }
    }
  }
}
```

### InputField
Single-line text input:
```json
{
  "type": "WIDGET_INPUT_FIELD",
  "properties": {
    "border": true,
    "title": "Input",
    "type_properties": {
      "input_field": {
        "label": "Name: ",
        "placeholder": "Enter your name",
        "field_width": 20
      }
    }
  }
}
```

### Button
Clickable button:
```json
{
  "type": "WIDGET_BUTTON",
  "properties": {
    "border": true,
    "type_properties": {
      "button": {
        "label": "Click Me!"
      }
    }
  }
}
```

### Checkbox
Toggle checkbox:
```json
{
  "type": "WIDGET_CHECKBOX",
  "properties": {
    "border": true,
    "type_properties": {
      "checkbox": {
        "label": "Enable feature",
        "checked": false
      }
    }
  }
}
```

## Widget Events

Subscribe to widget events:

```bash
grpcurl -plaintext -d '{
  "session_id": {"id": "YOUR_SESSION_ID"},
  "filter": {
    "receive_widget_events": true
  }
}' localhost:7755 industries.loosh.yutani.v1.EventService/Subscribe
```

Widget events include:
- **WIDGET_SUBMITTED** - Button pressed, Enter in InputField
- **WIDGET_CHANGED** - Checkbox toggled, InputField text changed
- **WIDGET_DONE** - InputField completed (Enter pressed)

## Common Operations

### Update Widget Properties

```bash
grpcurl -plaintext -d '{
  "session_id": {"id": "YOUR_SESSION_ID"},
  "widget_id": {"id": "YOUR_WIDGET_ID"},
  "properties": {
    "title": "Updated Title"
  }
}' localhost:7755 industries.loosh.yutani.v1.WidgetService/SetProperties
```

### Set Focus

```bash
grpcurl -plaintext -d '{
  "session_id": {"id": "YOUR_SESSION_ID"},
  "widget_id": {"id": "YOUR_WIDGET_ID"}
}' localhost:7755 industries.loosh.yutani.v1.WidgetService/SetFocus
```

### Delete Widget

```bash
grpcurl -plaintext -d '{
  "session_id": {"id": "YOUR_SESSION_ID"},
  "widget_id": {"id": "YOUR_WIDGET_ID"}
}' localhost:7755 industries.loosh.yutani.v1.WidgetService/DeleteWidget
```

## Next Steps

- See [PHASE3_COMPLETE.md](PHASE3_COMPLETE.md) for detailed documentation
- Check [PRD.md](PRD.md) for the complete API reference
- Look at [cmd/test-client/main.go](cmd/test-client/main.go) for Go client examples

