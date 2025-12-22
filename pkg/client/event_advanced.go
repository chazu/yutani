package client

import (
	"sync"
	"time"
)

// EventFilter provides advanced filtering for events.
type EventFilter struct {
	// Filter by event type
	Types []EventType

	// Filter by widget ID
	WidgetIDs []string

	// Filter by widget event type
	WidgetEventTypes []string

	// Custom filter function
	CustomFilter func(*Event) bool
}

// Matches returns true if the event passes this filter.
func (f *EventFilter) Matches(event *Event) bool {
	// Check event type filter
	if len(f.Types) > 0 {
		found := false
		for _, t := range f.Types {
			if event.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check widget ID filter
	if len(f.WidgetIDs) > 0 && event.Type == EventTypeWidget {
		if event.Widget == nil {
			return false
		}
		found := false
		for _, id := range f.WidgetIDs {
			if event.Widget.WidgetID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check widget event type filter
	if len(f.WidgetEventTypes) > 0 && event.Type == EventTypeWidget {
		if event.Widget == nil {
			return false
		}
		found := false
		for _, t := range f.WidgetEventTypes {
			if event.Widget.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check custom filter
	if f.CustomFilter != nil && !f.CustomFilter(event) {
		return false
	}

	return true
}

// EventMiddleware is a function that can intercept and modify events.
type EventMiddleware func(*Event) (*Event, bool)

// EventBatcher batches high-frequency events.
type EventBatcher struct {
	interval time.Duration
	buffer   []*Event
	mu       sync.Mutex
	flush    chan struct{}
	handler  EventHandler
}

// NewEventBatcher creates a new event batcher.
func NewEventBatcher(interval time.Duration, handler EventHandler) *EventBatcher {
	b := &EventBatcher{
		interval: interval,
		buffer:   make([]*Event, 0, 100),
		flush:    make(chan struct{}),
		handler:  handler,
	}
	go b.run()
	return b
}

// Add adds an event to the batch.
func (b *EventBatcher) Add(event *Event) {
	b.mu.Lock()
	b.buffer = append(b.buffer, event)
	b.mu.Unlock()
}

// run processes batches at regular intervals.
func (b *EventBatcher) run() {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flushBatch()
		case <-b.flush:
			return
		}
	}
}

// flushBatch sends all buffered events to the handler.
func (b *EventBatcher) flushBatch() {
	b.mu.Lock()
	if len(b.buffer) == 0 {
		b.mu.Unlock()
		return
	}

	events := make([]*Event, len(b.buffer))
	copy(events, b.buffer)
	b.buffer = b.buffer[:0]
	b.mu.Unlock()

	// Process all events
	for _, event := range events {
		b.handler(event)
	}
}

// Close stops the batcher.
func (b *EventBatcher) Close() {
	close(b.flush)
	b.flushBatch()
}

// EventRecorder records events for replay and debugging.
type EventRecorder struct {
	events    []*RecordedEvent
	mu        sync.RWMutex
	maxEvents int
	recording bool
}

// RecordedEvent is an event with timestamp.
type RecordedEvent struct {
	Event     *Event
	Timestamp time.Time
}

// NewEventRecorder creates a new event recorder.
func NewEventRecorder(maxEvents int) *EventRecorder {
	return &EventRecorder{
		events:    make([]*RecordedEvent, 0, maxEvents),
		maxEvents: maxEvents,
		recording: true,
	}
}

// Record records an event.
func (r *EventRecorder) Record(event *Event) {
	if !r.recording {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	recorded := &RecordedEvent{
		Event:     event,
		Timestamp: time.Now(),
	}

	r.events = append(r.events, recorded)

	// Trim if exceeds max
	if len(r.events) > r.maxEvents {
		r.events = r.events[len(r.events)-r.maxEvents:]
	}
}

// GetEvents returns all recorded events.
func (r *EventRecorder) GetEvents() []*RecordedEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*RecordedEvent, len(r.events))
	copy(events, r.events)
	return events
}

// GetEventsSince returns events since a specific time.
func (r *EventRecorder) GetEventsSince(since time.Time) []*RecordedEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*RecordedEvent
	for _, e := range r.events {
		if e.Timestamp.After(since) {
			result = append(result, e)
		}
	}
	return result
}

// Clear clears all recorded events.
func (r *EventRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = r.events[:0]
}

// Start starts recording.
func (r *EventRecorder) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recording = true
}

// Stop stops recording.
func (r *EventRecorder) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recording = false
}

// IsRecording returns true if currently recording.
func (r *EventRecorder) IsRecording() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.recording
}

// Replay replays recorded events to a handler.
func (r *EventRecorder) Replay(handler EventHandler, realtime bool) {
	events := r.GetEvents()
	if len(events) == 0 {
		return
	}

	if !realtime {
		// Replay immediately
		for _, e := range events {
			handler(e.Event)
		}
		return
	}

	// Replay with original timing
	startTime := events[0].Timestamp
	go func() {
		for _, e := range events {
			delay := e.Timestamp.Sub(startTime)
			time.Sleep(delay)
			handler(e.Event)
			startTime = e.Timestamp
		}
	}()
}

