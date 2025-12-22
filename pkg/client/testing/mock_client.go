// Package testing provides testing utilities for Yutani applications.
package testing

import (
	"context"
	"sync"

	"github.com/chazu/yutani/pkg/client"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// MockClient is an in-memory client for testing that doesn't require a server.
type MockClient struct {
	mu sync.RWMutex

	// Call recording
	calls     []MethodCall
	callIndex map[string]int // method name -> call count

	// Mock data
	widgets      map[string]*MockWidget
	sessionID    string
	eventStream  chan *client.Event
	eventHandler client.EventHandler

	// Configuration
	shouldFail map[string]error // method name -> error to return
}

// MethodCall records a method invocation.
type MethodCall struct {
	Method string
	Args   []interface{}
	Result interface{}
	Error  error
}

// MockWidget represents a mock widget for testing.
type MockWidget struct {
	ID         string
	Type       pb.WidgetType
	Properties map[string]interface{}
	Children   []string
	Parent     string
}

// NewMockClient creates a new mock client for testing.
func NewMockClient() *MockClient {
	return &MockClient{
		calls:        make([]MethodCall, 0),
		callIndex:    make(map[string]int),
		widgets:      make(map[string]*MockWidget),
		sessionID:    "mock-session-id",
		eventStream:  make(chan *client.Event, 100),
		shouldFail:   make(map[string]error),
	}
}

// RecordCall records a method call for verification.
func (m *MockClient) RecordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, MethodCall{
		Method: method,
		Args:   args,
	})
	m.callIndex[method]++
}

// GetCallCount returns the number of times a method was called.
func (m *MockClient) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callIndex[method]
}

// GetCalls returns all recorded calls for a method.
func (m *MockClient) GetCalls(method string) []MethodCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []MethodCall
	for _, call := range m.calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// GetAllCalls returns all recorded calls.
func (m *MockClient) GetAllCalls() []MethodCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]MethodCall{}, m.calls...)
}

// ClearCalls clears all recorded calls.
func (m *MockClient) ClearCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]MethodCall, 0)
	m.callIndex = make(map[string]int)
}

// SetError configures a method to return an error.
func (m *MockClient) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail[method] = err
}

// ClearError removes error configuration for a method.
func (m *MockClient) ClearError(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.shouldFail, method)
}

// checkError returns configured error if set.
func (m *MockClient) checkError(method string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shouldFail[method]
}

// Context returns a mock context.
func (m *MockClient) Context() context.Context {
	return context.Background()
}

// SessionID returns the mock session ID.
func (m *MockClient) SessionID() string {
	return m.sessionID
}

// Close closes the mock client (no-op).
func (m *MockClient) Close() error {
	m.RecordCall("Close")
	return m.checkError("Close")
}

// RegisterWidget registers a mock widget.
func (m *MockClient) RegisterWidget(id string, widgetType pb.WidgetType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.widgets[id] = &MockWidget{
		ID:         id,
		Type:       widgetType,
		Properties: make(map[string]interface{}),
		Children:   make([]string, 0),
	}
}

// GetWidget returns a mock widget by ID.
func (m *MockClient) GetWidget(id string) (*MockWidget, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	widget, ok := m.widgets[id]
	return widget, ok
}

// Event Simulation

// SimulateKeyPress simulates a key press event.
func (m *MockClient) SimulateKeyPress(key rune) {
	m.RecordCall("SimulateKeyPress", key)
	event := &client.Event{
		Type: client.EventTypeKey,
		Key: &client.KeyEvent{
			Key:  string(key),
			Rune: key,
		},
	}
	m.sendEvent(event)
}

// SimulateKeyEvent simulates a key event with full control.
func (m *MockClient) SimulateKeyEvent(key string, r rune) {
	m.RecordCall("SimulateKeyEvent", key, r)
	event := &client.Event{
		Type: client.EventTypeKey,
		Key: &client.KeyEvent{
			Key:  key,
			Rune: r,
		},
	}
	m.sendEvent(event)
}

// SimulateMouseClick simulates a mouse click event.
func (m *MockClient) SimulateMouseClick(x, y int, button int) {
	m.RecordCall("SimulateMouseClick", x, y, button)
	event := &client.Event{
		Type: client.EventTypeMouse,
		Mouse: &client.MouseEvent{
			X:      x,
			Y:      y,
			Button: button,
			Action: "click",
		},
	}
	m.sendEvent(event)
}

// SimulateResize simulates a terminal resize event.
func (m *MockClient) SimulateResize(width, height int) {
	m.RecordCall("SimulateResize", width, height)
	event := &client.Event{
		Type: client.EventTypeResize,
		Resize: &client.ResizeEvent{
			Width:  width,
			Height: height,
		},
	}
	m.sendEvent(event)
}

// SimulateWidgetEvent simulates a widget event.
func (m *MockClient) SimulateWidgetEvent(widgetID, eventType string) {
	m.RecordCall("SimulateWidgetEvent", widgetID, eventType)
	event := &client.Event{
		Type: client.EventTypeWidget,
		Widget: &client.WidgetEvent{
			WidgetID: widgetID,
			Type:     eventType,
		},
	}
	m.sendEvent(event)
}

// sendEvent sends an event to registered handlers.
func (m *MockClient) sendEvent(event *client.Event) {
	m.mu.RLock()
	handler := m.eventHandler
	m.mu.RUnlock()

	if handler != nil {
		handler(event)
	}

	// Also send to event stream channel
	select {
	case m.eventStream <- event:
	default:
		// Channel full, skip
	}
}

// OnEvent registers an event handler.
func (m *MockClient) OnEvent(handler client.EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RecordCall("OnEvent")
	m.eventHandler = handler
}

// StartEventStream starts the mock event stream (no-op).
func (m *MockClient) StartEventStream() error {
	m.RecordCall("StartEventStream")
	return m.checkError("StartEventStream")
}

// StopEventStream stops the mock event stream (no-op).
func (m *MockClient) StopEventStream() error {
	m.RecordCall("StopEventStream")
	return m.checkError("StopEventStream")
}

