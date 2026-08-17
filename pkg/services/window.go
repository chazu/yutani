package services

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WindowService implements the WindowService gRPC service.
type WindowService struct {
	pb.UnimplementedWindowServiceServer
	server *server.Server
}

// NewWindowService creates a new WindowService.
func NewWindowService(srv *server.Server) *WindowService {
	return &WindowService{
		server: srv,
	}
}

// WindowManagerAddWindow adds a window to a window manager.
func (s *WindowService) WindowManagerAddWindow(ctx context.Context, req *pb.WindowManagerAddWindowRequest) (*pb.WindowManagerAddWindowResponse, error) {
	if req.SessionId == nil || req.ManagerId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, manager_id, and window_id are required")
	}

	sessionID := req.SessionId.Id
	managerID := req.ManagerId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, managerID); err != nil {
		return nil, err
	}
	if err := s.verifyOwnership(sessionID, windowID); err != nil {
		return nil, err
	}

	var retErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		mgr, err := s.getWindowManager(managerID)
		if err != nil {
			retErr = err
			return
		}

		win, err := s.getWindow(windowID)
		if err != nil {
			retErr = err
			return
		}

		mgr.AddWindow(win)
		s.server.Widgets().AddChild(managerID, windowID)
		slog.Info("Window added to manager", "window_id", windowID, "manager_id", managerID)
	})
	<-done

	if retErr != nil {
		return nil, retErr
	}

	return &pb.WindowManagerAddWindowResponse{Success: true}, nil
}

// WindowManagerRemoveWindow removes a window from a window manager.
func (s *WindowService) WindowManagerRemoveWindow(ctx context.Context, req *pb.WindowManagerRemoveWindowRequest) (*pb.WindowManagerRemoveWindowResponse, error) {
	if req.SessionId == nil || req.ManagerId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, manager_id, and window_id are required")
	}

	sessionID := req.SessionId.Id
	managerID := req.ManagerId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, managerID); err != nil {
		return nil, err
	}

	var success bool
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		mgr, err := s.getWindowManager(managerID)
		if err != nil {
			return
		}
		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		success = mgr.RemoveWindow(win)
		if success {
			s.server.Widgets().RemoveChild(managerID, windowID)
			slog.Info("Window removed from manager", "window_id", windowID, "manager_id", managerID)
		}
	})
	<-done

	return &pb.WindowManagerRemoveWindowResponse{Success: success}, nil
}

// WindowMove moves a window programmatically.
func (s *WindowService) WindowMove(ctx context.Context, req *pb.WindowMoveRequest) (*pb.WindowMoveResponse, error) {
	if req.SessionId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and window_id are required")
	}

	sessionID := req.SessionId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, windowID); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		_, _, w, h := win.GetRect()
		win.SetRect(int(req.X), int(req.Y), w, h)
		if win.IsMovable() {
			win.SetRestoredRect(server.Rect{X: int(req.X), Y: int(req.Y), Width: w, Height: h})
		}
		slog.Info("Window moved", "window_id", windowID, "x", req.X, "y", req.Y)
	})
	<-done

	return &pb.WindowMoveResponse{Success: true}, nil
}

// WindowResize resizes a window programmatically.
func (s *WindowService) WindowResize(ctx context.Context, req *pb.WindowResizeRequest) (*pb.WindowResizeResponse, error) {
	if req.SessionId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and window_id are required")
	}

	sessionID := req.SessionId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, windowID); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		x, y, _, _ := win.GetRect()
		win.SetRect(x, y, int(req.Width), int(req.Height))
		slog.Info("Window resized", "window_id", windowID, "width", req.Width, "height", req.Height)
	})
	<-done

	return &pb.WindowResizeResponse{Success: true}, nil
}

// WindowSetState sets the window state.
func (s *WindowService) WindowSetState(ctx context.Context, req *pb.WindowSetStateRequest) (*pb.WindowSetStateResponse, error) {
	if req.SessionId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and window_id are required")
	}

	sessionID := req.SessionId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, windowID); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		switch req.State {
		case pb.WindowState_WINDOW_NORMAL:
			win.Restore()
		case pb.WindowState_WINDOW_MAXIMIZED:
			// Find the parent manager to get its rect
			parentRect := s.findManagerRect(windowID)
			win.Maximize(parentRect)
		case pb.WindowState_WINDOW_MINIMIZED:
			win.Minimize()
		}
		slog.Info("Window state set", "window_id", windowID, "state", req.State)
	})
	<-done

	return &pb.WindowSetStateResponse{Success: true}, nil
}

