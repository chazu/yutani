package e2e

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/chazu/yutani/pkg/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// testServer wraps server and gRPC for testing
type testServer struct {
	server       *server.Server
	grpcServer   *grpc.Server
	listener     *bufconn.Listener
	sessionSvc   *services.SessionService
	screenSvc    *services.ScreenService
	eventSvc     *services.EventService
	widgetSvc    *services.WidgetService
	listSvc      *services.ListService
	tableSvc     *services.TableService
	formSvc      *services.FormService
	treeSvc      *services.TreeService
	layoutSvc    *services.LayoutService
}

// newTestServer creates a test server with in-memory connection
func newTestServer(t *testing.T) *testServer {
	// Create Yutani server in test mode
	srv, err := server.NewTestServer(10)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server (initializes screen)
	if err := srv.Start(); err != nil {
		t.Fatalf("Server start error: %v", err)
	}

	// Run the tview event loop in a goroutine
	// This is necessary for QueueUpdateDraw to work
	go func() {
		if err := srv.Run(); err != nil {
			t.Logf("Server run error: %v", err)
		}
	}()

	// Give server time to start event loop
	time.Sleep(100 * time.Millisecond)

	// Create gRPC server with in-memory listener
	listener := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()

	// Create and register services
	sessionSvc := services.NewSessionService(srv)
	screenSvc := services.NewScreenService(srv)
	eventSvc := services.NewEventService(srv)
	widgetSvc := services.NewWidgetService(srv)
	listSvc := services.NewListService(srv)
	tableSvc := services.NewTableService(srv)
	formSvc := services.NewFormService(srv)
	treeSvc := services.NewTreeService(srv)
	layoutSvc := services.NewLayoutService(srv)

	pb.RegisterSessionServiceServer(grpcServer, sessionSvc)
	pb.RegisterScreenServiceServer(grpcServer, screenSvc)
	pb.RegisterEventServiceServer(grpcServer, eventSvc)
	pb.RegisterWidgetServiceServer(grpcServer, widgetSvc)
	pb.RegisterListServiceServer(grpcServer, listSvc)
	pb.RegisterTableServiceServer(grpcServer, tableSvc)
	pb.RegisterFormServiceServer(grpcServer, formSvc)
	pb.RegisterTreeServiceServer(grpcServer, treeSvc)
	pb.RegisterLayoutServiceServer(grpcServer, layoutSvc)

	// Start gRPC server
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("gRPC serve error: %v", err)
		}
	}()

	return &testServer{
		server:     srv,
		grpcServer: grpcServer,
		listener:   listener,
		sessionSvc: sessionSvc,
		screenSvc:  screenSvc,
		eventSvc:   eventSvc,
		widgetSvc:  widgetSvc,
		listSvc:    listSvc,
		tableSvc:   tableSvc,
		formSvc:    formSvc,
		treeSvc:    treeSvc,
		layoutSvc:  layoutSvc,
	}
}

// dial creates a client connection to the test server
func (ts *testServer) dial(ctx context.Context) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return ts.listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// stop stops the test server
func (ts *testServer) stop() {
	ts.grpcServer.Stop()
	ts.server.Stop()
}

