package server

import (
	"testing"

	"github.com/rivo/tview"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestWidgetRegistry_Register(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"
	widgetType := pb.WidgetType_WIDGET_BOX
	primitive := tview.NewBox()
	properties := &pb.WidgetProperties{
		Border: boolPtr(true),
		Title:  strPtr("Test Box"),
	}

	widgetID := registry.Register(sessionID, widgetType, primitive, properties)

	if widgetID == "" {
		t.Fatal("Expected non-empty widget ID")
	}

	// Verify widget was registered
	info, ok := registry.GetInfo(widgetID)
	if !ok || info == nil {
		t.Fatal("Expected widget info, got nil")
	}

	if info.ID != widgetID {
		t.Errorf("Expected ID %s, got %s", widgetID, info.ID)
	}

	if info.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, info.SessionID)
	}

	if info.Type != widgetType {
		t.Errorf("Expected type %v, got %v", widgetType, info.Type)
	}

	if info.Primitive != primitive {
		t.Error("Primitive not stored correctly")
	}

	if info.Properties != properties {
		t.Error("Properties not stored correctly")
	}
}

func TestWidgetRegistry_Get(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"
	primitive := tview.NewBox()

	widgetID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, primitive, nil)

	// Test successful get
	retrieved, ok := registry.Get(widgetID)
	if !ok || retrieved != primitive {
		t.Error("Retrieved primitive doesn't match registered primitive")
	}

	// Test non-existent widget
	nonExistent, ok := registry.Get("non-existent-id")
	if ok || nonExistent != nil {
		t.Error("Expected nil for non-existent widget")
	}
}

func TestWidgetRegistry_GetOwner(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"
	primitive := tview.NewBox()

	widgetID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, primitive, nil)

	// Test successful get owner
	owner, ok := registry.GetOwner(widgetID)
	if !ok {
		t.Fatal("Expected to find owner")
	}
	if owner != sessionID {
		t.Errorf("Expected owner %s, got %s", sessionID, owner)
	}

	// Test non-existent widget
	_, ok = registry.GetOwner("non-existent-id")
	if ok {
		t.Error("Expected not to find owner for non-existent widget")
	}
}

func TestWidgetRegistry_UpdateProperties(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"
	primitive := tview.NewBox()
	initialProps := &pb.WidgetProperties{
		Border: boolPtr(true),
		Title:  strPtr("Initial"),
	}

	widgetID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, primitive, initialProps)

	// Update properties
	newProps := &pb.WidgetProperties{
		Border: boolPtr(false),
		Title:  strPtr("Updated"),
	}

	ok := registry.UpdateProperties(widgetID, newProps)
	if !ok {
		t.Fatal("Expected UpdateProperties to succeed")
	}

	// Verify properties were updated
	info, ok := registry.GetInfo(widgetID)
	if !ok || info.Properties != newProps {
		t.Error("Properties not updated correctly")
	}

	// Test non-existent widget
	ok = registry.UpdateProperties("non-existent-id", newProps)
	if ok {
		t.Error("Expected UpdateProperties to fail for non-existent widget")
	}
}

func TestWidgetRegistry_AddChild(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"

	parentID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	childID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)

	// Add child
	ok := registry.AddChild(parentID, childID)
	if !ok {
		t.Fatal("Expected AddChild to succeed")
	}

	// Verify parent has child
	parentInfo, ok := registry.GetInfo(parentID)
	if !ok || len(parentInfo.ChildIDs) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(parentInfo.ChildIDs))
	}
	if parentInfo.ChildIDs[0] != childID {
		t.Errorf("Expected child ID %s, got %s", childID, parentInfo.ChildIDs[0])
	}

	// Verify child has parent
	childInfo, ok := registry.GetInfo(childID)
	if !ok || childInfo.ParentID != parentID {
		t.Errorf("Expected parent ID %s, got %s", parentID, childInfo.ParentID)
	}
}



func TestWidgetRegistry_RemoveChild(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"

	parentID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	childID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)

	// Add child first
	registry.AddChild(parentID, childID)

	// Remove child
	ok := registry.RemoveChild(parentID, childID)
	if !ok {
		t.Fatal("Expected RemoveChild to succeed")
	}

	// Verify parent no longer has child
	parentInfo, ok := registry.GetInfo(parentID)
	if !ok || len(parentInfo.ChildIDs) != 0 {
		t.Errorf("Expected 0 children, got %d", len(parentInfo.ChildIDs))
	}

	// Verify child no longer has parent
	childInfo, ok := registry.GetInfo(childID)
	if !ok || childInfo.ParentID != "" {
		t.Errorf("Expected empty parent ID, got %s", childInfo.ParentID)
	}
}

