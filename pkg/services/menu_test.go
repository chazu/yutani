package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
)

func TestMenuService_AddMenu(t *testing.T) {
	srv := setupTestServer(t)
	widgetSvc := NewWidgetService(srv)
	menuSvc := NewMenuService(srv)
	ctx := context.Background()

	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sid := &pb.SessionId{Id: session.ID}

	// Create menu bar
	mbResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_BAR,
	})
	if err != nil {
		t.Fatalf("Failed to create menu bar: %v", err)
	}

	// Create menu
	fileTitle := "File"
	menuResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Menu{
				Menu: &pb.MenuProperties{Title: &fileTitle},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create menu: %v", err)
	}

	// Add menu to menu bar
	addResp, err := menuSvc.AddMenu(ctx, &pb.MenuAddMenuRequest{
		SessionId: sid,
		MenuBarId: mbResp.WidgetId,
		MenuId:    menuResp.WidgetId,
	})
	if err != nil {
		t.Fatalf("AddMenu failed: %v", err)
	}
	if !addResp.Success {
		t.Error("Expected success=true")
	}

	// Verify the menu bar has the menu
	mbPrim, ok := srv.Widgets().Get(mbResp.WidgetId.Id)
	if !ok {
		t.Fatal("Menu bar not found in widget registry")
	}
	mb := mbPrim.(*server.MenuBarPrimitive)
	menus := mb.GetMenus()
	if len(menus) != 1 {
		t.Fatalf("Expected 1 menu, got %d", len(menus))
	}
	if menus[0].Title != "File" {
		t.Errorf("Expected menu title 'File', got %q", menus[0].Title)
	}
}

func TestMenuService_AddMenuItem(t *testing.T) {
	srv := setupTestServer(t)
	widgetSvc := NewWidgetService(srv)
	menuSvc := NewMenuService(srv)
	ctx := context.Background()

	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sid := &pb.SessionId{Id: session.ID}

	// Create menu bar
	mbResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_BAR,
	})
	if err != nil {
		t.Fatalf("Failed to create menu bar: %v", err)
	}

	// Create and add menu
	fileTitle := "File"
	menuResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Menu{
				Menu: &pb.MenuProperties{Title: &fileTitle},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create menu: %v", err)
	}

	_, err = menuSvc.AddMenu(ctx, &pb.MenuAddMenuRequest{
		SessionId: sid,
		MenuBarId: mbResp.WidgetId,
		MenuId:    menuResp.WidgetId,
	})
	if err != nil {
		t.Fatalf("AddMenu failed: %v", err)
	}

	// Create menu items
	newLabel := "New Scratchpad"
	newShortcut := "Ctrl+N"
	itemResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_ITEM,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_MenuItem{
				MenuItem: &pb.MenuItemProperties{
					Label:           &newLabel,
					ShortcutDisplay: &newShortcut,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Add item to menu
	addItemResp, err := menuSvc.AddMenuItem(ctx, &pb.MenuAddMenuItemRequest{
		SessionId:  sid,
		MenuId:     menuResp.WidgetId,
		MenuItemId: itemResp.WidgetId,
	})
	if err != nil {
		t.Fatalf("AddMenuItem failed: %v", err)
	}
	if !addItemResp.Success {
		t.Error("Expected success=true")
	}

	// Verify the menu has the item
	mbPrim, _ := srv.Widgets().Get(mbResp.WidgetId.Id)
	mb := mbPrim.(*server.MenuBarPrimitive)
	menus := mb.GetMenus()
	if len(menus[0].Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(menus[0].Items))
	}
	item := menus[0].Items[0]
	if item.Label != "New Scratchpad" {
		t.Errorf("Expected label 'New Scratchpad', got %q", item.Label)
	}
	if item.ShortcutDisplay != "Ctrl+N" {
		t.Errorf("Expected shortcut 'Ctrl+N', got %q", item.ShortcutDisplay)
	}
}

