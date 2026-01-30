// Package client provides a high-level Go client library for the Yutani Terminal Display Server.
//
// This package offers a fluent, idiomatic Go API for creating and managing terminal UIs
// through the Yutani server, abstracting away the low-level gRPC details.
//
// Example usage:
//
//	client, err := yutani.Connect("localhost:50051")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Create a list widget
//	list := client.NewList().
//		Title("Menu").
//		Border(true).
//		Build()
//
//	list.AddItem("New File", "Create a new file", "n")
//	list.AddItem("Open File", "Open an existing file", "o")
//	list.SetSelection(0)
//
//	// Handle events
//	client.OnEvent(func(event *Event) {
//		if event.Type == EventTypeWidget && event.Widget.Type == WidgetEventSelected {
//			fmt.Printf("Selected: %s\n", event.Widget.WidgetId)
//		}
//	})
//
//	client.Run()
package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a connection to the Yutani server.
type Client struct {
	conn      *grpc.ClientConn
	sessionID *pb.SessionId
	ctx       context.Context
	cancel    context.CancelFunc

	// gRPC clients
	sessionClient pb.SessionServiceClient
	widgetClient  pb.WidgetServiceClient
	screenClient  pb.ScreenServiceClient
	eventClient   pb.EventServiceClient
	listClient    pb.ListServiceClient
	tableClient   pb.TableServiceClient
	formClient    pb.FormServiceClient
	treeClient    pb.TreeServiceClient
	layoutClient  pb.LayoutServiceClient
	windowClient   pb.WindowServiceClient
	textAreaClient pb.TextAreaServiceClient
	menuClient     pb.MenuServiceClient

	// Event handling
	eventHandlers   []EventHandler
	eventMiddleware []EventMiddleware
	eventFilters    []*EventFilter
	eventRecorder   *EventRecorder
	eventMu         sync.RWMutex
	eventStream     pb.EventService_SubscribeClient
	eventDone       chan struct{}

	// Widget registry
	widgets   map[string]Widget
	widgetsMu sync.RWMutex

	// Connection management
	mu                 sync.RWMutex
	autoReconnect      bool
	reconnectAddress   string
	reconnectOpts      RetryOptions
	reconnectGrpcOpts  []grpc.DialOption
	connectionState    ConnectionState
	stateCallbacks     []ConnectionStateCallback
	healthCheckTicker  *time.Ticker
	healthCheckStop    chan struct{}
}

// ConnectionState represents the current connection state.
type ConnectionState int

const (
	// StateConnected indicates the client is connected.
	StateConnected ConnectionState = iota
	// StateDisconnected indicates the client is disconnected.
	StateDisconnected
	// StateReconnecting indicates the client is attempting to reconnect.
	StateReconnecting
)

// String returns the string representation of the connection state.
func (s ConnectionState) String() string {
	switch s {
	case StateConnected:
		return "Connected"
	case StateDisconnected:
		return "Disconnected"
	case StateReconnecting:
		return "Reconnecting"
	default:
		return "Unknown"
	}
}

// ConnectionStateCallback is called when connection state changes.
type ConnectionStateCallback func(oldState, newState ConnectionState)

// EventHandler is a function that handles events from the server.
type EventHandler func(*Event)

