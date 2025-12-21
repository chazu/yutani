package services

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SessionService implements the SessionService gRPC service
type SessionService struct {
	pb.UnimplementedSessionServiceServer
	server *server.Server
}

// NewSessionService creates a new SessionService
func NewSessionService(srv *server.Server) *SessionService {
	return &SessionService{
		server: srv,
	}
}

// CreateSession creates a new client session
func (s *SessionService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	slog.Info("Creating new session", "client_name", req.ClientName)

	// Convert preferences
	prefs := server.SessionPreferences{
		ReceiveMouseEvents:  true,
		ReceiveKeyEvents:    true,
		ReceiveResizeEvents: true,
	}

	if req.Preferences != nil {
		prefs.ReceiveMouseEvents = req.Preferences.ReceiveMouseEvents
		prefs.ReceiveKeyEvents = req.Preferences.ReceiveKeyEvents
		prefs.ReceiveResizeEvents = req.Preferences.ReceiveResizeEvents
	}

	// Create session
	session, err := s.server.Sessions().Create(req.ClientName, prefs)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return nil, status.Errorf(codes.ResourceExhausted, "failed to create session: %v", err)
	}

	slog.Info("Session created", "session_id", session.ID, "client_name", session.ClientName)

	return &pb.CreateSessionResponse{
		SessionId: &pb.SessionId{
			Id: session.ID,
		},
		ServerVersion: server.Version,
	}, nil
}

// DestroySession ends a session and cleans up its resources
func (s *SessionService) DestroySession(ctx context.Context, req *pb.DestroySessionRequest) (*pb.DestroySessionResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	slog.Info("Destroying session", "session_id", sessionID)

	// Clean up widgets owned by this session
	widgetCount := s.server.Widgets().DeleteBySession(sessionID)
	slog.Debug("Deleted widgets for session", "session_id", sessionID, "count", widgetCount)

	// Delete the session
	deleted := s.server.Sessions().Delete(sessionID)
	if !deleted {
		slog.Warn("Session not found", "session_id", sessionID)
		return &pb.DestroySessionResponse{Success: false}, nil
	}

	slog.Info("Session destroyed", "session_id", sessionID)
	return &pb.DestroySessionResponse{Success: true}, nil
}

// GetServerInfo returns server information
func (s *SessionService) GetServerInfo(ctx context.Context, req *pb.GetServerInfoRequest) (*pb.GetServerInfoResponse, error) {
	// Get screen size
	var screenSize *pb.Size
	if s.server.Screen() != nil {
		width, height := s.server.Screen().Size()
		screenSize = &pb.Size{
			Width:  int32(width),
			Height: int32(height),
		}
	} else {
		screenSize = &pb.Size{Width: 0, Height: 0}
	}

	// Build capabilities
	capabilities := &pb.ServerCapabilities{
		MouseSupport:  true,
		PasteSupport:  true,
		TrueColor:     true,
		SupportedWidgets: []string{
			"Box", "TextView", "InputField", "TextArea",
			"Button", "Checkbox", "DropDown", "List",
			"Table", "TreeView", "Form", "Flex",
			"Grid", "Pages", "Modal",
		},
	}

	return &pb.GetServerInfoResponse{
		Version:        server.Version,
		ActiveSessions: int32(s.server.Sessions().Count()),
		MaxSessions:    int32(s.server.MaxSessions()),
		ScreenSize:     screenSize,
		Capabilities:   capabilities,
	}, nil
}

// Ping responds to health check pings
func (s *SessionService) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Timestamp:       req.Timestamp,
		ServerTimestamp: time.Now().UnixNano(),
	}, nil
}

