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
	"google.golang.org/protobuf/proto"
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

	// Read live state from primitive (for editable widgets)
	// Must be done on tview's update thread since tview is not thread-safe
	var props *pb.WidgetProperties
	done := make(chan struct{})
	s.server.App().QueueUpdate(func() {
		defer close(done)
		props = s.enrichPropertiesFromPrimitive(info)
	})
	<-done

	return &pb.GetPropertiesResponse{
		Properties: props,
	}, nil
}

// enrichPropertiesFromPrimitive reads live state from primitives and returns
// a fresh WidgetProperties with both stored and live state.
// This is needed because some state (like text content) is managed by tview internally.
// We clone the stored properties to avoid mutating the registry's copy and to
// ensure proto serialization caches are clean.
func (s *WidgetService) enrichPropertiesFromPrimitive(info *server.WidgetInfo) *pb.WidgetProperties {
	var props *pb.WidgetProperties
	if info.Properties != nil {
		props = proto.Clone(info.Properties).(*pb.WidgetProperties)
	} else {
		props = &pb.WidgetProperties{}
	}

	switch info.Type {
	case pb.WidgetType_WIDGET_TEXT_AREA:
		if ta, ok := info.Primitive.(*tview.TextArea); ok {
			text := ta.GetText()
			if taProps, ok := props.TypeProperties.(*pb.WidgetProperties_TextArea); ok && taProps.TextArea != nil {
				taProps.TextArea.Text = &text
			} else {
				props.TypeProperties = &pb.WidgetProperties_TextArea{
					TextArea: &pb.TextAreaProperties{Text: &text},
				}
			}
		}
	case pb.WidgetType_WIDGET_INPUT_FIELD:
		if input, ok := info.Primitive.(*tview.InputField); ok {
			text := input.GetText()
			slog.Debug("enrichProperties: InputField", "widget_id", info.ID, "text", text, "textLen", len(text))
			if ifProps, ok := props.TypeProperties.(*pb.WidgetProperties_InputField); ok && ifProps.InputField != nil {
				ifProps.InputField.Text = &text
			} else {
				props.TypeProperties = &pb.WidgetProperties_InputField{
					InputField: &pb.InputFieldProperties{Text: &text},
				}
			}
		}
	case pb.WidgetType_WIDGET_TEXT_VIEW:
		if tv, ok := info.Primitive.(*tview.TextView); ok {
			text := tv.GetText(false)
			if tvProps, ok := props.TypeProperties.(*pb.WidgetProperties_TextView); ok && tvProps.TextView != nil {
				tvProps.TextView.Text = &text
			} else {
				props.TypeProperties = &pb.WidgetProperties_TextView{
					TextView: &pb.TextViewProperties{Text: &text},
				}
			}
		}
	}

	return props
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
	case pb.WidgetType_WIDGET_DROPDOWN:
		return s.createDropdown(props), nil
	case pb.WidgetType_WIDGET_TEXT_AREA:
		return s.createTextArea(props), nil
	case pb.WidgetType_WIDGET_MODAL:
		return s.createModal(props), nil
	case pb.WidgetType_WIDGET_IMAGE:
		return s.createImage(props), nil
	case pb.WidgetType_WIDGET_PROGRESS_BAR:
		return s.createProgressBar(props), nil
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
	case pb.WidgetType_WIDGET_DROPDOWN:
		if dd, ok := primitive.(*tview.DropDown); ok {
			return s.applyDropdownProperties(dd, props)
		}
	case pb.WidgetType_WIDGET_TEXT_AREA:
		if ta, ok := primitive.(*tview.TextArea); ok {
			return s.applyTextAreaProperties(ta, props)
		}
	case pb.WidgetType_WIDGET_MODAL:
		if modal, ok := primitive.(*tview.Modal); ok {
			return s.applyModalProperties(modal, props)
		}
	case pb.WidgetType_WIDGET_IMAGE:
		if img, ok := primitive.(*tview.Image); ok {
			return s.applyImageProperties(img, props)
		}
	case pb.WidgetType_WIDGET_PROGRESS_BAR:
		if pb, ok := primitive.(*ProgressBar); ok {
			return s.applyProgressBarProperties(pb, props)
		}
	}

	return nil
}


// AddChild adds a child widget to a container.
// Note: For Flex, Grid, and Pages containers, use LayoutService methods instead
// (AddFlexChild, AddGridChild, AddPage).
func (s *WidgetService) AddChild(ctx context.Context, req *pb.AddChildRequest) (*pb.AddChildResponse, error) {
	return nil, status.Error(codes.Unimplemented,
		"use LayoutService for container operations: AddFlexChild, AddGridChild, or AddPage")
}

// RemoveChild removes a child widget from a container.
// Note: For Flex, Grid, and Pages containers, use LayoutService methods instead
// (RemoveFlexChild, RemoveGridChild, RemovePage).
func (s *WidgetService) RemoveChild(ctx context.Context, req *pb.RemoveChildRequest) (*pb.RemoveChildResponse, error) {
	return nil, status.Error(codes.Unimplemented,
		"use LayoutService for container operations: RemoveFlexChild, RemoveGridChild, or RemovePage")
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
	case pb.WidgetType_WIDGET_DROPDOWN:
		if dd, ok := primitive.(*tview.DropDown); ok {
			dd.SetSelectedFunc(func(text string, index int) {
				data := map[string]string{
					"text":  text,
					"index": fmt.Sprintf("%d", index),
				}
				s.emitWidgetEvent(sessionID, widgetID, pb.WidgetEventType_WIDGET_SELECTED, data)
			})
		}
	case pb.WidgetType_WIDGET_TEXT_AREA:
		if ta, ok := primitive.(*tview.TextArea); ok {
			ta.SetChangedFunc(func() {
				data := map[string]string{"text": ta.GetText()}
				s.emitWidgetEvent(sessionID, widgetID, pb.WidgetEventType_WIDGET_CHANGED, data)
			})
		}
	case pb.WidgetType_WIDGET_MODAL:
		if modal, ok := primitive.(*tview.Modal); ok {
			modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				data := map[string]string{
					"button_index": fmt.Sprintf("%d", buttonIndex),
					"button_label": buttonLabel,
				}
				s.emitWidgetEvent(sessionID, widgetID, pb.WidgetEventType_WIDGET_SUBMITTED, data)
			})
		}
	}
}

// emitWidgetEvent emits a widget-specific event
func (s *WidgetService) emitWidgetEvent(sessionID, widgetID string, eventType pb.WidgetEventType, data map[string]string) {
	event := &pb.Event{
		SessionId: &pb.SessionId{Id: sessionID},
		Timestamp: time.Now().UnixNano(),
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




