package services

import (
	"context"
	"log/slog"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EventService implements the EventService gRPC service
type EventService struct {
	pb.UnimplementedEventServiceServer
	server *server.Server
}

// NewEventService creates a new EventService
func NewEventService(srv *server.Server) *EventService {
	return &EventService{
		server: srv,
	}
}

// Subscribe creates a streaming subscription to events
func (s *EventService) Subscribe(req *pb.SubscribeRequest, stream pb.EventService_SubscribeServer) error {
	if req.SessionId == nil {
		return status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id

	// Verify session exists
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return status.Error(codes.NotFound, "session not found")
	}

	slog.Info("Client subscribing to events", "session_id", sessionID)

	// Create subscription with filter
	filter := req.Filter
	if filter == nil {
		// Default filter: receive all events
		filter = &pb.EventFilter{
			ReceiveKeyEvents:    true,
			ReceiveMouseEvents:  true,
			ReceiveResizeEvents: true,
			ReceiveFocusEvents:  true,
			ReceiveWidgetEvents: true,
		}
	}

	sub := s.server.Events().Subscribe(sessionID, filter)
	defer s.server.Events().Unsubscribe(sessionID)

	// Stream events to client
	for {
		select {
		case event, ok := <-sub.Events:
			if !ok {
				// Channel closed
				slog.Info("Event channel closed", "session_id", sessionID)
				return nil
			}

			// Send event to client
			if err := stream.Send(event); err != nil {
				slog.Error("Failed to send event", "session_id", sessionID, "error", err)
				return status.Errorf(codes.Internal, "failed to send event: %v", err)
			}

		case <-stream.Context().Done():
			// Client disconnected
			slog.Info("Client disconnected from event stream", "session_id", sessionID)
			return stream.Context().Err()

		case <-sub.Done:
			// Subscription cancelled
			slog.Info("Subscription cancelled", "session_id", sessionID)
			return nil
		}
	}
}

// InjectEvent allows injecting synthetic events (for testing)
func (s *EventService) InjectEvent(ctx context.Context, req *pb.InjectEventRequest) (*pb.InjectEventResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id

	// Verify session exists
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	// Ensure event has correct session ID
	req.Event.SessionId = req.SessionId

	// Dispatch the event
	s.server.Events().Dispatch(req.Event)

	slog.Debug("Injected event", "session_id", sessionID, "event_type", req.Event.Event)

	return &pb.InjectEventResponse{Success: true}, nil
}

// SetEventFilter updates the event filter for a session
func (s *EventService) SetEventFilter(ctx context.Context, req *pb.SetEventFilterRequest) (*pb.SetEventFilterResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id

	// Verify session exists
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	if req.Filter == nil {
		return nil, status.Error(codes.InvalidArgument, "filter is required")
	}

	// Update the filter
	success := s.server.Events().UpdateFilter(sessionID, req.Filter)
	if !success {
		return nil, status.Error(codes.NotFound, "no active subscription for session")
	}

	slog.Info("Updated event filter", "session_id", sessionID)

	return &pb.SetEventFilterResponse{Success: true}, nil
}

