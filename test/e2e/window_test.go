package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/testutil"
)

// createE2ESession creates a session and returns the session ID and cleanup func.
func createE2ESession(t *testing.T, ctx context.Context, sessionClient pb.SessionServiceClient, name string) *pb.SessionId {
	t.Helper()
	resp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: name,
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	return resp.SessionId
}

// createE2EWindow creates a window widget via the widget service and returns its ID.
func createE2EWindow(t *testing.T, ctx context.Context, widgetClient pb.WidgetServiceClient, sessionID *pb.SessionId, title string, x, y, w, h int32) *pb.WidgetId {
	t.Helper()
	resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
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
	return resp.WidgetId
}

// createE2EWindowManager creates a window manager widget and returns its ID.
func createE2EWindowManager(t *testing.T, ctx context.Context, widgetClient pb.WidgetServiceClient, sessionID *pb.SessionId) *pb.WidgetId {
	t.Helper()
	resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_WINDOW_MANAGER,
	})
	if err != nil {
		t.Fatalf("Failed to create window manager: %v", err)
	}
	return resp.WidgetId
}

func TestE2E_WindowManagerLifecycle(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	windowClient := pb.NewWindowServiceClient(conn)

	sessionID := createE2ESession(t, ctx, sessionClient, "window-e2e")

	// Create a window manager
	mgrID := createE2EWindowManager(t, ctx, widgetClient, sessionID)

	// Create two windows
	win1ID := createE2EWindow(t, ctx, widgetClient, sessionID, "Window 1", 5, 5, 30, 15)
	win2ID := createE2EWindow(t, ctx, widgetClient, sessionID, "Window 2", 10, 10, 25, 12)

	// Add windows to manager
	addResp, err := windowClient.WindowManagerAddWindow(ctx, &pb.WindowManagerAddWindowRequest{
		SessionId: sessionID,
		ManagerId: mgrID,
		WindowId:  win1ID,
	})
	if err != nil {
		t.Fatalf("WindowManagerAddWindow failed: %v", err)
	}
	if !addResp.Success {
		t.Error("expected success adding win1")
	}

	addResp2, err := windowClient.WindowManagerAddWindow(ctx, &pb.WindowManagerAddWindowRequest{
		SessionId: sessionID,
		ManagerId: mgrID,
		WindowId:  win2ID,
	})
	if err != nil {
		t.Fatalf("WindowManagerAddWindow (2) failed: %v", err)
	}
	if !addResp2.Success {
		t.Error("expected success adding win2")
	}

	// Verify z-order: win2 should be in front
	zResp, err := windowClient.WindowManagerGetZOrder(ctx, &pb.WindowManagerGetZOrderRequest{
		SessionId: sessionID,
		ManagerId: mgrID,
	})
	if err != nil {
		t.Fatalf("WindowManagerGetZOrder failed: %v", err)
	}
	if len(zResp.WindowIds) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(zResp.WindowIds))
	}
	if zResp.WindowIds[0].Id != win2ID.Id {
		t.Errorf("expected win2 at front, got %s", zResp.WindowIds[0].Id)
	}

	// Bring win1 to front
	_, err = windowClient.WindowBringToFront(ctx, &pb.WindowBringToFrontRequest{
		SessionId: sessionID,
		ManagerId: mgrID,
		WindowId:  win1ID,
	})
	if err != nil {
		t.Fatalf("WindowBringToFront failed: %v", err)
	}

	// Verify win1 is now in front
	zResp2, err := windowClient.WindowManagerGetZOrder(ctx, &pb.WindowManagerGetZOrderRequest{
		SessionId: sessionID,
		ManagerId: mgrID,
	})
	if err != nil {
		t.Fatalf("WindowManagerGetZOrder (2) failed: %v", err)
	}
	if zResp2.WindowIds[0].Id != win1ID.Id {
		t.Errorf("expected win1 at front after BringToFront, got %s", zResp2.WindowIds[0].Id)
	}

	// Remove win1
	rmResp, err := windowClient.WindowManagerRemoveWindow(ctx, &pb.WindowManagerRemoveWindowRequest{
		SessionId: sessionID,
		ManagerId: mgrID,
		WindowId:  win1ID,
	})
	if err != nil {
		t.Fatalf("WindowManagerRemoveWindow failed: %v", err)
	}
	if !rmResp.Success {
		t.Error("expected success removing win1")
	}

	// Verify only 1 window remains
	zResp3, err := windowClient.WindowManagerGetZOrder(ctx, &pb.WindowManagerGetZOrderRequest{
		SessionId: sessionID,
		ManagerId: mgrID,
	})
	if err != nil {
		t.Fatalf("WindowManagerGetZOrder (3) failed: %v", err)
	}
	if len(zResp3.WindowIds) != 1 {
		t.Errorf("expected 1 window after removal, got %d", len(zResp3.WindowIds))
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

func TestE2E_WindowMoveResize(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	windowClient := pb.NewWindowServiceClient(conn)

	sessionID := createE2ESession(t, ctx, sessionClient, "window-move-e2e")

	winID := createE2EWindow(t, ctx, widgetClient, sessionID, "Movable", 5, 5, 30, 15)

	// Move
	moveResp, err := windowClient.WindowMove(ctx, &pb.WindowMoveRequest{
		SessionId: sessionID,
		WindowId:  winID,
		X:         20,
		Y:         10,
	})
	if err != nil {
		t.Fatalf("WindowMove failed: %v", err)
	}
	if !moveResp.Success {
		t.Error("expected successful move")
	}

	// Verify via GetState
	state, err := windowClient.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: sessionID,
		WindowId:  winID,
	})
	if err != nil {
		t.Fatalf("WindowGetState failed: %v", err)
	}
	if state.Rect.X != 20 || state.Rect.Y != 10 {
		t.Errorf("expected position (20,10), got (%d,%d)", state.Rect.X, state.Rect.Y)
	}
	if state.Rect.Width != 30 || state.Rect.Height != 15 {
		t.Errorf("expected size (30,15), got (%d,%d)", state.Rect.Width, state.Rect.Height)
	}

	// Resize
	resizeResp, err := windowClient.WindowResize(ctx, &pb.WindowResizeRequest{
		SessionId: sessionID,
		WindowId:  winID,
		Width:     40,
		Height:    20,
	})
	if err != nil {
		t.Fatalf("WindowResize failed: %v", err)
	}
	if !resizeResp.Success {
		t.Error("expected successful resize")
	}

	// Verify
	state2, err := windowClient.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: sessionID,
		WindowId:  winID,
	})
	if err != nil {
		t.Fatalf("WindowGetState (2) failed: %v", err)
	}
	if state2.Rect.Width != 40 || state2.Rect.Height != 20 {
		t.Errorf("expected size (40,20), got (%d,%d)", state2.Rect.Width, state2.Rect.Height)
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

func TestE2E_WindowStateTransitions(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	windowClient := pb.NewWindowServiceClient(conn)

	sessionID := createE2ESession(t, ctx, sessionClient, "window-state-e2e")

	winID := createE2EWindow(t, ctx, widgetClient, sessionID, "StateTest", 5, 5, 30, 15)

	// Initial state should be NORMAL
	state, err := windowClient.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: sessionID,
		WindowId:  winID,
	})
	if err != nil {
		t.Fatalf("WindowGetState failed: %v", err)
	}
	if state.State != pb.WindowState_WINDOW_NORMAL {
		t.Errorf("expected NORMAL state, got %v", state.State)
	}

	// Minimize
	_, err = windowClient.WindowSetState(ctx, &pb.WindowSetStateRequest{
		SessionId: sessionID,
		WindowId:  winID,
		State:     pb.WindowState_WINDOW_MINIMIZED,
	})
	if err != nil {
		t.Fatalf("WindowSetState (minimize) failed: %v", err)
	}

	state2, err := windowClient.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: sessionID,
		WindowId:  winID,
	})
	if err != nil {
		t.Fatalf("WindowGetState (2) failed: %v", err)
	}
	if state2.State != pb.WindowState_WINDOW_MINIMIZED {
		t.Errorf("expected MINIMIZED, got %v", state2.State)
	}

	// Restore
	_, err = windowClient.WindowSetState(ctx, &pb.WindowSetStateRequest{
		SessionId: sessionID,
		WindowId:  winID,
		State:     pb.WindowState_WINDOW_NORMAL,
	})
	if err != nil {
		t.Fatalf("WindowSetState (restore) failed: %v", err)
	}

	state3, err := windowClient.WindowGetState(ctx, &pb.WindowGetStateRequest{
		SessionId: sessionID,
		WindowId:  winID,
	})
	if err != nil {
		t.Fatalf("WindowGetState (3) failed: %v", err)
	}
	if state3.State != pb.WindowState_WINDOW_NORMAL {
		t.Errorf("expected NORMAL after restore, got %v", state3.State)
	}
	// Verify position was restored
	if state3.Rect.X != 5 || state3.Rect.Y != 5 {
		t.Errorf("expected restored position (5,5), got (%d,%d)", state3.Rect.X, state3.Rect.Y)
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

func TestE2E_WindowConstraints(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	windowClient := pb.NewWindowServiceClient(conn)

	sessionID := createE2ESession(t, ctx, sessionClient, "window-constraints-e2e")

	winID := createE2EWindow(t, ctx, widgetClient, sessionID, "Constrained", 5, 5, 30, 15)

	// Set constraints
	resizable := false
	movable := false
	minW := int32(20)
	minH := int32(10)
	maxW := int32(80)
	maxH := int32(40)

	_, err = windowClient.WindowSetConstraints(ctx, &pb.WindowSetConstraintsRequest{
		SessionId: sessionID,
		WindowId:  winID,
		Constraints: &pb.WindowConstraints{
			Resizable: &resizable,
			Movable:   &movable,
			MinWidth:  &minW,
			MinHeight: &minH,
			MaxWidth:  &maxW,
			MaxHeight: &maxH,
		},
	})
	if err != nil {
		t.Fatalf("WindowSetConstraints failed: %v", err)
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

func TestE2E_WindowCreationWithProperties(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)

	sessionID := createE2ESession(t, ctx, sessionClient, "window-props-e2e")

	// Create a window with full properties
	title := "Full Props Window"
	resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_WINDOW,
		Properties: &pb.WidgetProperties{
			Title: &title,
			TypeProperties: &pb.WidgetProperties_Window{
				Window: &pb.WindowProperties{
					InitialRect: &pb.Rect{X: 2, Y: 2, Width: 40, Height: 20},
					Resizable:   testutil.BoolPtr(true),
					Movable:     testutil.BoolPtr(true),
					Closable:    testutil.BoolPtr(true),
					Minimizable: testutil.BoolPtr(true),
					Maximizable: testutil.BoolPtr(true),
					MinWidth:    testutil.Int32Ptr(10),
					MinHeight:   testutil.Int32Ptr(5),
					MaxWidth:    testutil.Int32Ptr(60),
					MaxHeight:   testutil.Int32Ptr(30),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Window with props) failed: %v", err)
	}
	if resp.WidgetId.Id == "" {
		t.Error("expected non-empty widget ID")
	}

	// Create a window manager with properties
	wmResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_WINDOW_MANAGER,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_WindowManager{
				WindowManager: &pb.WindowManagerProperties{
					BackgroundColor: &pb.Color{
						Color: &pb.Color_Name{Name: "darkblue"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (WindowManager with props) failed: %v", err)
	}
	if wmResp.WidgetId.Id == "" {
		t.Error("expected non-empty window manager widget ID")
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

func TestE2E_WindowValidation(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	windowClient := pb.NewWindowServiceClient(conn)

	// Missing session_id should fail
	_, err = windowClient.WindowMove(ctx, &pb.WindowMoveRequest{})
	if err == nil {
		t.Error("expected error for missing session_id")
	}

	// Non-existent session/widget should fail
	_, err = windowClient.WindowMove(ctx, &pb.WindowMoveRequest{
		SessionId: &pb.SessionId{Id: "fake-session"},
		WindowId:  &pb.WidgetId{Id: "fake-widget"},
		X:         10,
		Y:         10,
	})
	if err == nil {
		t.Error("expected error for non-existent widget")
	}

	// Non-existent manager should fail for add window
	_, err = windowClient.WindowManagerAddWindow(ctx, &pb.WindowManagerAddWindowRequest{
		SessionId: &pb.SessionId{Id: "fake-session"},
		ManagerId: &pb.WidgetId{Id: "fake-manager"},
		WindowId:  &pb.WidgetId{Id: "fake-window"},
	})
	if err == nil {
		t.Error("expected error for non-existent manager")
	}
}
