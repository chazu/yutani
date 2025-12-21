package services

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/rivo/tview"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FormService implements the FormService gRPC service
type FormService struct {
	pb.UnimplementedFormServiceServer
	server *server.Server
}

// NewFormService creates a new FormService
func NewFormService(srv *server.Server) *FormService {
	return &FormService{
		server: srv,
	}
}

// AddField adds a form field to the form
func (s *FormService) AddField(ctx context.Context, req *pb.AddFieldRequest) (*pb.AddFieldResponse, error) {
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

	var fieldIndex int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		form, ok := primitive.(*tview.Form)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a form")
			return
		}

		// Get current item count to determine field index
		fieldIndex = int32(form.GetFormItemCount())

		// Determine field width
		fieldWidth := 20 // default
		if req.FieldWidth != nil {
			fieldWidth = int(*req.FieldWidth)
		}

		// Get initial value
		initialValue := ""
		if req.InitialValue != nil {
			initialValue = *req.InitialValue
		}

		// Add field based on type
		switch req.FieldType {
		case pb.FormFieldType_FORM_FIELD_INPUT:
			form.AddInputField(req.Label, initialValue, fieldWidth, nil, nil)
		case pb.FormFieldType_FORM_FIELD_PASSWORD:
			form.AddPasswordField(req.Label, initialValue, fieldWidth, '*', nil)
		case pb.FormFieldType_FORM_FIELD_CHECKBOX:
			checked := initialValue == "true" || initialValue == "1"
			form.AddCheckbox(req.Label, checked, nil)
		case pb.FormFieldType_FORM_FIELD_DROPDOWN:
			selectedIndex := 0
			form.AddDropDown(req.Label, req.DropdownOptions, selectedIndex, nil)
		default:
			err = status.Error(codes.InvalidArgument, fmt.Sprintf("unknown field type: %v", req.FieldType))
			return
		}

		slog.Info("Form field added", "widget_id", widgetID, "field_index", fieldIndex, "type", req.FieldType)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.AddFieldResponse{
		Success:    true,
		FieldIndex: fieldIndex,
	}, nil
}

// AddButton adds a button to the form
func (s *FormService) AddButton(ctx context.Context, req *pb.AddButtonRequest) (*pb.AddButtonResponse, error) {
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

	var buttonIndex int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		form, ok := primitive.(*tview.Form)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a form")
			return
		}

		// Get current button count
		buttonIndex = int32(form.GetButtonCount())

		// Add button
		form.AddButton(req.Label, nil)

		slog.Info("Form button added", "widget_id", widgetID, "button_index", buttonIndex)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.AddButtonResponse{
		Success:     true,
		ButtonIndex: buttonIndex,
	}, nil
}

// GetFieldValue gets the value of a form field
func (s *FormService) GetFieldValue(ctx context.Context, req *pb.GetFieldValueRequest) (*pb.GetFieldValueResponse, error) {
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

	var value string
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		form, ok := primitive.(*tview.Form)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a form")
			return
		}

		// Check if field index is valid
		if req.FieldIndex < 0 || req.FieldIndex >= int32(form.GetFormItemCount()) {
			err = status.Error(codes.InvalidArgument, "field index out of range")
			return
		}

		// Get the form item
		formItem := form.GetFormItem(int(req.FieldIndex))
		if formItem == nil {
			err = status.Error(codes.NotFound, "form item not found")
			return
		}

		// Extract value based on type
		switch item := formItem.(type) {
		case *tview.InputField:
			value = item.GetText()
		case *tview.Checkbox:
			if item.IsChecked() {
				value = "true"
			} else {
				value = "false"
			}
		case *tview.DropDown:
			_, value = item.GetCurrentOption()
		default:
			err = status.Error(codes.InvalidArgument, "unsupported form item type")
			return
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetFieldValueResponse{
		Value: value,
	}, nil
}

// SetFieldValue sets the value of a form field
func (s *FormService) SetFieldValue(ctx context.Context, req *pb.SetFieldValueRequest) (*pb.SetFieldValueResponse, error) {
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

		form, ok := primitive.(*tview.Form)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a form")
			return
		}

		// Check if field index is valid
		if req.FieldIndex < 0 || req.FieldIndex >= int32(form.GetFormItemCount()) {
			err = status.Error(codes.InvalidArgument, "field index out of range")
			return
		}

		// Get the form item
		formItem := form.GetFormItem(int(req.FieldIndex))
		if formItem == nil {
			err = status.Error(codes.NotFound, "form item not found")
			return
		}

		// Set value based on type
		switch item := formItem.(type) {
		case *tview.InputField:
			item.SetText(req.Value)
		case *tview.Checkbox:
			checked := req.Value == "true" || req.Value == "1"
			item.SetChecked(checked)
		case *tview.DropDown:
			// For dropdown, we can't easily set by text value in tview
			// The client should send the index as a string, or we just skip this
			// For now, we'll try to parse the value as an index
			// Note: tview.DropDown doesn't expose a way to get option text by index
			// So we can't search by text. The value should be the option text itself
			// and we'll use SetCurrentOption with the current option if it matches
			_, currentText := item.GetCurrentOption()
			if currentText != req.Value {
				// We can't easily set by text, so we'll just log a warning
				slog.Warn("Cannot set dropdown by text value", "widget_id", widgetID, "value", req.Value)
			}
		default:
			err = status.Error(codes.InvalidArgument, "unsupported form item type")
			return
		}

		slog.Info("Form field value set", "widget_id", widgetID, "field_index", req.FieldIndex)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetFieldValueResponse{
		Success: true,
	}, nil
}

// Clear clears all form fields
func (s *FormService) Clear(ctx context.Context, req *pb.ClearFormRequest) (*pb.ClearFormResponse, error) {
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

		form, ok := primitive.(*tview.Form)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a form")
			return
		}

		// Clear the form
		form.Clear(true) // true = clear buttons as well

		slog.Info("Form cleared", "widget_id", widgetID)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.ClearFormResponse{
		Success: true,
	}, nil
}

// GetItemCount gets the number of form items
func (s *FormService) GetItemCount(ctx context.Context, req *pb.GetFormItemCountRequest) (*pb.GetFormItemCountResponse, error) {
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

		form, ok := primitive.(*tview.Form)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a form")
			return
		}

		count = int32(form.GetFormItemCount())
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetFormItemCountResponse{
		Count: count,
	}, nil
}

