package client

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEventFilter tests event filtering
func TestEventFilter(t *testing.T) {
	// Test type filter
	filter := &EventFilter{
		Types: []EventType{EventTypeKey, EventTypeWidget},
	}

	keyEvent := &Event{Type: EventTypeKey}
	mouseEvent := &Event{Type: EventTypeMouse}
	widgetEvent := &Event{Type: EventTypeWidget, Widget: &WidgetEvent{WidgetID: "widget1"}}

	if !filter.Matches(keyEvent) {
		t.Error("Filter should match key event")
	}
	if filter.Matches(mouseEvent) {
		t.Error("Filter should not match mouse event")
	}
	if !filter.Matches(widgetEvent) {
		t.Error("Filter should match widget event")
	}
}

// TestEventFilterWidgetID tests widget ID filtering
func TestEventFilterWidgetID(t *testing.T) {
	filter := &EventFilter{
		Types:     []EventType{EventTypeWidget},
		WidgetIDs: []string{"widget1", "widget2"},
	}

	event1 := &Event{Type: EventTypeWidget, Widget: &WidgetEvent{WidgetID: "widget1"}}
	event2 := &Event{Type: EventTypeWidget, Widget: &WidgetEvent{WidgetID: "widget3"}}

	if !filter.Matches(event1) {
		t.Error("Filter should match widget1")
	}
	if filter.Matches(event2) {
		t.Error("Filter should not match widget3")
	}
}

// TestEventFilterWidgetEventType tests widget event type filtering
func TestEventFilterWidgetEventType(t *testing.T) {
	filter := &EventFilter{
		Types:            []EventType{EventTypeWidget},
		WidgetEventTypes: []string{"WIDGET_SELECTED", "WIDGET_CHANGED"},
	}

	event1 := &Event{Type: EventTypeWidget, Widget: &WidgetEvent{Type: "WIDGET_SELECTED"}}
	event2 := &Event{Type: EventTypeWidget, Widget: &WidgetEvent{Type: "WIDGET_SUBMITTED"}}

	if !filter.Matches(event1) {
		t.Error("Filter should match WIDGET_SELECTED")
	}
	if filter.Matches(event2) {
		t.Error("Filter should not match WIDGET_SUBMITTED")
	}
}

// TestEventFilterCustom tests custom filter function
func TestEventFilterCustom(t *testing.T) {
	filter := &EventFilter{
		CustomFilter: func(e *Event) bool {
			return e.Type == EventTypeKey && e.Key != nil && e.Key.Rune == 'a'
		},
	}

	event1 := &Event{Type: EventTypeKey, Key: &KeyEvent{Rune: 'a'}}
	event2 := &Event{Type: EventTypeKey, Key: &KeyEvent{Rune: 'b'}}

	if !filter.Matches(event1) {
		t.Error("Filter should match 'a' key")
	}
	if filter.Matches(event2) {
		t.Error("Filter should not match 'b' key")
	}
}

// TestEventBatcher tests event batching
func TestEventBatcher(t *testing.T) {
	var receivedCount atomic.Int32
	handler := func(e *Event) {
		receivedCount.Add(1)
	}

	batcher := NewEventBatcher(50*time.Millisecond, handler)
	defer batcher.Close()

	// Add events
	for i := 0; i < 10; i++ {
		batcher.Add(&Event{Type: EventTypeKey})
	}

	// Wait for batch to flush
	time.Sleep(100 * time.Millisecond)

	if receivedCount.Load() != 10 {
		t.Errorf("Expected 10 events, got %d", receivedCount.Load())
	}
}

// TestEventRecorder tests event recording
func TestEventRecorder(t *testing.T) {
	recorder := NewEventRecorder(100)

	// Record some events
	event1 := &Event{Type: EventTypeKey}
	event2 := &Event{Type: EventTypeMouse}
	event3 := &Event{Type: EventTypeWidget}

	recorder.Record(event1)
	time.Sleep(10 * time.Millisecond)
	recorder.Record(event2)
	time.Sleep(10 * time.Millisecond)
	recorder.Record(event3)

	// Check recorded events
	events := recorder.GetEvents()
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Test GetEventsSince
	since := time.Now().Add(-15 * time.Millisecond)
	recentEvents := recorder.GetEventsSince(since)
	if len(recentEvents) < 2 {
		t.Errorf("Expected at least 2 recent events, got %d", len(recentEvents))
	}

	// Test Clear
	recorder.Clear()
	events = recorder.GetEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events after clear, got %d", len(events))
	}
}

// TestEventRecorderStartStop tests recording control
func TestEventRecorderStartStop(t *testing.T) {
	recorder := NewEventRecorder(100)

	recorder.Record(&Event{Type: EventTypeKey})
	
	recorder.Stop()
	if recorder.IsRecording() {
		t.Error("Recorder should be stopped")
	}

	recorder.Record(&Event{Type: EventTypeMouse})
	
	events := recorder.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event (second should not be recorded), got %d", len(events))
	}

	recorder.Start()
	if !recorder.IsRecording() {
		t.Error("Recorder should be recording")
	}
}

