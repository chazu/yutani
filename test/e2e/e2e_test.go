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

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			t.Logf("Server start error: %v", err)
		}
	}()

	// Give server time to start
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

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

