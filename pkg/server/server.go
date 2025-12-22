package server

import (
	"fmt"
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

// Start initializes the TUI (does not block)
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	// For test mode, create and initialize a simulation screen
	if s.testMode {
		screen := tcell.NewSimulationScreen("UTF-8")
		if err := screen.Init(); err != nil {
			return err
		}
		screen.SetSize(80, 24)
		s.screen = screen
		s.app.SetScreen(screen)
	}
	// For normal mode, let tview create and manage the screen in Run()

	// Create a placeholder widget to show when no client is connected
	placeholder := tview.NewTextView().
		SetText("[yellow::b]Yutani Display Server[-::-]\n\n" +
			"Waiting for client connections...\n\n" +
			"Server is running on " + "localhost:7755" + "\n" +
			"Press [red::b]Ctrl+C[-::-] to exit").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	placeholder.SetBorder(true).SetTitle(" Yutani v" + Version + " ")

	// Set the placeholder as the initial root
	s.app.SetRoot(placeholder, true)

	// Set up input capture for event dispatching AND Ctrl+C handling
	s.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle Ctrl+C to allow graceful shutdown
		if event.Key() == tcell.KeyCtrlC {
			s.app.Stop()
			return nil
		}
		// Dispatch other key events
		return s.handleInputEvent(event)
	})

	return nil
}

// Run starts the TUI application (blocking call)
// This should be called from the main goroutine
func (s *Server) Run() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("server not started")
	}
	s.mu.Unlock()

	// Set a callback to capture the screen after tview initializes it
	// This is needed for event polling
	s.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		if s.screen == nil && !s.testMode {
			s.mu.Lock()
			s.screen = screen
			s.mu.Unlock()
			// Start event polling now that we have the screen
			go s.pollScreenEvents()
		}
		return false
	})

	// Run the application (blocks until app.Stop() is called)
	// tview will create and initialize the screen automatically
	return s.app.Run()
}

// Stop stops the server
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

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
					// Ignore panic - expected in test mode
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
