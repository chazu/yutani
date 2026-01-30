package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/gdamore/tcell/v2"
)

// createTextAreaWidget is a test helper that creates a session and text area widget.
func createTextAreaWidget(t *testing.T, srv *server.Server, initialText string) (string, string) {
	t.Helper()

	widgetSvc := NewWidgetService(srv)
	ctx := context.Background()

	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	var props *pb.WidgetProperties
	if initialText != "" {
		props = &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_TextArea{
				TextArea: &pb.TextAreaProperties{
					Text: &initialText,
				},
			},
		}
	}

	createResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId:  &pb.SessionId{Id: session.ID},
		Type:       pb.WidgetType_WIDGET_TEXT_AREA,
		Properties: props,
	})
	if err != nil {
		t.Fatalf("Failed to create text area: %v", err)
	}

	return session.ID, createResp.WidgetId.Id
}

func TestTextAreaService_GetCursorPosition(t *testing.T) {
	srv := setupTestServer(t)
	taSvc := NewTextAreaService(srv)
	ctx := context.Background()

	sessionID, widgetID := createTextAreaWidget(t, srv, "hello\nworld")

	resp, err := taSvc.GetCursorPosition(ctx, &pb.TextAreaGetCursorRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetCursorPosition failed: %v", err)
	}

	// tview places cursor at end of text when initialized
	// "hello\nworld" -> row 1, col 5
	if resp.Row != 1 || resp.Column != 5 {
		t.Errorf("Expected cursor at end of text (1,5), got (%d,%d)", resp.Row, resp.Column)
	}
}

func TestTextAreaService_SetCursorPosition(t *testing.T) {
	srv := setupTestServer(t)
	taSvc := NewTextAreaService(srv)
	ctx := context.Background()

	sessionID, widgetID := createTextAreaWidget(t, srv, "hello\nworld")

	// Move cursor to row 1, column 3 ("wor|ld")
	setResp, err := taSvc.SetCursorPosition(ctx, &pb.TextAreaSetCursorRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       1,
		Column:    3,
	})
	if err != nil {
		t.Fatalf("SetCursorPosition failed: %v", err)
	}
	if !setResp.Success {
		t.Error("Expected success=true")
	}
}

func TestTextAreaService_HasSelection(t *testing.T) {
	srv := setupTestServer(t)
	taSvc := NewTextAreaService(srv)
	ctx := context.Background()

	sessionID, widgetID := createTextAreaWidget(t, srv, "hello world")

	resp, err := taSvc.HasSelection(ctx, &pb.TextAreaHasSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("HasSelection failed: %v", err)
	}

	// No selection initially
	if resp.HasSelection {
		t.Error("Expected no selection initially")
	}
}

func TestTextAreaService_SetAndGetSelection(t *testing.T) {
	srv := setupTestServer(t)
	taSvc := NewTextAreaService(srv)
	ctx := context.Background()

	sessionID, widgetID := createTextAreaWidget(t, srv, "hello world")

	// Select bytes 0..5 ("hello")
	setResp, err := taSvc.SetSelection(ctx, &pb.TextAreaSetSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Start:     0,
		End:       5,
	})
	if err != nil {
		t.Fatalf("SetSelection failed: %v", err)
	}
	if !setResp.Success {
		t.Error("Expected success=true")
	}

	// Verify selection exists
	hasResp, err := taSvc.HasSelection(ctx, &pb.TextAreaHasSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("HasSelection failed: %v", err)
	}
	if !hasResp.HasSelection {
		t.Error("Expected selection after SetSelection")
	}

	// Get selected text
	textResp, err := taSvc.GetSelectedText(ctx, &pb.TextAreaGetSelectedTextRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetSelectedText failed: %v", err)
	}
	if textResp.Text != "hello" {
		t.Errorf("Expected selected text 'hello', got %q", textResp.Text)
	}

	// Get full selection info
	selResp, err := taSvc.GetSelection(ctx, &pb.TextAreaGetSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetSelection failed: %v", err)
	}
	if !selResp.HasSelection {
		t.Error("Expected HasSelection=true")
	}
	if selResp.SelectedText != "hello" {
		t.Errorf("Expected selected text 'hello', got %q", selResp.SelectedText)
	}
}

func TestTextAreaService_InsertTextAtCursor(t *testing.T) {
	srv := setupTestServer(t)
	taSvc := NewTextAreaService(srv)
	widgetSvc := NewWidgetService(srv)
	ctx := context.Background()

	sessionID, widgetID := createTextAreaWidget(t, srv, "hello world")

	// Select "hello" (bytes 0..5)
	_, err := taSvc.SetSelection(ctx, &pb.TextAreaSetSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Start:     0,
		End:       5,
	})
	if err != nil {
		t.Fatalf("SetSelection failed: %v", err)
	}

	// Replace selection with "goodbye"
	insertResp, err := taSvc.InsertTextAtCursor(ctx, &pb.TextAreaInsertTextRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Text:      "goodbye",
	})
	if err != nil {
		t.Fatalf("InsertTextAtCursor failed: %v", err)
	}
	if !insertResp.Success {
		t.Error("Expected success=true")
	}

	// Verify text changed via GetProperties
	propsResp, err := widgetSvc.GetProperties(ctx, &pb.GetPropertiesRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetProperties failed: %v", err)
	}
	text := propsResp.Properties.GetTextArea().GetText()
	if text != "goodbye world" {
		t.Errorf("Expected 'goodbye world', got %q", text)
	}
}

