# Yutani - Terminal Display Server
## Product Requirements Document

---

## 1. Executive Summary

**Yutani** is a Go-based terminal display server that provides networked, widget-based windowing capabilities for text-mode applications. Inspired by TWIN, Yutani uses gRPC for client-server communication and leverages tcell/tview for the underlying TUI rendering. Clients can control the terminal UI through both low-level cell operations and high-level widget abstractions.

### Goals
- Provide a networked terminal UI server accessible via gRPC
- Support both low-level drawing (cells, styles) and high-level widgets (forms, lists, tables)
- Enable multiple clients to interact with and control the UI
- Expose all services via gRPC reflection for easy introspection and tooling

### Non-Goals (Initial Release)
- Multi-display/multi-head support
- Built-in terminal emulator (clients handle their own PTY if needed)
- Compression (gRPC handles this at transport level)

---

## 2. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         YUTANI SERVER                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────┐  │
│  │ gRPC Server  │◄──►│ Widget       │◄──►│ tview/tcell      │  │
│  │              │    │ Registry     │    │ Application      │  │
│  │ - Screen     │    │              │    │                  │  │
│  │ - Widget     │    │ ID → Widget  │    │ - Event Loop     │  │
│  │ - Event      │    │ mapping      │    │ - Rendering      │  │
│  │ - Session    │    │              │    │ - Input          │  │
│  └──────────────┘    └──────────────┘    └──────────────────┘  │
│         ▲                                         │             │
│         │                                         ▼             │
│         │                                 ┌──────────────┐      │
│         │                                 │   Terminal   │      │
│         │                                 │   (stdout)   │      │
└─────────┼─────────────────────────────────┴──────────────┴──────┘
          │
          │ gRPC (TCP)
          │
