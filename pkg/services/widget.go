package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/rivo/tview"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WidgetService implements the WidgetService gRPC service
type WidgetService struct {
	pb.UnimplementedWidgetServiceServer
	server *server.Server
}

// NewWidgetService creates a new WidgetService
func NewWidgetService(srv *server.Server) *WidgetService {
	return &WidgetService{
		server: srv,
	}
}

// CreateWidget creates a new widget
func (s *WidgetService) CreateWidget(ctx context.Context, req *pb.CreateWidgetRequest) (*pb.CreateWidgetResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	var widgetID string
	var err error

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Create the widget based on type
		primitive, createErr := s.createPrimitive(req.Type, req.Properties)
		if createErr != nil {
			err = createErr
			return
		}

		// Register the widget
		widgetID = s.server.Widgets().Register(sessionID, req.Type, primitive, req.Properties)

		// Wire up event handlers
		s.wireWidgetEvents(sessionID, widgetID, req.Type, primitive)

		slog.Info("Widget created", "widget_id", widgetID, "type", req.Type, "session_id", sessionID)
	})
	<-done

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateWidgetResponse{
		WidgetId: &pb.WidgetId{Id: widgetID},
		Success:  true,
	}, nil
}

// DeleteWidget deletes a widget
func (s *WidgetService) DeleteWidget(ctx context.Context, req *pb.DeleteWidgetRequest) (*pb.DeleteWidgetResponse, error) {
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

	var success bool
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)
		success = s.server.Widgets().Delete(widgetID)
		if success {
			slog.Info("Widget deleted", "widget_id", widgetID, "session_id", sessionID)
		}
	})
	<-done

	return &pb.DeleteWidgetResponse{
		Success: success,
	}, nil
}

// SetProperties sets widget properties
func (s *WidgetService) SetProperties(ctx context.Context, req *pb.SetPropertiesRequest) (*pb.SetPropertiesResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil || req.Properties == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, widget_id, and properties are required")
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

		info, ok := s.server.Widgets().GetInfo(widgetID)
		if !ok {
			err = fmt.Errorf("widget not found")
			return
		}

		// Apply properties to the primitive
		err = s.applyProperties(info.Primitive, info.Type, req.Properties)
		if err == nil {
			// Update stored properties
			s.server.Widgets().UpdateProperties(widgetID, req.Properties)
			slog.Info("Widget properties updated", "widget_id", widgetID, "session_id", sessionID)
		}
	})
	<-done

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SetPropertiesResponse{
		Success: true,
	}, nil
}

// GetProperties gets widget properties
func (s *WidgetService) GetProperties(ctx context.Context, req *pb.GetPropertiesRequest) (*pb.GetPropertiesResponse, error) {
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

	info, ok := s.server.Widgets().GetInfo(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}

	return &pb.GetPropertiesResponse{
		Properties: info.Properties,
	}, nil
}

// SetRoot sets a widget as the root (displays it)
func (s *WidgetService) SetRoot(ctx context.Context, req *pb.SetRootRequest) (*pb.SetRootResponse, error) {
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

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			return
		}

		// Set as application root
		s.server.App().SetRoot(primitive, true)
		s.server.Widgets().SetRoot(sessionID, widgetID)
		slog.Info("Widget set as root", "widget_id", widgetID, "session_id", sessionID)
	})
	<-done

	return &pb.SetRootResponse{
		Success: true,
	}, nil
}

// ListWidgets lists all widgets for a session
func (s *WidgetService) ListWidgets(ctx context.Context, req *pb.ListWidgetsRequest) (*pb.ListWidgetsResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	infos := s.server.Widgets().ListBySession(sessionID)
	widgets := make([]*pb.WidgetInfo, len(infos))

	for i, info := range infos {
		widgetInfo := &pb.WidgetInfo{
			WidgetId: &pb.WidgetId{Id: info.ID},
			Type:     info.Type,
			Children: make([]*pb.WidgetId, len(info.ChildIDs)),
		}

		if info.ParentID != "" {
			widgetInfo.ParentId = &pb.WidgetId{Id: info.ParentID}
		}

		for j, childID := range info.ChildIDs {
			widgetInfo.Children[j] = &pb.WidgetId{Id: childID}
		}

		widgets[i] = widgetInfo
	}

	return &pb.ListWidgetsResponse{
		Widgets: widgets,
	}, nil
}

