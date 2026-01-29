package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
)

func createWindowTestSetup(t *testing.T) (*server.Server, *WidgetService, *WindowService, string) {
	t.Helper()
	srv := setupTestServer(t)
	widgetSvc := NewWidgetService(srv)
	windowSvc := NewWindowService(srv)

	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	return srv, widgetSvc, windowSvc, session.ID
}

func createTestWindow(t *testing.T, widgetSvc *WidgetService, sessionID string, title string, x, y, w, h int32) string {
	t.Helper()
	ctx := context.Background()
	resp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_WINDOW,
		Properties: &pb.WidgetProperties{
			Title: &title,
			TypeProperties: &pb.WidgetProperties_Window{
				Window: &pb.WindowProperties{
					InitialRect: &pb.Rect{X: x, Y: y, Width: w, Height: h},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create window: %v", err)
	}
	return resp.WidgetId.Id
}

func createTestWindowManager(t *testing.T, widgetSvc *WidgetService, sessionID string) string {
	t.Helper()
	ctx := context.Background()
	resp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_WINDOW_MANAGER,
	})
	if err != nil {
		t.Fatalf("Failed to create window manager: %v", err)
	}
	return resp.WidgetId.Id
}

func TestWindowService_AddRemoveWindow(t *testing.T) {
	_, widgetSvc, windowSvc, sessionID := createWindowTestSetup(t)
	ctx := context.Background()

	mgrID := createTestWindowManager(t, widgetSvc, sessionID)
	winID := createTestWindow(t, widgetSvc, sessionID, "Test", 5, 5, 30, 15)

	// Add window to manager
	addResp, err := windowSvc.WindowManagerAddWindow(ctx, &pb.WindowManagerAddWindowRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		ManagerId: &pb.WidgetId{Id: mgrID},
		WindowId:  &pb.WidgetId{Id: winID},
	})
	if err != nil {
		t.Fatalf("Failed to add window: %v", err)
	}
	if !addResp.Success {
		t.Error("expected success")
	}

	// Remove window
	rmResp, err := windowSvc.WindowManagerRemoveWindow(ctx, &pb.WindowManagerRemoveWindowRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		ManagerId: &pb.WidgetId{Id: mgrID},
		WindowId:  &pb.WidgetId{Id: winID},
	})
	if err != nil {
		t.Fatalf("Failed to remove window: %v", err)
	}
	if !rmResp.Success {
		t.Error("expected success")
	}
}

func TestWindowService_MoveResize(t *testing.T) {
	_, widgetSvc, windowSvc, sessionID := createWindowTestSetup(t)
	ctx := context.Background()

	winID := createTestWindow(t, widgetSvc, sessionID, "Movable", 5, 5, 30, 15)

	// Move
	moveResp, err := windowSvc.WindowMove(ctx, &pb.WindowMoveRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
		X:         20,
		Y:         10,
	})
	if err != nil {
		t.Fatalf("Failed to move: %v", err)
	}
	if !moveResp.Success {
		t.Error("expected success")
	}

	// Verify via GetState
	stateResp, err := windowSvc.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
	})
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if stateResp.Rect.X != 20 || stateResp.Rect.Y != 10 {
		t.Errorf("expected position (20,10), got (%d,%d)", stateResp.Rect.X, stateResp.Rect.Y)
	}

	// Resize
	resizeResp, err := windowSvc.WindowResize(ctx, &pb.WindowResizeRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
		Width:     40,
		Height:    20,
	})
	if err != nil {
		t.Fatalf("Failed to resize: %v", err)
	}
	if !resizeResp.Success {
		t.Error("expected success")
	}

	// Verify
	stateResp2, err := windowSvc.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
	})
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if stateResp2.Rect.Width != 40 || stateResp2.Rect.Height != 20 {
		t.Errorf("expected size (40,20), got (%d,%d)", stateResp2.Rect.Width, stateResp2.Rect.Height)
	}
}

