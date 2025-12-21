package services

import (
	"context"
	"log/slog"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/rivo/tview"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListService implements the ListService gRPC service
type ListService struct {
	pb.UnimplementedListServiceServer
	server *server.Server
}

// NewListService creates a new ListService
func NewListService(srv *server.Server) *ListService {
	return &ListService{
		server: srv,
	}
}

// AddItem adds an item to the list
func (s *ListService) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var index int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		list, ok := primitive.(*tview.List)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a list")
			return
		}

		index = int32(list.GetItemCount())
		list.AddItem(req.MainText, req.SecondaryText, 0, nil)

		slog.Info("List item added", "widget_id", widgetID, "index", index)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.AddItemResponse{
		Success: true,
		Index:   index,
	}, nil
}

// RemoveItem removes an item from the list
func (s *ListService) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		list, ok := primitive.(*tview.List)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a list")
			return
		}

		if req.Index < 0 || req.Index >= int32(list.GetItemCount()) {
			err = status.Error(codes.InvalidArgument, "index out of range")
			return
		}

		list.RemoveItem(int(req.Index))

		slog.Info("List item removed", "widget_id", widgetID, "index", req.Index)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.RemoveItemResponse{
		Success: true,
	}, nil
}

// Clear clears all items from the list
func (s *ListService) Clear(ctx context.Context, req *pb.ClearListRequest) (*pb.ClearListResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		list, ok := primitive.(*tview.List)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a list")
			return
		}

		list.Clear()

		slog.Info("List cleared", "widget_id", widgetID)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.ClearListResponse{
		Success: true,
	}, nil
}

// GetItemCount returns the number of items in the list
func (s *ListService) GetItemCount(ctx context.Context, req *pb.GetItemCountRequest) (*pb.GetItemCountResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var count int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		list, ok := primitive.(*tview.List)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a list")
			return
		}

		count = int32(list.GetItemCount())
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetItemCountResponse{
		Count: count,
	}, nil
}

// GetSelected returns the currently selected item index
func (s *ListService) GetSelected(ctx context.Context, req *pb.GetSelectedRequest) (*pb.GetSelectedResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var index int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		list, ok := primitive.(*tview.List)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a list")
			return
		}

		index = int32(list.GetCurrentItem())
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetSelectedResponse{
		Index: index,
	}, nil
}

// SetSelected sets the selected item index
func (s *ListService) SetSelected(ctx context.Context, req *pb.SetSelectedRequest) (*pb.SetSelectedResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		list, ok := primitive.(*tview.List)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a list")
			return
		}

		if req.Index < 0 || req.Index >= int32(list.GetItemCount()) {
			err = status.Error(codes.InvalidArgument, "index out of range")
			return
		}

		list.SetCurrentItem(int(req.Index))

		slog.Info("List selection changed", "widget_id", widgetID, "index", req.Index)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetSelectedResponse{
		Success: true,
	}, nil
}

// GetItem returns an item's text
func (s *ListService) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var mainText, secondaryText string
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		list, ok := primitive.(*tview.List)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a list")
			return
		}

		if req.Index < 0 || req.Index >= int32(list.GetItemCount()) {
			err = status.Error(codes.InvalidArgument, "index out of range")
			return
		}

		mainText, secondaryText = list.GetItemText(int(req.Index))
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetItemResponse{
		MainText:      mainText,
		SecondaryText: secondaryText,
		Shortcut:      "", // tview doesn't expose shortcut text
	}, nil
}


