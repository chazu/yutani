package services

import (
	"context"
	"log/slog"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MenuService implements the MenuService gRPC service.
type MenuService struct {
	pb.UnimplementedMenuServiceServer
	server *server.Server
}

// NewMenuService creates a new MenuService.
func NewMenuService(srv *server.Server) *MenuService {
	return &MenuService{
		server: srv,
	}
}

// verifyOwnership checks that the widget is owned by the session.
func (s *MenuService) verifyOwnership(sessionID, widgetID string) error {
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return status.Error(codes.PermissionDenied, "widget not owned by session")
	}
	return nil
}

// AddMenu adds a menu to a menu bar.
func (s *MenuService) AddMenu(ctx context.Context, req *pb.MenuAddMenuRequest) (*pb.MenuAddMenuResponse, error) {
	if req.SessionId == nil || req.MenuBarId == nil || req.MenuId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, menu_bar_id, and menu_id are required")
	}

	sessionID := req.SessionId.Id
	menuBarID := req.MenuBarId.Id
	menuID := req.MenuId.Id

	if err := s.verifyOwnership(sessionID, menuBarID); err != nil {
		return nil, err
	}
	if err := s.verifyOwnership(sessionID, menuID); err != nil {
		return nil, err
	}

	var retErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the MenuBarPrimitive
		mbPrim, ok := s.server.Widgets().Get(menuBarID)
		if !ok {
			retErr = status.Error(codes.NotFound, "menu bar widget not found")
			return
		}
		mb, ok := mbPrim.(*server.MenuBarPrimitive)
		if !ok {
			retErr = status.Error(codes.InvalidArgument, "widget is not a menu bar")
			return
		}

		// Get menu widget info for the title
		menuInfo, ok := s.server.Widgets().GetInfo(menuID)
		if !ok || menuInfo == nil {
			retErr = status.Error(codes.NotFound, "menu widget not found")
			return
		}

		title := ""
		if menuInfo.Properties != nil {
			if menuProps := menuInfo.Properties.GetMenu(); menuProps != nil {
				if menuProps.Title != nil {
					title = *menuProps.Title
				}
			}
		}
		// Fallback to widget title
		if title == "" && menuInfo.Properties != nil && menuInfo.Properties.Title != nil {
			title = *menuInfo.Properties.Title
		}

		// Add as a MenuEntry
		entry := &server.MenuEntry{
			ID:    menuID,
			Title: title,
		}
		mb.AddMenu(entry)
		s.server.Widgets().AddChild(menuBarID, menuID)
		slog.Info("Menu added to menu bar", "menu_id", menuID, "menu_bar_id", menuBarID, "title", title)
	})
	<-done

	if retErr != nil {
		return nil, retErr
	}
	return &pb.MenuAddMenuResponse{Success: true}, nil
}

// AddMenuItem adds a menu item to a menu.
func (s *MenuService) AddMenuItem(ctx context.Context, req *pb.MenuAddMenuItemRequest) (*pb.MenuAddMenuItemResponse, error) {
	if req.SessionId == nil || req.MenuId == nil || req.MenuItemId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, menu_id, and menu_item_id are required")
	}

	sessionID := req.SessionId.Id
	menuID := req.MenuId.Id
	menuItemID := req.MenuItemId.Id

	if err := s.verifyOwnership(sessionID, menuID); err != nil {
		return nil, err
	}
	if err := s.verifyOwnership(sessionID, menuItemID); err != nil {
		return nil, err
	}

	var retErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Find the MenuEntry in any MenuBar that contains this menu.
		// We need to walk MenuBarPrimitives to find the one containing this menu.
		menuEntry := s.findMenuEntry(menuID)
		if menuEntry == nil {
			retErr = status.Error(codes.FailedPrecondition, "menu not added to any menu bar")
			return
		}

		// Get menu item properties
		itemInfo, ok := s.server.Widgets().GetInfo(menuItemID)
		if !ok || itemInfo == nil {
			retErr = status.Error(codes.NotFound, "menu item widget not found")
			return
		}

		entry := &server.MenuItemEntry{
			ID: menuItemID,
		}

		if itemInfo.Properties != nil {
			if miProps := itemInfo.Properties.GetMenuItem(); miProps != nil {
				if miProps.Label != nil {
					entry.Label = *miProps.Label
				}
				if miProps.ShortcutDisplay != nil {
					entry.ShortcutDisplay = *miProps.ShortcutDisplay
				}
				if miProps.Separator != nil {
					entry.Separator = *miProps.Separator
				}
				if miProps.Disabled != nil {
					entry.Disabled = *miProps.Disabled
				}
			}
		}
		// Fallback label from widget title
		if entry.Label == "" && itemInfo.Properties != nil && itemInfo.Properties.Title != nil {
			entry.Label = *itemInfo.Properties.Title
		}

		menuEntry.Items = append(menuEntry.Items, entry)
		s.server.Widgets().AddChild(menuID, menuItemID)
		slog.Info("Menu item added to menu", "menu_item_id", menuItemID, "menu_id", menuID, "label", entry.Label)
	})
	<-done

	if retErr != nil {
		return nil, retErr
	}
	return &pb.MenuAddMenuItemResponse{Success: true}, nil
}

// RemoveMenuItem removes a menu item from a menu.
func (s *MenuService) RemoveMenuItem(ctx context.Context, req *pb.MenuRemoveMenuItemRequest) (*pb.MenuRemoveMenuItemResponse, error) {
	if req.SessionId == nil || req.MenuId == nil || req.MenuItemId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, menu_id, and menu_item_id are required")
	}

	sessionID := req.SessionId.Id
	menuID := req.MenuId.Id
	menuItemID := req.MenuItemId.Id

	if err := s.verifyOwnership(sessionID, menuID); err != nil {
		return nil, err
	}

	var retErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		menuEntry := s.findMenuEntry(menuID)
		if menuEntry == nil {
			retErr = status.Error(codes.FailedPrecondition, "menu not added to any menu bar")
			return
		}

		found := false
		for i, item := range menuEntry.Items {
			if item.ID == menuItemID {
				menuEntry.Items = append(menuEntry.Items[:i], menuEntry.Items[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			retErr = status.Error(codes.NotFound, "menu item not found in menu")
			return
		}

		s.server.Widgets().RemoveChild(menuID, menuItemID)
		slog.Info("Menu item removed from menu", "menu_item_id", menuItemID, "menu_id", menuID)
	})
	<-done

	if retErr != nil {
		return nil, retErr
	}
	return &pb.MenuRemoveMenuItemResponse{Success: true}, nil
}

// findMenuEntry searches MenuBarPrimitives for a MenuEntry with the given ID.
func (s *MenuService) findMenuEntry(menuID string) *server.MenuEntry {
	// Look up the menu's parent (which should be the menu bar)
	menuInfo, ok := s.server.Widgets().GetInfo(menuID)
	if !ok {
		return nil
	}
	if menuInfo.ParentID != "" {
		prim, ok := s.server.Widgets().Get(menuInfo.ParentID)
		if ok {
			if mb, ok := prim.(*server.MenuBarPrimitive); ok {
				return mb.GetMenu(menuID)
			}
		}
	}
	return nil
}
