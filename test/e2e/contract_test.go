package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------

// requireSession creates a session and returns the SessionId, failing the test
// on error.
func requireSession(t *testing.T, ctx context.Context, client pb.SessionServiceClient, name string) *pb.SessionId {
	t.Helper()
	resp, err := client.CreateSession(ctx, &pb.CreateSessionRequest{ClientName: name})
	if err != nil {
		t.Fatalf("CreateSession(%s) failed: %v", name, err)
	}
	if resp.SessionId == nil || resp.SessionId.Id == "" {
		t.Fatal("CreateSession returned nil/empty session_id")
	}
	return resp.SessionId
}

// requireWidget creates a widget and returns its WidgetId, failing the test on
// error.
func requireWidget(t *testing.T, ctx context.Context, client pb.WidgetServiceClient, sessionID *pb.SessionId, wtype pb.WidgetType, title string) *pb.WidgetId {
	t.Helper()
	props := &pb.WidgetProperties{
		Border: testutil.BoolPtr(true),
		Title:  testutil.StrPtr(title),
	}
	resp, err := client.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId:  sessionID,
		Type:       wtype,
		Properties: props,
	})
	if err != nil {
		t.Fatalf("CreateWidget(%v, %s) failed: %v", wtype, title, err)
	}
	if resp.WidgetId == nil || resp.WidgetId.Id == "" {
		t.Fatal("CreateWidget returned nil/empty widget_id")
	}
	return resp.WidgetId
}

// assertGRPCCode checks that err carries the expected gRPC status code.
func assertGRPCCode(t *testing.T, err error, expected codes.Code, label string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error with code %v, got nil", label, expected)
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Errorf("%s: expected gRPC status error, got %v", label, err)
		return
	}
	if st.Code() != expected {
		t.Errorf("%s: expected code %v, got %v (%s)", label, expected, st.Code(), st.Message())
	}
}

// fakeSessionID returns a SessionId that does not correspond to any real session.
func fakeSessionID() *pb.SessionId {
	return &pb.SessionId{Id: "non-existent-session-00000000"}
}

// fakeWidgetID returns a WidgetId that does not correspond to any real widget.
func fakeWidgetID() *pb.WidgetId {
	return &pb.WidgetId{Id: "non-existent-widget-00000000"}
}

// --------------------------------------------------------------------
// 1. SessionService
// --------------------------------------------------------------------

func TestContract_SessionService(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewSessionServiceClient(conn)

	// --- CreateSession ---
	t.Run("CreateSession_valid", func(t *testing.T) {
		resp, err := client.CreateSession(ctx, &pb.CreateSessionRequest{
			ClientName: "contract-session",
			Preferences: &pb.SessionPreferences{
				ReceiveKeyEvents:    true,
				ReceiveMouseEvents:  true,
				ReceiveResizeEvents: true,
			},
		})
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}
		if resp.SessionId == nil || resp.SessionId.Id == "" {
			t.Fatal("Expected non-empty session_id")
		}
		if resp.ServerVersion == "" {
			t.Error("Expected non-empty server_version")
		}

		// Clean up
		client.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: resp.SessionId})
	})

	// --- DestroySession ---
	t.Run("DestroySession_valid", func(t *testing.T) {
		sid := requireSession(t, ctx, client, "destroy-valid")
		resp, err := client.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid})
		if err != nil {
			t.Fatalf("DestroySession failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	t.Run("DestroySession_nil_session_id", func(t *testing.T) {
		_, err := client.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: nil})
		assertGRPCCode(t, err, codes.InvalidArgument, "DestroySession(nil)")
	})

	t.Run("DestroySession_non_existent", func(t *testing.T) {
		resp, err := client.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: fakeSessionID()})
		// The implementation returns success=false (not an error) for non-existent sessions
		if err != nil {
			t.Fatalf("DestroySession(non-existent) unexpected error: %v", err)
		}
		if resp.Success {
			t.Error("Expected success=false for non-existent session")
		}
	})

	// --- ListSessions ---
	t.Run("ListSessions_returns_created", func(t *testing.T) {
		sid1 := requireSession(t, ctx, client, "list-sess-1")
		sid2 := requireSession(t, ctx, client, "list-sess-2")
		defer client.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid1})
		defer client.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid2})

		resp, err := client.ListSessions(ctx, &pb.ListSessionsRequest{})
		if err != nil {
			t.Fatalf("ListSessions failed: %v", err)
		}

		found := map[string]bool{}
		for _, s := range resp.Sessions {
			found[s.SessionId.Id] = true
		}
		if !found[sid1.Id] {
			t.Errorf("ListSessions missing session %s", sid1.Id)
		}
		if !found[sid2.Id] {
			t.Errorf("ListSessions missing session %s", sid2.Id)
		}
	})

	// --- GetServerInfo ---
	t.Run("GetServerInfo", func(t *testing.T) {
		resp, err := client.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
		if err != nil {
			t.Fatalf("GetServerInfo failed: %v", err)
		}
		if resp.Version == "" {
			t.Error("Expected non-empty version")
		}
		if resp.MaxSessions != 10 {
			t.Errorf("Expected max_sessions=10, got %d", resp.MaxSessions)
		}
		if resp.Capabilities == nil {
			t.Error("Expected non-nil capabilities")
		} else if len(resp.Capabilities.SupportedWidgets) == 0 {
			t.Error("Expected non-empty supported_widgets list")
		}
	})

	// --- Ping ---
	t.Run("Ping_returns_timestamps", func(t *testing.T) {
		now := time.Now().UnixNano()
		resp, err := client.Ping(ctx, &pb.PingRequest{Timestamp: now})
		if err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
		if resp.Timestamp != now {
			t.Errorf("Expected echoed timestamp %d, got %d", now, resp.Timestamp)
		}
		if resp.ServerTimestamp == 0 {
			t.Error("Expected non-zero server_timestamp")
		}
	})
}

// --------------------------------------------------------------------
// 2. WidgetService
// --------------------------------------------------------------------