// Connect creates a new client connection to the Yutani server.
func Connect(address string) (*Client, error) {
	return ConnectWithOptions(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// ConnectWithOptions creates a new client connection with custom gRPC dial options.
func ConnectWithOptions(address string, opts ...grpc.DialOption) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return ConnectWithConn(conn)
}

// ConnectWithConn creates a new client using an existing gRPC connection.
// This is useful for testing with a test server.
func ConnectWithConn(conn *grpc.ClientConn) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		conn:            conn,
		ctx:             ctx,
		cancel:          cancel,
		sessionClient:   pb.NewSessionServiceClient(conn),
		widgetClient:    pb.NewWidgetServiceClient(conn),
		screenClient:    pb.NewScreenServiceClient(conn),
		eventClient:     pb.NewEventServiceClient(conn),
		listClient:      pb.NewListServiceClient(conn),
		tableClient:     pb.NewTableServiceClient(conn),
		formClient:      pb.NewFormServiceClient(conn),
		treeClient:      pb.NewTreeServiceClient(conn),
		layoutClient:    pb.NewLayoutServiceClient(conn),
		windowClient:    pb.NewWindowServiceClient(conn),
		textAreaClient:  pb.NewTextAreaServiceClient(conn),
		menuClient:      pb.NewMenuServiceClient(conn),
		eventDone:       make(chan struct{}),
		widgets:         make(map[string]Widget),
		connectionState: StateConnected,
		stateCallbacks:  make([]ConnectionStateCallback, 0),
	}

	// Create session
	sessionResp, err := client.sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "yutani-go-client",
	})
	if err != nil {
		conn.Close()
		cancel()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	client.sessionID = sessionResp.SessionId

	return client, nil
}

// Close closes the client connection and cleans up resources.
func (c *Client) Close() error {
	// Stop health check
	c.StopHealthCheck()

	// Update state
	c.setConnectionState(StateDisconnected)

	// Destroy session
	if c.sessionID != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{
			SessionId: c.sessionID,
		})
	}

	// Stop event stream
	if c.eventStream != nil {
		c.eventStream.CloseSend()
		<-c.eventDone
	}

	// Close connection
	c.cancel()
	return c.conn.Close()
}

// SessionID returns the current session ID.
func (c *Client) SessionID() *pb.SessionId {
	return c.sessionID
}

// Context returns the client's context.
func (c *Client) Context() context.Context {
	return c.ctx
}

// OnEvent registers an event handler.
func (c *Client) OnEvent(handler EventHandler) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()
	c.eventHandlers = append(c.eventHandlers, handler)
}

// OnEventFiltered registers an event handler with a filter.
func (c *Client) OnEventFiltered(handler EventHandler, filter *EventFilter) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()

	// Wrap handler with filter
	wrappedHandler := func(event *Event) {
		if filter.Matches(event) {
			handler(event)
		}
	}
	c.eventHandlers = append(c.eventHandlers, wrappedHandler)
}

// OnWidgetEvent registers a handler for events from a specific widget.
func (c *Client) OnWidgetEvent(widgetID string, handler EventHandler) {
	filter := &EventFilter{
		Types:     []EventType{EventTypeWidget},
		WidgetIDs: []string{widgetID},
	}
	c.OnEventFiltered(handler, filter)
}

// OnEventType registers a handler for specific event types.
func (c *Client) OnEventType(eventType EventType, handler EventHandler) {
	filter := &EventFilter{
		Types: []EventType{eventType},
	}
	c.OnEventFiltered(handler, filter)
}

// AddEventMiddleware adds middleware to the event processing pipeline.
func (c *Client) AddEventMiddleware(middleware EventMiddleware) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()
	c.eventMiddleware = append(c.eventMiddleware, middleware)
}

// EnableEventRecording enables event recording with the specified max events.
func (c *Client) EnableEventRecording(maxEvents int) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()
	c.eventRecorder = NewEventRecorder(maxEvents)
}

// GetEventRecorder returns the event recorder if enabled.
func (c *Client) GetEventRecorder() *EventRecorder {
	c.eventMu.RLock()
	defer c.eventMu.RUnlock()
	return c.eventRecorder
}

// SetServerEventFilter updates the server-side event filter.
// This controls which events are sent from the server to the client.
func (c *Client) SetServerEventFilter(filter *pb.EventFilter) error {
	_, err := c.eventClient.SetEventFilter(c.ctx, &pb.SetEventFilterRequest{
		SessionId: c.sessionID,
		Filter:    filter,
	})
	return err
}

// SetServerEventFilterSimple is a convenience method to set basic event type filters.
func (c *Client) SetServerEventFilterSimple(key, mouse, resize, focus, widget bool) error {
	return c.SetServerEventFilter(&pb.EventFilter{
		ReceiveKeyEvents:    key,
		ReceiveMouseEvents:  mouse,
		ReceiveResizeEvents: resize,
		ReceiveFocusEvents:  focus,
		ReceiveWidgetEvents: widget,
	})
}

