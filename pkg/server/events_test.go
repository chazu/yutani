package server

import (
	"testing"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestEventDispatcher_Subscribe(t *testing.T) {
	dispatcher := NewEventDispatcher()
	sessionID := "test-session"

	filter := &pb.EventFilter{
		ReceiveKeyEvents: true,
	}

	sub := dispatcher.Subscribe(sessionID, filter)

	if sub == nil {
		t.Fatal("Expected subscription, got nil")
	}

	if sub.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, sub.SessionID)
	}

	if sub.Filter != filter {
		t.Error("Filter not set correctly")
	}

	if sub.Events == nil {
		t.Error("Events channel is nil")
	}

	if sub.Done == nil {
		t.Error("Done channel is nil")
	}
}

func TestEventDispatcher_Dispatch(t *testing.T) {
	dispatcher := NewEventDispatcher()
	sessionID := "test-session"

	filter := &pb.EventFilter{
		ReceiveKeyEvents: true,
	}

	sub := dispatcher.Subscribe(sessionID, filter)

	// Create a key event
	event := CreateEvent(sessionID, &pb.KeyEvent{
		Key:  pb.Key_KEY_ENTER,
		Rune: "\n",
	})

	// Dispatch the event
	dispatcher.Dispatch(event)

	// Receive the event with timeout
	select {
	case received := <-sub.Events:
		if received.SessionId.Id != sessionID {
			t.Errorf("Expected session ID %s, got %s", sessionID, received.SessionId.Id)
		}
		if received.Event == nil {
			t.Error("Event is nil")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

func TestEventDispatcher_FilterKeyEvents(t *testing.T) {
	dispatcher := NewEventDispatcher()
	sessionID := "test-session"

	// Filter that only accepts mouse events
	filter := &pb.EventFilter{
		ReceiveKeyEvents:   false,
		ReceiveMouseEvents: true,
	}

	sub := dispatcher.Subscribe(sessionID, filter)

	// Dispatch a key event (should be filtered out)
	keyEvent := CreateEvent(sessionID, &pb.KeyEvent{
		Key: pb.Key_KEY_ENTER,
	})
	dispatcher.Dispatch(keyEvent)

	// Should not receive the event
	select {
	case <-sub.Events:
		t.Error("Received key event when filter should have blocked it")
	case <-time.After(50 * time.Millisecond):
		// Expected - no event received
	}

	// Dispatch a mouse event (should pass through)
	mouseEvent := CreateEvent(sessionID, &pb.MouseEvent{
		Action: pb.MouseAction_MOUSE_CLICK,
		Position: &pb.Position{
			X: 10,
			Y: 20,
		},
	})
	dispatcher.Dispatch(mouseEvent)

	// Should receive the event
	select {
	case received := <-sub.Events:
		if _, ok := received.Event.(*pb.Event_Mouse); !ok {
			t.Error("Expected mouse event")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for mouse event")
	}
}

func TestEventDispatcher_UpdateFilter(t *testing.T) {
	dispatcher := NewEventDispatcher()
	sessionID := "test-session"

	// Initial filter
	filter1 := &pb.EventFilter{
		ReceiveKeyEvents: true,
	}

	sub := dispatcher.Subscribe(sessionID, filter1)

	// Update filter
	filter2 := &pb.EventFilter{
		ReceiveMouseEvents: true,
	}

	success := dispatcher.UpdateFilter(sessionID, filter2)
	if !success {
		t.Error("Failed to update filter")
	}

	if sub.Filter != filter2 {
		t.Error("Filter not updated")
	}
}

func TestEventDispatcher_Unsubscribe(t *testing.T) {
	dispatcher := NewEventDispatcher()
	sessionID := "test-session"

	filter := &pb.EventFilter{
		ReceiveKeyEvents: true,
	}

	sub := dispatcher.Subscribe(sessionID, filter)

	// Unsubscribe
	dispatcher.Unsubscribe(sessionID)

	// Channels should be closed
	select {
	case _, ok := <-sub.Events:
		if ok {
			t.Error("Events channel should be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for channel close")
	}
}