// TestEventRecorderMaxEvents tests max events limit
func TestEventRecorderMaxEvents(t *testing.T) {
	recorder := NewEventRecorder(5)

	// Record more than max
	for i := 0; i < 10; i++ {
		recorder.Record(&Event{Type: EventTypeKey})
	}

	events := recorder.GetEvents()
	if len(events) != 5 {
		t.Errorf("Expected 5 events (max), got %d", len(events))
	}
}

// TestEventRecorderReplay tests event replay
func TestEventRecorderReplay(t *testing.T) {
	recorder := NewEventRecorder(100)

	// Record events
	for i := 0; i < 5; i++ {
		recorder.Record(&Event{Type: EventTypeKey})
	}

	// Replay immediately
	replayCount := 0
	handler := func(e *Event) {
		replayCount++
	}

	recorder.Replay(handler, false)

	if replayCount != 5 {
		t.Errorf("Expected 5 replayed events, got %d", replayCount)
	}
}

// TestEventRecorderReplayRealtime tests realtime replay
func TestEventRecorderReplayRealtime(t *testing.T) {
	recorder := NewEventRecorder(100)

	// Record events with delays
	recorder.Record(&Event{Type: EventTypeKey})
	time.Sleep(20 * time.Millisecond)
	recorder.Record(&Event{Type: EventTypeMouse})
	time.Sleep(20 * time.Millisecond)
	recorder.Record(&Event{Type: EventTypeWidget})

	// Replay with original timing
	replayCount := 0
	var mu sync.Mutex
	var replayTimes []time.Time

	handler := func(e *Event) {
		mu.Lock()
		replayCount++
		replayTimes = append(replayTimes, time.Now())
		mu.Unlock()
	}

	recorder.Replay(handler, true)

	// Wait for replay to complete
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := replayCount
	times := make([]time.Time, len(replayTimes))
	copy(times, replayTimes)
	mu.Unlock()

	if count != 3 {
		t.Errorf("Expected 3 replayed events, got %d", count)
	}

	// Verify timing (events should be spaced out)
	if len(times) >= 2 {
		delay := times[1].Sub(times[0])
		if delay < 15*time.Millisecond {
			t.Errorf("Expected delay of at least 15ms between events, got %v", delay)
		}
	}
}

// TestEventRecorderReplayEmpty tests replay with no events
func TestEventRecorderReplayEmpty(t *testing.T) {
	recorder := NewEventRecorder(100)

	replayCount := 0
	handler := func(e *Event) {
		replayCount++
	}

	// Should not panic or call handler
	recorder.Replay(handler, false)
	recorder.Replay(handler, true)

	if replayCount != 0 {
		t.Errorf("Expected 0 replayed events, got %d", replayCount)
	}
}

// TestEventMiddleware tests middleware functionality
func TestEventMiddleware(t *testing.T) {
	// Middleware that modifies events
	middleware1 := func(e *Event) (*Event, bool) {
		if e.Type == EventTypeKey && e.Key != nil {
			e.Key.Rune = 'X' // Modify the rune
		}
		return e, true
	}

	// Middleware that filters events
	middleware2 := func(e *Event) (*Event, bool) {
		if e.Type == EventTypeMouse {
			return e, false // Block mouse events
		}
		return e, true
	}

	// Test middleware1
	event := &Event{Type: EventTypeKey, Key: &KeyEvent{Rune: 'a'}}
	modifiedEvent, cont := middleware1(event)
	if !cont {
		t.Error("Middleware1 should continue")
	}
	if modifiedEvent.Key.Rune != 'X' {
		t.Error("Middleware1 should modify rune to 'X'")
	}

	// Test middleware2
	mouseEvent := &Event{Type: EventTypeMouse}
	_, cont = middleware2(mouseEvent)
	if cont {
		t.Error("Middleware2 should block mouse events")
	}

	keyEvent := &Event{Type: EventTypeKey}
	_, cont = middleware2(keyEvent)
	if !cont {
		t.Error("Middleware2 should allow key events")
	}
}

// TestCombinedFilters tests combining multiple filter criteria
func TestCombinedFilters(t *testing.T) {
	filter := &EventFilter{
		Types:            []EventType{EventTypeWidget},
		WidgetIDs:        []string{"widget1"},
		WidgetEventTypes: []string{"WIDGET_SELECTED"},
	}

	// Should match: correct type, widget ID, and event type
	event1 := &Event{
		Type: EventTypeWidget,
		Widget: &WidgetEvent{
			WidgetID: "widget1",
			Type:     "WIDGET_SELECTED",
		},
	}

	// Should not match: wrong widget ID
	event2 := &Event{
		Type: EventTypeWidget,
		Widget: &WidgetEvent{
			WidgetID: "widget2",
			Type:     "WIDGET_SELECTED",
		},
	}

	// Should not match: wrong event type
	event3 := &Event{
		Type: EventTypeWidget,
		Widget: &WidgetEvent{
			WidgetID: "widget1",
			Type:     "WIDGET_CHANGED",
		},
	}

	if !filter.Matches(event1) {
		t.Error("Filter should match event1")
	}
	if filter.Matches(event2) {
		t.Error("Filter should not match event2 (wrong widget ID)")
	}
	if filter.Matches(event3) {
		t.Error("Filter should not match event3 (wrong event type)")
	}
}