func TestContract_WidgetService(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)

	sid := requireSession(t, ctx, sessionClient, "widget-contract")
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid})

	// --- CreateWidget ---
	t.Run("CreateWidget_valid", func(t *testing.T) {
		resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
			SessionId: sid,
			Type:      pb.WidgetType_WIDGET_BOX,
			Properties: &pb.WidgetProperties{
				Border: testutil.BoolPtr(true),
				Title:  testutil.StrPtr("Contract Box"),
			},
		})
		if err != nil {
			t.Fatalf("CreateWidget failed: %v", err)
		}
		if resp.WidgetId == nil || resp.WidgetId.Id == "" {
			t.Fatal("Expected non-empty widget_id")
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	t.Run("CreateWidget_nil_session_id", func(t *testing.T) {
		_, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
			SessionId: nil,
			Type:      pb.WidgetType_WIDGET_BOX,
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "CreateWidget(nil session)")
	})

	t.Run("CreateWidget_non_existent_session", func(t *testing.T) {
		_, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
			SessionId: fakeSessionID(),
			Type:      pb.WidgetType_WIDGET_BOX,
		})
		assertGRPCCode(t, err, codes.NotFound, "CreateWidget(fake session)")
	})

	// --- DeleteWidget ---
	t.Run("DeleteWidget_valid", func(t *testing.T) {
		wid := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Del Box")
		resp, err := widgetClient.DeleteWidget(ctx, &pb.DeleteWidgetRequest{
			SessionId: sid,
			WidgetId:  wid,
		})
		if err != nil {
			t.Fatalf("DeleteWidget failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	t.Run("DeleteWidget_nil_session_id", func(t *testing.T) {
		_, err := widgetClient.DeleteWidget(ctx, &pb.DeleteWidgetRequest{
			SessionId: nil,
			WidgetId:  fakeWidgetID(),
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "DeleteWidget(nil session)")
	})

	t.Run("DeleteWidget_non_existent_widget", func(t *testing.T) {
		_, err := widgetClient.DeleteWidget(ctx, &pb.DeleteWidgetRequest{
			SessionId: sid,
			WidgetId:  fakeWidgetID(),
		})
		assertGRPCCode(t, err, codes.NotFound, "DeleteWidget(fake widget)")
	})

	// --- SetProperties / GetProperties round-trip ---
	t.Run("SetProperties_GetProperties_roundtrip", func(t *testing.T) {
		wid := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Props Box")

		// SetProperties
		setResp, err := widgetClient.SetProperties(ctx, &pb.SetPropertiesRequest{
			SessionId: sid,
			WidgetId:  wid,
			Properties: &pb.WidgetProperties{
				Title: testutil.StrPtr("Updated Title"),
			},
		})
		if err != nil {
			t.Fatalf("SetProperties failed: %v", err)
		}
		if !setResp.Success {
			t.Error("Expected success=true from SetProperties")
		}

		// GetProperties
		getResp, err := widgetClient.GetProperties(ctx, &pb.GetPropertiesRequest{
			SessionId: sid,
			WidgetId:  wid,
		})
		if err != nil {
			t.Fatalf("GetProperties failed: %v", err)
		}
		if getResp.Properties == nil {
			t.Fatal("Expected non-nil properties")
		}
		if getResp.Properties.GetTitle() != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", getResp.Properties.GetTitle())
		}
	})

	t.Run("SetProperties_nil_session_id", func(t *testing.T) {
		_, err := widgetClient.SetProperties(ctx, &pb.SetPropertiesRequest{
			SessionId:  nil,
			WidgetId:   fakeWidgetID(),
			Properties: &pb.WidgetProperties{},
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetProperties(nil session)")
	})

	t.Run("GetProperties_non_existent_widget", func(t *testing.T) {
		_, err := widgetClient.GetProperties(ctx, &pb.GetPropertiesRequest{
			SessionId: sid,
			WidgetId:  fakeWidgetID(),
		})
		assertGRPCCode(t, err, codes.NotFound, "GetProperties(fake widget)")
	})

	// --- SetRoot ---
	t.Run("SetRoot_valid", func(t *testing.T) {
		wid := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Root Box")
		resp, err := widgetClient.SetRoot(ctx, &pb.SetRootRequest{
			SessionId: sid,
			WidgetId:  wid,
		})
		if err != nil {
			t.Fatalf("SetRoot failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true from SetRoot")
		}
	})

	t.Run("SetRoot_non_existent_widget", func(t *testing.T) {
		_, err := widgetClient.SetRoot(ctx, &pb.SetRootRequest{
			SessionId: sid,
			WidgetId:  fakeWidgetID(),
		})
		assertGRPCCode(t, err, codes.NotFound, "SetRoot(fake widget)")
	})

	t.Run("SetRoot_nil_session_id", func(t *testing.T) {
		_, err := widgetClient.SetRoot(ctx, &pb.SetRootRequest{
			SessionId: nil,
			WidgetId:  fakeWidgetID(),
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetRoot(nil session)")
	})

	// --- SetFocus / GetFocus ---
	t.Run("SetFocus_GetFocus_roundtrip", func(t *testing.T) {
		wid := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Focus Box")

		// First set as root so it can receive focus
		widgetClient.SetRoot(ctx, &pb.SetRootRequest{SessionId: sid, WidgetId: wid})

		setResp, err := widgetClient.SetFocus(ctx, &pb.SetFocusRequest{
			SessionId: sid,
			WidgetId:  wid,
		})
		if err != nil {
			t.Fatalf("SetFocus failed: %v", err)
		}
		if !setResp.Success {
			t.Error("Expected success=true from SetFocus")
		}

		getResp, err := widgetClient.GetFocus(ctx, &pb.GetFocusRequest{SessionId: sid})
		if err != nil {
			t.Fatalf("GetFocus failed: %v", err)
		}
		if getResp.WidgetId == nil {
			t.Error("Expected non-nil focused widget_id")
		} else if getResp.WidgetId.Id != wid.Id {
			t.Errorf("Expected focused widget %s, got %s", wid.Id, getResp.WidgetId.Id)
		}
	})

	t.Run("SetFocus_nil_session_id", func(t *testing.T) {
		_, err := widgetClient.SetFocus(ctx, &pb.SetFocusRequest{SessionId: nil, WidgetId: fakeWidgetID()})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetFocus(nil session)")
	})

	t.Run("GetFocus_non_existent_session", func(t *testing.T) {
		_, err := widgetClient.GetFocus(ctx, &pb.GetFocusRequest{SessionId: fakeSessionID()})
		assertGRPCCode(t, err, codes.NotFound, "GetFocus(fake session)")
	})

	// --- ListWidgets ---
	t.Run("ListWidgets_returns_all", func(t *testing.T) {
		// Create a fresh session to have a known widget count
		freshSid := requireSession(t, ctx, sessionClient, "list-widgets-sess")
		defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: freshSid})

		w1 := requireWidget(t, ctx, widgetClient, freshSid, pb.WidgetType_WIDGET_BOX, "LW1")
		w2 := requireWidget(t, ctx, widgetClient, freshSid, pb.WidgetType_WIDGET_LIST, "LW2")

		resp, err := widgetClient.ListWidgets(ctx, &pb.ListWidgetsRequest{SessionId: freshSid})
		if err != nil {
			t.Fatalf("ListWidgets failed: %v", err)
		}
		if len(resp.Widgets) != 2 {
			t.Errorf("Expected 2 widgets, got %d", len(resp.Widgets))
		}
		found := map[string]bool{}
		for _, w := range resp.Widgets {
			found[w.WidgetId.Id] = true
		}
		if !found[w1.Id] || !found[w2.Id] {
			t.Errorf("ListWidgets missing expected widgets")
		}
	})

	t.Run("ListWidgets_nil_session_id", func(t *testing.T) {
		_, err := widgetClient.ListWidgets(ctx, &pb.ListWidgetsRequest{SessionId: nil})
		assertGRPCCode(t, err, codes.InvalidArgument, "ListWidgets(nil session)")
	})

	t.Run("ListWidgets_non_existent_session", func(t *testing.T) {
		_, err := widgetClient.ListWidgets(ctx, &pb.ListWidgetsRequest{SessionId: fakeSessionID()})
		assertGRPCCode(t, err, codes.NotFound, "ListWidgets(fake session)")
	})
}

// --------------------------------------------------------------------
// 3. ListService
// --------------------------------------------------------------------

func TestContract_ListService(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	listClient := pb.NewListServiceClient(conn)

	sid := requireSession(t, ctx, sessionClient, "list-contract")
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid})

	listID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_LIST, "Contract List")

	// --- AddItem ---
	t.Run("AddItem_valid", func(t *testing.T) {
		resp, err := listClient.AddItem(ctx, &pb.ListAddItemRequest{
			SessionId:     sid,
			WidgetId:      listID,
			MainText:      "Item A",
			SecondaryText: "Desc A",
		})
		if err != nil {
			t.Fatalf("AddItem failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
		if resp.Index != 0 {
			t.Errorf("Expected index 0, got %d", resp.Index)
		}
	})

	// Add more items for subsequent tests
	listClient.AddItem(ctx, &pb.ListAddItemRequest{SessionId: sid, WidgetId: listID, MainText: "Item B", SecondaryText: "Desc B"})
	listClient.AddItem(ctx, &pb.ListAddItemRequest{SessionId: sid, WidgetId: listID, MainText: "Item C", SecondaryText: "Desc C"})

	// --- GetItemCount ---
	t.Run("GetItemCount", func(t *testing.T) {
		resp, err := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{SessionId: sid, WidgetId: listID})
		if err != nil {
			t.Fatalf("GetItemCount failed: %v", err)
		}
		if resp.Count != 3 {
			t.Errorf("Expected count 3, got %d", resp.Count)
		}
	})

	// --- GetItem ---
	t.Run("GetItem_valid", func(t *testing.T) {
		resp, err := listClient.GetItem(ctx, &pb.ListGetItemRequest{SessionId: sid, WidgetId: listID, Index: 0})
		if err != nil {
			t.Fatalf("GetItem failed: %v", err)
		}
		if resp.MainText != "Item A" {
			t.Errorf("Expected 'Item A', got '%s'", resp.MainText)
		}
		if resp.SecondaryText != "Desc A" {
			t.Errorf("Expected 'Desc A', got '%s'", resp.SecondaryText)
		}
	})

	// --- SetSelection / GetSelection ---
	t.Run("SetSelection_GetSelection_roundtrip", func(t *testing.T) {
		setResp, err := listClient.SetSelection(ctx, &pb.ListSetSelectionRequest{SessionId: sid, WidgetId: listID, Index: 2})
		if err != nil {
			t.Fatalf("SetSelection failed: %v", err)
		}
		if !setResp.Success {
			t.Error("Expected success=true from SetSelection")
		}

		getResp, err := listClient.GetSelection(ctx, &pb.ListGetSelectionRequest{SessionId: sid, WidgetId: listID})
		if err != nil {
			t.Fatalf("GetSelection failed: %v", err)
		}
		if getResp.Index != 2 {
			t.Errorf("Expected selected index 2, got %d", getResp.Index)
		}
	})

	// --- RemoveItem ---
	t.Run("RemoveItem_valid", func(t *testing.T) {
		resp, err := listClient.RemoveItem(ctx, &pb.ListRemoveItemRequest{SessionId: sid, WidgetId: listID, Index: 1})
		if err != nil {
			t.Fatalf("RemoveItem failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true from RemoveItem")
		}

		// Verify count decreased
		countResp, _ := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{SessionId: sid, WidgetId: listID})
		if countResp.Count != 2 {
			t.Errorf("Expected count 2 after removal, got %d", countResp.Count)
		}
	})

	// --- Clear ---
	t.Run("Clear_valid", func(t *testing.T) {
		resp, err := listClient.Clear(ctx, &pb.ListClearRequest{SessionId: sid, WidgetId: listID})
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true from Clear")
		}

		countResp, _ := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{SessionId: sid, WidgetId: listID})
		if countResp.Count != 0 {
			t.Errorf("Expected count 0 after clear, got %d", countResp.Count)
		}
	})

	// --- Error cases ---
	t.Run("AddItem_nil_session_id", func(t *testing.T) {
		_, err := listClient.AddItem(ctx, &pb.ListAddItemRequest{SessionId: nil, WidgetId: listID, MainText: "X"})
		assertGRPCCode(t, err, codes.InvalidArgument, "AddItem(nil session)")
	})

	t.Run("AddItem_non_existent_widget", func(t *testing.T) {
		_, err := listClient.AddItem(ctx, &pb.ListAddItemRequest{SessionId: sid, WidgetId: fakeWidgetID(), MainText: "X"})
		assertGRPCCode(t, err, codes.NotFound, "AddItem(fake widget)")
	})

	t.Run("GetItemCount_nil_session_id", func(t *testing.T) {
		_, err := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{SessionId: nil, WidgetId: listID})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetItemCount(nil session)")
	})

	t.Run("GetSelection_nil_session_id", func(t *testing.T) {
		_, err := listClient.GetSelection(ctx, &pb.ListGetSelectionRequest{SessionId: nil, WidgetId: listID})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetSelection(nil session)")
	})

	t.Run("RemoveItem_nil_session_id", func(t *testing.T) {
		_, err := listClient.RemoveItem(ctx, &pb.ListRemoveItemRequest{SessionId: nil, WidgetId: listID, Index: 0})
		assertGRPCCode(t, err, codes.InvalidArgument, "RemoveItem(nil session)")
	})

	t.Run("Clear_nil_session_id", func(t *testing.T) {
		_, err := listClient.Clear(ctx, &pb.ListClearRequest{SessionId: nil, WidgetId: listID})
		assertGRPCCode(t, err, codes.InvalidArgument, "Clear(nil session)")
	})

	t.Run("SetSelection_nil_session_id", func(t *testing.T) {
		_, err := listClient.SetSelection(ctx, &pb.ListSetSelectionRequest{SessionId: nil, WidgetId: listID, Index: 0})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetSelection(nil session)")
	})

	t.Run("GetItem_nil_session_id", func(t *testing.T) {
		_, err := listClient.GetItem(ctx, &pb.ListGetItemRequest{SessionId: nil, WidgetId: listID, Index: 0})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetItem(nil session)")
	})
}