┌─────────┴─────────┐
│   YUTANI CLIENT   │
│                   │
│ - Create widgets  │
│ - Handle events   │
│ - Draw content    │
└───────────────────┘
```

---

## 3. Core Concepts

### 3.1 Sessions
A **Session** represents a client connection to the server. Each session:
- Has a unique session ID
- Owns widgets it creates
- Receives events for its widgets
- Can be configured with preferences (e.g., event filtering)

### 3.2 Widgets
Widgets are UI elements with unique IDs. The server maintains a registry mapping IDs to tview primitives:

| Widget Type | tview Primitive | Description |
|-------------|-----------------|-------------|
| Box | Box | Base container with border/title |
| TextView | TextView | Scrollable text display |
| InputField | InputField | Single-line text input |
| TextArea | TextArea | Multi-line text editor |
| Button | Button | Clickable button |
| Checkbox | Checkbox | Boolean toggle |
| DropDown | DropDown | Selection dropdown |
| List | List | Navigable item list |
| Table | Table | 2D tabular data |
| TreeView | TreeView | Hierarchical tree |
| Form | Form | Form container |
| Flex | Flex | Flexbox layout |
| Grid | Grid | Grid layout |
| Pages | Pages | Stacked pages/tabs |
| Modal | Modal | Dialog overlay |
| Image | Image | Terminal image display |
| ProgressBar | ProgressBar | Progress indicator |

### 3.3 Events
Events flow from server to client via gRPC streaming:
- **KeyEvent** - Keyboard input (key, modifiers, rune)
- **MouseEvent** - Mouse actions (click, drag, scroll, position)
- **ResizeEvent** - Terminal resize
- **FocusEvent** - Widget focus changes
- **WidgetEvent** - Widget-specific (selection, change, submit)

### 3.4 Styles
Styling uses tcell's color and attribute system:
- Foreground/background colors (named, 256-color, true-color)
- Attributes: bold, italic, underline, reverse, blink, dim, strikethrough

---

## 4. gRPC Services

### 4.1 SessionService
Manages client sessions and server lifecycle.

```protobuf
service SessionService {
  // Create a new session
  rpc CreateSession(CreateSessionRequest) returns (CreateSessionResponse);

  // End a session and cleanup its resources
  rpc DestroySession(DestroySessionRequest) returns (DestroySessionResponse);

  // Get server information
  rpc GetServerInfo(GetServerInfoRequest) returns (GetServerInfoResponse);

  // Ping for health check
  rpc Ping(PingRequest) returns (PingResponse);
}
```

### 4.2 ScreenService
Low-level screen/cell operations (tcell layer).

```protobuf
service ScreenService {
  // Get screen dimensions
  rpc GetSize(GetSizeRequest) returns (GetSizeResponse);

  // Set a single cell
  rpc SetCell(SetCellRequest) returns (SetCellResponse);

  // Set multiple cells (batch)
  rpc SetCells(SetCellsRequest) returns (SetCellsResponse);

  // Fill a region with a cell
  rpc Fill(FillRequest) returns (FillResponse);

  // Clear the screen or region
  rpc Clear(ClearRequest) returns (ClearResponse);

  // Draw text at position
  rpc DrawText(DrawTextRequest) returns (DrawTextResponse);

  // Draw a box/border
  rpc DrawBox(DrawBoxRequest) returns (DrawBoxResponse);

  // Force screen sync/refresh
  rpc Sync(SyncRequest) returns (SyncResponse);

  // Get cell content at position
  rpc GetCell(GetCellRequest) returns (GetCellResponse);
}
```

### 4.3 WidgetService
High-level widget operations (tview layer).

```protobuf
service WidgetService {
  // Create a widget
  rpc CreateWidget(CreateWidgetRequest) returns (CreateWidgetResponse);

  // Delete a widget
  rpc DeleteWidget(DeleteWidgetRequest) returns (DeleteWidgetResponse);

  // Set widget properties
  rpc SetProperties(SetPropertiesRequest) returns (SetPropertiesResponse);

  // Get widget properties
  rpc GetProperties(GetPropertiesRequest) returns (GetPropertiesResponse);

  // Set widget as root (display it)
  rpc SetRoot(SetRootRequest) returns (SetRootResponse);

  // Set focus to widget
  rpc SetFocus(SetFocusRequest) returns (SetFocusResponse);

  // Get currently focused widget
  rpc GetFocus(GetFocusRequest) returns (GetFocusResponse);

  // List all widgets for session
  rpc ListWidgets(ListWidgetsRequest) returns (ListWidgetsResponse);
}
```

### 4.4 EventService
Event streaming from server to client.

```protobuf
service EventService {
  // Subscribe to events (server-streaming)
  rpc Subscribe(SubscribeRequest) returns (stream Event);

  // Inject a synthetic event (for testing)
  rpc InjectEvent(InjectEventRequest) returns (InjectEventResponse);

  // Configure event filtering
  rpc SetEventFilter(SetEventFilterRequest) returns (SetEventFilterResponse);
}
```

### 4.5 TextViewService
Operations specific to TextView widgets.

```protobuf
service TextViewService {
  // Set text content
  rpc SetText(SetTextRequest) returns (SetTextResponse);

  // Append text (streaming input)
  rpc Write(stream WriteRequest) returns (WriteResponse);

  // Get current text
  rpc GetText(GetTextRequest) returns (GetTextResponse);

  // Scroll to position
  rpc ScrollTo(ScrollToRequest) returns (ScrollToResponse);

  // Set/get regions and highlights
  rpc SetHighlight(SetHighlightRequest) returns (SetHighlightResponse);

  // Clear content
  rpc Clear(ClearRequest) returns (ClearResponse);
}
```

### 4.6 ListService
Operations specific to List widgets.

```protobuf
service ListService {
  // Add item to list
  rpc AddItem(AddItemRequest) returns (AddItemResponse);

  // Remove item
  rpc RemoveItem(RemoveItemRequest) returns (RemoveItemResponse);

  // Clear all items
  rpc Clear(ClearRequest) returns (ClearResponse);

  // Get item count
  rpc GetItemCount(GetItemCountRequest) returns (GetItemCountResponse);

  // Get/set selection
  rpc GetSelection(ListGetSelectionRequest) returns (ListGetSelectionResponse);
  rpc SetSelection(ListSetSelectionRequest) returns (ListSetSelectionResponse);

  // Get item text
  rpc GetItem(ListGetItemRequest) returns (ListGetItemResponse);
}
```

### 4.7 TableService
Operations specific to Table widgets.

```protobuf
service TableService {
  // Set cell content
  rpc SetCell(SetCellRequest) returns (SetCellResponse);

  // Get cell content
  rpc GetCell(GetCellRequest) returns (GetCellResponse);

  // Set multiple cells (batch)
  rpc SetCells(SetCellsRequest) returns (SetCellsResponse);

  // Clear table
  rpc Clear(ClearRequest) returns (ClearResponse);

  // Get row/column count
  rpc GetDimensions(GetDimensionsRequest) returns (GetDimensionsResponse);

  // Get/set selection
  rpc GetSelection(GetSelectionRequest) returns (GetSelectionResponse);
  rpc SetSelection(SetSelectionRequest) returns (SetSelectionResponse);

  // Configure fixed rows/columns (headers)
  rpc SetFixed(SetFixedRequest) returns (SetFixedResponse);
}
```

### 4.8 FormService
Operations specific to Form widgets.

```protobuf
service FormService {
  // Add form field
  rpc AddField(AddFieldRequest) returns (AddFieldResponse);

  // Add button
  rpc AddButton(AddButtonRequest) returns (AddButtonResponse);

  // Get field value
  rpc GetFieldValue(GetFieldValueRequest) returns (GetFieldValueResponse);

  // Set field value
  rpc SetFieldValue(SetFieldValueRequest) returns (SetFieldValueResponse);

  // Clear form
  rpc Clear(ClearRequest) returns (ClearResponse);

  // Get form item count
  rpc GetItemCount(GetItemCountRequest) returns (GetItemCountResponse);
}
```

### 4.9 TreeService
Operations specific to TreeView widgets.

```protobuf
service TreeService {
  // Set root node
  rpc SetRoot(SetRootRequest) returns (SetRootResponse);

  // Add child node
  rpc AddChild(AddChildRequest) returns (AddChildResponse);

  // Remove node
  rpc RemoveNode(RemoveNodeRequest) returns (RemoveNodeResponse);

  // Expand/collapse node
  rpc SetExpanded(SetExpandedRequest) returns (SetExpandedResponse);

  // Get/set selection
  rpc GetSelection(TreeGetSelectionRequest) returns (TreeGetSelectionResponse);
  rpc SetSelection(TreeSetSelectionRequest) returns (TreeSetSelectionResponse);

  // Get node children
  rpc GetChildren(TreeGetChildrenRequest) returns (TreeGetChildrenResponse);
}
```

### 4.10 LayoutService
Operations for layout containers (Flex, Grid, Pages).

```protobuf
service LayoutService {
  // Flex operations
  rpc FlexAddItem(FlexAddItemRequest) returns (FlexAddItemResponse);
  rpc FlexRemoveItem(FlexRemoveItemRequest) returns (FlexRemoveItemResponse);
  rpc FlexSetDirection(FlexSetDirectionRequest) returns (FlexSetDirectionResponse);

  // Grid operations
  rpc GridAddItem(GridAddItemRequest) returns (GridAddItemResponse);
  rpc GridRemoveItem(GridRemoveItemRequest) returns (GridRemoveItemResponse);
  rpc GridSetRows(GridSetRowsRequest) returns (GridSetRowsResponse);
  rpc GridSetColumns(GridSetColumnsRequest) returns (GridSetColumnsResponse);

  // Pages operations
  rpc PagesAddPage(PagesAddPageRequest) returns (PagesAddPageResponse);
  rpc PagesRemovePage(PagesRemovePageRequest) returns (PagesRemovePageResponse);
  rpc PagesSwitchTo(PagesSwitchToRequest) returns (PagesSwitchToResponse);
  rpc PagesGetCurrent(PagesGetCurrentRequest) returns (PagesGetCurrentResponse);
}
```

---

## 5. Message Types

### 5.1 Core Types

```protobuf
message WidgetId {
  string id = 1;
}