func TestWidgetRegistry_SetRoot(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"

	widgetID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)

	// Set root
	ok := registry.SetRoot(sessionID, widgetID)
	if !ok {
		t.Fatal("Expected SetRoot to succeed")
	}

	// Verify root was set
	rootID, ok := registry.GetRoot(sessionID)
	if !ok || rootID != widgetID {
		t.Errorf("Expected root ID %s, got %s", widgetID, rootID)
	}

	// Test non-existent widget
	ok = registry.SetRoot(sessionID, "non-existent-id")
	if ok {
		t.Error("Expected SetRoot to fail for non-existent widget")
	}
}

func TestWidgetRegistry_GetRoot(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"

	// Test no root set
	rootID, ok := registry.GetRoot(sessionID)
	if ok || rootID != "" {
		t.Errorf("Expected empty root ID, got %s", rootID)
	}

	// Set root and test
	widgetID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	registry.SetRoot(sessionID, widgetID)

	rootID, ok = registry.GetRoot(sessionID)
	if !ok || rootID != widgetID {
		t.Errorf("Expected root ID %s, got %s", widgetID, rootID)
	}
}

func TestWidgetRegistry_Delete(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"

	widgetID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)

	// Delete widget
	ok := registry.Delete(widgetID)
	if !ok {
		t.Fatal("Expected Delete to succeed")
	}

	// Verify widget was deleted
	_, ok = registry.GetInfo(widgetID)
	if ok {
		t.Error("Expected widget to be deleted")
	}

	// Test deleting non-existent widget
	ok = registry.Delete("non-existent-id")
	if ok {
		t.Error("Expected Delete to fail for non-existent widget")
	}
}

func TestWidgetRegistry_DeleteRecursive(t *testing.T) {
	registry := NewWidgetRegistry()
	sessionID := "test-session"

	// Create parent with two children
	parentID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	child1ID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	child2ID := registry.Register(sessionID, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)

	registry.AddChild(parentID, child1ID)
	registry.AddChild(parentID, child2ID)

	// Delete parent (should delete children too)
	ok := registry.Delete(parentID)
	if !ok {
		t.Fatal("Expected Delete to succeed")
	}

	// Verify parent was deleted
	if _, ok := registry.GetInfo(parentID); ok {
		t.Error("Expected parent to be deleted")
	}

	// Verify children were deleted
	if _, ok := registry.GetInfo(child1ID); ok {
		t.Error("Expected child1 to be deleted")
	}
	if _, ok := registry.GetInfo(child2ID); ok {
		t.Error("Expected child2 to be deleted")
	}
}

func TestWidgetRegistry_DeleteBySession(t *testing.T) {
	registry := NewWidgetRegistry()
	session1 := "session-1"
	session2 := "session-2"

	// Register widgets for two sessions
	widget1 := registry.Register(session1, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	widget2 := registry.Register(session1, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	widget3 := registry.Register(session2, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)

	// Delete session1 widgets
	registry.DeleteBySession(session1)

	// Verify session1 widgets were deleted
	if _, ok := registry.GetInfo(widget1); ok {
		t.Error("Expected widget1 to be deleted")
	}
	if _, ok := registry.GetInfo(widget2); ok {
		t.Error("Expected widget2 to be deleted")
	}

	// Verify session2 widget still exists
	if _, ok := registry.GetInfo(widget3); !ok {
		t.Error("Expected widget3 to still exist")
	}
}

func TestWidgetRegistry_ListBySession(t *testing.T) {
	registry := NewWidgetRegistry()
	session1 := "session-1"
	session2 := "session-2"

	// Register widgets for two sessions
	widget1 := registry.Register(session1, pb.WidgetType_WIDGET_BOX, tview.NewBox(), nil)
	widget2 := registry.Register(session1, pb.WidgetType_WIDGET_TEXT_VIEW, tview.NewTextView(), nil)
	widget3 := registry.Register(session2, pb.WidgetType_WIDGET_BUTTON, tview.NewButton("Test"), nil)

	// List session1 widgets
	widgets := registry.ListBySession(session1)
	if len(widgets) != 2 {
		t.Fatalf("Expected 2 widgets for session1, got %d", len(widgets))
	}

	// Verify widget IDs
	foundWidget1 := false
	foundWidget2 := false
	for _, info := range widgets {
		if info.ID == widget1 {
			foundWidget1 = true
			if info.Type != pb.WidgetType_WIDGET_BOX {
				t.Error("Widget1 type mismatch")
			}
		}
		if info.ID == widget2 {
			foundWidget2 = true
			if info.Type != pb.WidgetType_WIDGET_TEXT_VIEW {
				t.Error("Widget2 type mismatch")
			}
		}
	}

	if !foundWidget1 || !foundWidget2 {
		t.Error("Not all session1 widgets found in list")
	}

	// List session2 widgets
	widgets = registry.ListBySession(session2)
	if len(widgets) != 1 {
		t.Fatalf("Expected 1 widget for session2, got %d", len(widgets))
	}
	if widgets[0].ID != widget3 {
		t.Error("Widget3 not found in session2 list")
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

