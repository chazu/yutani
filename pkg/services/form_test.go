package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
)

func TestFormService_AddField(t *testing.T) {
	srv := setupTestServer(t)

	formService := NewFormService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a form widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FORM,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Test AddField - Input field
	fieldWidth := int32(30)
	initialValue := "test value"
	resp, err := formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId:    &pb.SessionId{Id: sessionID},
		WidgetId:     &pb.WidgetId{Id: widgetID},
		Label:        "Username",
		FieldType:    pb.FormFieldType_FORM_FIELD_INPUT,
		FieldWidth:   &fieldWidth,
		InitialValue: &initialValue,
	})
	if err != nil {
		t.Fatalf("AddField failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.FieldIndex != 0 {
		t.Errorf("Expected field index 0, got %d", resp.FieldIndex)
	}

	// Test AddField - Password field
	resp2, err := formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Password",
		FieldType: pb.FormFieldType_FORM_FIELD_PASSWORD,
	})
	if err != nil {
		t.Fatalf("AddField (password) failed: %v", err)
	}
	if resp2.FieldIndex != 1 {
		t.Errorf("Expected field index 1, got %d", resp2.FieldIndex)
	}

	// Test AddField - Checkbox
	resp3, err := formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Remember me",
		FieldType: pb.FormFieldType_FORM_FIELD_CHECKBOX,
	})
	if err != nil {
		t.Fatalf("AddField (checkbox) failed: %v", err)
	}
	if resp3.FieldIndex != 2 {
		t.Errorf("Expected field index 2, got %d", resp3.FieldIndex)
	}

	// Test AddField - Dropdown
	resp4, err := formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId:       &pb.SessionId{Id: sessionID},
		WidgetId:        &pb.WidgetId{Id: widgetID},
		Label:           "Country",
		FieldType:       pb.FormFieldType_FORM_FIELD_DROPDOWN,
		DropdownOptions: []string{"USA", "Canada", "Mexico"},
	})
	if err != nil {
		t.Fatalf("AddField (dropdown) failed: %v", err)
	}
	if resp4.FieldIndex != 3 {
		t.Errorf("Expected field index 3, got %d", resp4.FieldIndex)
	}
}

func TestFormService_AddButton(t *testing.T) {
	srv := setupTestServer(t)

	formService := NewFormService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a form widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FORM,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Test AddButton
	resp, err := formService.AddButton(ctx, &pb.AddButtonRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Submit",
	})
	if err != nil {
		t.Fatalf("AddButton failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.ButtonIndex != 0 {
		t.Errorf("Expected button index 0, got %d", resp.ButtonIndex)
	}

	// Add another button
	resp2, err := formService.AddButton(ctx, &pb.AddButtonRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Cancel",
	})
	if err != nil {
		t.Fatalf("AddButton (2nd) failed: %v", err)
	}
	if resp2.ButtonIndex != 1 {
		t.Errorf("Expected button index 1, got %d", resp2.ButtonIndex)
	}
}

