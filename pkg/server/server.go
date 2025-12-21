package server

import (
	"log/slog"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

const Version = "0.1.0-alpha"

// Server represents the Yutani display server
type Server struct {
	// Configuration
	maxSessions int
	mouseEnable bool
	pasteEnable bool
	testMode    bool // Use simulation screen for testing

	// TUI components
	app    *tview.Application
	screen tcell.Screen

	// Registries
	sessions *SessionRegistry
	widgets  *WidgetRegistry
	events   *EventDispatcher

	// Synchronization
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// NewServer creates a new Yutani server
func NewServer(maxSessions int, mouseEnable, pasteEnable bool) (*Server, error) {
	app := tview.NewApplication()

	s := &Server{
		maxSessions: maxSessions,
		mouseEnable: mouseEnable,
		pasteEnable: pasteEnable,
		testMode:    false,
		app:         app,
		sessions:    NewSessionRegistry(maxSessions),
		widgets:     NewWidgetRegistry(),
		events:      NewEventDispatcher(),
		stopCh:      make(chan struct{}),
	}

	return s, nil
}

// NewTestServer creates a server for testing with simulation screen
func NewTestServer(maxSessions int) (*Server, error) {
	app := tview.NewApplication()

	s := &Server{
		maxSessions: maxSessions,
		mouseEnable: true,
		pasteEnable: true,
		testMode:    true,
		app:         app,
		sessions:    NewSessionRegistry(maxSessions),
		widgets:     NewWidgetRegistry(),
		events:      NewEventDispatcher(),
		stopCh:      make(chan struct{}),
	}

	return s, nil
}

// Start starts the TUI event loop in a goroutine
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	var screen tcell.Screen
	var err error

	// Use simulation screen for testing
	if s.testMode {
		screen = tcell.NewSimulationScreen("UTF-8")
		if err := screen.Init(); err != nil {
			return err
		}
		// Set a reasonable default size for testing
		screen.SetSize(80, 24)
	} else {
		// Initialize the real screen
		screen, err = tcell.NewScreen()
		if err != nil {
			return err
		}

		if err := screen.Init(); err != nil {
			return err
		}
	}

	// Configure screen
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	s.screen = screen
	s.app.SetScreen(screen)

	// Set up input capture for event dispatching
	s.app.SetInputCapture(s.handleInputEvent)

	// Start the application in a goroutine
	go func() {
		slog.Info("Starting tview application event loop")
		if err := s.app.Run(); err != nil {
			slog.Error("Application error", "error", err)
		}
		slog.Info("tview application event loop stopped")
	}()

	// Start event polling for mouse and resize events
	go s.pollScreenEvents()

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	slog.Info("Stopping server")
	s.running = false

	// Close stop channel (only if not already closed)
	select {
	case <-s.stopCh:
		// Already closed
	default:
		close(s.stopCh)
	}

	if s.app != nil {
		s.app.Stop()
	}

	if s.screen != nil {
		// Fini() can panic if called twice on simulation screen
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Debug("Screen Fini() panic (expected in test mode)", "error", r)
				}
			}()
			s.screen.Fini()
		}()
		s.screen = nil
	}
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// App returns the tview application
func (s *Server) App() *tview.Application {
	return s.app
}

// Screen returns the tcell screen
func (s *Server) Screen() tcell.Screen {
	return s.screen
}

// Sessions returns the session registry
func (s *Server) Sessions() *SessionRegistry {
	return s.sessions
}

// Widgets returns the widget registry
func (s *Server) Widgets() *WidgetRegistry {
	return s.widgets
}

// MaxSessions returns the maximum number of sessions
func (s *Server) MaxSessions() int {
	return s.maxSessions
}

// Events returns the event dispatcher
func (s *Server) Events() *EventDispatcher {
	return s.events
}

// handleInputEvent captures keyboard events and dispatches them
func (s *Server) handleInputEvent(event *tcell.EventKey) *tcell.EventKey {
	// Dispatch key event to all sessions
	s.sessions.mu.RLock()
	sessionIDs := make([]string, 0, len(s.sessions.sessions))
	for id := range s.sessions.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	s.sessions.mu.RUnlock()

	// Convert tcell event to protobuf event
	for _, sessionID := range sessionIDs {
		pbEvent := s.convertKeyEvent(sessionID, event)
		s.events.Dispatch(pbEvent)
	}

	return event // Pass through to tview
}

// pollScreenEvents polls for mouse and resize events
func (s *Server) pollScreenEvents() {
	var lastWidth, lastHeight int

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		if s.screen == nil {
			continue
		}

		// Check for resize
		width, height := s.screen.Size()
		if width != lastWidth || height != lastHeight {
			lastWidth, lastHeight = width, height

			// Dispatch resize event to all sessions
			s.sessions.mu.RLock()
			for sessionID := range s.sessions.sessions {
				pbEvent := CreateEvent(sessionID, &pb.ResizeEvent{
					NewSize: &pb.Size{
						Width:  int32(width),
						Height: int32(height),
					},
				})
				s.events.Dispatch(pbEvent)
			}
			s.sessions.mu.RUnlock()
		}

		// Poll for events (non-blocking)
		ev := s.screen.PollEvent()
		if ev == nil {
			continue
		}

		// Handle mouse events
		if mouseEv, ok := ev.(*tcell.EventMouse); ok {
			s.handleMouseEvent(mouseEv)
		}
	}
}

// handleMouseEvent converts and dispatches mouse events
func (s *Server) handleMouseEvent(event *tcell.EventMouse) {
	// Dispatch to all sessions
	s.sessions.mu.RLock()
	sessionIDs := make([]string, 0, len(s.sessions.sessions))
	for id := range s.sessions.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	s.sessions.mu.RUnlock()

	for _, sessionID := range sessionIDs {
		pbEvent := s.convertMouseEvent(sessionID, event)
		s.events.Dispatch(pbEvent)
	}
}