func TestTextAreaService_ValidationErrors(t *testing.T) {
	srv := setupTestServer(t)
	taSvc := NewTextAreaService(srv)
	ctx := context.Background()

	// Test nil session/widget
	_, err := taSvc.GetCursorPosition(ctx, &pb.TextAreaGetCursorRequest{})
	if err == nil {
		t.Error("Expected error for nil session/widget")
	}

	// Test non-existent widget
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	_, err = taSvc.GetCursorPosition(ctx, &pb.TextAreaGetCursorRequest{
		SessionId: &pb.SessionId{Id: session.ID},
		WidgetId:  &pb.WidgetId{Id: "nonexistent"},
	})
	if err == nil {
		t.Error("Expected error for non-existent widget")
	}
}

func TestTextAreaService_WrongWidgetType(t *testing.T) {
	srv := setupTestServer(t)
	widgetSvc := NewWidgetService(srv)
	taSvc := NewTextAreaService(srv)
	ctx := context.Background()

	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create a Box widget (not a TextArea)
	createResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: session.ID},
		Type:      pb.WidgetType_WIDGET_BOX,
	})
	if err != nil {
		t.Fatalf("Failed to create box: %v", err)
	}

	// Try TextArea operations on a Box - should fail
	_, err = taSvc.GetCursorPosition(ctx, &pb.TextAreaGetCursorRequest{
		SessionId: &pb.SessionId{Id: session.ID},
		WidgetId:  createResp.WidgetId,
	})
	if err == nil {
		t.Error("Expected error when using TextArea service on a Box widget")
	}
}

func TestRowColToOffset(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		row, col int
		expected int
	}{
		{"start of text", "hello\nworld", 0, 0, 0},
		{"middle of first line", "hello\nworld", 0, 3, 3},
		{"end of first line", "hello\nworld", 0, 5, 5},
		{"start of second line", "hello\nworld", 1, 0, 6},
		{"middle of second line", "hello\nworld", 1, 3, 9},
		{"end of text", "hello\nworld", 1, 5, 11},
		{"past end of line", "hello\nworld", 0, 10, 5},
		{"past end of text", "hello\nworld", 2, 0, 11},
		{"empty text", "", 0, 0, 0},
		{"single line", "hello", 0, 3, 3},
		{"three lines", "aa\nbb\ncc", 2, 1, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rowColToOffset(tt.text, tt.row, tt.col)
			if result != tt.expected {
				t.Errorf("rowColToOffset(%q, %d, %d) = %d, expected %d",
					tt.text, tt.row, tt.col, result, tt.expected)
			}
		})
	}
}

func TestParseKeyString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKey  tcell.Key
		wantRune rune
		wantMod  tcell.ModMask
		wantOk   bool
	}{
		// tcell encodes ctrl+letter as a specific key constant, not KeyRune+ModCtrl
		{"ctrl+a", "ctrl+a", tcell.KeyCtrlA, 0, 0, true},
		{"ctrl+d", "ctrl+d", tcell.KeyCtrlD, 0, 0, true},
		{"ctrl+z", "ctrl+z", tcell.KeyCtrlZ, 0, 0, true},
		{"ctrl+space", "ctrl+space", tcell.KeyCtrlSpace, 0, 0, true},
		{"ctrl+backslash", "ctrl+backslash", tcell.KeyCtrlBackslash, 0, 0, true},
		// alt+letter uses KeyRune with ModAlt
		{"alt+x", "alt+x", tcell.KeyRune, 'x', tcell.ModAlt, true},
		{"empty string", "", 0, 0, 0, false},
		{"just ctrl", "ctrl+", 0, 0, 0, false},
		{"unknown", "ctrl+unknown", 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk, ok := parseKeyString(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("parseKeyString(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if sk.key != tt.wantKey {
				t.Errorf("parseKeyString(%q) key = %v, want %v", tt.input, sk.key, tt.wantKey)
			}
			if sk.rune != tt.wantRune {
				t.Errorf("parseKeyString(%q) rune = %v, want %v", tt.input, sk.rune, tt.wantRune)
			}
			if sk.mod != tt.wantMod {
				t.Errorf("parseKeyString(%q) mod = %v, want %v", tt.input, sk.mod, tt.wantMod)
			}
		})
	}
}

func TestParseSuppressedKeys(t *testing.T) {
	keys := parseSuppressedKeys([]string{"ctrl+d", "ctrl+p", "ctrl+i", "invalid"})

	// Should parse 3 valid keys, skip invalid
	if len(keys) != 3 {
		t.Errorf("Expected 3 parsed keys, got %d", len(keys))
	}
}

func TestSuppressedKeyMatches(t *testing.T) {
	sk, ok := parseKeyString("ctrl+d")
	if !ok {
		t.Fatal("Failed to parse ctrl+d")
	}

	// Should match Ctrl+D
	event := tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl)
	if !sk.matches(event) {
		t.Error("Expected ctrl+d to match KeyCtrlD event")
	}

	// Should not match Ctrl+A
	event2 := tcell.NewEventKey(tcell.KeyCtrlA, 0, tcell.ModCtrl)
	if sk.matches(event2) {
		t.Error("Expected ctrl+d to NOT match KeyCtrlA event")
	}
}