message SessionId {
  string id = 1;
}

message Position {
  int32 x = 1;
  int32 y = 2;
}

message Size {
  int32 width = 1;
  int32 height = 2;
}

message Rect {
  int32 x = 1;
  int32 y = 2;
  int32 width = 3;
  int32 height = 4;
}

message Color {
  oneof color {
    string name = 1;      // Named color: "red", "blue", etc.
    int32 index = 2;      // 256-color palette index
    RGB rgb = 3;          // True color
  }
}

message RGB {
  int32 r = 1;
  int32 g = 2;
  int32 b = 3;
}

message Style {
  Color foreground = 1;
  Color background = 2;
  repeated Attribute attributes = 3;
}

enum Attribute {
  ATTR_NONE = 0;
  ATTR_BOLD = 1;
  ATTR_ITALIC = 2;
  ATTR_UNDERLINE = 3;
  ATTR_REVERSE = 4;
  ATTR_BLINK = 5;
  ATTR_DIM = 6;
  ATTR_STRIKETHROUGH = 7;
}

message Cell {
  string rune = 1;        // UTF-8 character
  Style style = 2;
}
```

### 5.2 Widget Types

```protobuf
enum WidgetType {
  WIDGET_BOX = 0;
  WIDGET_TEXT_VIEW = 1;
  WIDGET_INPUT_FIELD = 2;
  WIDGET_TEXT_AREA = 3;
  WIDGET_BUTTON = 4;
  WIDGET_CHECKBOX = 5;
  WIDGET_DROPDOWN = 6;
  WIDGET_LIST = 7;
  WIDGET_TABLE = 8;
  WIDGET_TREE_VIEW = 9;
  WIDGET_FORM = 10;
  WIDGET_FLEX = 11;
  WIDGET_GRID = 12;
  WIDGET_PAGES = 13;
  WIDGET_MODAL = 14;
  WIDGET_IMAGE = 15;
}