func TestWindowService_SetState(t *testing.T) {
	_, widgetSvc, windowSvc, sessionID := createWindowTestSetup(t)
	ctx := context.Background()

	winID := createTestWindow(t, widgetSvc, sessionID, "StateTest", 5, 5, 30, 15)

	// Minimize
	_, err := windowSvc.WindowSetState(ctx, &pb.WindowSetStateRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
		State:     pb.WindowState_WINDOW_MINIMIZED,
	})
	if err != nil {
		t.Fatalf("Failed to set state: %v", err)
	}

	stateResp, err := windowSvc.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
	})
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if stateResp.State != pb.WindowState_WINDOW_MINIMIZED {
		t.Errorf("expected MINIMIZED, got %v", stateResp.State)
	}

	// Restore
	_, err = windowSvc.WindowSetState(ctx, &pb.WindowSetStateRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
		State:     pb.WindowState_WINDOW_NORMAL,
	})
	if err != nil {
		t.Fatalf("Failed to restore: %v", err)
	}

	stateResp2, err := windowSvc.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
	})
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if stateResp2.State != pb.WindowState_WINDOW_NORMAL {
		t.Errorf("expected NORMAL, got %v", stateResp2.State)
	}
}

func TestWindowService_ZOrder(t *testing.T) {
	_, widgetSvc, windowSvc, sessionID := createWindowTestSetup(t)
	ctx := context.Background()

	mgrID := createTestWindowManager(t, widgetSvc, sessionID)
	win1ID := createTestWindow(t, widgetSvc, sessionID, "Win1", 5, 5, 20, 10)
	win2ID := createTestWindow(t, widgetSvc, sessionID, "Win2", 10, 10, 20, 10)

	// Add both windows
	windowSvc.WindowManagerAddWindow(ctx, &pb.WindowManagerAddWindowRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		ManagerId: &pb.WidgetId{Id: mgrID},
		WindowId:  &pb.WidgetId{Id: win1ID},
	})
	windowSvc.WindowManagerAddWindow(ctx, &pb.WindowManagerAddWindowRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		ManagerId: &pb.WidgetId{Id: mgrID},
		WindowId:  &pb.WidgetId{Id: win2ID},
	})

	// Get z-order (front to back)
	zResp, err := windowSvc.WindowManagerGetZOrder(ctx, &pb.WindowManagerGetZOrderRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		ManagerId: &pb.WidgetId{Id: mgrID},
	})
	if err != nil {
		t.Fatalf("Failed to get z-order: %v", err)
	}
	if len(zResp.WindowIds) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(zResp.WindowIds))
	}
	// win2 should be in front (added last)
	if zResp.WindowIds[0].Id != win2ID {
		t.Errorf("expected win2 at front, got %s", zResp.WindowIds[0].Id)
	}

	// Bring win1 to front
	windowSvc.WindowBringToFront(ctx, &pb.WindowBringToFrontRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		ManagerId: &pb.WidgetId{Id: mgrID},
		WindowId:  &pb.WidgetId{Id: win1ID},
	})

	zResp2, err := windowSvc.WindowManagerGetZOrder(ctx, &pb.WindowManagerGetZOrderRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		ManagerId: &pb.WidgetId{Id: mgrID},
	})
	if err != nil {
		t.Fatalf("Failed to get z-order after bring to front: %v", err)
	}
	if zResp2.WindowIds[0].Id != win1ID {
		t.Errorf("expected win1 at front after BringToFront, got %s", zResp2.WindowIds[0].Id)
	}
}

func TestWindowService_SetConstraints(t *testing.T) {
	_, widgetSvc, windowSvc, sessionID := createWindowTestSetup(t)
	ctx := context.Background()

	winID := createTestWindow(t, widgetSvc, sessionID, "Constrained", 5, 5, 30, 15)

	resizable := false
	movable := false
	minW := int32(20)
	minH := int32(10)

	_, err := windowSvc.WindowSetConstraints(ctx, &pb.WindowSetConstraintsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WindowId:  &pb.WidgetId{Id: winID},
		Constraints: &pb.WindowConstraints{
			Resizable: &resizable,
			Movable:   &movable,
			MinWidth:  &minW,
			MinHeight: &minH,
		},
	})
	if err != nil {
		t.Fatalf("Failed to set constraints: %v", err)
	}
}

func TestWindowService_ValidationErrors(t *testing.T) {
	_, _, windowSvc, _ := createWindowTestSetup(t)
	ctx := context.Background()

	// Missing session_id
	_, err := windowSvc.WindowMove(ctx, &pb.WindowMoveRequest{})
	if err == nil {
		t.Error("expected error for missing session_id")
	}

	// Non-existent widget
	_, err = windowSvc.WindowMove(ctx, &pb.WindowMoveRequest{
		SessionId: &pb.SessionId{Id: "fake-session"},
		WindowId:  &pb.WidgetId{Id: "fake-widget"},
		X:         10,
		Y:         10,
	})
	if err == nil {
		t.Error("expected error for non-existent widget")
	}
}