// --------------------------------------------------------------------
// 4. TableService
// --------------------------------------------------------------------

func TestContract_TableService(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	tableClient := pb.NewTableServiceClient(conn)

	sid := requireSession(t, ctx, sessionClient, "table-contract")
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid})

	tableID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_TABLE, "Contract Table")

	// --- SetCell ---
	t.Run("SetCell_valid", func(t *testing.T) {
		resp, err := tableClient.SetCell(ctx, &pb.SetTableCellRequest{
			SessionId: sid,
			WidgetId:  tableID,
			Row:       0,
			Column:    0,
			Cell:      &pb.TableCell{Text: "Header 1"},
		})
		if err != nil {
			t.Fatalf("SetCell failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	// Populate more cells for subsequent tests
	tableClient.SetCell(ctx, &pb.SetTableCellRequest{SessionId: sid, WidgetId: tableID, Row: 0, Column: 1, Cell: &pb.TableCell{Text: "Header 2"}})
	tableClient.SetCell(ctx, &pb.SetTableCellRequest{SessionId: sid, WidgetId: tableID, Row: 1, Column: 0, Cell: &pb.TableCell{Text: "R1C1"}})
	tableClient.SetCell(ctx, &pb.SetTableCellRequest{SessionId: sid, WidgetId: tableID, Row: 1, Column: 1, Cell: &pb.TableCell{Text: "R1C2"}})

	// --- GetCell ---
	t.Run("GetCell_valid", func(t *testing.T) {
		resp, err := tableClient.GetCell(ctx, &pb.GetTableCellRequest{SessionId: sid, WidgetId: tableID, Row: 0, Column: 0})
		if err != nil {
			t.Fatalf("GetCell failed: %v", err)
		}
		if resp.Cell == nil {
			t.Fatal("Expected non-nil cell")
		}
		if resp.Cell.Text != "Header 1" {
			t.Errorf("Expected 'Header 1', got '%s'", resp.Cell.Text)
		}
	})

	// --- SetCells ---
	t.Run("SetCells_valid", func(t *testing.T) {
		resp, err := tableClient.SetCells(ctx, &pb.SetTableCellsRequest{
			SessionId: sid,
			WidgetId:  tableID,
			Cells: []*pb.TableCellUpdate{
				{Row: 2, Column: 0, Cell: &pb.TableCell{Text: "R2C1"}},
				{Row: 2, Column: 1, Cell: &pb.TableCell{Text: "R2C2"}},
			},
		})
		if err != nil {
			t.Fatalf("SetCells failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
		if resp.CellsSet != 2 {
			t.Errorf("Expected cells_set=2, got %d", resp.CellsSet)
		}
	})

	// --- GetDimensions ---
	t.Run("GetDimensions_valid", func(t *testing.T) {
		resp, err := tableClient.GetDimensions(ctx, &pb.TableGetDimensionsRequest{SessionId: sid, WidgetId: tableID})
		if err != nil {
			t.Fatalf("GetDimensions failed: %v", err)
		}
		if resp.Rows != 3 {
			t.Errorf("Expected 3 rows, got %d", resp.Rows)
		}
		if resp.Columns != 2 {
			t.Errorf("Expected 2 columns, got %d", resp.Columns)
		}
	})

	// --- SetSelection / GetSelection ---
	t.Run("SetSelection_GetSelection_roundtrip", func(t *testing.T) {
		setResp, err := tableClient.SetSelection(ctx, &pb.TableSetSelectionRequest{SessionId: sid, WidgetId: tableID, Row: 1, Column: 1})
		if err != nil {
			t.Fatalf("SetSelection failed: %v", err)
		}
		if !setResp.Success {
			t.Error("Expected success=true")
		}

		getResp, err := tableClient.GetSelection(ctx, &pb.TableGetSelectionRequest{SessionId: sid, WidgetId: tableID})
		if err != nil {
			t.Fatalf("GetSelection failed: %v", err)
		}
		if getResp.Row != 1 || getResp.Column != 1 {
			t.Errorf("Expected selection (1,1), got (%d,%d)", getResp.Row, getResp.Column)
		}
	})

	// --- SetFixed ---
	t.Run("SetFixed_valid", func(t *testing.T) {
		resp, err := tableClient.SetFixed(ctx, &pb.TableSetFixedRequest{
			SessionId:    sid,
			WidgetId:     tableID,
			FixedRows:    1,
			FixedColumns: 0,
		})
		if err != nil {
			t.Fatalf("SetFixed failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	// --- Clear ---
	t.Run("Clear_valid", func(t *testing.T) {
		resp, err := tableClient.Clear(ctx, &pb.TableClearRequest{SessionId: sid, WidgetId: tableID})
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}

		dimsResp, _ := tableClient.GetDimensions(ctx, &pb.TableGetDimensionsRequest{SessionId: sid, WidgetId: tableID})
		if dimsResp.Rows != 0 || dimsResp.Columns != 0 {
			t.Errorf("Expected 0x0 after clear, got %dx%d", dimsResp.Rows, dimsResp.Columns)
		}
	})

	// --- Error cases ---
	t.Run("SetCell_nil_session_id", func(t *testing.T) {
		_, err := tableClient.SetCell(ctx, &pb.SetTableCellRequest{
			SessionId: nil, WidgetId: tableID, Row: 0, Column: 0,
			Cell: &pb.TableCell{Text: "X"},
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetCell(nil session)")
	})

	t.Run("GetCell_nil_session_id", func(t *testing.T) {
		_, err := tableClient.GetCell(ctx, &pb.GetTableCellRequest{SessionId: nil, WidgetId: tableID, Row: 0, Column: 0})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetCell(nil session)")
	})

	t.Run("SetCells_nil_session_id", func(t *testing.T) {
		_, err := tableClient.SetCells(ctx, &pb.SetTableCellsRequest{
			SessionId: nil, WidgetId: tableID,
			Cells: []*pb.TableCellUpdate{{Row: 0, Column: 0, Cell: &pb.TableCell{Text: "X"}}},
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetCells(nil session)")
	})

	t.Run("Clear_nil_session_id", func(t *testing.T) {
		_, err := tableClient.Clear(ctx, &pb.TableClearRequest{SessionId: nil, WidgetId: tableID})
		assertGRPCCode(t, err, codes.InvalidArgument, "Clear(nil session)")
	})

	t.Run("GetDimensions_nil_session_id", func(t *testing.T) {
		_, err := tableClient.GetDimensions(ctx, &pb.TableGetDimensionsRequest{SessionId: nil, WidgetId: tableID})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetDimensions(nil session)")
	})

	t.Run("GetSelection_nil_session_id", func(t *testing.T) {
		_, err := tableClient.GetSelection(ctx, &pb.TableGetSelectionRequest{SessionId: nil, WidgetId: tableID})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetSelection(nil session)")
	})

	t.Run("SetSelection_nil_session_id", func(t *testing.T) {
		_, err := tableClient.SetSelection(ctx, &pb.TableSetSelectionRequest{SessionId: nil, WidgetId: tableID, Row: 0, Column: 0})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetSelection(nil session)")
	})

	t.Run("SetFixed_nil_session_id", func(t *testing.T) {
		_, err := tableClient.SetFixed(ctx, &pb.TableSetFixedRequest{SessionId: nil, WidgetId: tableID, FixedRows: 1, FixedColumns: 0})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetFixed(nil session)")
	})

	t.Run("SetCell_non_existent_widget", func(t *testing.T) {
		_, err := tableClient.SetCell(ctx, &pb.SetTableCellRequest{
			SessionId: sid, WidgetId: fakeWidgetID(), Row: 0, Column: 0,
			Cell: &pb.TableCell{Text: "X"},
		})
		assertGRPCCode(t, err, codes.NotFound, "SetCell(fake widget)")
	})

	t.Run("GetCell_non_existent_widget", func(t *testing.T) {
		_, err := tableClient.GetCell(ctx, &pb.GetTableCellRequest{SessionId: sid, WidgetId: fakeWidgetID(), Row: 0, Column: 0})
		assertGRPCCode(t, err, codes.NotFound, "GetCell(fake widget)")
	})
}

// --------------------------------------------------------------------
// 5. FormService
// --------------------------------------------------------------------

func TestContract_FormService(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	formClient := pb.NewFormServiceClient(conn)

	sid := requireSession(t, ctx, sessionClient, "form-contract")
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid})

	formID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_FORM, "Contract Form")

	// --- AddField ---
	t.Run("AddField_input", func(t *testing.T) {
		resp, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
			SessionId:    sid,
			WidgetId:     formID,
			Label:        "Username",
			FieldType:    pb.FormFieldType_FORM_FIELD_INPUT,
			FieldWidth:   testutil.Int32Ptr(20),
			InitialValue: testutil.StrPtr("admin"),
		})
		if err != nil {
			t.Fatalf("AddField(input) failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
		if resp.FieldIndex != 0 {
			t.Errorf("Expected field_index 0, got %d", resp.FieldIndex)
		}
	})

	t.Run("AddField_password", func(t *testing.T) {
		resp, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
			SessionId:  sid,
			WidgetId:   formID,
			Label:      "Password",
			FieldType:  pb.FormFieldType_FORM_FIELD_PASSWORD,
			FieldWidth: testutil.Int32Ptr(20),
		})
		if err != nil {
			t.Fatalf("AddField(password) failed: %v", err)
		}
		if resp.FieldIndex != 1 {
			t.Errorf("Expected field_index 1, got %d", resp.FieldIndex)
		}
	})

	t.Run("AddField_checkbox", func(t *testing.T) {
		resp, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
			SessionId: sid,
			WidgetId:  formID,
			Label:     "Remember",
			FieldType: pb.FormFieldType_FORM_FIELD_CHECKBOX,
		})
		if err != nil {
			t.Fatalf("AddField(checkbox) failed: %v", err)
		}
		if resp.FieldIndex != 2 {
			t.Errorf("Expected field_index 2, got %d", resp.FieldIndex)
		}
	})

	t.Run("AddField_dropdown", func(t *testing.T) {
		resp, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
			SessionId:       sid,
			WidgetId:        formID,
			Label:           "Role",
			FieldType:       pb.FormFieldType_FORM_FIELD_DROPDOWN,
			DropdownOptions: []string{"Admin", "User", "Guest"},
			InitialValue:    testutil.StrPtr("User"),
		})
		if err != nil {
			t.Fatalf("AddField(dropdown) failed: %v", err)
		}
		if resp.FieldIndex != 3 {
			t.Errorf("Expected field_index 3, got %d", resp.FieldIndex)
		}
	})

	// --- GetItemCount ---
	t.Run("GetItemCount", func(t *testing.T) {
		resp, err := formClient.GetItemCount(ctx, &pb.FormGetItemCountRequest{SessionId: sid, WidgetId: formID})
		if err != nil {
			t.Fatalf("GetItemCount failed: %v", err)
		}
		if resp.Count != 4 {
			t.Errorf("Expected count 4, got %d", resp.Count)
		}
	})

	// --- GetFieldValue / SetFieldValue ---
	t.Run("GetFieldValue_initial", func(t *testing.T) {
		resp, err := formClient.GetFieldValue(ctx, &pb.FormGetFieldValueRequest{
			SessionId: sid, WidgetId: formID, FieldIndex: 0,
		})
		if err != nil {
			t.Fatalf("GetFieldValue failed: %v", err)
		}
		if resp.Value != "admin" {
			t.Errorf("Expected 'admin', got '%s'", resp.Value)
		}
	})

	t.Run("SetFieldValue_GetFieldValue_roundtrip", func(t *testing.T) {
		setResp, err := formClient.SetFieldValue(ctx, &pb.FormSetFieldValueRequest{
			SessionId: sid, WidgetId: formID, FieldIndex: 0, Value: "newuser",
		})
		if err != nil {
			t.Fatalf("SetFieldValue failed: %v", err)
		}
		if !setResp.Success {
			t.Error("Expected success=true")
		}

		getResp, err := formClient.GetFieldValue(ctx, &pb.FormGetFieldValueRequest{
			SessionId: sid, WidgetId: formID, FieldIndex: 0,
		})
		if err != nil {
			t.Fatalf("GetFieldValue after set failed: %v", err)
		}
		if getResp.Value != "newuser" {
			t.Errorf("Expected 'newuser', got '%s'", getResp.Value)
		}
	})

	// --- AddButton ---
	t.Run("AddButton_valid", func(t *testing.T) {
		resp, err := formClient.AddButton(ctx, &pb.FormAddButtonRequest{
			SessionId: sid, WidgetId: formID, Label: "Submit",
		})
		if err != nil {
			t.Fatalf("AddButton failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	// --- Clear ---
	t.Run("Clear_valid", func(t *testing.T) {
		resp, err := formClient.Clear(ctx, &pb.FormClearRequest{SessionId: sid, WidgetId: formID})
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}

		countResp, _ := formClient.GetItemCount(ctx, &pb.FormGetItemCountRequest{SessionId: sid, WidgetId: formID})
		if countResp.Count != 0 {
			t.Errorf("Expected count 0 after clear, got %d", countResp.Count)
		}
	})

	// --- Error cases ---
	t.Run("AddField_nil_session_id", func(t *testing.T) {
		_, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
			SessionId: nil, WidgetId: formID, Label: "X", FieldType: pb.FormFieldType_FORM_FIELD_INPUT,
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "AddField(nil session)")
	})

	t.Run("AddButton_nil_session_id", func(t *testing.T) {
		_, err := formClient.AddButton(ctx, &pb.FormAddButtonRequest{SessionId: nil, WidgetId: formID, Label: "X"})
		assertGRPCCode(t, err, codes.InvalidArgument, "AddButton(nil session)")
	})

	t.Run("GetFieldValue_nil_session_id", func(t *testing.T) {
		_, err := formClient.GetFieldValue(ctx, &pb.FormGetFieldValueRequest{SessionId: nil, WidgetId: formID, FieldIndex: 0})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetFieldValue(nil session)")
	})

	t.Run("SetFieldValue_nil_session_id", func(t *testing.T) {
		_, err := formClient.SetFieldValue(ctx, &pb.FormSetFieldValueRequest{SessionId: nil, WidgetId: formID, FieldIndex: 0, Value: "x"})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetFieldValue(nil session)")
	})

	t.Run("Clear_nil_session_id", func(t *testing.T) {
		_, err := formClient.Clear(ctx, &pb.FormClearRequest{SessionId: nil, WidgetId: formID})
		assertGRPCCode(t, err, codes.InvalidArgument, "Clear(nil session)")
	})

	t.Run("GetItemCount_nil_session_id", func(t *testing.T) {
		_, err := formClient.GetItemCount(ctx, &pb.FormGetItemCountRequest{SessionId: nil, WidgetId: formID})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetItemCount(nil session)")
	})

	t.Run("AddField_non_existent_widget", func(t *testing.T) {
		_, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
			SessionId: sid, WidgetId: fakeWidgetID(), Label: "X", FieldType: pb.FormFieldType_FORM_FIELD_INPUT,
		})
		assertGRPCCode(t, err, codes.NotFound, "AddField(fake widget)")
	})

	t.Run("AddButton_non_existent_widget", func(t *testing.T) {
		_, err := formClient.AddButton(ctx, &pb.FormAddButtonRequest{SessionId: sid, WidgetId: fakeWidgetID(), Label: "X"})
		assertGRPCCode(t, err, codes.NotFound, "AddButton(fake widget)")
	})
}

// --------------------------------------------------------------------
// 6. TreeService
// --------------------------------------------------------------------

func TestContract_TreeService(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	treeClient := pb.NewTreeServiceClient(conn)

	sid := requireSession(t, ctx, sessionClient, "tree-contract")
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid})

	treeID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_TREE_VIEW, "Contract Tree")

	// --- SetRoot ---
	var rootNodeID *pb.TreeNodeId
	t.Run("SetRoot_valid", func(t *testing.T) {
		resp, err := treeClient.SetRoot(ctx, &pb.SetTreeRootRequest{
			SessionId: sid,
			WidgetId:  treeID,
			Node: &pb.TreeNode{
				Text:       "Root",
				Selectable: testutil.BoolPtr(true),
				Expanded:   testutil.BoolPtr(true),
				Reference:  testutil.StrPtr("root-ref"),
			},
		})
		if err != nil {
			t.Fatalf("SetRoot failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
		if resp.NodeId == nil || resp.NodeId.Id == "" {
			t.Fatal("Expected non-empty node_id")
		}
		rootNodeID = resp.NodeId
	})

	// --- AddChild ---
	var child1ID, child2ID *pb.TreeNodeId
	t.Run("AddChild_level1", func(t *testing.T) {
		resp, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
			SessionId: sid,
			WidgetId:  treeID,
			ParentId:  rootNodeID,
			Node: &pb.TreeNode{
				Text:       "Child 1",
				Selectable: testutil.BoolPtr(true),
				Reference:  testutil.StrPtr("c1-ref"),
			},
		})
		if err != nil {
			t.Fatalf("AddChild(1) failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
		child1ID = resp.NodeId
	})

	t.Run("AddChild_level1_second", func(t *testing.T) {
		resp, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
			SessionId: sid,
			WidgetId:  treeID,
			ParentId:  rootNodeID,
			Node: &pb.TreeNode{
				Text:      "Child 2",
				Expanded:  testutil.BoolPtr(true),
				Reference: testutil.StrPtr("c2-ref"),
			},
		})
		if err != nil {
			t.Fatalf("AddChild(2) failed: %v", err)
		}
		child2ID = resp.NodeId
	})

	// Add grandchild
	t.Run("AddChild_grandchild", func(t *testing.T) {
		resp, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
			SessionId: sid,
			WidgetId:  treeID,
			ParentId:  child2ID,
			Node: &pb.TreeNode{
				Text:      "Grandchild 1",
				Reference: testutil.StrPtr("gc1-ref"),
			},
		})
		if err != nil {
			t.Fatalf("AddChild(grandchild) failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	// --- GetChildren ---
	t.Run("GetChildren_root", func(t *testing.T) {
		resp, err := treeClient.GetChildren(ctx, &pb.TreeGetChildrenRequest{
			SessionId: sid, WidgetId: treeID, NodeId: rootNodeID,
		})
		if err != nil {
			t.Fatalf("GetChildren failed: %v", err)
		}
		if len(resp.Children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(resp.Children))
		}
	})

	t.Run("GetChildren_child2_has_grandchild", func(t *testing.T) {
		resp, err := treeClient.GetChildren(ctx, &pb.TreeGetChildrenRequest{
			SessionId: sid, WidgetId: treeID, NodeId: child2ID,
		})
		if err != nil {
			t.Fatalf("GetChildren(child2) failed: %v", err)
		}
		if len(resp.Children) != 1 {
			t.Errorf("Expected 1 grandchild, got %d", len(resp.Children))
		}
	})

	// --- SetExpanded ---
	t.Run("SetExpanded_collapse", func(t *testing.T) {
		resp, err := treeClient.SetExpanded(ctx, &pb.TreeSetExpandedRequest{
			SessionId: sid, WidgetId: treeID, NodeId: child2ID, Expanded: false,
		})
		if err != nil {
			t.Fatalf("SetExpanded failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}
	})

	// --- SetSelection / GetSelection ---
	t.Run("SetSelection_GetSelection_roundtrip", func(t *testing.T) {
		setResp, err := treeClient.SetSelection(ctx, &pb.TreeSetSelectionRequest{
			SessionId: sid, WidgetId: treeID, NodeId: child1ID,
		})
		if err != nil {
			t.Fatalf("SetSelection failed: %v", err)
		}
		if !setResp.Success {
			t.Error("Expected success=true")
		}

		getResp, err := treeClient.GetSelection(ctx, &pb.TreeGetSelectionRequest{
			SessionId: sid, WidgetId: treeID,
		})
		if err != nil {
			t.Fatalf("GetSelection failed: %v", err)
		}
		if getResp.NodeId == nil || getResp.NodeId.Id != child1ID.Id {
			t.Errorf("Expected selected node %s, got %v", child1ID.Id, getResp.NodeId)
		}
		if getResp.Text != "Child 1" {
			t.Errorf("Expected text 'Child 1', got '%s'", getResp.Text)
		}
	})

	// --- RemoveNode ---
	t.Run("RemoveNode_valid", func(t *testing.T) {
		resp, err := treeClient.RemoveNode(ctx, &pb.RemoveTreeNodeRequest{
			SessionId: sid, WidgetId: treeID, NodeId: child2ID,
		})
		if err != nil {
			t.Fatalf("RemoveNode failed: %v", err)
		}
		if !resp.Success {
			t.Error("Expected success=true")
		}

		// Verify only one child remains
		childrenResp, _ := treeClient.GetChildren(ctx, &pb.TreeGetChildrenRequest{
			SessionId: sid, WidgetId: treeID, NodeId: rootNodeID,
		})
		if len(childrenResp.Children) != 1 {
			t.Errorf("Expected 1 child after removal, got %d", len(childrenResp.Children))
		}
	})

	// --- Error cases ---
	t.Run("SetRoot_nil_session_id", func(t *testing.T) {
		_, err := treeClient.SetRoot(ctx, &pb.SetTreeRootRequest{
			SessionId: nil, WidgetId: treeID,
			Node: &pb.TreeNode{Text: "X"},
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetRoot(nil session)")
	})

	t.Run("AddChild_nil_session_id", func(t *testing.T) {
		_, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
			SessionId: nil, WidgetId: treeID, ParentId: rootNodeID,
			Node: &pb.TreeNode{Text: "X"},
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "AddChild(nil session)")
	})

	t.Run("RemoveNode_nil_session_id", func(t *testing.T) {
		_, err := treeClient.RemoveNode(ctx, &pb.RemoveTreeNodeRequest{
			SessionId: nil, WidgetId: treeID, NodeId: rootNodeID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "RemoveNode(nil session)")
	})

	t.Run("SetExpanded_nil_session_id", func(t *testing.T) {
		_, err := treeClient.SetExpanded(ctx, &pb.TreeSetExpandedRequest{
			SessionId: nil, WidgetId: treeID, NodeId: rootNodeID, Expanded: true,
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetExpanded(nil session)")
	})

	t.Run("GetSelection_nil_session_id", func(t *testing.T) {
		_, err := treeClient.GetSelection(ctx, &pb.TreeGetSelectionRequest{
			SessionId: nil, WidgetId: treeID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetSelection(nil session)")
	})

	t.Run("SetSelection_nil_session_id", func(t *testing.T) {
		_, err := treeClient.SetSelection(ctx, &pb.TreeSetSelectionRequest{
			SessionId: nil, WidgetId: treeID, NodeId: rootNodeID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "SetSelection(nil session)")
	})

	t.Run("GetChildren_nil_session_id", func(t *testing.T) {
		_, err := treeClient.GetChildren(ctx, &pb.TreeGetChildrenRequest{
			SessionId: nil, WidgetId: treeID, NodeId: rootNodeID,
		})
		assertGRPCCode(t, err, codes.InvalidArgument, "GetChildren(nil session)")
	})

	t.Run("SetRoot_non_existent_widget", func(t *testing.T) {
		_, err := treeClient.SetRoot(ctx, &pb.SetTreeRootRequest{
			SessionId: sid, WidgetId: fakeWidgetID(),
			Node: &pb.TreeNode{Text: "X"},
		})
		assertGRPCCode(t, err, codes.NotFound, "SetRoot(fake widget)")
	})

	t.Run("AddChild_non_existent_widget", func(t *testing.T) {
		_, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
			SessionId: sid, WidgetId: fakeWidgetID(), ParentId: rootNodeID,
			Node: &pb.TreeNode{Text: "X"},
		})
		assertGRPCCode(t, err, codes.NotFound, "AddChild(fake widget)")
	})
}

// --------------------------------------------------------------------
// 7. LayoutService
// --------------------------------------------------------------------

func TestContract_LayoutService(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	layoutClient := pb.NewLayoutServiceClient(conn)

	sid := requireSession(t, ctx, sessionClient, "layout-contract")
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sid})

	// ====================================
	// Flex
	// ====================================

	t.Run("Flex", func(t *testing.T) {
		flexID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_FLEX, "Contract Flex")
		box1ID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Flex Child 1")
		box2ID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Flex Child 2")

		// --- FlexAddItem ---
		t.Run("FlexAddItem_proportional", func(t *testing.T) {
			resp, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
				SessionId: sid, FlexId: flexID, ItemId: box1ID,
				Proportion: 1, Focus: false,
			})
			if err != nil {
				t.Fatalf("FlexAddItem failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		t.Run("FlexAddItem_fixed_size", func(t *testing.T) {
			resp, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
				SessionId: sid, FlexId: flexID, ItemId: box2ID,
				Proportion: 0, FixedSize: 10, Focus: false,
			})
			if err != nil {
				t.Fatalf("FlexAddItem(fixed) failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- FlexSetDirection ---
		t.Run("FlexSetDirection_column", func(t *testing.T) {
			resp, err := layoutClient.FlexSetDirection(ctx, &pb.FlexSetDirectionRequest{
				SessionId: sid, FlexId: flexID,
				Direction: pb.FlexDirection_FLEX_COLUMN,
			})
			if err != nil {
				t.Fatalf("FlexSetDirection failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		t.Run("FlexSetDirection_row", func(t *testing.T) {
			resp, err := layoutClient.FlexSetDirection(ctx, &pb.FlexSetDirectionRequest{
				SessionId: sid, FlexId: flexID,
				Direction: pb.FlexDirection_FLEX_ROW,
			})
			if err != nil {
				t.Fatalf("FlexSetDirection(row) failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- FlexRemoveItem ---
		t.Run("FlexRemoveItem_valid", func(t *testing.T) {
			resp, err := layoutClient.FlexRemoveItem(ctx, &pb.FlexRemoveItemRequest{
				SessionId: sid, FlexId: flexID, ItemId: box2ID,
			})
			if err != nil {
				t.Fatalf("FlexRemoveItem failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- Flex error cases ---
		t.Run("FlexAddItem_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
				SessionId: nil, FlexId: flexID, ItemId: box1ID, Proportion: 1,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "FlexAddItem(nil session)")
		})

		t.Run("FlexRemoveItem_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.FlexRemoveItem(ctx, &pb.FlexRemoveItemRequest{
				SessionId: nil, FlexId: flexID, ItemId: box1ID,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "FlexRemoveItem(nil session)")
		})

		t.Run("FlexSetDirection_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.FlexSetDirection(ctx, &pb.FlexSetDirectionRequest{
				SessionId: nil, FlexId: flexID, Direction: pb.FlexDirection_FLEX_ROW,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "FlexSetDirection(nil session)")
		})

		t.Run("FlexAddItem_non_existent_flex", func(t *testing.T) {
			_, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
				SessionId: sid, FlexId: fakeWidgetID(), ItemId: box1ID, Proportion: 1,
			})
			assertGRPCCode(t, err, codes.NotFound, "FlexAddItem(fake flex)")
		})

		t.Run("FlexAddItem_non_existent_item", func(t *testing.T) {
			_, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
				SessionId: sid, FlexId: flexID, ItemId: fakeWidgetID(), Proportion: 1,
			})
			assertGRPCCode(t, err, codes.NotFound, "FlexAddItem(fake item)")
		})
	})

	// ====================================
	// Grid
	// ====================================

	t.Run("Grid", func(t *testing.T) {
		gridID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_GRID, "Contract Grid")
		cell1ID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Grid Cell 1")
		cell2ID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Grid Cell 2")

		// --- GridSetRows ---
		t.Run("GridSetRows_valid", func(t *testing.T) {
			resp, err := layoutClient.GridSetRows(ctx, &pb.GridSetRowsRequest{
				SessionId: sid, GridId: gridID,
				RowSizes: []int32{-1, -1},
			})
			if err != nil {
				t.Fatalf("GridSetRows failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- GridSetColumns ---
		t.Run("GridSetColumns_valid", func(t *testing.T) {
			resp, err := layoutClient.GridSetColumns(ctx, &pb.GridSetColumnsRequest{
				SessionId: sid, GridId: gridID,
				ColumnSizes: []int32{20, -1},
			})
			if err != nil {
				t.Fatalf("GridSetColumns failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- GridAddItem ---
		t.Run("GridAddItem_valid", func(t *testing.T) {
			resp, err := layoutClient.GridAddItem(ctx, &pb.GridAddItemRequest{
				SessionId: sid, GridId: gridID, ItemId: cell1ID,
				Row: 0, Column: 0, RowSpan: 1, ColumnSpan: 1,
				MinWidth: 10, MinHeight: 5, Focus: false,
			})
			if err != nil {
				t.Fatalf("GridAddItem(1) failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		t.Run("GridAddItem_second", func(t *testing.T) {
			resp, err := layoutClient.GridAddItem(ctx, &pb.GridAddItemRequest{
				SessionId: sid, GridId: gridID, ItemId: cell2ID,
				Row: 0, Column: 1, RowSpan: 1, ColumnSpan: 1,
				MinWidth: 10, MinHeight: 5, Focus: false,
			})
			if err != nil {
				t.Fatalf("GridAddItem(2) failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- GridRemoveItem ---
		t.Run("GridRemoveItem_valid", func(t *testing.T) {
			resp, err := layoutClient.GridRemoveItem(ctx, &pb.GridRemoveItemRequest{
				SessionId: sid, GridId: gridID, ItemId: cell2ID,
			})
			if err != nil {
				t.Fatalf("GridRemoveItem failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- Grid error cases ---
		t.Run("GridAddItem_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.GridAddItem(ctx, &pb.GridAddItemRequest{
				SessionId: nil, GridId: gridID, ItemId: cell1ID,
				Row: 0, Column: 0, RowSpan: 1, ColumnSpan: 1,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "GridAddItem(nil session)")
		})

		t.Run("GridRemoveItem_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.GridRemoveItem(ctx, &pb.GridRemoveItemRequest{
				SessionId: nil, GridId: gridID, ItemId: cell1ID,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "GridRemoveItem(nil session)")
		})

		t.Run("GridSetRows_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.GridSetRows(ctx, &pb.GridSetRowsRequest{
				SessionId: nil, GridId: gridID, RowSizes: []int32{-1},
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "GridSetRows(nil session)")
		})

		t.Run("GridSetColumns_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.GridSetColumns(ctx, &pb.GridSetColumnsRequest{
				SessionId: nil, GridId: gridID, ColumnSizes: []int32{-1},
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "GridSetColumns(nil session)")
		})

		t.Run("GridAddItem_non_existent_grid", func(t *testing.T) {
			_, err := layoutClient.GridAddItem(ctx, &pb.GridAddItemRequest{
				SessionId: sid, GridId: fakeWidgetID(), ItemId: cell1ID,
				Row: 0, Column: 0, RowSpan: 1, ColumnSpan: 1,
			})
			assertGRPCCode(t, err, codes.NotFound, "GridAddItem(fake grid)")
		})

		t.Run("GridAddItem_non_existent_item", func(t *testing.T) {
			_, err := layoutClient.GridAddItem(ctx, &pb.GridAddItemRequest{
				SessionId: sid, GridId: gridID, ItemId: fakeWidgetID(),
				Row: 0, Column: 0, RowSpan: 1, ColumnSpan: 1,
			})
			assertGRPCCode(t, err, codes.NotFound, "GridAddItem(fake item)")
		})
	})

	// ====================================
	// Pages
	// ====================================

	t.Run("Pages", func(t *testing.T) {
		pagesID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_PAGES, "Contract Pages")
		page1ID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Page 1")
		page2ID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Page 2")
		page3ID := requireWidget(t, ctx, widgetClient, sid, pb.WidgetType_WIDGET_BOX, "Page 3")

		// --- PagesAddPage ---
		t.Run("PagesAddPage_home", func(t *testing.T) {
			resp, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "home",
				ItemId: page1ID, Resize: true, Visible: true,
			})
			if err != nil {
				t.Fatalf("PagesAddPage(home) failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		t.Run("PagesAddPage_settings", func(t *testing.T) {
			resp, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "settings",
				ItemId: page2ID, Resize: true, Visible: true,
			})
			if err != nil {
				t.Fatalf("PagesAddPage(settings) failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		t.Run("PagesAddPage_about", func(t *testing.T) {
			resp, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "about",
				ItemId: page3ID, Resize: true, Visible: true,
			})
			if err != nil {
				t.Fatalf("PagesAddPage(about) failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- PagesShowPage ---
		t.Run("PagesShowPage_valid", func(t *testing.T) {
			resp, err := layoutClient.PagesShowPage(ctx, &pb.PagesShowPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "settings",
			})
			if err != nil {
				t.Fatalf("PagesShowPage failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- PagesGetCurrentPage ---
		t.Run("PagesGetCurrentPage_valid", func(t *testing.T) {
			resp, err := layoutClient.PagesGetCurrentPage(ctx, &pb.PagesGetCurrentPageRequest{
				SessionId: sid, PagesId: pagesID,
			})
			if err != nil {
				t.Fatalf("PagesGetCurrentPage failed: %v", err)
			}
			if resp.PageName != "settings" {
				t.Errorf("Expected page 'settings', got '%s'", resp.PageName)
			}
		})

		// Switch and verify
		t.Run("PagesShowPage_switch_to_about", func(t *testing.T) {
			layoutClient.PagesShowPage(ctx, &pb.PagesShowPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "about",
			})

			resp, err := layoutClient.PagesGetCurrentPage(ctx, &pb.PagesGetCurrentPageRequest{
				SessionId: sid, PagesId: pagesID,
			})
			if err != nil {
				t.Fatalf("PagesGetCurrentPage after switch failed: %v", err)
			}
			if resp.PageName != "about" {
				t.Errorf("Expected page 'about', got '%s'", resp.PageName)
			}
		})

		// --- PagesRemovePage ---
		t.Run("PagesRemovePage_valid", func(t *testing.T) {
			resp, err := layoutClient.PagesRemovePage(ctx, &pb.PagesRemovePageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "settings",
			})
			if err != nil {
				t.Fatalf("PagesRemovePage failed: %v", err)
			}
			if !resp.Success {
				t.Error("Expected success=true")
			}
		})

		// --- Pages error cases ---
		t.Run("PagesAddPage_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
				SessionId: nil, PagesId: pagesID, PageName: "x",
				ItemId: page1ID, Resize: true, Visible: true,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "PagesAddPage(nil session)")
		})

		t.Run("PagesRemovePage_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.PagesRemovePage(ctx, &pb.PagesRemovePageRequest{
				SessionId: nil, PagesId: pagesID, PageName: "home",
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "PagesRemovePage(nil session)")
		})

		t.Run("PagesShowPage_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.PagesShowPage(ctx, &pb.PagesShowPageRequest{
				SessionId: nil, PagesId: pagesID, PageName: "home",
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "PagesShowPage(nil session)")
		})

		t.Run("PagesGetCurrentPage_nil_session_id", func(t *testing.T) {
			_, err := layoutClient.PagesGetCurrentPage(ctx, &pb.PagesGetCurrentPageRequest{
				SessionId: nil, PagesId: pagesID,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "PagesGetCurrentPage(nil session)")
		})

		t.Run("PagesAddPage_non_existent_pages", func(t *testing.T) {
			_, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
				SessionId: sid, PagesId: fakeWidgetID(), PageName: "x",
				ItemId: page1ID, Resize: true, Visible: true,
			})
			assertGRPCCode(t, err, codes.NotFound, "PagesAddPage(fake pages)")
		})

		t.Run("PagesAddPage_non_existent_item", func(t *testing.T) {
			_, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "x",
				ItemId: fakeWidgetID(), Resize: true, Visible: true,
			})
			assertGRPCCode(t, err, codes.NotFound, "PagesAddPage(fake item)")
		})

		t.Run("PagesAddPage_empty_page_name", func(t *testing.T) {
			_, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "",
				ItemId: page1ID, Resize: true, Visible: true,
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "PagesAddPage(empty name)")
		})

		t.Run("PagesRemovePage_empty_page_name", func(t *testing.T) {
			_, err := layoutClient.PagesRemovePage(ctx, &pb.PagesRemovePageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "",
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "PagesRemovePage(empty name)")
		})

		t.Run("PagesShowPage_empty_page_name", func(t *testing.T) {
			_, err := layoutClient.PagesShowPage(ctx, &pb.PagesShowPageRequest{
				SessionId: sid, PagesId: pagesID, PageName: "",
			})
			assertGRPCCode(t, err, codes.InvalidArgument, "PagesShowPage(empty name)")
		})
	})
}