message WidgetProperties {
  // Common properties (from Box)
  optional Rect rect = 1;
  optional bool border = 2;
  optional string title = 3;
  optional Alignment title_align = 4;
  optional Color background_color = 5;
  optional Color border_color = 6;
  optional Color title_color = 7;
  optional Padding padding = 8;

  // Type-specific properties
  oneof type_properties {
    TextViewProperties text_view = 20;
    InputFieldProperties input_field = 21;
    ButtonProperties button = 22;
    CheckboxProperties checkbox = 23;
    ListProperties list = 24;
    TableProperties table = 25;
    FlexProperties flex = 26;
    GridProperties grid = 27;
  }
}

message Padding {
  int32 top = 1;
  int32 bottom = 2;
  int32 left = 3;
  int32 right = 4;
}

enum Alignment {
  ALIGN_LEFT = 0;
  ALIGN_CENTER = 1;
  ALIGN_RIGHT = 2;
}
```

### 5.3 Event Types

```protobuf
message Event {
  SessionId session_id = 1;
  int64 timestamp = 2;

  oneof event {
    KeyEvent key = 10;
    MouseEvent mouse = 11;
    ResizeEvent resize = 12;
    FocusEvent focus = 13;
    WidgetEvent widget = 14;
  }
}

message KeyEvent {
  WidgetId widget_id = 1;
  Key key = 2;
  string rune = 3;           // UTF-8 character if printable
  repeated Modifier modifiers = 4;
}

enum Key {
  KEY_RUNE = 0;              // Regular character (check rune field)
  KEY_UP = 1;
  KEY_DOWN = 2;
  KEY_LEFT = 3;
  KEY_RIGHT = 4;
  KEY_ENTER = 5;
  KEY_ESCAPE = 6;
  KEY_TAB = 7;
  KEY_BACKTAB = 8;
  KEY_BACKSPACE = 9;
  KEY_DELETE = 10;
  KEY_INSERT = 11;
  KEY_HOME = 12;
  KEY_END = 13;
  KEY_PGUP = 14;
  KEY_PGDN = 15;
  KEY_F1 = 16;
  KEY_F2 = 17;
  // ... F3-F12
}

enum Modifier {
  MOD_NONE = 0;
  MOD_SHIFT = 1;
  MOD_CTRL = 2;
  MOD_ALT = 3;
  MOD_META = 4;
}

message MouseEvent {
  WidgetId widget_id = 1;
  MouseAction action = 2;
  Position position = 3;       // Relative to widget
  Position screen_position = 4; // Absolute screen position
  repeated Modifier modifiers = 5;
}

enum MouseAction {
  MOUSE_CLICK = 0;
  MOUSE_DOUBLE_CLICK = 1;
  MOUSE_RIGHT_CLICK = 2;
  MOUSE_MIDDLE_CLICK = 3;
  MOUSE_SCROLL_UP = 4;
  MOUSE_SCROLL_DOWN = 5;
  MOUSE_DRAG = 6;
  MOUSE_RELEASE = 7;
  MOUSE_MOVE = 8;
}

message ResizeEvent {
  Size new_size = 1;
}

message FocusEvent {
  WidgetId old_widget = 1;
  WidgetId new_widget = 2;
}

message WidgetEvent {
  WidgetId widget_id = 1;
  WidgetEventType type = 2;
  map<string, string> data = 3;  // Event-specific key-value data
}