// WindowGetState returns the window state and geometry.
func (s *WindowService) WindowGetState(ctx context.Context, req *pb.WindowGetStateRequest) (*pb.WindowGetStateResponse, error) {
	if req.SessionId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and window_id are required")
	}

	sessionID := req.SessionId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, windowID); err != nil {
		return nil, err
	}

	var resp *pb.WindowGetStateResponse
	done := make(chan struct{})
	s.server.App().QueueUpdate(func() {
		defer close(done)

		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		x, y, w, h := win.GetRect()
		rr := win.GetRestoredRect()

		state := pb.WindowState_WINDOW_NORMAL
		switch win.GetWindowState() {
		case server.WindowStateMaximized:
			state = pb.WindowState_WINDOW_MAXIMIZED
		case server.WindowStateMinimized:
			state = pb.WindowState_WINDOW_MINIMIZED
		}

		resp = &pb.WindowGetStateResponse{
			State: state,
			Rect: &pb.Rect{
				X: int32(x), Y: int32(y),
				Width: int32(w), Height: int32(h),
			},
			RestoredRect: &pb.Rect{
				X: int32(rr.X), Y: int32(rr.Y),
				Width: int32(rr.Width), Height: int32(rr.Height),
			},
		}
	})
	<-done

	if resp == nil {
		return nil, status.Error(codes.Internal, "failed to get window state")
	}

	return resp, nil
}

// WindowBringToFront brings a window to the front.
func (s *WindowService) WindowBringToFront(ctx context.Context, req *pb.WindowBringToFrontRequest) (*pb.WindowBringToFrontResponse, error) {
	if req.SessionId == nil || req.ManagerId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, manager_id, and window_id are required")
	}

	sessionID := req.SessionId.Id
	managerID := req.ManagerId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, managerID); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		mgr, err := s.getWindowManager(managerID)
		if err != nil {
			return
		}
		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		mgr.BringToFront(win)
		slog.Info("Window brought to front", "window_id", windowID, "manager_id", managerID)
	})
	<-done

	return &pb.WindowBringToFrontResponse{Success: true}, nil
}

// WindowSendToBack sends a window to the back.
func (s *WindowService) WindowSendToBack(ctx context.Context, req *pb.WindowSendToBackRequest) (*pb.WindowSendToBackResponse, error) {
	if req.SessionId == nil || req.ManagerId == nil || req.WindowId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, manager_id, and window_id are required")
	}

	sessionID := req.SessionId.Id
	managerID := req.ManagerId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, managerID); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		mgr, err := s.getWindowManager(managerID)
		if err != nil {
			return
		}
		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		mgr.SendToBack(win)
		slog.Info("Window sent to back", "window_id", windowID, "manager_id", managerID)
	})
	<-done

	return &pb.WindowSendToBackResponse{Success: true}, nil
}

// WindowManagerGetZOrder returns the z-order of windows.
func (s *WindowService) WindowManagerGetZOrder(ctx context.Context, req *pb.WindowManagerGetZOrderRequest) (*pb.WindowManagerGetZOrderResponse, error) {
	if req.SessionId == nil || req.ManagerId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and manager_id are required")
	}

	sessionID := req.SessionId.Id
	managerID := req.ManagerId.Id

	if err := s.verifyOwnership(sessionID, managerID); err != nil {
		return nil, err
	}

	var windowIDs []*pb.WidgetId
	done := make(chan struct{})
	s.server.App().QueueUpdate(func() {
		defer close(done)

		mgr, err := s.getWindowManager(managerID)
		if err != nil {
			return
		}

		zorder := mgr.GetZOrder()
		windowIDs = make([]*pb.WidgetId, 0, len(zorder))

		for _, win := range zorder {
			// Find widget ID for this primitive
			widgetID := s.findWidgetIDForPrimitive(win)
			if widgetID != "" {
				windowIDs = append(windowIDs, &pb.WidgetId{Id: widgetID})
			}
		}
	})
	<-done

	return &pb.WindowManagerGetZOrderResponse{WindowIds: windowIDs}, nil
}

