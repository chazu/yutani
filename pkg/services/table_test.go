package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTableService_SetCell(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	tableService := NewTableService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a table widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TABLE,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Test SetCell
	text := "Test Cell"
	align := pb.Alignment_ALIGN_CENTER
	selectable := true
	expansion := int32(2)

	resp, err := tableService.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       0,
		Column:    0,
		Cell: &pb.TableCell{
			Text:       text,
			Align:      &align,
			Selectable: &selectable,
			Expansion:  &expansion,
		},
	})
	if err != nil {
		t.Fatalf("SetCell failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Test with invalid session
	_, err = tableService.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: &pb.SessionId{Id: "invalid"},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       0,
		Column:    0,
		Cell:      &pb.TableCell{Text: "test"},
	})
	if err == nil {
		t.Error("Expected error for invalid session")
	}
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", status.Code(err))
	}

	// Test with invalid widget
	_, err = tableService.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: "invalid"},
		Row:       0,
		Column:    0,
		Cell:      &pb.TableCell{Text: "test"},
	})
	if err == nil {
		t.Error("Expected error for invalid widget")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("Expected NotFound, got %v", status.Code(err))
	}
}

func TestTableService_GetCell(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	tableService := NewTableService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a table widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TABLE,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Set a cell first
	text := "Test Cell"
	_, err = tableService.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       1,
		Column:    2,
		Cell:      &pb.TableCell{Text: text},
	})
	if err != nil {
		t.Fatalf("SetCell failed: %v", err)
	}

	// Test GetCell
	resp, err := tableService.GetCell(ctx, &pb.GetTableCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       1,
		Column:    2,
	})
	if err != nil {
		t.Fatalf("GetCell failed: %v", err)
	}
	if resp.Cell.Text != text {
		t.Errorf("Expected text %q, got %q", text, resp.Cell.Text)
	}

	// Test GetCell for non-existent cell (tview creates cells on demand, so this won't error)
	// Instead, verify it returns an empty cell
	emptyResp, err := tableService.GetCell(ctx, &pb.GetTableCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       100,
		Column:    100,
	})
	if err != nil {
		t.Fatalf("GetCell failed for empty cell: %v", err)
	}
	if emptyResp.Cell.Text != "" {
		t.Errorf("Expected empty text for non-existent cell, got %q", emptyResp.Cell.Text)
	}
}

func TestTableService_SetCells(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	tableService := NewTableService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a table widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TABLE,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Test SetCells with multiple cells
	cells := []*pb.TableCellUpdate{
		{Row: 0, Column: 0, Cell: &pb.TableCell{Text: "A1"}},
		{Row: 0, Column: 1, Cell: &pb.TableCell{Text: "B1"}},
		{Row: 1, Column: 0, Cell: &pb.TableCell{Text: "A2"}},
		{Row: 1, Column: 1, Cell: &pb.TableCell{Text: "B2"}},
	}

	resp, err := tableService.SetCells(ctx, &pb.SetTableCellsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Cells:     cells,
	})
	if err != nil {
		t.Fatalf("SetCells failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.CellsSet != 4 {
		t.Errorf("Expected 4 cells set, got %d", resp.CellsSet)
	}

	// Verify cells were set
	getResp, err := tableService.GetCell(ctx, &pb.GetTableCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       1,
		Column:    1,
	})
	if err != nil {
		t.Fatalf("GetCell failed: %v", err)
	}
	if getResp.Cell.Text != "B2" {
		t.Errorf("Expected text 'B2', got %q", getResp.Cell.Text)
	}
}

func TestTableService_Clear(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	tableService := NewTableService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a table widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TABLE,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Add some cells
	_, err = tableService.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       0,
		Column:    0,
		Cell:      &pb.TableCell{Text: "Test"},
	})
	if err != nil {
		t.Fatalf("SetCell failed: %v", err)
	}

	// Test Clear
	resp, err := tableService.Clear(ctx, &pb.ClearTableRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Verify table is cleared
	dimResp, err := tableService.GetDimensions(ctx, &pb.GetDimensionsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetDimensions failed: %v", err)
	}
	if dimResp.Rows != 0 || dimResp.Columns != 0 {
		t.Errorf("Expected 0 rows and columns after clear, got %d rows, %d columns", dimResp.Rows, dimResp.Columns)
	}
}

func TestTableService_GetDimensions(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	tableService := NewTableService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a table widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TABLE,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Add cells to create dimensions
	cells := []*pb.TableCellUpdate{
		{Row: 0, Column: 0, Cell: &pb.TableCell{Text: "A1"}},
		{Row: 0, Column: 1, Cell: &pb.TableCell{Text: "B1"}},
		{Row: 1, Column: 0, Cell: &pb.TableCell{Text: "A2"}},
		{Row: 2, Column: 2, Cell: &pb.TableCell{Text: "C3"}},
	}
	_, err = tableService.SetCells(ctx, &pb.SetTableCellsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Cells:     cells,
	})
	if err != nil {
		t.Fatalf("SetCells failed: %v", err)
	}

	// Test GetDimensions
	resp, err := tableService.GetDimensions(ctx, &pb.GetDimensionsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetDimensions failed: %v", err)
	}
	if resp.Rows != 3 {
		t.Errorf("Expected 3 rows, got %d", resp.Rows)
	}
	if resp.Columns != 3 {
		t.Errorf("Expected 3 columns, got %d", resp.Columns)
	}
}

func TestTableService_GetSetSelection(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	tableService := NewTableService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a table widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TABLE,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Add some cells
	cells := []*pb.TableCellUpdate{
		{Row: 0, Column: 0, Cell: &pb.TableCell{Text: "A1"}},
		{Row: 1, Column: 1, Cell: &pb.TableCell{Text: "B2"}},
	}
	_, err = tableService.SetCells(ctx, &pb.SetTableCellsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Cells:     cells,
	})
	if err != nil {
		t.Fatalf("SetCells failed: %v", err)
	}

	// Test SetSelection
	setResp, err := tableService.SetSelection(ctx, &pb.SetSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Row:       1,
		Column:    1,
	})
	if err != nil {
		t.Fatalf("SetSelection failed: %v", err)
	}
	if !setResp.Success {
		t.Error("Expected success to be true")
	}

	// Test GetSelection
	getResp, err := tableService.GetSelection(ctx, &pb.GetSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetSelection failed: %v", err)
	}
	if getResp.Row != 1 {
		t.Errorf("Expected row 1, got %d", getResp.Row)
	}
	if getResp.Column != 1 {
		t.Errorf("Expected column 1, got %d", getResp.Column)
	}
}

func TestTableService_SetFixed(t *testing.T) {
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	tableService := NewTableService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a table widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TABLE,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Test SetFixed
	resp, err := tableService.SetFixed(ctx, &pb.SetFixedRequest{
		SessionId:    &pb.SessionId{Id: sessionID},
		WidgetId:     &pb.WidgetId{Id: widgetID},
		FixedRows:    1,
		FixedColumns: 2,
	})
	if err != nil {
		t.Fatalf("SetFixed failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Note: tview doesn't expose GetFixed, so we can't verify the values
	// But we can verify the call succeeded without error
}