enum WidgetEventType {
  WIDGET_SELECTED = 0;      // Item selected (List, Table, Tree)
  WIDGET_CHANGED = 1;       // Value changed (Input, Checkbox)
  WIDGET_SUBMITTED = 2;     // Form submitted, button pressed
  WIDGET_CANCELLED = 3;     // Escape pressed
  WIDGET_DONE = 4;          // Input complete (Enter pressed)
}
```

---

## 6. Server Implementation

### 6.1 Component Structure

```
yutani/
├── cmd/
│   ├── yutani-server/
│   │   └── main.go           # Server entry point
│   ├── test-client/
│   │   └── main.go           # Test client
│   └── phase4-demo/
│       └── main.go           # Phase 4 demo
├── pkg/
│   ├── server/
│   │   ├── server.go         # Core server (tview app, screen)
│   │   ├── session.go        # Session management
│   │   ├── registry.go       # Widget registry
│   │   ├── events.go         # Event dispatcher
│   │   └── event_convert.go  # tcell-to-protobuf conversion
│   ├── services/
│   │   ├── session.go        # SessionService impl
│   │   ├── screen.go         # ScreenService impl
│   │   ├── widget.go         # WidgetService impl
│   │   ├── widget_factory.go # Widget creation and properties
│   │   ├── event.go          # EventService impl
│   │   ├── list.go           # ListService impl
│   │   ├── table.go          # TableService impl
│   │   ├── form.go           # FormService impl
│   │   ├── tree.go           # TreeService impl
│   │   ├── layout.go         # LayoutService impl
│   │   ├── debug.go          # DebugService impl
│   │   └── test.go           # TestService impl
│   ├── client/
│   │   ├── client.go         # Go client library
│   │   └── testing/          # Integration test helpers
│   ├── cli/
│   │   ├── root.go           # CLI entry point (cobra)
│   │   ├── debug.go          # Debug commands
│   │   ├── session.go        # Session commands
│   │   └── ...               # Other CLI commands
│   ├── config/               # Configuration loading
│   ├── testutil/             # Shared test utilities
│   └── proto/
│       └── yutani/           # Generated Go code
├── api/
│   └── proto/
│       └── industries/
│           └── loosh/
│               └── yutani/
│                   └── v1/
│                       ├── session.proto
│                       ├── screen.proto
│                       ├── widget.proto
│                       ├── event.proto
│                       ├── list.proto
│                       ├── table.proto
│                       ├── form.proto
│                       ├── tree.proto
│                       ├── layout.proto
│                       ├── debug.proto
│                       ├── test.proto
│                       └── types.proto
├── test/
│   └── e2e/                  # E2E, acceptance, and contract tests
├── examples/                 # 8 example applications
├── go.mod
└── go.sum
```

### 6.2 Thread Safety

tview is not thread-safe. All gRPC handlers must:
1. Queue updates via `app.QueueUpdate()` or `app.QueueUpdateDraw()`
2. Use channels to communicate with the main event loop
3. Never directly modify tview primitives from gRPC goroutines

```go
// Example: Safe widget property update
func (s *WidgetService) SetProperties(ctx context.Context, req *pb.SetPropertiesRequest) (*pb.SetPropertiesResponse, error) {
    done := make(chan error, 1)

    s.app.QueueUpdate(func() {
        widget, ok := s.registry.Get(req.WidgetId)
        if !ok {
            done <- ErrWidgetNotFound
            return
        }

        if err := applyProperties(widget, req.Properties); err != nil {
            done <- err
            return
        }
        done <- nil
    })

    if err := <-done; err != nil {
        return nil, err
    }

    return &pb.SetPropertiesResponse{Success: true}, nil
}
```

### 6.3 Event Dispatch

Events captured by tview must be forwarded to subscribed clients:

```go
type EventDispatcher struct {
    subscribers map[string]chan *pb.Event  // sessionId -> event channel
    mu          sync.RWMutex
}

func (d *EventDispatcher) Dispatch(sessionId string, event *pb.Event) {
    d.mu.RLock()
    defer d.mu.RUnlock()

    if ch, ok := d.subscribers[sessionId]; ok {
        select {
        case ch <- event:
        default:
            // Channel full, drop event or log warning
        }
    }
}
```

### 6.4 Widget Registry

```go
type WidgetRegistry struct {
    widgets   map[string]tview.Primitive  // widgetId -> primitive
    owners    map[string]string           // widgetId -> sessionId
    mu        sync.RWMutex
}