// createPrimitive creates a tview primitive based on widget type
func (s *WidgetService) createPrimitive(widgetType pb.WidgetType, props *pb.WidgetProperties) (tview.Primitive, error) {
	switch widgetType {
	case pb.WidgetType_WIDGET_BOX:
		return s.createBox(props), nil
	case pb.WidgetType_WIDGET_TEXT_VIEW:
		return s.createTextView(props), nil
	case pb.WidgetType_WIDGET_INPUT_FIELD:
		return s.createInputField(props), nil
	case pb.WidgetType_WIDGET_BUTTON:
		return s.createButton(props), nil
	case pb.WidgetType_WIDGET_CHECKBOX:
		return s.createCheckbox(props), nil
	case pb.WidgetType_WIDGET_LIST:
		return s.createList(props), nil
	case pb.WidgetType_WIDGET_TABLE:
		return s.createTable(props), nil
	case pb.WidgetType_WIDGET_TREE_VIEW:
		return s.createTreeView(props), nil
	case pb.WidgetType_WIDGET_FORM:
		return s.createForm(props), nil
	case pb.WidgetType_WIDGET_FLEX:
		return s.createFlex(props), nil
	case pb.WidgetType_WIDGET_GRID:
		return s.createGrid(props), nil
	case pb.WidgetType_WIDGET_PAGES:
		return s.createPages(props), nil
	default:
		return nil, fmt.Errorf("unsupported widget type: %v", widgetType)
	}
}

// applyProperties applies properties to an existing primitive
func (s *WidgetService) applyProperties(primitive tview.Primitive, widgetType pb.WidgetType, props *pb.WidgetProperties) error {
	// Apply common Box properties
	if box, ok := primitive.(*tview.Box); ok {
		s.applyBoxProperties(box, props)
	}

	// Apply type-specific properties
	switch widgetType {
	case pb.WidgetType_WIDGET_BOX:
		// Box properties already applied above
		return nil
	case pb.WidgetType_WIDGET_TEXT_VIEW:
		if tv, ok := primitive.(*tview.TextView); ok {
			return s.applyTextViewProperties(tv, props)
		}
	case pb.WidgetType_WIDGET_INPUT_FIELD:
		if input, ok := primitive.(*tview.InputField); ok {
			return s.applyInputFieldProperties(input, props)
		}
	case pb.WidgetType_WIDGET_BUTTON:
		if btn, ok := primitive.(*tview.Button); ok {
			return s.applyButtonProperties(btn, props)
		}
	case pb.WidgetType_WIDGET_CHECKBOX:
		if cb, ok := primitive.(*tview.Checkbox); ok {
			return s.applyCheckboxProperties(cb, props)
		}
	}

	return nil
}


// AddChild adds a child widget to a container
func (s *WidgetService) AddChild(ctx context.Context, req *pb.AddChildRequest) (*pb.AddChildResponse, error) {
	if req.SessionId == nil || req.ParentId == nil || req.ChildId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, parent_id, and child_id are required")
	}

	sessionID := req.SessionId.Id
	parentID := req.ParentId.Id
	childID := req.ChildId.Id

	// Verify ownership of both widgets
	parentOwner, ok := s.server.Widgets().GetOwner(parentID)
	if !ok {
		return nil, status.Error(codes.NotFound, "parent widget not found")
	}
	if parentOwner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "parent widget not owned by session")
	}

	childOwner, ok := s.server.Widgets().GetOwner(childID)
	if !ok {
		return nil, status.Error(codes.NotFound, "child widget not found")
	}
	if childOwner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "child widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		parentInfo, ok := s.server.Widgets().GetInfo(parentID)
		if !ok {
			err = fmt.Errorf("parent widget not found")
			return
		}

		_, ok = s.server.Widgets().Get(childID)
		if !ok {
			err = fmt.Errorf("child widget not found")
			return
		}

		// Add child based on parent type
		switch parentInfo.Type {
		case pb.WidgetType_WIDGET_BOX:
			// Box doesn't support children directly
			err = fmt.Errorf("Box widget does not support children")
			return
		case pb.WidgetType_WIDGET_FLEX:
			// TODO: Implement Flex support in Phase 4
			err = fmt.Errorf("Flex widget not yet implemented")
			return
		default:
			err = fmt.Errorf("widget type %v does not support children", parentInfo.Type)
			return
		}
	})
	<-done

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Update registry
	if !s.server.Widgets().AddChild(parentID, childID) {
		return nil, status.Error(codes.Internal, "failed to update widget hierarchy")
	}

	slog.Info("Child added to widget", "parent_id", parentID, "child_id", childID, "session_id", sessionID)

	return &pb.AddChildResponse{
		Success: true,
	}, nil
}

