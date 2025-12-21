package server

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// TestServerLifecycle tests server start/stop
func TestServerLifecycle(t *testing.T) {
	// Use test server with simulation screen
	server, err := NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is running
	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// Stop server
	server.Stop()

	// Verify server stopped
	time.Sleep(100 * time.Millisecond)
	if server.IsRunning() {
		t.Error("Server should be stopped")
	}

	// Check for errors
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Server error: %v", err)
		}
	default:
	}
}

// TestServerWithSimulationScreen tests server with simulated screen
func TestServerWithSimulationScreen(t *testing.T) {
	// Create simulation screen
	simScreen := tcell.NewSimulationScreen("UTF-8")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("Failed to init simulation screen: %v", err)
	}
	defer simScreen.Fini()

	// Set screen size
	simScreen.SetSize(80, 24)

	// Create server
	server, err := NewServer(10, true, true)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Replace screen with simulation screen
	server.screen = simScreen

	// Test screen operations
	width, height := server.Screen().Size()
	if width != 80 || height != 24 {
		t.Errorf("Expected screen size 80x24, got %dx%d", width, height)
	}

	// Test setting a cell
	simScreen.SetContent(5, 5, 'X', nil, tcell.StyleDefault)
	simScreen.Show()

	// Verify cell content
	mainc, _, _, _ := simScreen.GetContent(5, 5)
	if mainc != 'X' {
		t.Errorf("Expected 'X' at (5,5), got '%c'", mainc)
	}
}

// TestSessionRegistryIntegration tests session registry with server
func TestSessionRegistryIntegration(t *testing.T) {
	server, err := NewServer(2, true, true)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create sessions
	prefs := SessionPreferences{
		ReceiveKeyEvents:    true,
		ReceiveMouseEvents:  true,
		ReceiveResizeEvents: true,
	}

	session1, err := server.Sessions().Create("client1", prefs)
	if err != nil {
		t.Fatalf("Failed to create session1: %v", err)
	}

	session2, err := server.Sessions().Create("client2", prefs)
	if err != nil {
		t.Fatalf("Failed to create session2: %v", err)
	}

	// Verify count
	if count := server.Sessions().Count(); count != 2 {
		t.Errorf("Expected 2 sessions, got %d", count)
	}

	// Test max sessions
	_, err = server.Sessions().Create("client3", prefs)
	if err == nil {
		t.Error("Expected error when exceeding max sessions")
	}

	// Delete session
	if !server.Sessions().Delete(session1.ID) {
		t.Error("Failed to delete session1")
	}

	// Verify count after deletion
	if count := server.Sessions().Count(); count != 1 {
		t.Errorf("Expected 1 session after deletion, got %d", count)
	}

	// Verify correct session remains
	if _, ok := server.Sessions().Get(session2.ID); !ok {
		t.Error("Session2 should still exist")
	}
}

// TestEventDispatcherIntegration tests event dispatcher with server
func TestEventDispatcherIntegration(t *testing.T) {
	server, err := NewServer(10, true, true)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Verify event dispatcher is initialized
	if server.Events() == nil {
		t.Fatal("Event dispatcher should be initialized")
	}

	// Test event subscription
	sessionID := "test-session"
	sub := server.Events().Subscribe(sessionID, nil)

	if sub == nil {
		t.Fatal("Expected non-nil subscription")
	}
	if sub.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, sub.SessionID)
	}
	if sub.Events == nil {
		t.Error("Expected non-nil event channel")
	}

	// Cleanup
	server.Events().Unsubscribe(sessionID)
}