func (r *WidgetRegistry) Register(sessionId string, widget tview.Primitive) string {
    r.mu.Lock()
    defer r.mu.Unlock()

    id := uuid.New().String()
    r.widgets[id] = widget
    r.owners[id] = sessionId
    return id
}

func (r *WidgetRegistry) Get(id string) (tview.Primitive, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.widgets[id]
}

func (r *WidgetRegistry) DeleteBySession(sessionId string) {
    r.mu.Lock()
    defer r.mu.Unlock()

    for id, owner := range r.owners {
        if owner == sessionId {
            delete(r.widgets, id)
            delete(r.owners, id)
        }
    }
}
```

---

## 7. Client Library

A Go client library will be provided for convenient server interaction:

```go
package yutani

type Client struct {
    conn      *grpc.ClientConn
    session   pb.SessionServiceClient
    screen    pb.ScreenServiceClient
    widget    pb.WidgetServiceClient
    event     pb.EventServiceClient
    // ... other service clients

    sessionId string
    events    <-chan *pb.Event
}

// Connect to Yutani server
func Connect(address string) (*Client, error)

// Close connection
func (c *Client) Close() error

// Create widgets
func (c *Client) NewTextView() (*TextView, error)
func (c *Client) NewList() (*List, error)
func (c *Client) NewTable() (*Table, error)
func (c *Client) NewForm() (*Form, error)
func (c *Client) NewFlex() (*Flex, error)
// ... etc

// Event handling
func (c *Client) Events() <-chan *pb.Event

// Screen operations
func (c *Client) ScreenSize() (width, height int, err error)
func (c *Client) SetCell(x, y int, r rune, style Style) error
func (c *Client) DrawText(x, y int, text string, style Style) error
```

---

## 8. gRPC Reflection

All services will have gRPC reflection enabled:

```go
import "google.golang.org/grpc/reflection"

func main() {
    server := grpc.NewServer()

    // Register services
    pb.RegisterSessionServiceServer(server, &sessionService{})
    pb.RegisterScreenServiceServer(server, &screenService{})
    pb.RegisterWidgetServiceServer(server, &widgetService{})
    // ... etc

    // Enable reflection
    reflection.Register(server)

    server.Serve(listener)
}
```

This allows tools like `grpcurl` and `grpcui` to introspect and interact with the server without pre-compiled stubs.

---

## 9. Configuration

Server configuration follows a three-tier override system:
1. `.yutani.conf` file (lowest priority)
2. Environment variables (middle priority)
3. Command-line flags (highest priority)

Configuration file format (`.yutani.conf`):

```conf
# Server settings
YUTANI_ADDRESS=:7755
YUTANI_MAX_SESSIONS=100

# TUI settings
YUTANI_MOUSE=true
YUTANI_PASTE=true

# Logging settings
YUTANI_LOG_LEVEL=info
```

Command-line flags:
```bash
yutani-server \
  --address=:7755 \
  --max-sessions=100 \
  --mouse=true \
  --paste=true \
  --log-level=info