func TestFormService_GetSetFieldValue(t *testing.T) {
	srv := setupTestServer(t)

	formService := NewFormService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a form widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FORM,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Add an input field
	initialValue := "initial"
	_, err = formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId:    &pb.SessionId{Id: sessionID},
		WidgetId:     &pb.WidgetId{Id: widgetID},
		Label:        "Name",
		FieldType:    pb.FormFieldType_FORM_FIELD_INPUT,
		InitialValue: &initialValue,
	})
	if err != nil {
		t.Fatalf("AddField failed: %v", err)
	}

	// Test GetFieldValue
	getResp, err := formService.GetFieldValue(ctx, &pb.GetFieldValueRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		WidgetId:   &pb.WidgetId{Id: widgetID},
		FieldIndex: 0,
	})
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if getResp.Value != "initial" {
		t.Errorf("Expected value 'initial', got %q", getResp.Value)
	}

	// Test SetFieldValue
	setResp, err := formService.SetFieldValue(ctx, &pb.SetFieldValueRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		WidgetId:   &pb.WidgetId{Id: widgetID},
		FieldIndex: 0,
		Value:      "updated",
	})
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if !setResp.Success {
		t.Error("Expected success to be true")
	}

	// Verify the value was updated
	getResp2, err := formService.GetFieldValue(ctx, &pb.GetFieldValueRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		WidgetId:   &pb.WidgetId{Id: widgetID},
		FieldIndex: 0,
	})
	if err != nil {
		t.Fatalf("GetFieldValue (2nd) failed: %v", err)
	}
	if getResp2.Value != "updated" {
		t.Errorf("Expected value 'updated', got %q", getResp2.Value)
	}

	// Test checkbox field
	_, err = formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Enabled",
		FieldType: pb.FormFieldType_FORM_FIELD_CHECKBOX,
	})
	if err != nil {
		t.Fatalf("AddField (checkbox) failed: %v", err)
	}

	// Get checkbox value (should be false)
	getResp3, err := formService.GetFieldValue(ctx, &pb.GetFieldValueRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		WidgetId:   &pb.WidgetId{Id: widgetID},
		FieldIndex: 1,
	})
	if err != nil {
		t.Fatalf("GetFieldValue (checkbox) failed: %v", err)
	}
	if getResp3.Value != "false" {
		t.Errorf("Expected value 'false', got %q", getResp3.Value)
	}

	// Set checkbox to true
	_, err = formService.SetFieldValue(ctx, &pb.SetFieldValueRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		WidgetId:   &pb.WidgetId{Id: widgetID},
		FieldIndex: 1,
		Value:      "true",
	})
	if err != nil {
		t.Fatalf("SetFieldValue (checkbox) failed: %v", err)
	}

	// Verify checkbox is now true
	getResp4, err := formService.GetFieldValue(ctx, &pb.GetFieldValueRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		WidgetId:   &pb.WidgetId{Id: widgetID},
		FieldIndex: 1,
	})
	if err != nil {
		t.Fatalf("GetFieldValue (checkbox 2nd) failed: %v", err)
	}
	if getResp4.Value != "true" {
		t.Errorf("Expected value 'true', got %q", getResp4.Value)
	}
}

func TestFormService_Clear(t *testing.T) {
	srv := setupTestServer(t)

	formService := NewFormService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a form widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FORM,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Add some fields and buttons
	_, err = formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Name",
		FieldType: pb.FormFieldType_FORM_FIELD_INPUT,
	})
	if err != nil {
		t.Fatalf("AddField failed: %v", err)
	}

	_, err = formService.AddButton(ctx, &pb.AddButtonRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Submit",
	})
	if err != nil {
		t.Fatalf("AddButton failed: %v", err)
	}

	// Test Clear
	resp, err := formService.Clear(ctx, &pb.ClearFormRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Verify form is cleared
	countResp, err := formService.GetItemCount(ctx, &pb.GetFormItemCountRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetItemCount failed: %v", err)
	}
	if countResp.Count != 0 {
		t.Errorf("Expected 0 items after clear, got %d", countResp.Count)
	}
}

func TestFormService_GetItemCount(t *testing.T) {
	srv := setupTestServer(t)

	formService := NewFormService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a form widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FORM,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Test GetItemCount with empty form
	resp, err := formService.GetItemCount(ctx, &pb.GetFormItemCountRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetItemCount failed: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("Expected 0 items, got %d", resp.Count)
	}

	// Add some fields
	_, err = formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Field 1",
		FieldType: pb.FormFieldType_FORM_FIELD_INPUT,
	})
	if err != nil {
		t.Fatalf("AddField (1) failed: %v", err)
	}

	_, err = formService.AddField(ctx, &pb.AddFieldRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Label:     "Field 2",
		FieldType: pb.FormFieldType_FORM_FIELD_CHECKBOX,
	})
	if err != nil {
		t.Fatalf("AddField (2) failed: %v", err)
	}

	// Test GetItemCount with 2 fields
	resp2, err := formService.GetItemCount(ctx, &pb.GetFormItemCountRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetItemCount (2nd) failed: %v", err)
	}
	if resp2.Count != 2 {
		t.Errorf("Expected 2 items, got %d", resp2.Count)
	}
}