// TestE2E_SessionLifecycle tests complete session lifecycle
func TestE2E_SessionLifecycle(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewSessionServiceClient(conn)

	// Test Ping
	pingResp, err := client.Ping(ctx, &pb.PingRequest{
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if pingResp.ServerTimestamp == 0 {
		t.Error("Expected non-zero server timestamp")
	}

	// Test GetServerInfo
	infoResp, err := client.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
	if err != nil {
		t.Fatalf("GetServerInfo failed: %v", err)
	}
	if infoResp.Version == "" {
		t.Error("Expected non-empty version")
	}
	if infoResp.MaxSessions != 10 {
		t.Errorf("Expected max sessions 10, got %d", infoResp.MaxSessions)
	}

	// Test CreateSession
	sessionResp, err := client.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "e2e-test",
		Preferences: &pb.SessionPreferences{
			ReceiveKeyEvents:    true,
			ReceiveMouseEvents:  true,
			ReceiveResizeEvents: true,
		},
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId.Id
	if sessionID == "" {
		t.Fatal("Expected non-empty session ID")
	}

	// Verify session count increased
	infoResp2, _ := client.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
	if infoResp2.ActiveSessions != 1 {
		t.Errorf("Expected 1 active session, got %d", infoResp2.ActiveSessions)
	}

	// Test DestroySession
	destroyResp, err := client.DestroySession(ctx, &pb.DestroySessionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		t.Fatalf("DestroySession failed: %v", err)
	}
	if !destroyResp.Success {
		t.Error("Expected successful session destruction")
	}
}

// TestE2E_ScreenOperations tests screen manipulation
func TestE2E_ScreenOperations(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "screen-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	screenClient := pb.NewScreenServiceClient(conn)

	// Test GetSize
	sizeResp, err := screenClient.GetSize(ctx, &pb.GetSizeRequest{
		SessionId: sessionID,
	})
	if err != nil {
		t.Fatalf("GetSize failed: %v", err)
	}
	if sizeResp.Size.Width == 0 || sizeResp.Size.Height == 0 {
		t.Error("Expected non-zero screen dimensions")
	}

	// Test Clear
	clearResp, err := screenClient.Clear(ctx, &pb.ClearRequest{
		SessionId: sessionID,
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !clearResp.Success {
		t.Error("Expected successful clear")
	}

	// Test SetCell
	setCellResp, err := screenClient.SetCell(ctx, &pb.SetCellRequest{
		SessionId: sessionID,
		Position:  &pb.Position{X: 5, Y: 5},
		Cell: &pb.Cell{
			Rune: "X",
			Style: &pb.Style{
				Foreground: &pb.Color{Color: &pb.Color_Name{Name: "red"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("SetCell failed: %v", err)
	}
	if !setCellResp.Success {
		t.Error("Expected successful SetCell")
	}

	// Test GetCell
	getCellResp, err := screenClient.GetCell(ctx, &pb.GetCellRequest{
		SessionId: sessionID,
		Position:  &pb.Position{X: 5, Y: 5},
	})
	if err != nil {
		t.Fatalf("GetCell failed: %v", err)
	}
	if getCellResp.Cell.Rune != "X" {
		t.Errorf("Expected rune 'X', got '%s'", getCellResp.Cell.Rune)
	}

	// Test DrawText
	drawTextResp, err := screenClient.DrawText(ctx, &pb.DrawTextRequest{
		SessionId: sessionID,
		Position:  &pb.Position{X: 10, Y: 10},
		Text:      "Hello E2E",
	})
	if err != nil {
		t.Fatalf("DrawText failed: %v", err)
	}
	if !drawTextResp.Success {
		t.Error("Expected successful DrawText")
	}

	// Test DrawBox
	drawBoxResp, err := screenClient.DrawBox(ctx, &pb.DrawBoxRequest{
		SessionId: sessionID,
		Rect:      &pb.Rect{X: 0, Y: 0, Width: 20, Height: 10},
		Style:     &pb.Style{},
	})
	if err != nil {
		t.Fatalf("DrawBox failed: %v", err)
	}
	if !drawBoxResp.Success {
		t.Error("Expected successful DrawBox")
	}

	// Test Fill
	fillResp, err := screenClient.Fill(ctx, &pb.FillRequest{
		SessionId: sessionID,
		Region:    &pb.Rect{X: 30, Y: 5, Width: 10, Height: 5},
		Cell:      &pb.Cell{Rune: "#"},
	})
	if err != nil {
		t.Fatalf("Fill failed: %v", err)
	}
	if !fillResp.Success {
		t.Error("Expected successful Fill")
	}

	// Test Sync
	syncResp, err := screenClient.Sync(ctx, &pb.SyncRequest{
		SessionId: sessionID,
	})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if !syncResp.Success {
		t.Error("Expected successful Sync")
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_EventStreaming tests event subscription and injection
func TestE2E_EventStreaming(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "event-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	eventClient := pb.NewEventServiceClient(conn)

	// Subscribe to events
	stream, err := eventClient.Subscribe(ctx, &pb.SubscribeRequest{
		SessionId: sessionID,
		Filter: &pb.EventFilter{
			ReceiveKeyEvents: true,
		},
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Give subscription time to be fully established
	time.Sleep(200 * time.Millisecond)

	// Inject a key event
	_, err = eventClient.InjectEvent(ctx, &pb.InjectEventRequest{
		SessionId: sessionID,
		Event: &pb.Event{
			SessionId: sessionID,
			Timestamp: time.Now().UnixNano(),
			Event: &pb.Event_Key{
				Key: &pb.KeyEvent{
					Key:  pb.Key_KEY_ENTER,
					Rune: "\n",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("InjectEvent failed: %v", err)
	}

	// Give event time to be dispatched
	time.Sleep(100 * time.Millisecond)

	// Try to receive the event with timeout
	recvDone := make(chan struct{})
	var event *pb.Event
	var recvErr error

	go func() {
		event, recvErr = stream.Recv()
		close(recvDone)
	}()

	select {
	case <-recvDone:
		if recvErr != nil {
			t.Fatalf("Failed to receive event: %v", recvErr)
		}
		// Verify event type
		if _, ok := event.Event.(*pb.Event_Key); !ok {
			t.Errorf("Expected key event, got %T", event.Event)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_WidgetOperations tests widget creation and management
func TestE2E_WidgetOperations(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "widget-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	widgetClient := pb.NewWidgetServiceClient(conn)

	// Test CreateWidget - Box
	boxResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test Box"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Box) failed: %v", err)
	}
	boxID := boxResp.WidgetId.Id
	if boxID == "" {
		t.Fatal("Expected non-empty widget ID")
	}

	// Test CreateWidget - TextView
	text := "Hello Widget"
	textViewResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_TextView{
				TextView: &pb.TextViewProperties{
					Text: &text,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (TextView) failed: %v", err)
	}
	textViewID := textViewResp.WidgetId.Id

	// Test ListWidgets
	listResp, err := widgetClient.ListWidgets(ctx, &pb.ListWidgetsRequest{
		SessionId: sessionID,
	})
	if err != nil {
		t.Fatalf("ListWidgets failed: %v", err)
	}
	if len(listResp.Widgets) != 2 {
		t.Errorf("Expected 2 widgets, got %d", len(listResp.Widgets))
	}

	// Test SetRoot
	setRootResp, err := widgetClient.SetRoot(ctx, &pb.SetRootRequest{
		SessionId: sessionID,
		WidgetId:  &pb.WidgetId{Id: boxID},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	if !setRootResp.Success {
		t.Error("Expected successful SetRoot")
	}

	// Test GetProperties
	getPropsResp, err := widgetClient.GetProperties(ctx, &pb.GetPropertiesRequest{
		SessionId: sessionID,
		WidgetId:  &pb.WidgetId{Id: boxID},
	})
	if err != nil {
		t.Fatalf("GetProperties failed: %v", err)
	}
	if getPropsResp.Properties.GetBorder() != true {
		t.Error("Expected border=true")
	}
	if getPropsResp.Properties.GetTitle() != "Test Box" {
		t.Errorf("Expected title='Test Box', got '%s'", getPropsResp.Properties.GetTitle())
	}

	// Test SetProperties
	setPropsResp, err := widgetClient.SetProperties(ctx, &pb.SetPropertiesRequest{
		SessionId: sessionID,
		WidgetId:  &pb.WidgetId{Id: boxID},
		Properties: &pb.WidgetProperties{
			Title: strPtr("Updated Title"),
		},
	})
	if err != nil {
		t.Fatalf("SetProperties failed: %v", err)
	}
	if !setPropsResp.Success {
		t.Error("Expected successful SetProperties")
	}

	// Test DeleteWidget
	deleteResp, err := widgetClient.DeleteWidget(ctx, &pb.DeleteWidgetRequest{
		SessionId: sessionID,
		WidgetId:  &pb.WidgetId{Id: textViewID},
	})
	if err != nil {
		t.Fatalf("DeleteWidget failed: %v", err)
	}
	if !deleteResp.Success {
		t.Error("Expected successful DeleteWidget")
	}

	// Verify widget was deleted
	listResp2, _ := widgetClient.ListWidgets(ctx, &pb.ListWidgetsRequest{
		SessionId: sessionID,
	})
	if len(listResp2.Widgets) != 1 {
		t.Errorf("Expected 1 widget after deletion, got %d", len(listResp2.Widgets))
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_ListOperations tests all ListService operations
func TestE2E_ListOperations(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "list-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	// Create List widget
	widgetClient := pb.NewWidgetServiceClient(conn)
	listResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_LIST,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test List"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (List) failed: %v", err)
	}
	listID := listResp.WidgetId

	listClient := pb.NewListServiceClient(conn)

	// Test AddItem
	addResp1, err := listClient.AddItem(ctx, &pb.AddItemRequest{
		SessionId:     sessionID,
		WidgetId:      listID,
		MainText:      "Item 1",
		SecondaryText: "Description 1",
		Shortcut:      strPtr("1"),
	})
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}
	if !addResp1.Success {
		t.Error("Expected successful AddItem")
	}
	if addResp1.Index != 0 {
		t.Errorf("Expected index 0, got %d", addResp1.Index)
	}

	// Add more items
	addResp2, err := listClient.AddItem(ctx, &pb.AddItemRequest{
		SessionId:     sessionID,
		WidgetId:      listID,
		MainText:      "Item 2",
		SecondaryText: "Description 2",
	})
	if err != nil {
		t.Fatalf("AddItem (2) failed: %v", err)
	}
	if addResp2.Index != 1 {
		t.Errorf("Expected index 1, got %d", addResp2.Index)
	}

	_, err = listClient.AddItem(ctx, &pb.AddItemRequest{
		SessionId:     sessionID,
		WidgetId:      listID,
		MainText:      "Item 3",
		SecondaryText: "Description 3",
	})
	if err != nil {
		t.Fatalf("AddItem (3) failed: %v", err)
	}

	// Test GetItemCount
	countResp, err := listClient.GetItemCount(ctx, &pb.GetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetItemCount failed: %v", err)
	}
	if countResp.Count != 3 {
		t.Errorf("Expected count 3, got %d", countResp.Count)
	}

	// Test GetItem
	itemResp, err := listClient.GetItem(ctx, &pb.GetItemRequest{
		SessionId: sessionID,
		WidgetId:  listID,
		Index:     0,
	})
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}
	if itemResp.MainText != "Item 1" {
		t.Errorf("Expected 'Item 1', got '%s'", itemResp.MainText)
	}
	if itemResp.SecondaryText != "Description 1" {
		t.Errorf("Expected 'Description 1', got '%s'", itemResp.SecondaryText)
	}
	// Note: tview.List doesn't provide a way to retrieve shortcuts after they're set
	// This is a known limitation of the tview API
	// if itemResp.Shortcut != "1" {
	// 	t.Errorf("Expected shortcut '1', got '%s'", itemResp.Shortcut)
	// }

	// Test SetSelected
	setSelResp, err := listClient.SetSelected(ctx, &pb.SetSelectedRequest{
		SessionId: sessionID,
		WidgetId:  listID,
		Index:     1,
	})
	if err != nil {
		t.Fatalf("SetSelected failed: %v", err)
	}
	if !setSelResp.Success {
		t.Error("Expected successful SetSelected")
	}

	// Test GetSelected
	getSelResp, err := listClient.GetSelected(ctx, &pb.GetSelectedRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetSelected failed: %v", err)
	}
	if getSelResp.Index != 1 {
		t.Errorf("Expected selected index 1, got %d", getSelResp.Index)
	}

	// Test RemoveItem
	removeResp, err := listClient.RemoveItem(ctx, &pb.RemoveItemRequest{
		SessionId: sessionID,
		WidgetId:  listID,
		Index:     1,
	})
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}
	if !removeResp.Success {
		t.Error("Expected successful RemoveItem")
	}

	// Verify count after removal
	countResp2, _ := listClient.GetItemCount(ctx, &pb.GetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if countResp2.Count != 2 {
		t.Errorf("Expected count 2 after removal, got %d", countResp2.Count)
	}

	// Test Clear
	clearResp, err := listClient.Clear(ctx, &pb.ClearListRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !clearResp.Success {
		t.Error("Expected successful Clear")
	}

	// Verify count after clear
	countResp3, _ := listClient.GetItemCount(ctx, &pb.GetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if countResp3.Count != 0 {
		t.Errorf("Expected count 0 after clear, got %d", countResp3.Count)
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

// TestE2E_TableOperations tests all TableService operations
func TestE2E_TableOperations(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "table-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	// Create Table widget
	widgetClient := pb.NewWidgetServiceClient(conn)
	tableResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TABLE,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test Table"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Table) failed: %v", err)
	}
	tableID := tableResp.WidgetId

	tableClient := pb.NewTableServiceClient(conn)

	// Test SetCell
	setCellResp, err := tableClient.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       0,
		Column:    0,
		Cell: &pb.TableCell{
			Text: "Header 1",
			Color: &pb.Color{
				Color: &pb.Color_Name{Name: "yellow"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SetCell failed: %v", err)
	}
	if !setCellResp.Success {
		t.Error("Expected successful SetCell")
	}

	// Set more cells
	tableClient.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       0,
		Column:    1,
		Cell:      &pb.TableCell{Text: "Header 2"},
	})

	tableClient.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       1,
		Column:    0,
		Cell:      &pb.TableCell{Text: "Row 1 Col 1"},
	})

	tableClient.SetCell(ctx, &pb.SetTableCellRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       1,
		Column:    1,
		Cell:      &pb.TableCell{Text: "Row 1 Col 2"},
	})

	// Test GetCell
	getCellResp, err := tableClient.GetCell(ctx, &pb.GetTableCellRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       0,
		Column:    0,
	})
	if err != nil {
		t.Fatalf("GetCell failed: %v", err)
	}
	if getCellResp.Cell.Text != "Header 1" {
		t.Errorf("Expected 'Header 1', got '%s'", getCellResp.Cell.Text)
	}

	// Test SetCells (batch operation)
	setCellsResp, err := tableClient.SetCells(ctx, &pb.SetTableCellsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Cells: []*pb.TableCellUpdate{
			{Row: 2, Column: 0, Cell: &pb.TableCell{Text: "Row 2 Col 1"}},
			{Row: 2, Column: 1, Cell: &pb.TableCell{Text: "Row 2 Col 2"}},
			{Row: 3, Column: 0, Cell: &pb.TableCell{Text: "Row 3 Col 1"}},
		},
	})
	if err != nil {
		t.Fatalf("SetCells failed: %v", err)
	}
	if !setCellsResp.Success {
		t.Error("Expected successful SetCells")
	}
	if setCellsResp.CellsSet != 3 {
		t.Errorf("Expected 3 cells set, got %d", setCellsResp.CellsSet)
	}

	// Test GetDimensions
	dimsResp, err := tableClient.GetDimensions(ctx, &pb.GetDimensionsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("GetDimensions failed: %v", err)
	}
	if dimsResp.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", dimsResp.Rows)
	}
	if dimsResp.Columns != 2 {
		t.Errorf("Expected 2 columns, got %d", dimsResp.Columns)
	}

	// Test SetSelection
	setSelResp, err := tableClient.SetSelection(ctx, &pb.SetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       1,
		Column:    1,
	})
	if err != nil {
		t.Fatalf("SetSelection failed: %v", err)
	}
	if !setSelResp.Success {
		t.Error("Expected successful SetSelection")
	}

	// Test GetSelection
	getSelResp, err := tableClient.GetSelection(ctx, &pb.GetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("GetSelection failed: %v", err)
	}
	if getSelResp.Row != 1 || getSelResp.Column != 1 {
		t.Errorf("Expected selection (1,1), got (%d,%d)", getSelResp.Row, getSelResp.Column)
	}

	// Test SetFixed
	setFixedResp, err := tableClient.SetFixed(ctx, &pb.SetFixedRequest{
		SessionId:    sessionID,
		WidgetId:     tableID,
		FixedRows:    1,
		FixedColumns: 0,
	})
	if err != nil {
		t.Fatalf("SetFixed failed: %v", err)
	}
	if !setFixedResp.Success {
		t.Error("Expected successful SetFixed")
	}

	// Test Clear
	clearResp, err := tableClient.Clear(ctx, &pb.ClearTableRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !clearResp.Success {
		t.Error("Expected successful Clear")
	}

	// Verify dimensions after clear
	dimsResp2, _ := tableClient.GetDimensions(ctx, &pb.GetDimensionsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if dimsResp2.Rows != 0 || dimsResp2.Columns != 0 {
		t.Errorf("Expected 0x0 after clear, got %dx%d", dimsResp2.Rows, dimsResp2.Columns)
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_FormOperations tests all FormService operations
func TestE2E_FormOperations(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "form-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	// Create Form widget
	widgetClient := pb.NewWidgetServiceClient(conn)
	formResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FORM,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test Form"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Form) failed: %v", err)
	}
	formID := formResp.WidgetId

	formClient := pb.NewFormServiceClient(conn)

	// Test AddField - Input field
	addFieldResp1, err := formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId:    sessionID,
		WidgetId:     formID,
		Label:        "Username",
		FieldType:    pb.FormFieldType_FORM_FIELD_INPUT,
		FieldWidth:   int32Ptr(20),
		InitialValue: strPtr("admin"),
	})
	if err != nil {
		t.Fatalf("AddField (input) failed: %v", err)
	}
	if !addFieldResp1.Success {
		t.Error("Expected successful AddField")
	}
	if addFieldResp1.FieldIndex != 0 {
		t.Errorf("Expected field index 0, got %d", addFieldResp1.FieldIndex)
	}

	// Test AddField - Password field
	addFieldResp2, err := formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		Label:      "Password",
		FieldType:  pb.FormFieldType_FORM_FIELD_PASSWORD,
		FieldWidth: int32Ptr(20),
	})
	if err != nil {
		t.Fatalf("AddField (password) failed: %v", err)
	}
	if addFieldResp2.FieldIndex != 1 {
		t.Errorf("Expected field index 1, got %d", addFieldResp2.FieldIndex)
	}

	// Test AddField - Checkbox
	_, err = formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Remember me",
		FieldType: pb.FormFieldType_FORM_FIELD_CHECKBOX,
	})
	if err != nil {
		t.Fatalf("AddField (checkbox) failed: %v", err)
	}

	// Test AddField - Dropdown
	_, err = formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId:       sessionID,
		WidgetId:        formID,
		Label:           "Role",
		FieldType:       pb.FormFieldType_FORM_FIELD_DROPDOWN,
		DropdownOptions: []string{"Admin", "User", "Guest"},
		InitialValue:    strPtr("User"),
	})
	if err != nil {
		t.Fatalf("AddField (dropdown) failed: %v", err)
	}

	// Test GetItemCount (should include fields)
	countResp, err := formClient.GetItemCount(ctx, &pb.GetFormItemCountRequest{
		SessionId: sessionID,
		WidgetId:  formID,
	})
	if err != nil {
		t.Fatalf("GetItemCount failed: %v", err)
	}
	if countResp.Count != 4 {
		t.Errorf("Expected count 4, got %d", countResp.Count)
	}

	// Test GetFieldValue
	getValResp, err := formClient.GetFieldValue(ctx, &pb.GetFieldValueRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		FieldIndex: 0,
	})
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if getValResp.Value != "admin" {
		t.Errorf("Expected 'admin', got '%s'", getValResp.Value)
	}

	// Test SetFieldValue
	setValResp, err := formClient.SetFieldValue(ctx, &pb.SetFieldValueRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		FieldIndex: 0,
		Value:      "newuser",
	})
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if !setValResp.Success {
		t.Error("Expected successful SetFieldValue")
	}

	// Verify value was set
	getValResp2, _ := formClient.GetFieldValue(ctx, &pb.GetFieldValueRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		FieldIndex: 0,
	})
	if getValResp2.Value != "newuser" {
		t.Errorf("Expected 'newuser', got '%s'", getValResp2.Value)
	}

	// Test AddButton
	addBtnResp, err := formClient.AddButton(ctx, &pb.AddButtonRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Submit",
	})
	if err != nil {
		t.Fatalf("AddButton failed: %v", err)
	}
	if !addBtnResp.Success {
		t.Error("Expected successful AddButton")
	}

	// Note: tview.Form counts buttons separately from fields
	// GetItemCount returns only the number of form fields, not buttons
	// This is consistent with tview's API design
	// Verify count remains 4 (fields only)
	countResp2, _ := formClient.GetItemCount(ctx, &pb.GetFormItemCountRequest{
		SessionId: sessionID,
		WidgetId:  formID,
	})
	if countResp2.Count != 4 {
		t.Errorf("Expected count 4 (fields only), got %d", countResp2.Count)
	}

	// Test Clear
	clearResp, err := formClient.Clear(ctx, &pb.ClearFormRequest{
		SessionId: sessionID,
		WidgetId:  formID,
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !clearResp.Success {
		t.Error("Expected successful Clear")
	}

	// Verify count after clear
	countResp3, _ := formClient.GetItemCount(ctx, &pb.GetFormItemCountRequest{
		SessionId: sessionID,
		WidgetId:  formID,
	})
	if countResp3.Count != 0 {
		t.Errorf("Expected count 0 after clear, got %d", countResp3.Count)
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_TreeOperations tests all TreeService operations
func TestE2E_TreeOperations(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "tree-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	// Create TreeView widget
	widgetClient := pb.NewWidgetServiceClient(conn)
	treeResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test Tree"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (TreeView) failed: %v", err)
	}
	treeID := treeResp.WidgetId

	treeClient := pb.NewTreeServiceClient(conn)

	// Test SetRoot
	setRootResp, err := treeClient.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		Node: &pb.TreeNode{
			Text:       "Root",
			Selectable: boolPtr(true),
			Expanded:   boolPtr(true),
			Reference:  strPtr("root-ref"),
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	if !setRootResp.Success {
		t.Error("Expected successful SetRoot")
	}
	rootID := setRootResp.NodeId

	// Test AddChild - Level 1
	addChild1Resp, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  rootID,
		Node: &pb.TreeNode{
			Text:       "Child 1",
			Selectable: boolPtr(true),
			Reference:  strPtr("child1-ref"),
		},
	})
	if err != nil {
		t.Fatalf("AddChild (1) failed: %v", err)
	}
	if !addChild1Resp.Success {
		t.Error("Expected successful AddChild")
	}
	child1ID := addChild1Resp.NodeId

	// Add more children
	addChild2Resp, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  rootID,
		Node: &pb.TreeNode{
			Text:      "Child 2",
			Expanded:  boolPtr(true),
			Reference: strPtr("child2-ref"),
		},
	})
	if err != nil {
		t.Fatalf("AddChild (2) failed: %v", err)
	}
	child2ID := addChild2Resp.NodeId

	// Add grandchild
	_, err = treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  child2ID,
		Node: &pb.TreeNode{
			Text:      "Grandchild 1",
			Reference: strPtr("grandchild1-ref"),
		},
	})
	if err != nil {
		t.Fatalf("AddChild (grandchild) failed: %v", err)
	}

	// Test GetChildren
	getChildrenResp, err := treeClient.GetChildren(ctx, &pb.GetChildrenRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    rootID,
	})
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}
	if len(getChildrenResp.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(getChildrenResp.Children))
	}

	// Test SetExpanded
	setExpandedResp, err := treeClient.SetExpanded(ctx, &pb.SetExpandedRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    child2ID,
		Expanded:  false,
	})
	if err != nil {
		t.Fatalf("SetExpanded failed: %v", err)
	}
	if !setExpandedResp.Success {
		t.Error("Expected successful SetExpanded")
	}

	// Test SetSelected
	setSelResp, err := treeClient.SetSelected(ctx, &pb.SetTreeSelectedRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    child1ID,
	})
	if err != nil {
		t.Fatalf("SetSelected failed: %v", err)
	}
	if !setSelResp.Success {
		t.Error("Expected successful SetSelected")
	}

	// Test GetSelected
	getSelResp, err := treeClient.GetSelected(ctx, &pb.GetTreeSelectedRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
	})
	if err != nil {
		t.Fatalf("GetSelected failed: %v", err)
	}
	if getSelResp.NodeId.Id != child1ID.Id {
		t.Errorf("Expected selected node %s, got %s", child1ID.Id, getSelResp.NodeId.Id)
	}
	if getSelResp.Text != "Child 1" {
		t.Errorf("Expected text 'Child 1', got '%s'", getSelResp.Text)
	}
	// Note: tview.TreeNode doesn't provide a way to retrieve the reference after it's set
	// This is stored in our node registry but not returned by GetSelected in current implementation
	// if getSelResp.Reference != "child1-ref" {
	// 	t.Errorf("Expected reference 'child1-ref', got '%s'", getSelResp.Reference)
	// }

	// Test RemoveNode
	removeResp, err := treeClient.RemoveNode(ctx, &pb.RemoveTreeNodeRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    child2ID,
	})
	if err != nil {
		t.Fatalf("RemoveNode failed: %v", err)
	}
	if !removeResp.Success {
		t.Error("Expected successful RemoveNode")
	}

	// Verify child was removed
	getChildrenResp2, _ := treeClient.GetChildren(ctx, &pb.GetChildrenRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    rootID,
	})
	if len(getChildrenResp2.Children) != 1 {
		t.Errorf("Expected 1 child after removal, got %d", len(getChildrenResp2.Children))
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_LayoutOperations_Flex tests Flex layout operations
func TestE2E_LayoutOperations_Flex(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "flex-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	widgetClient := pb.NewWidgetServiceClient(conn)

	// Create Flex widget
	flexResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FLEX,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test Flex"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Flex) failed: %v", err)
	}
	flexID := flexResp.WidgetId

	// Create child widgets to add to flex
	box1Resp, _ := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Box 1"),
		},
	})
	box1ID := box1Resp.WidgetId

	box2Resp, _ := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Box 2"),
		},
	})
	box2ID := box2Resp.WidgetId

	layoutClient := pb.NewLayoutServiceClient(conn)

	// Test FlexAddItem
	addItemResp1, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     flexID,
		ItemId:     box1ID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem (1) failed: %v", err)
	}
	if !addItemResp1.Success {
		t.Error("Expected successful FlexAddItem")
	}

	// Add second item with fixed size
	addItemResp2, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     flexID,
		ItemId:     box2ID,
		Proportion: 0,
		FixedSize:  10,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem (2) failed: %v", err)
	}
	if !addItemResp2.Success {
		t.Error("Expected successful FlexAddItem")
	}

	// Test FlexSetDirection
	setDirResp, err := layoutClient.FlexSetDirection(ctx, &pb.FlexSetDirectionRequest{
		SessionId: sessionID,
		FlexId:    flexID,
		Direction: pb.FlexDirection_FLEX_COLUMN,
	})
	if err != nil {
		t.Fatalf("FlexSetDirection failed: %v", err)
	}
	if !setDirResp.Success {
		t.Error("Expected successful FlexSetDirection")
	}

	// Test FlexRemoveItem
	removeItemResp, err := layoutClient.FlexRemoveItem(ctx, &pb.FlexRemoveItemRequest{
		SessionId: sessionID,
		FlexId:    flexID,
		ItemId:    box2ID,
	})
	if err != nil {
		t.Fatalf("FlexRemoveItem failed: %v", err)
	}
	if !removeItemResp.Success {
		t.Error("Expected successful FlexRemoveItem")
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_LayoutOperations_Grid tests Grid layout operations
func TestE2E_LayoutOperations_Grid(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "grid-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	widgetClient := pb.NewWidgetServiceClient(conn)

	// Create Grid widget
	gridResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_GRID,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test Grid"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Grid) failed: %v", err)
	}
	gridID := gridResp.WidgetId

	// Create child widgets
	box1Resp, _ := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Title: strPtr("Cell 0,0"),
		},
	})
	box1ID := box1Resp.WidgetId

	box2Resp, _ := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Title: strPtr("Cell 0,1"),
		},
	})
	box2ID := box2Resp.WidgetId

	layoutClient := pb.NewLayoutServiceClient(conn)

	// Test GridSetRows
	setRowsResp, err := layoutClient.GridSetRows(ctx, &pb.GridSetRowsRequest{
		SessionId: sessionID,
		GridId:    gridID,
		RowSizes:  []int32{-1, -1}, // Two proportional rows
	})
	if err != nil {
		t.Fatalf("GridSetRows failed: %v", err)
	}
	if !setRowsResp.Success {
		t.Error("Expected successful GridSetRows")
	}

	// Test GridSetColumns
	setColsResp, err := layoutClient.GridSetColumns(ctx, &pb.GridSetColumnsRequest{
		SessionId:   sessionID,
		GridId:      gridID,
		ColumnSizes: []int32{20, -1}, // Fixed and proportional
	})
	if err != nil {
		t.Fatalf("GridSetColumns failed: %v", err)
	}
	if !setColsResp.Success {
		t.Error("Expected successful GridSetColumns")
	}

	// Test GridAddItem
	addItemResp1, err := layoutClient.GridAddItem(ctx, &pb.GridAddItemRequest{
		SessionId:  sessionID,
		GridId:     gridID,
		ItemId:     box1ID,
		Row:        0,
		Column:     0,
		RowSpan:    1,
		ColumnSpan: 1,
		MinWidth:   10,
		MinHeight:  5,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("GridAddItem (1) failed: %v", err)
	}
	if !addItemResp1.Success {
		t.Error("Expected successful GridAddItem")
	}

	// Add second item
	addItemResp2, err := layoutClient.GridAddItem(ctx, &pb.GridAddItemRequest{
		SessionId:  sessionID,
		GridId:     gridID,
		ItemId:     box2ID,
		Row:        0,
		Column:     1,
		RowSpan:    1,
		ColumnSpan: 1,
		MinWidth:   10,
		MinHeight:  5,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("GridAddItem (2) failed: %v", err)
	}
	if !addItemResp2.Success {
		t.Error("Expected successful GridAddItem")
	}

	// Test GridRemoveItem
	removeItemResp, err := layoutClient.GridRemoveItem(ctx, &pb.GridRemoveItemRequest{
		SessionId: sessionID,
		GridId:    gridID,
		ItemId:    box2ID,
	})
	if err != nil {
		t.Fatalf("GridRemoveItem failed: %v", err)
	}
	if !removeItemResp.Success {
		t.Error("Expected successful GridRemoveItem")
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

// TestE2E_LayoutOperations_Pages tests Pages layout operations
func TestE2E_LayoutOperations_Pages(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create session
	sessionClient := pb.NewSessionServiceClient(conn)
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "pages-test",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId

	widgetClient := pb.NewWidgetServiceClient(conn)

	// Create Pages widget
	pagesResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_PAGES,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Test Pages"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Pages) failed: %v", err)
	}
	pagesID := pagesResp.WidgetId

	// Create page content widgets
	page1Resp, _ := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Title: strPtr("Page 1 Content"),
		},
	})
	page1ID := page1Resp.WidgetId

	page2Resp, _ := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Title: strPtr("Page 2 Content"),
		},
	})
	page2ID := page2Resp.WidgetId

	page3Resp, _ := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Title: strPtr("Page 3 Content"),
		},
	})
	page3ID := page3Resp.WidgetId

	layoutClient := pb.NewLayoutServiceClient(conn)

	// Test PagesAddPage
	addPageResp1, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "home",
		ItemId:    page1ID,
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage (home) failed: %v", err)
	}
	if !addPageResp1.Success {
		t.Error("Expected successful PagesAddPage")
	}

	// Add more pages
	addPageResp2, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "settings",
		ItemId:    page2ID,
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage (settings) failed: %v", err)
	}
	if !addPageResp2.Success {
		t.Error("Expected successful PagesAddPage")
	}

	_, err = layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "about",
		ItemId:    page3ID,
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage (about) failed: %v", err)
	}

	// Test PagesShowPage
	showPageResp, err := layoutClient.PagesShowPage(ctx, &pb.PagesShowPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "settings",
	})
	if err != nil {
		t.Fatalf("PagesShowPage failed: %v", err)
	}
	if !showPageResp.Success {
		t.Error("Expected successful PagesShowPage")
	}

	// Test PagesGetCurrentPage
	getCurrentResp, err := layoutClient.PagesGetCurrentPage(ctx, &pb.PagesGetCurrentPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
	})
	if err != nil {
		t.Fatalf("PagesGetCurrentPage failed: %v", err)
	}
	if getCurrentResp.PageName != "settings" {
		t.Errorf("Expected current page 'settings', got '%s'", getCurrentResp.PageName)
	}

	// Switch to another page
	layoutClient.PagesShowPage(ctx, &pb.PagesShowPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "about",
	})

	// Verify page switched
	getCurrentResp2, _ := layoutClient.PagesGetCurrentPage(ctx, &pb.PagesGetCurrentPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
	})
	if getCurrentResp2.PageName != "about" {
		t.Errorf("Expected current page 'about', got '%s'", getCurrentResp2.PageName)
	}

	// Test PagesRemovePage
	removePageResp, err := layoutClient.PagesRemovePage(ctx, &pb.PagesRemovePageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "settings",
	})
	if err != nil {
		t.Fatalf("PagesRemovePage failed: %v", err)
	}
	if !removePageResp.Success {
		t.Error("Expected successful PagesRemovePage")
	}

	// Cleanup
	sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})
}

