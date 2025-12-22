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
//	list.SetSelected(0)
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

	pb "industries/loosh/yutani/pkg/proto/industries/loosh/yutani/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a connection to the Yutani server.
type Client struct {
	conn      *grpc.ClientConn
	sessionID *pb.SessionID
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

	// Event handling
	eventHandlers []EventHandler
	eventMu       sync.RWMutex
	eventStream   pb.EventService_StreamEventsClient
	eventDone     chan struct{}

	// Widget registry
	widgets   map[string]Widget
	widgetsMu sync.RWMutex
}

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

	client := &Client{
		conn:          conn,
		ctx:           ctx,
		cancel:        cancel,
		sessionClient: pb.NewSessionServiceClient(conn),
		widgetClient:  pb.NewWidgetServiceClient(conn),
		screenClient:  pb.NewScreenServiceClient(conn),
		eventClient:   pb.NewEventServiceClient(conn),
		listClient:    pb.NewListServiceClient(conn),
		tableClient:   pb.NewTableServiceClient(conn),
		formClient:    pb.NewFormServiceClient(conn),
		treeClient:    pb.NewTreeServiceClient(conn),
		layoutClient:  pb.NewLayoutServiceClient(conn),
		eventDone:     make(chan struct{}),
		widgets:       make(map[string]Widget),
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
func (c *Client) SessionID() *pb.SessionID {
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

// StartEventStream starts listening for events from the server.
func (c *Client) StartEventStream() error {
	stream, err := c.eventClient.StreamEvents(c.ctx, &pb.StreamEventsRequest{
		SessionId: c.sessionID,
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
	handlers := make([]EventHandler, len(c.eventHandlers))
	copy(handlers, c.eventHandlers)
	c.eventMu.RUnlock()

	for _, handler := range handlers {
		handler(event)
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
	return int(resp.Width), int(resp.Height), nil
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

