package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
)

func TestListService_AddItem(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	widgetSvc := NewWidgetService(srv)
	listSvc := NewListService(srv)
	ctx := context.Background()

	// Create session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create list widget
	createResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_LIST,
	})
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	widgetID := createResp.WidgetId.Id

	// Add item
	addResp, err := listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId:     &pb.SessionId{Id: sessionID},
		WidgetId:      &pb.WidgetId{Id: widgetID},
		MainText:      "Item 1",
		SecondaryText: "Description 1",
	})
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	if !addResp.Success {
		t.Error("Expected success=true")
	}
	if addResp.Index != 0 {
		t.Errorf("Expected index=0, got %d", addResp.Index)
	}

	// Verify item count
	countResp, err := listSvc.GetItemCount(ctx, &pb.GetItemCountRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("Failed to get item count: %v", err)
	}

	if countResp.Count != 1 {
		t.Errorf("Expected count=1, got %d", countResp.Count)
	}
}

func TestListService_RemoveItem(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	widgetSvc := NewWidgetService(srv)
	listSvc := NewListService(srv)
	ctx := context.Background()

	// Create session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create list widget
	createResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_LIST,
	})
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	widgetID := createResp.WidgetId.Id

	// Add items
	listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		MainText:  "Item 1",
	})
	listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		MainText:  "Item 2",
	})

	// Remove first item
	removeResp, err := listSvc.RemoveItem(ctx, &pb.RemoveItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Index:     0,
	})
	if err != nil {
		t.Fatalf("Failed to remove item: %v", err)
	}

	if !removeResp.Success {
		t.Error("Expected success=true")
	}

	// Verify item count
	countResp, err := listSvc.GetItemCount(ctx, &pb.GetItemCountRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("Failed to get item count: %v", err)
	}

	if countResp.Count != 1 {
		t.Errorf("Expected count=1, got %d", countResp.Count)
	}
}

func TestListService_Clear(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	widgetSvc := NewWidgetService(srv)
	listSvc := NewListService(srv)
	ctx := context.Background()

	// Create session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create list widget
	createResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_LIST,
	})
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	widgetID := createResp.WidgetId.Id

	// Add items
	listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		MainText:  "Item 1",
	})
	listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		MainText:  "Item 2",
	})

	// Clear list
	clearResp, err := listSvc.Clear(ctx, &pb.ClearListRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("Failed to clear list: %v", err)
	}

	if !clearResp.Success {
		t.Error("Expected success=true")
	}

	// Verify item count
	countResp, err := listSvc.GetItemCount(ctx, &pb.GetItemCountRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("Failed to get item count: %v", err)
	}

	if countResp.Count != 0 {
		t.Errorf("Expected count=0, got %d", countResp.Count)
	}
}

func TestListService_GetSetSelected(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	widgetSvc := NewWidgetService(srv)
	listSvc := NewListService(srv)
	ctx := context.Background()

	// Create session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create list widget
	createResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_LIST,
	})
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	widgetID := createResp.WidgetId.Id

	// Add items
	listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		MainText:  "Item 1",
	})
	listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		MainText:  "Item 2",
	})

	// Set selected
	setResp, err := listSvc.SetSelected(ctx, &pb.SetSelectedRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Index:     1,
	})
	if err != nil {
		t.Fatalf("Failed to set selected: %v", err)
	}

	if !setResp.Success {
		t.Error("Expected success=true")
	}

	// Get selected
	getResp, err := listSvc.GetSelected(ctx, &pb.GetSelectedRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("Failed to get selected: %v", err)
	}

	if getResp.Index != 1 {
		t.Errorf("Expected index=1, got %d", getResp.Index)
	}
}

func TestListService_GetItem(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	widgetSvc := NewWidgetService(srv)
	listSvc := NewListService(srv)
	ctx := context.Background()

	// Create session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create list widget
	createResp, err := widgetSvc.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_LIST,
	})
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	widgetID := createResp.WidgetId.Id

	// Add item
	listSvc.AddItem(ctx, &pb.AddItemRequest{
		SessionId:     &pb.SessionId{Id: sessionID},
		WidgetId:      &pb.WidgetId{Id: widgetID},
		MainText:      "Item 1",
		SecondaryText: "Description 1",
	})

	// Get item
	getResp, err := listSvc.GetItem(ctx, &pb.GetItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Index:     0,
	})
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}

	if getResp.MainText != "Item 1" {
		t.Errorf("Expected main text 'Item 1', got '%s'", getResp.MainText)
	}
	if getResp.SecondaryText != "Description 1" {
		t.Errorf("Expected secondary text 'Description 1', got '%s'", getResp.SecondaryText)
	}
}