```

Environment variables use the `YUTANI_` prefix and match the flag names in uppercase.

---

## 10. Security Considerations

### Initial Release
- No authentication (local use assumed)
- Session IDs provide basic isolation

### Future Considerations
- TLS for encrypted connections
- Token-based authentication
- Per-session permissions/capabilities
- Rate limiting

---

## 11. Success Metrics

- Widget creation latency < 5ms
- Event delivery latency < 10ms
- Support 10+ concurrent sessions
- Full tview widget coverage
- Clean session cleanup on disconnect

---

## 12. Implementation Phases

### Phase 1: Foundation ✅ **COMPLETE**
- [x] Project structure and build system
- [x] Proto definitions for core types
- [x] SessionService implementation
- [x] Basic ScreenService (size, clear, sync)
- [x] gRPC server with reflection

### Phase 2: Low-Level API ✅ **COMPLETE**
- [x] Complete ScreenService (cells, text, box drawing)
- [x] EventService with key/mouse/resize events
- [x] Event streaming and filtering

### Phase 3: Widget System ✅ **COMPLETE**
- [x] WidgetService core operations
- [x] Box, TextView, InputField widgets
- [x] Button, Checkbox widgets
- [x] Focus management
- [x] Widget hierarchy infrastructure
- [x] Widget event emission

### Phase 4: Complex Widgets ✅ **COMPLETE**
- [x] List, Table, TreeView services
- [x] Form service
- [x] Layout services (Flex, Grid, Pages)
- [x] All 32 RPCs across 5 services
- [x] Comprehensive E2E tests

### Phase 5: Client Library, Documentation, and Examples ✅ **COMPLETE**
- [x] Go client library with fluent API
- [x] Widget builders (Box, TextView, List, Table, Form)
- [x] Event handling with callbacks
- [x] 3 complete example applications
- [x] Comprehensive tutorial (5 lessons)
- [x] Full API documentation

### Phase 6: Advanced Features and Optimization (Future)

#### 6.1 Additional Widget Builders ✅ **COMPLETE**
- [x] TreeView widget builder with fluent API
- [x] Flex layout widget builder
- [x] Grid layout widget builder
- [x] Pages layout widget builder
- [x] InputField and Button widget builders
- [x] Checkbox widget builder

#### 6.2 Advanced Event Handling ✅ **COMPLETE**
- [x] Event filtering by widget ID or type
- [x] Event middleware/interceptors
- [x] Event batching for high-frequency events
- [x] Custom event types and handlers
- [x] Event replay/history for debugging

#### 6.3 Connection Management ✅ **COMPLETE**
- [x] Automatic reconnection on disconnect
- [x] Connection pooling for multiple concurrent clients
- [x] Connection health checks and monitoring
- [x] Graceful degradation on connection loss
- [x] Connection state callbacks

#### 6.4 Performance Optimization
- [ ] Benchmarking suite for all operations
- [ ] Profiling and performance analysis
- [ ] Batch operation optimizations
- [ ] Memory usage optimization
- [ ] Rendering performance improvements

#### 6.5 Additional Examples ✅ **COMPLETE**
- [x] File browser application
- [x] System dashboard with real-time metrics
- [x] Chat application with multiple users
- [x] Text editor with syntax highlighting
- [x] Process monitor/task manager

#### 6.6 Testing Utilities ✅ **COMPLETE**
- [x] Mock client for unit testing
- [x] Test helpers for common scenarios
- [x] Integration test framework
- [ ] Performance regression tests

#### 6.7 Developer Tools ✅ **COMPLETE**
- [x] CLI tool for quick prototyping
- [x] Widget inspector/debugger
- [x] Event monitor/logger
- [x] Performance profiler

#### 6.8 Additional Features
- [ ] Widget templates/presets
- [ ] Theme system for consistent styling
- [ ] Modal dialog support
- [ ] Progress bar widget
- [ ] Notification/toast system
- [ ] Context menu support

#### 6.9 Documentation Enhancements
- [ ] Video tutorials
- [ ] Interactive examples
- [ ] Architecture deep-dive
- [ ] Performance tuning guide
- [ ] Cookbook with common recipes

#### 6.10 Community and Ecosystem
- [ ] Plugin system for custom widgets
- [ ] Widget marketplace/registry
- [ ] Community examples repository
- [ ] Contributing guidelines

---

## 13. Implementation Decisions

1. **Multi-client widget sharing**: Widgets are strictly owned by their creator session.

2. **Custom drawing**: Clients can batch calls to set cell values and attributes, or to draw lines, rectangles, fill rectangles, erase, etc. This is not a goal for MVP.

3. **Widget-level event filtering**: Clients can subscribe to events for specific widgets only.

4. **State synchronization**: Clients can subscribe to state changes for widgets and receive updates.

5. **Error handling**: Keep it simple - per-operation errors. More granular error reporting can be added later if needed.

6. **Proto package naming**: Use `industries.loosh.yutani.v1` for protobuf package namespace. All messages follow `{Service}{Action}Request/Response` naming (standardized in Phase 7 quality initiative).

7. **Default port**: gRPC server listens on `:7755` by default.

8. **Configuration**: Three-tier system: `.yutani.conf` file → environment variables → command-line flags.

9. **Logging**: Use Go's standard `log/slog` package.

10. **Testing**: Unit tests implemented in Phase 2 and beyond, focused on business logic.

11. **Widget Event Emission**: Interactive widgets (Button, Checkbox, InputField) automatically emit events through the EventDispatcher when user interactions occur.

12. **Widget Hierarchy**: Parent-child relationships tracked in WidgetRegistry. Container widgets (Flex, Grid) will be implemented in Phase 4.

---

## 14. References

- [TWIN Protocol Analysis](./wish_report.md)
- [tview Documentation](https://github.com/rivo/tview/wiki)
- [tcell Documentation](https://pkg.go.dev/github.com/gdamore/tcell/v2)
- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