func TestMenuService_SeparatorAndDisabled(t *testing.T) {
	srv := setupTestServer(t)
	widgetSvc := NewWidgetService(srv)
	menuSvc := NewMenuService(srv)
	ctx := context.Background()

	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sid := &pb.SessionId{Id: session.ID}

	// Create menu bar + menu
	mbResp, _ := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_BAR,
	})
	fileTitle := "File"
	menuResp, _ := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Menu{
				Menu: &pb.MenuProperties{Title: &fileTitle},
			},
		},
	})
	menuSvc.AddMenu(ctx, &pb.MenuAddMenuRequest{
		SessionId: sid,
		MenuBarId: mbResp.WidgetId,
		MenuId:    menuResp.WidgetId,
	})

	// Add separator
	sepTrue := true
	sepResp, _ := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_ITEM,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_MenuItem{
				MenuItem: &pb.MenuItemProperties{
					Separator: &sepTrue,
				},
			},
		},
	})
	menuSvc.AddMenuItem(ctx, &pb.MenuAddMenuItemRequest{
		SessionId:  sid,
		MenuId:     menuResp.WidgetId,
		MenuItemId: sepResp.WidgetId,
	})

	// Add disabled item
	disLabel := "Disabled Action"
	disTrue := true
	disResp, _ := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_ITEM,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_MenuItem{
				MenuItem: &pb.MenuItemProperties{
					Label:    &disLabel,
					Disabled: &disTrue,
				},
			},
		},
	})
	menuSvc.AddMenuItem(ctx, &pb.MenuAddMenuItemRequest{
		SessionId:  sid,
		MenuId:     menuResp.WidgetId,
		MenuItemId: disResp.WidgetId,
	})

	// Verify
	mbPrim, _ := srv.Widgets().Get(mbResp.WidgetId.Id)
	mb := mbPrim.(*server.MenuBarPrimitive)
	items := mb.GetMenus()[0].Items
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}
	if !items[0].Separator {
		t.Error("Expected first item to be a separator")
	}
	if !items[1].Disabled {
		t.Error("Expected second item to be disabled")
	}
}

func TestMenuService_RemoveMenuItem(t *testing.T) {
	srv := setupTestServer(t)
	widgetSvc := NewWidgetService(srv)
	menuSvc := NewMenuService(srv)
	ctx := context.Background()

	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sid := &pb.SessionId{Id: session.ID}

	// Create full menu structure
	mbResp, _ := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_BAR,
	})
	fileTitle := "File"
	menuResp, _ := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Menu{
				Menu: &pb.MenuProperties{Title: &fileTitle},
			},
		},
	})
	menuSvc.AddMenu(ctx, &pb.MenuAddMenuRequest{
		SessionId: sid,
		MenuBarId: mbResp.WidgetId,
		MenuId:    menuResp.WidgetId,
	})

	newLabel := "New"
	itemResp, _ := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sid,
		Type:      pb.WidgetType_WIDGET_MENU_ITEM,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_MenuItem{
				MenuItem: &pb.MenuItemProperties{Label: &newLabel},
			},
		},
	})
	menuSvc.AddMenuItem(ctx, &pb.MenuAddMenuItemRequest{
		SessionId:  sid,
		MenuId:     menuResp.WidgetId,
		MenuItemId: itemResp.WidgetId,
	})

	// Remove item
	removeResp, err := menuSvc.RemoveMenuItem(ctx, &pb.MenuRemoveMenuItemRequest{
		SessionId:  sid,
		MenuId:     menuResp.WidgetId,
		MenuItemId: itemResp.WidgetId,
	})
	if err != nil {
		t.Fatalf("RemoveMenuItem failed: %v", err)
	}
	if !removeResp.Success {
		t.Error("Expected success=true")
	}

	// Verify empty
	mbPrim, _ := srv.Widgets().Get(mbResp.WidgetId.Id)
	mb := mbPrim.(*server.MenuBarPrimitive)
	items := mb.GetMenus()[0].Items
	if len(items) != 0 {
		t.Errorf("Expected 0 items after removal, got %d", len(items))
	}
}

func TestMenuService_ValidationErrors(t *testing.T) {
	srv := setupTestServer(t)
	menuSvc := NewMenuService(srv)
	ctx := context.Background()

	// Nil session/widget
	_, err := menuSvc.AddMenu(ctx, &pb.MenuAddMenuRequest{})
	if err == nil {
		t.Error("Expected error for nil fields")
	}

	_, err = menuSvc.AddMenuItem(ctx, &pb.MenuAddMenuItemRequest{})
	if err == nil {
		t.Error("Expected error for nil fields")
	}

	_, err = menuSvc.RemoveMenuItem(ctx, &pb.MenuRemoveMenuItemRequest{})
	if err == nil {
		t.Error("Expected error for nil fields")
	}
}
