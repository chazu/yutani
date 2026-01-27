package services

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultWaitTimeout = 5000 // 5 seconds

// TestService implements the TestService gRPC service for headless UI testing.
// All methods require the server to be in test mode (using SimulationScreen).
type TestService struct {
	pb.UnimplementedTestServiceServer
	server *server.Server
}

// NewTestService creates a new TestService
func NewTestService(srv *server.Server) *TestService {
	return &TestService{
		server: srv,
	}
}

// InjectKey injects a single keystroke into the tview input queue
func (s *TestService) InjectKey(ctx context.Context, req *pb.InjectKeyRequest) (*pb.InjectKeyResponse, error) {
	if !s.server.IsTestMode() {
		return nil, status.Error(codes.FailedPrecondition, "server not in test mode")
	}

	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	sim := s.server.GetSimulationScreen()
	if sim == nil {
		return nil, status.Error(codes.Internal, "simulation screen not available")
	}

	// Convert protobuf key to tcell key
	tcellKey := convertProtoKeyToTcell(req.Key)
	tcellMod := convertProtoModifiersToTcell(req.Modifiers)

	// Get the rune for KEY_RUNE events
	var r rune
	if req.Key == pb.Key_KEY_RUNE && len(req.Rune) > 0 {
		r = []rune(req.Rune)[0]
	}

	// Inject the key
	sim.InjectKey(tcellKey, r, tcellMod)

	return &pb.InjectKeyResponse{Success: true}, nil
}

// InjectText injects a string as a sequence of KEY_RUNE events
func (s *TestService) InjectText(ctx context.Context, req *pb.InjectTextRequest) (*pb.InjectTextResponse, error) {
	if !s.server.IsTestMode() {
		return nil, status.Error(codes.FailedPrecondition, "server not in test mode")
	}

	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	sim := s.server.GetSimulationScreen()
	if sim == nil {
		return nil, status.Error(codes.Internal, "simulation screen not available")
	}

	// Inject each character as a KEY_RUNE event
	runes := []rune(req.Text)
	for _, r := range runes {
		sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	}

	return &pb.InjectTextResponse{
		Success:            true,
		CharactersInjected: int32(len(runes)),
	}, nil
}

// InjectMouse injects a mouse event into the tview input queue
func (s *TestService) InjectMouse(ctx context.Context, req *pb.InjectMouseRequest) (*pb.InjectMouseResponse, error) {
	if !s.server.IsTestMode() {
		return nil, status.Error(codes.FailedPrecondition, "server not in test mode")
	}

	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	sim := s.server.GetSimulationScreen()
	if sim == nil {
		return nil, status.Error(codes.Internal, "simulation screen not available")
	}

	// Convert protobuf button to tcell ButtonMask
	tcellButton := convertProtoButtonToTcell(req.Button)
	tcellMod := convertProtoModifiersToTcell(req.Modifiers)

	// Inject the mouse event
	sim.InjectMouse(int(req.X), int(req.Y), tcellButton, tcellMod)

	return &pb.InjectMouseResponse{Success: true}, nil
}

// WaitForIdle blocks until tview has processed all pending events
func (s *TestService) WaitForIdle(ctx context.Context, req *pb.WaitForIdleRequest) (*pb.WaitForIdleResponse, error) {
	if !s.server.IsTestMode() {
		return nil, status.Error(codes.FailedPrecondition, "server not in test mode")
	}

	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	timeout := req.TimeoutMs
	if timeout <= 0 {
		timeout = defaultWaitTimeout
	}

	// Create a done channel for the QueueUpdate callback
	done := make(chan struct{})

	// Queue an update - when this executes, all prior events have been processed
	s.server.App().QueueUpdate(func() {
		close(done)
	})

	// Wait for the callback or timeout
	select {
	case <-done:
		return &pb.WaitForIdleResponse{Success: true, TimedOut: false}, nil
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return &pb.WaitForIdleResponse{Success: false, TimedOut: true}, nil
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "request canceled")
	}
}

// IsTestMode checks if the server is running in test mode
func (s *TestService) IsTestMode(ctx context.Context, req *pb.IsTestModeRequest) (*pb.IsTestModeResponse, error) {
	return &pb.IsTestModeResponse{TestMode: s.server.IsTestMode()}, nil
}