// WindowSetConstraints sets window constraints.
func (s *WindowService) WindowSetConstraints(ctx context.Context, req *pb.WindowSetConstraintsRequest) (*pb.WindowSetConstraintsResponse, error) {
	if req.SessionId == nil || req.WindowId == nil || req.Constraints == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, window_id, and constraints are required")
	}

	sessionID := req.SessionId.Id
	windowID := req.WindowId.Id

	if err := s.verifyOwnership(sessionID, windowID); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		win, err := s.getWindow(windowID)
		if err != nil {
			return
		}

		c := req.Constraints
		if c.Resizable != nil {
			win.SetResizable(*c.Resizable)
		}
		if c.Movable != nil {
			win.SetMovable(*c.Movable)
		}
		if c.Closable != nil {
			win.SetClosable(*c.Closable)
		}
		if c.Minimizable != nil {
			win.SetMinimizable(*c.Minimizable)
		}
		if c.Maximizable != nil {
			win.SetMaximizable(*c.Maximizable)
		}
		if c.MinWidth != nil {
			win.SetMinWidth(int(*c.MinWidth))
		}
		if c.MinHeight != nil {
			win.SetMinHeight(int(*c.MinHeight))
		}
		if c.MaxWidth != nil {
			win.SetMaxWidth(int(*c.MaxWidth))
		}
		if c.MaxHeight != nil {
			win.SetMaxHeight(int(*c.MaxHeight))
		}
		slog.Info("Window constraints set", "window_id", windowID)
	})
	<-done

	return &pb.WindowSetConstraintsResponse{Success: true}, nil
}

// WindowSetContent sets the child widget content of a window.
func (s *WindowService) WindowSetContent(ctx context.Context, req *pb.WindowSetContentRequest) (*pb.WindowSetContentResponse, error) {
	if req.SessionId == nil || req.WindowId == nil || req.ChildId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, window_id, and child_id are required")
	}

	sessionID := req.SessionId.Id
	windowID := req.WindowId.Id
	childID := req.ChildId.Id

	if err := s.verifyOwnership(sessionID, windowID); err != nil {
		return nil, err
	}
	if err := s.verifyOwnership(sessionID, childID); err != nil {
		return nil, err
	}

	var retErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		win, err := s.getWindow(windowID)
		if err != nil {
			retErr = err
			return
		}

		childPrimitive, ok := s.server.Widgets().Get(childID)
		if !ok {
			retErr = status.Error(codes.NotFound, "child widget not found")
			return
		}

		win.SetChild(childPrimitive)
		s.server.Widgets().AddChild(windowID, childID)
		slog.Info("Window content set", "window_id", windowID, "child_id", childID)
	})
	<-done

	if retErr != nil {
		return nil, retErr
	}

	return &pb.WindowSetContentResponse{Success: true}, nil
}

// Helper methods

func (s *WindowService) verifyOwnership(sessionID, widgetID string) error {
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return status.Error(codes.PermissionDenied, "widget not owned by session")
	}
	return nil
}

func (s *WindowService) getWindow(widgetID string) (*server.WindowPrimitive, error) {
	primitive, ok := s.server.Widgets().Get(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "window widget not found")
	}
	win, ok := primitive.(*server.WindowPrimitive)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("widget %s is not a window", widgetID))
	}
	return win, nil
}

func (s *WindowService) getWindowManager(widgetID string) (*server.WindowManagerPrimitive, error) {
	primitive, ok := s.server.Widgets().Get(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "window manager widget not found")
	}
	mgr, ok := primitive.(*server.WindowManagerPrimitive)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("widget %s is not a window manager", widgetID))
	}
	return mgr, nil
}

// findManagerRect finds the manager's rect for a window widget.
func (s *WindowService) findManagerRect(windowID string) server.Rect {
	info, ok := s.server.Widgets().GetInfo(windowID)
	if !ok || info.ParentID == "" {
		return server.Rect{X: 0, Y: 0, Width: 80, Height: 24} // fallback
	}

	parentPrimitive, ok := s.server.Widgets().Get(info.ParentID)
	if !ok {
		return server.Rect{X: 0, Y: 0, Width: 80, Height: 24}
	}

	x, y, w, h := parentPrimitive.GetRect()
	return server.Rect{X: x, Y: y, Width: w, Height: h}
}

// findWidgetIDForPrimitive finds the widget registry ID for a given primitive
// by searching the manager's children.
func (s *WindowService) findWidgetIDForPrimitive(primitive *server.WindowPrimitive) string {
	// We don't have a direct primitive->ID lookup, so iterate the manager's children.
	// This is called from the tview goroutine.
	return s.server.Widgets().FindByPrimitive(primitive)
}