// SetServerWidgetFilter sets the server to only send events for specific widgets.
func (c *Client) SetServerWidgetFilter(widgetIDs []string) error {
	return c.SetServerEventFilter(&pb.EventFilter{
		ReceiveKeyEvents:    true,
		ReceiveMouseEvents:  true,
		ReceiveResizeEvents: true,
		ReceiveFocusEvents:  true,
		ReceiveWidgetEvents: true,
		WidgetIds:           widgetIDs,
	})
}

// StartEventStream starts listening for events from the server.
func (c *Client) StartEventStream() error {
	stream, err := c.eventClient.Subscribe(c.ctx, &pb.SubscribeRequest{
		SessionId: c.sessionID,
		Filter: &pb.EventFilter{
			ReceiveKeyEvents:    true,
			ReceiveMouseEvents:  true,
			ReceiveResizeEvents: true,
			ReceiveFocusEvents:  true,
			ReceiveWidgetEvents: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to start event stream: %w", err)
	}

	c.eventStream = stream

	go func() {
		defer close(c.eventDone)
		for {
			event, err := stream.Recv()
			if err != nil {
				return
			}

			// Convert to client event and dispatch
			clientEvent := convertEvent(event)
			c.dispatchEvent(clientEvent)
		}
	}()

	return nil
}

// dispatchEvent sends an event to all registered handlers.
func (c *Client) dispatchEvent(event *Event) {
	c.eventMu.RLock()

	// Apply middleware
	processedEvent := event
	shouldContinue := true
	for _, middleware := range c.eventMiddleware {
		var cont bool
		processedEvent, cont = middleware(processedEvent)
		if !cont {
			shouldContinue = false
			break
		}
	}

	// Record event if recorder is enabled
	if c.eventRecorder != nil {
		c.eventRecorder.Record(processedEvent)
	}

	handlers := make([]EventHandler, len(c.eventHandlers))
	copy(handlers, c.eventHandlers)
	c.eventMu.RUnlock()

	if !shouldContinue {
		return
	}

	for _, handler := range handlers {
		handler(processedEvent)
	}
}

// GetScreenSize returns the current screen dimensions.
func (c *Client) GetScreenSize() (width, height int, err error) {
	resp, err := c.screenClient.GetSize(c.ctx, &pb.GetSizeRequest{
		SessionId: c.sessionID,
	})
	if err != nil {
		return 0, 0, err
	}
	if resp.Size == nil {
		return 0, 0, nil
	}
	return int(resp.Size.Width), int(resp.Size.Height), nil
}

// ClearScreen clears the screen.
func (c *Client) ClearScreen() error {
	_, err := c.screenClient.Clear(c.ctx, &pb.ClearRequest{
		SessionId: c.sessionID,
	})
	return err
}

// Sync synchronizes the screen display.
func (c *Client) Sync() error {
	_, err := c.screenClient.Sync(c.ctx, &pb.SyncRequest{
		SessionId: c.sessionID,
	})
	return err
}

// registerWidget adds a widget to the client's registry.
func (c *Client) registerWidget(widget Widget) {
	c.widgetsMu.Lock()
	defer c.widgetsMu.Unlock()
	c.widgets[widget.ID()] = widget
}

// unregisterWidget removes a widget from the client's registry.
func (c *Client) unregisterWidget(widgetID string) {
	c.widgetsMu.Lock()
	defer c.widgetsMu.Unlock()
	delete(c.widgets, widgetID)
}

// GetWidget returns a widget by ID.
func (c *Client) GetWidget(widgetID string) (Widget, bool) {
	c.widgetsMu.RLock()
	defer c.widgetsMu.RUnlock()
	widget, ok := c.widgets[widgetID]
	return widget, ok
}

// SetRoot sets a widget as the root widget displayed on the server.
// This makes the widget visible on the server's terminal.
func (c *Client) SetRoot(widget Widget) error {
	_, err := c.widgetClient.SetRoot(c.ctx, &pb.SetRootRequest{
		SessionId: c.sessionID,
		WidgetId:  &pb.WidgetId{Id: widget.ID()},
	})
	return err
}

// Connection State Management

// GetConnectionState returns the current connection state.
func (c *Client) GetConnectionState() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connectionState
}

// IsConnected returns true if the client is connected.
func (c *Client) IsConnected() bool {
	return c.GetConnectionState() == StateConnected
}

// IsHealthy checks if the connection is healthy by pinging the server.
func (c *Client) IsHealthy() bool {
	if !c.IsConnected() {
		return false
	}

	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()

	_, err := c.sessionClient.Ping(ctx, &pb.PingRequest{
		Timestamp: time.Now().UnixNano(),
	})

	return err == nil
}

// OnConnectionStateChange registers a callback for connection state changes.
func (c *Client) OnConnectionStateChange(callback ConnectionStateCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stateCallbacks = append(c.stateCallbacks, callback)
}

// setConnectionState updates the connection state and notifies callbacks.
func (c *Client) setConnectionState(newState ConnectionState) {
	c.mu.Lock()
	oldState := c.connectionState
	c.connectionState = newState
	callbacks := make([]ConnectionStateCallback, len(c.stateCallbacks))
	copy(callbacks, c.stateCallbacks)
	c.mu.Unlock()

	// Notify callbacks
	if oldState != newState {
		for _, callback := range callbacks {
			callback(oldState, newState)
		}
	}
}

// Reconnect manually triggers a reconnection attempt.
func (c *Client) Reconnect() error {
	c.mu.RLock()
	if !c.autoReconnect {
		c.mu.RUnlock()
		return fmt.Errorf("auto-reconnect not enabled")
	}
	address := c.reconnectAddress
	opts := c.reconnectOpts
	grpcOpts := c.reconnectGrpcOpts
	c.mu.RUnlock()

	c.setConnectionState(StateReconnecting)

	// Close existing connection
	if c.conn != nil {
		c.conn.Close()
	}

	// Try to reconnect
	newClient, err := ConnectWithRetryAndOptions(address, opts, grpcOpts...)
	if err != nil {
		c.setConnectionState(StateDisconnected)
		return err
	}

	// Update connection
	c.mu.Lock()
	c.conn = newClient.conn
	c.sessionID = newClient.sessionID
	c.sessionClient = newClient.sessionClient
	c.widgetClient = newClient.widgetClient
	c.screenClient = newClient.screenClient
	c.eventClient = newClient.eventClient
	c.listClient = newClient.listClient
	c.tableClient = newClient.tableClient
	c.formClient = newClient.formClient
	c.treeClient = newClient.treeClient
	c.layoutClient = newClient.layoutClient
	c.windowClient = newClient.windowClient
	c.textAreaClient = newClient.textAreaClient
	c.menuClient = newClient.menuClient
	c.mu.Unlock()

	c.setConnectionState(StateConnected)

	// Restart event stream if it was running
	if c.eventStream != nil {
		c.StartEventStream()
	}

	return nil
}

// StartHealthCheck starts periodic health checks.
func (c *Client) StartHealthCheck(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop existing health check if running
	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
		close(c.healthCheckStop)
	}

	c.healthCheckTicker = time.NewTicker(interval)
	c.healthCheckStop = make(chan struct{})

	// Capture in local variables to avoid data race with StopHealthCheck
	ticker := c.healthCheckTicker
	stopCh := c.healthCheckStop

	go func() {
		for {
			select {
			case <-ticker.C:
				if !c.IsHealthy() && c.autoReconnect {
					c.Reconnect()
				}
			case <-stopCh:
				return
			}
		}
	}()
}

// StopHealthCheck stops periodic health checks.
func (c *Client) StopHealthCheck() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
		close(c.healthCheckStop)
		c.healthCheckTicker = nil
		c.healthCheckStop = nil
	}
}

