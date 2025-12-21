package services

import (
	"context"
	"testing"

	"github.com/chazu/yutani/pkg/server"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestLayoutService_FlexAddItem(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a flex container
	flexResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FLEX,
	})
	if err != nil {
		t.Fatalf("CreateWidget (flex) failed: %v", err)
	}
	flexID := flexResp.WidgetId.Id

	// Create a text view to add to the flex
	textResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text) failed: %v", err)
	}
	textID := textResp.WidgetId.Id

	// Test FlexAddItem with proportion
	resp, err := layoutService.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		FlexId:     &pb.WidgetId{Id: flexID},
		ItemId:     &pb.WidgetId{Id: textID},
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_FlexRemoveItem(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a flex container
	flexResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FLEX,
	})
	if err != nil {
		t.Fatalf("CreateWidget (flex) failed: %v", err)
	}
	flexID := flexResp.WidgetId.Id

	// Create a text view to add to the flex
	textResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text) failed: %v", err)
	}
	textID := textResp.WidgetId.Id

	// Add the item first
	_, err = layoutService.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		FlexId:     &pb.WidgetId{Id: flexID},
		ItemId:     &pb.WidgetId{Id: textID},
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem failed: %v", err)
	}

	// Test FlexRemoveItem
	resp, err := layoutService.FlexRemoveItem(ctx, &pb.FlexRemoveItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		FlexId:    &pb.WidgetId{Id: flexID},
		ItemId:    &pb.WidgetId{Id: textID},
	})
	if err != nil {
		t.Fatalf("FlexRemoveItem failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_FlexSetDirection(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a flex container
	flexResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_FLEX,
	})
	if err != nil {
		t.Fatalf("CreateWidget (flex) failed: %v", err)
	}
	flexID := flexResp.WidgetId.Id

	// Test FlexSetDirection to column
	resp, err := layoutService.FlexSetDirection(ctx, &pb.FlexSetDirectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		FlexId:    &pb.WidgetId{Id: flexID},
		Direction: pb.FlexDirection_FLEX_COLUMN,
	})
	if err != nil {
		t.Fatalf("FlexSetDirection (column) failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Test FlexSetDirection to row
	resp2, err := layoutService.FlexSetDirection(ctx, &pb.FlexSetDirectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		FlexId:    &pb.WidgetId{Id: flexID},
		Direction: pb.FlexDirection_FLEX_ROW,
	})
	if err != nil {
		t.Fatalf("FlexSetDirection (row) failed: %v", err)
	}
	if !resp2.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_GridAddItem(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a grid container
	gridResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_GRID,
	})
	if err != nil {
		t.Fatalf("CreateWidget (grid) failed: %v", err)
	}
	gridID := gridResp.WidgetId.Id

	// Create a text view to add to the grid
	textResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text) failed: %v", err)
	}
	textID := textResp.WidgetId.Id

	// Test GridAddItem
	resp, err := layoutService.GridAddItem(ctx, &pb.GridAddItemRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		GridId:     &pb.WidgetId{Id: gridID},
		ItemId:     &pb.WidgetId{Id: textID},
		Row:        0,
		Column:     0,
		RowSpan:    1,
		ColumnSpan: 1,
		MinWidth:   10,
		MinHeight:  5,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("GridAddItem failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_GridRemoveItem(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a grid container
	gridResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_GRID,
	})
	if err != nil {
		t.Fatalf("CreateWidget (grid) failed: %v", err)
	}
	gridID := gridResp.WidgetId.Id

	// Create a text view to add to the grid
	textResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text) failed: %v", err)
	}
	textID := textResp.WidgetId.Id

	// Add the item first
	_, err = layoutService.GridAddItem(ctx, &pb.GridAddItemRequest{
		SessionId:  &pb.SessionId{Id: sessionID},
		GridId:     &pb.WidgetId{Id: gridID},
		ItemId:     &pb.WidgetId{Id: textID},
		Row:        0,
		Column:     0,
		RowSpan:    1,
		ColumnSpan: 1,
		MinWidth:   10,
		MinHeight:  5,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("GridAddItem failed: %v", err)
	}

	// Test GridRemoveItem
	resp, err := layoutService.GridRemoveItem(ctx, &pb.GridRemoveItemRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		GridId:    &pb.WidgetId{Id: gridID},
		ItemId:    &pb.WidgetId{Id: textID},
	})
	if err != nil {
		t.Fatalf("GridRemoveItem failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_GridSetRows(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a grid container
	gridResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_GRID,
	})
	if err != nil {
		t.Fatalf("CreateWidget (grid) failed: %v", err)
	}
	gridID := gridResp.WidgetId.Id

	// Test GridSetRows
	resp, err := layoutService.GridSetRows(ctx, &pb.GridSetRowsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		GridId:    &pb.WidgetId{Id: gridID},
		RowSizes:  []int32{10, -1, 5}, // Fixed, proportional, fixed
	})
	if err != nil {
		t.Fatalf("GridSetRows failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_GridSetColumns(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a grid container
	gridResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_GRID,
	})
	if err != nil {
		t.Fatalf("CreateWidget (grid) failed: %v", err)
	}
	gridID := gridResp.WidgetId.Id

	// Test GridSetColumns
	resp, err := layoutService.GridSetColumns(ctx, &pb.GridSetColumnsRequest{
		SessionId:   &pb.SessionId{Id: sessionID},
		GridId:      &pb.WidgetId{Id: gridID},
		ColumnSizes: []int32{20, -2, -1}, // Fixed, proportional, proportional
	})
	if err != nil {
		t.Fatalf("GridSetColumns failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_PagesAddPage(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a pages container
	pagesResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_PAGES,
	})
	if err != nil {
		t.Fatalf("CreateWidget (pages) failed: %v", err)
	}
	pagesID := pagesResp.WidgetId.Id

	// Create a text view to add as a page
	textResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text) failed: %v", err)
	}
	textID := textResp.WidgetId.Id

	// Test PagesAddPage
	resp, err := layoutService.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		PagesId:   &pb.WidgetId{Id: pagesID},
		PageName:  "page1",
		ItemId:    &pb.WidgetId{Id: textID},
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_PagesRemovePage(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a pages container
	pagesResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_PAGES,
	})
	if err != nil {
		t.Fatalf("CreateWidget (pages) failed: %v", err)
	}
	pagesID := pagesResp.WidgetId.Id

	// Create a text view to add as a page
	textResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text) failed: %v", err)
	}
	textID := textResp.WidgetId.Id

	// Add the page first
	_, err = layoutService.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		PagesId:   &pb.WidgetId{Id: pagesID},
		PageName:  "page1",
		ItemId:    &pb.WidgetId{Id: textID},
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage failed: %v", err)
	}

	// Test PagesRemovePage
	resp, err := layoutService.PagesRemovePage(ctx, &pb.PagesRemovePageRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		PagesId:   &pb.WidgetId{Id: pagesID},
		PageName:  "page1",
	})
	if err != nil {
		t.Fatalf("PagesRemovePage failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestLayoutService_PagesShowGetPage(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	layoutService := NewLayoutService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a pages container
	pagesResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_PAGES,
	})
	if err != nil {
		t.Fatalf("CreateWidget (pages) failed: %v", err)
	}
	pagesID := pagesResp.WidgetId.Id

	// Create two text views to add as pages
	text1Resp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text1) failed: %v", err)
	}
	text1ID := text1Resp.WidgetId.Id

	text2Resp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget (text2) failed: %v", err)
	}
	text2ID := text2Resp.WidgetId.Id

	// Add two pages
	_, err = layoutService.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		PagesId:   &pb.WidgetId{Id: pagesID},
		PageName:  "page1",
		ItemId:    &pb.WidgetId{Id: text1ID},
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage (page1) failed: %v", err)
	}

	_, err = layoutService.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		PagesId:   &pb.WidgetId{Id: pagesID},
		PageName:  "page2",
		ItemId:    &pb.WidgetId{Id: text2ID},
		Resize:    true,
		Visible:   false,
	})
	if err != nil {
		t.Fatalf("PagesAddPage (page2) failed: %v", err)
	}

	// Test PagesShowPage
	showResp, err := layoutService.PagesShowPage(ctx, &pb.PagesShowPageRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		PagesId:   &pb.WidgetId{Id: pagesID},
		PageName:  "page2",
	})
	if err != nil {
		t.Fatalf("PagesShowPage failed: %v", err)
	}
	if !showResp.Success {
		t.Error("Expected success to be true")
	}

	// Test PagesGetCurrentPage
	getResp, err := layoutService.PagesGetCurrentPage(ctx, &pb.PagesGetCurrentPageRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		PagesId:   &pb.WidgetId{Id: pagesID},
	})
	if err != nil {
		t.Fatalf("PagesGetCurrentPage failed: %v", err)
	}
	if getResp.PageName != "page2" {
		t.Errorf("Expected current page 'page2', got %q", getResp.PageName)
	}
}

