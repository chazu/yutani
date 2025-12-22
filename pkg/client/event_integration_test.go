package client

import (
	"sync"
	"testing"
	"time"
)

// TestClientEventFilteredIntegration tests OnEventFiltered integration with Client
func TestClientEventFilteredIntegration(t *testing.T) {
	c := &Client{
		eventHandlers:   []EventHandler{},
		eventMiddleware: []EventMiddleware{},
		eventFilters:    []*EventFilter{},
	}

	receivedEvents := 0
	var mu sync.Mutex

	// Register filtered handler
	filter := &EventFilter{
		Types: []EventType{EventTypeKey},
	}
	c.OnEventFiltered(func(e *Event) {
		mu.Lock()
		receivedEvents++
		mu.Unlock()
	}, filter)

	// Dispatch key event (should be received)
	c.dispatchEvent(&Event{Type: EventTypeKey})

	// Dispatch mouse event (should be filtered out)
	c.dispatchEvent(&Event{Type: EventTypeMouse})

	mu.Lock()
	count := receivedEvents
	mu.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 event, got %d", count)
	}
}

// TestClientOnWidgetEventIntegration tests OnWidgetEvent integration
func TestClientOnWidgetEventIntegration(t *testing.T) {
	c := &Client{
		eventHandlers:   []EventHandler{},
		eventMiddleware: []EventMiddleware{},
		eventFilters:    []*EventFilter{},
	}

	receivedEvents := 0
	var mu sync.Mutex

	c.OnWidgetEvent("widget1", func(e *Event) {
		mu.Lock()
		receivedEvents++
		mu.Unlock()
	})

	// Should receive
	c.dispatchEvent(&Event{
		Type:   EventTypeWidget,
		Widget: &WidgetEvent{WidgetID: "widget1"},
	})

	// Should not receive (different widget)
	c.dispatchEvent(&Event{
		Type:   EventTypeWidget,
		Widget: &WidgetEvent{WidgetID: "widget2"},
	})

	// Should not receive (not a widget event)
	c.dispatchEvent(&Event{Type: EventTypeKey})

	mu.Lock()
	count := receivedEvents
	mu.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 event, got %d", count)
	}
}

// TestClientOnEventTypeIntegration tests OnEventType integration
func TestClientOnEventTypeIntegration(t *testing.T) {
	c := &Client{
		eventHandlers:   []EventHandler{},
		eventMiddleware: []EventMiddleware{},
		eventFilters:    []*EventFilter{},
	}

	keyCount := 0
	mouseCount := 0
	var mu sync.Mutex

	c.OnEventType(EventTypeKey, func(e *Event) {
		mu.Lock()
		keyCount++
		mu.Unlock()
	})

	c.OnEventType(EventTypeMouse, func(e *Event) {
		mu.Lock()
		mouseCount++
		mu.Unlock()
	})

	c.dispatchEvent(&Event{Type: EventTypeKey})
	c.dispatchEvent(&Event{Type: EventTypeKey})
	c.dispatchEvent(&Event{Type: EventTypeMouse})

	mu.Lock()
	kc := keyCount
	mc := mouseCount
	mu.Unlock()

	if kc != 2 {
		t.Errorf("Expected 2 key events, got %d", kc)
	}
	if mc != 1 {
		t.Errorf("Expected 1 mouse event, got %d", mc)
	}
}

// TestClientMiddlewareIntegration tests middleware integration with dispatchEvent
func TestClientMiddlewareIntegration(t *testing.T) {
	c := &Client{
		eventHandlers:   []EventHandler{},
		eventMiddleware: []EventMiddleware{},
		eventFilters:    []*EventFilter{},
	}

	middlewareCalled := false
	handlerCalled := false
	var mu sync.Mutex

	// Add middleware that modifies events
	c.AddEventMiddleware(func(e *Event) (*Event, bool) {
		mu.Lock()
		middlewareCalled = true
		mu.Unlock()
		if e.Type == EventTypeKey && e.Key != nil {
			e.Key.Rune = 'Z'
		}
		return e, true
	})

	c.OnEvent(func(e *Event) {
		mu.Lock()
		handlerCalled = true
		mu.Unlock()
		if e.Type == EventTypeKey && e.Key != nil && e.Key.Rune != 'Z' {
			t.Error("Middleware should have modified the rune to 'Z'")
		}
	})

	c.dispatchEvent(&Event{Type: EventTypeKey, Key: &KeyEvent{Rune: 'A'}})

	mu.Lock()
	mw := middlewareCalled
	h := handlerCalled
	mu.Unlock()

	if !mw {
		t.Error("Middleware should have been called")
	}
	if !h {
		t.Error("Handler should have been called")
	}
}

