package server

import (
	"log/slog"
	"sync"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// EventDispatcher manages event subscriptions and dispatching
type EventDispatcher struct {
	subscribers map[string]*EventSubscriber // sessionId -> subscriber
	mu          sync.RWMutex
}

// EventSubscriber represents a client's event subscription
type EventSubscriber struct {
	SessionID string
	Filter    *pb.EventFilter
	Events    chan *pb.Event
	Done      chan struct{}
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		subscribers: make(map[string]*EventSubscriber),
	}
}

// Subscribe creates a new event subscription for a session
func (d *EventDispatcher) Subscribe(sessionID string, filter *pb.EventFilter) *EventSubscriber {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Close existing subscription if any
	if existing, ok := d.subscribers[sessionID]; ok {
		close(existing.Done)
		close(existing.Events)
	}

	// Create new subscriber with buffered channel
	sub := &EventSubscriber{
		SessionID: sessionID,
		Filter:    filter,
		Events:    make(chan *pb.Event, 100), // Buffer up to 100 events
		Done:      make(chan struct{}),
	}

	d.subscribers[sessionID] = sub
	return sub
}

// Unsubscribe removes a session's event subscription
func (d *EventDispatcher) Unsubscribe(sessionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if sub, ok := d.subscribers[sessionID]; ok {
		close(sub.Done)
		close(sub.Events)
		delete(d.subscribers, sessionID)
	}
}

// Dispatch sends an event to the appropriate subscriber(s)
func (d *EventDispatcher) Dispatch(event *pb.Event) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if event.SessionId == nil {
		slog.Debug("Event dispatch: no session ID")
		return
	}

	sessionID := event.SessionId.Id
	sub, ok := d.subscribers[sessionID]
	if !ok {
		slog.Debug("Event dispatch: no subscriber", "session_id", sessionID)
		return
	}

	// Check if event passes the filter
	if !d.passesFilter(event, sub.Filter) {
		slog.Debug("Event dispatch: filtered out", "session_id", sessionID, "event_type", getEventTypeName(event))
		return
	}

	// Try to send event, drop if channel is full
	select {
	case sub.Events <- event:
		slog.Debug("Event dispatched", "session_id", sessionID, "event_type", getEventTypeName(event))
	default:
		// Event dropped - channel full
		slog.Warn("Event dropped: channel full", "session_id", sessionID, "event_type", getEventTypeName(event))
	}
}

// getEventTypeName returns a human-readable event type name
func getEventTypeName(event *pb.Event) string {
	switch event.Event.(type) {
	case *pb.Event_Key:
		return "KEY"
	case *pb.Event_Mouse:
		return "MOUSE"
	case *pb.Event_Resize:
		return "RESIZE"
	case *pb.Event_Focus:
		return "FOCUS"
	case *pb.Event_Widget:
		return "WIDGET"
	default:
		return "UNKNOWN"
	}
}

// passesFilter checks if an event passes the subscriber's filter
func (d *EventDispatcher) passesFilter(event *pb.Event, filter *pb.EventFilter) bool {
	if filter == nil {
		return true // No filter, accept all events
	}

	// Check event type filters
	switch event.Event.(type) {
	case *pb.Event_Key:
		if !filter.ReceiveKeyEvents {
			return false
		}
	case *pb.Event_Mouse:
		if !filter.ReceiveMouseEvents {
			return false
		}
	case *pb.Event_Resize:
		if !filter.ReceiveResizeEvents {
			return false
		}
	case *pb.Event_Focus:
		if !filter.ReceiveFocusEvents {
			return false
		}
	case *pb.Event_Widget:
		if !filter.ReceiveWidgetEvents {
			return false
		}
	}

	// Check widget ID filter if specified
	if len(filter.WidgetIds) > 0 {
		widgetID := d.getEventWidgetID(event)
		if widgetID == "" {
			return false
		}

		found := false
		for _, id := range filter.WidgetIds {
			if id == widgetID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// getEventWidgetID extracts the widget ID from an event if present
func (d *EventDispatcher) getEventWidgetID(event *pb.Event) string {
	switch e := event.Event.(type) {
	case *pb.Event_Key:
		if e.Key.WidgetId != nil {
			return e.Key.WidgetId.Id
		}
	case *pb.Event_Mouse:
		if e.Mouse.WidgetId != nil {
			return e.Mouse.WidgetId.Id
		}
	case *pb.Event_Widget:
		if e.Widget.WidgetId != nil {
			return e.Widget.WidgetId.Id
		}
	}
	return ""
}

// UpdateFilter updates the filter for a session's subscription
func (d *EventDispatcher) UpdateFilter(sessionID string, filter *pb.EventFilter) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	sub, ok := d.subscribers[sessionID]
	if !ok {
		return false
	}

	sub.Filter = filter
	return true
}

// CreateEvent creates a new event with timestamp
func CreateEvent(sessionID string, eventData interface{}) *pb.Event {
	event := &pb.Event{
		SessionId: &pb.SessionId{Id: sessionID},
		Timestamp: time.Now().UnixNano(),
	}

	switch e := eventData.(type) {
	case *pb.KeyEvent:
		event.Event = &pb.Event_Key{Key: e}
	case *pb.MouseEvent:
		event.Event = &pb.Event_Mouse{Mouse: e}
	case *pb.ResizeEvent:
		event.Event = &pb.Event_Resize{Resize: e}
	case *pb.FocusEvent:
		event.Event = &pb.Event_Focus{Focus: e}
	case *pb.WidgetEvent:
		event.Event = &pb.Event_Widget{Widget: e}
	}

	return event
}