// RemoveChild removes a child widget from a container
func (s *WidgetService) RemoveChild(ctx context.Context, req *pb.RemoveChildRequest) (*pb.RemoveChildResponse, error) {
	if req.SessionId == nil || req.ParentId == nil || req.ChildId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, parent_id, and child_id are required")
	}

	sessionID := req.SessionId.Id
	parentID := req.ParentId.Id
	childID := req.ChildId.Id

	// Verify ownership
	parentOwner, ok := s.server.Widgets().GetOwner(parentID)
	if !ok {
		return nil, status.Error(codes.NotFound, "parent widget not found")
	}
	if parentOwner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "parent widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		parentInfo, ok := s.server.Widgets().GetInfo(parentID)
		if !ok {
			err = fmt.Errorf("parent widget not found")
			return
		}

		// Remove child based on parent type
		switch parentInfo.Type {
		case pb.WidgetType_WIDGET_FLEX:
			// TODO: Implement Flex support in Phase 4
			err = fmt.Errorf("Flex widget not yet implemented")
			return
		default:
			// For now, just update the registry
			// Actual removal from tview will be implemented with container widgets
		}
	})
	<-done

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Update registry
	if !s.server.Widgets().RemoveChild(parentID, childID) {
		return nil, status.Error(codes.NotFound, "child not found in parent")
	}

	slog.Info("Child removed from widget", "parent_id", parentID, "child_id", childID, "session_id", sessionID)

	return &pb.RemoveChildResponse{
		Success: true,
	}, nil
}


// SetFocus sets focus to a widget
func (s *WidgetService) SetFocus(ctx context.Context, req *pb.SetFocusRequest) (*pb.SetFocusResponse, error) {
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

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			return
		}

		// Set focus to the widget
		s.server.App().SetFocus(primitive)
		slog.Info("Focus set to widget", "widget_id", widgetID, "session_id", sessionID)
	})
	<-done

	return &pb.SetFocusResponse{
		Success: true,
	}, nil
}

// GetFocus gets the currently focused widget
func (s *WidgetService) GetFocus(ctx context.Context, req *pb.GetFocusRequest) (*pb.GetFocusResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	var focusedWidgetID string
	done := make(chan struct{})
	s.server.App().QueueUpdate(func() {
		defer close(done)

		focused := s.server.App().GetFocus()
		if focused == nil {
			return
		}

		// Find the widget ID for this primitive
		infos := s.server.Widgets().ListBySession(sessionID)
		for _, info := range infos {
			if info.Primitive == focused {
				focusedWidgetID = info.ID
				break
			}
		}
	})
	<-done

	if focusedWidgetID == "" {
		return &pb.GetFocusResponse{}, nil
	}

	return &pb.GetFocusResponse{
		WidgetId: &pb.WidgetId{Id: focusedWidgetID},
	}, nil
}


// wireWidgetEvents wires up event handlers for interactive widgets
func (s *WidgetService) wireWidgetEvents(sessionID, widgetID string, widgetType pb.WidgetType, primitive tview.Primitive) {
	switch widgetType {
	case pb.WidgetType_WIDGET_BUTTON:
		if btn, ok := primitive.(*tview.Button); ok {
			btn.SetSelectedFunc(func() {
				s.emitWidgetEvent(sessionID, widgetID, pb.WidgetEventType_WIDGET_SUBMITTED, nil)
			})
		}
	case pb.WidgetType_WIDGET_CHECKBOX:
		if cb, ok := primitive.(*tview.Checkbox); ok {
			cb.SetChangedFunc(func(checked bool) {
				data := map[string]string{"checked": "false"}
				if checked {
					data["checked"] = "true"
				}
				s.emitWidgetEvent(sessionID, widgetID, pb.WidgetEventType_WIDGET_CHANGED, data)
			})
		}
	case pb.WidgetType_WIDGET_INPUT_FIELD:
		if input, ok := primitive.(*tview.InputField); ok {
			input.SetChangedFunc(func(text string) {
				data := map[string]string{"text": text}
				s.emitWidgetEvent(sessionID, widgetID, pb.WidgetEventType_WIDGET_CHANGED, data)
			})
			input.SetDoneFunc(func(key tcell.Key) {
				s.emitWidgetEvent(sessionID, widgetID, pb.WidgetEventType_WIDGET_DONE, nil)
			})
		}
	}
}

// emitWidgetEvent emits a widget-specific event
func (s *WidgetService) emitWidgetEvent(sessionID, widgetID string, eventType pb.WidgetEventType, data map[string]string) {
	event := &pb.Event{
		SessionId: &pb.SessionId{Id: sessionID},
		Timestamp: int64(time.Now().Unix()),
		Event: &pb.Event_Widget{
			Widget: &pb.WidgetEvent{
				WidgetId: &pb.WidgetId{Id: widgetID},
				Type:     eventType,
				Data:     data,
			},
		},
	}
	s.server.Events().Dispatch(event)
}