// TestClientMiddlewareBlocking tests middleware that blocks events
func TestClientMiddlewareBlocking(t *testing.T) {
	c := &Client{
		eventHandlers:   []EventHandler{},
		eventMiddleware: []EventMiddleware{},
		eventFilters:    []*EventFilter{},
	}

	handlerCalled := false
	var mu sync.Mutex

	// Add middleware that blocks mouse events
	c.AddEventMiddleware(func(e *Event) (*Event, bool) {
		if e.Type == EventTypeMouse {
			return e, false // Block
		}
		return e, true
	})

	c.OnEvent(func(e *Event) {
		mu.Lock()
		handlerCalled = true
		mu.Unlock()
	})

	// This should be blocked
	c.dispatchEvent(&Event{Type: EventTypeMouse})

	mu.Lock()
	called := handlerCalled
	mu.Unlock()

	if called {
		t.Error("Handler should not have been called (event blocked by middleware)")
	}

	// This should pass through
	c.dispatchEvent(&Event{Type: EventTypeKey})

	mu.Lock()
	called = handlerCalled
	mu.Unlock()

	if !called {
		t.Error("Handler should have been called for key event")
	}
}

// TestClientEventRecorderIntegration tests event recorder integration
func TestClientEventRecorderIntegration(t *testing.T) {
	c := &Client{
		eventHandlers:   []EventHandler{},
		eventMiddleware: []EventMiddleware{},
		eventFilters:    []*EventFilter{},
	}

	c.EnableEventRecording(100)

	// Dispatch some events
	c.dispatchEvent(&Event{Type: EventTypeKey})
	c.dispatchEvent(&Event{Type: EventTypeMouse})
	c.dispatchEvent(&Event{Type: EventTypeWidget})

	recorder := c.GetEventRecorder()
	if recorder == nil {
		t.Fatal("Recorder should be enabled")
	}

	events := recorder.GetEvents()
	if len(events) != 3 {
		t.Errorf("Expected 3 recorded events, got %d", len(events))
	}
}

// TestClientMultipleMiddleware tests multiple middleware in sequence
func TestClientMultipleMiddleware(t *testing.T) {
	c := &Client{
		eventHandlers:   []EventHandler{},
		eventMiddleware: []EventMiddleware{},
		eventFilters:    []*EventFilter{},
	}

	order := []int{}
	var mu sync.Mutex

	// First middleware
	c.AddEventMiddleware(func(e *Event) (*Event, bool) {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		return e, true
	})

	// Second middleware
	c.AddEventMiddleware(func(e *Event) (*Event, bool) {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		return e, true
	})

	// Third middleware
	c.AddEventMiddleware(func(e *Event) (*Event, bool) {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		return e, true
	})

	c.OnEvent(func(e *Event) {
		mu.Lock()
		order = append(order, 4)
		mu.Unlock()
	})

	c.dispatchEvent(&Event{Type: EventTypeKey})

	mu.Lock()
	finalOrder := make([]int, len(order))
	copy(finalOrder, order)
	mu.Unlock()

	expected := []int{1, 2, 3, 4}
	if len(finalOrder) != len(expected) {
		t.Errorf("Expected %d items in order, got %d", len(expected), len(finalOrder))
	}
	for i, v := range expected {
		if i >= len(finalOrder) || finalOrder[i] != v {
			t.Errorf("Expected order %v, got %v", expected, finalOrder)
			break
		}
	}
}

// TestEventFilterNilEvent tests filter with nil event (edge case)
func TestEventFilterNilEvent(t *testing.T) {
	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Filter.Matches panicked with nil event: %v", r)
		}
	}()

	// Create event with nil widget (edge case for widget filters)
	event := &Event{Type: EventTypeWidget, Widget: nil}
	filter := &EventFilter{
		WidgetIDs: []string{"widget1"},
	}

	result := filter.Matches(event)
	if result {
		t.Error("Filter should not match event with nil widget")
	}
}

// TestEventBatcherConcurrency tests concurrent access to EventBatcher
func TestEventBatcherConcurrency(t *testing.T) {
	receivedCount := 0
	var mu sync.Mutex

	handler := func(e *Event) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	batcher := NewEventBatcher(50*time.Millisecond, handler)
	defer batcher.Close()

	// Add events concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				batcher.Add(&Event{Type: EventTypeKey})
			}
		}()
	}

	wg.Wait()
	time.Sleep(150 * time.Millisecond) // Wait for batches to flush

	mu.Lock()
	count := receivedCount
	mu.Unlock()

	if count != 100 {
		t.Errorf("Expected 100 events, got %d", count)
	}
}

// TestEventRecorderConcurrency tests concurrent access to EventRecorder
func TestEventRecorderConcurrency(t *testing.T) {
	recorder := NewEventRecorder(1000)

	// Record events concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				recorder.Record(&Event{Type: EventTypeKey})
			}
		}()
	}

	wg.Wait()

	events := recorder.GetEvents()
	if len(events) != 100 {
		t.Errorf("Expected 100 events, got %d", len(events))
	}
}