// convertProtoKeyToTcell converts a protobuf Key to tcell.Key
func convertProtoKeyToTcell(key pb.Key) tcell.Key {
	switch key {
	case pb.Key_KEY_ENTER:
		return tcell.KeyEnter
	case pb.Key_KEY_TAB:
		return tcell.KeyTab
	case pb.Key_KEY_BACKTAB:
		return tcell.KeyBacktab
	case pb.Key_KEY_BACKSPACE:
		return tcell.KeyBackspace2
	case pb.Key_KEY_ESCAPE:
		return tcell.KeyEscape
	case pb.Key_KEY_UP:
		return tcell.KeyUp
	case pb.Key_KEY_DOWN:
		return tcell.KeyDown
	case pb.Key_KEY_LEFT:
		return tcell.KeyLeft
	case pb.Key_KEY_RIGHT:
		return tcell.KeyRight
	case pb.Key_KEY_HOME:
		return tcell.KeyHome
	case pb.Key_KEY_END:
		return tcell.KeyEnd
	case pb.Key_KEY_PGUP:
		return tcell.KeyPgUp
	case pb.Key_KEY_PGDN:
		return tcell.KeyPgDn
	case pb.Key_KEY_DELETE:
		return tcell.KeyDelete
	case pb.Key_KEY_INSERT:
		return tcell.KeyInsert
	case pb.Key_KEY_F1:
		return tcell.KeyF1
	case pb.Key_KEY_F2:
		return tcell.KeyF2
	case pb.Key_KEY_F3:
		return tcell.KeyF3
	case pb.Key_KEY_F4:
		return tcell.KeyF4
	case pb.Key_KEY_F5:
		return tcell.KeyF5
	case pb.Key_KEY_F6:
		return tcell.KeyF6
	case pb.Key_KEY_F7:
		return tcell.KeyF7
	case pb.Key_KEY_F8:
		return tcell.KeyF8
	case pb.Key_KEY_F9:
		return tcell.KeyF9
	case pb.Key_KEY_F10:
		return tcell.KeyF10
	case pb.Key_KEY_F11:
		return tcell.KeyF11
	case pb.Key_KEY_F12:
		return tcell.KeyF12
	case pb.Key_KEY_RUNE:
		return tcell.KeyRune
	default:
		return tcell.KeyRune
	}
}

// convertProtoModifiersToTcell converts protobuf modifiers to tcell.ModMask
func convertProtoModifiersToTcell(mods []pb.Modifier) tcell.ModMask {
	var result tcell.ModMask

	for _, mod := range mods {
		switch mod {
		case pb.Modifier_MOD_SHIFT:
			result |= tcell.ModShift
		case pb.Modifier_MOD_CTRL:
			result |= tcell.ModCtrl
		case pb.Modifier_MOD_ALT:
			result |= tcell.ModAlt
		case pb.Modifier_MOD_META:
			result |= tcell.ModMeta
		}
	}

	return result
}

// convertProtoButtonToTcell converts a protobuf MouseButton to tcell.ButtonMask
func convertProtoButtonToTcell(button pb.MouseButton) tcell.ButtonMask {
	switch button {
	case pb.MouseButton_MOUSE_PRIMARY:
		return tcell.ButtonPrimary
	case pb.MouseButton_MOUSE_SECONDARY:
		return tcell.ButtonSecondary
	case pb.MouseButton_MOUSE_MIDDLE:
		return tcell.ButtonMiddle
	case pb.MouseButton_MOUSE_WHEEL_UP:
		return tcell.WheelUp
	case pb.MouseButton_MOUSE_WHEEL_DOWN:
		return tcell.WheelDown
	case pb.MouseButton_MOUSE_WHEEL_LEFT:
		return tcell.WheelLeft
	case pb.MouseButton_MOUSE_WHEEL_RIGHT:
		return tcell.WheelRight
	case pb.MouseButton_MOUSE_NONE:
		return tcell.ButtonNone
	default:
		return tcell.ButtonNone
	}
}
