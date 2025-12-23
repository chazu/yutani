package services

import (
	"testing"
	"time"

	"github.com/chazu/yutani/pkg/server"
)

// testServerHelper creates a test server with the tview event loop running.
// This is required because QueueUpdateDraw blocks until tview processes
// the queued function, which only happens when Run() is active.
func setupTestServer(t *testing.T) *server.Server {
	t.Helper()

	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Run the tview event loop in a goroutine
	// This is necessary for QueueUpdateDraw to work
	go func() {
		_ = srv.Run()
	}()

	// Give server time to start event loop
	time.Sleep(50 * time.Millisecond)

	t.Cleanup(func() {
		srv.Stop()
	})

	return srv
}
