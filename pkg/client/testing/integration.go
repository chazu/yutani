package testing

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/chazu/yutani/pkg/client"
	"github.com/chazu/yutani/pkg/config"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/chazu/yutani/pkg/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestServer represents a test server instance.
type TestServer struct {
	t          *testing.T
	server     *server.Server
	grpcServer *grpc.Server
	listener   net.Listener
	address    string
	stopCh     chan struct{}
}

// StartTestServer starts an ephemeral test server.
func StartTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}

	address := listener.Addr().String()
	t.Logf("Test server listening on %s", address)

	// Create server in test mode
	cfg := &config.Config{
		Address:     address,
		MaxSessions: 10,
		Mouse:       true,
		Paste:       true,
		LogLevel:    "error",
	}

	srv, err := server.New(cfg, server.WithTestMode(true))
	if err != nil {
		listener.Close()
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	sessionService := services.NewSessionService(srv)
	widgetService := services.NewWidgetService(srv)
	screenService := services.NewScreenService(srv)
	eventService := services.NewEventService(srv)
	listService := services.NewListService(srv)
	tableService := services.NewTableService(srv)
	formService := services.NewFormService(srv)
	treeService := services.NewTreeService(srv)
	layoutService := services.NewLayoutService(srv)

	pb.RegisterSessionServiceServer(grpcServer, sessionService)
	pb.RegisterWidgetServiceServer(grpcServer, widgetService)
	pb.RegisterScreenServiceServer(grpcServer, screenService)
	pb.RegisterEventServiceServer(grpcServer, eventService)
	pb.RegisterListServiceServer(grpcServer, listService)
	pb.RegisterTableServiceServer(grpcServer, tableService)
	pb.RegisterFormServiceServer(grpcServer, formService)
	pb.RegisterTreeServiceServer(grpcServer, treeService)
	pb.RegisterLayoutServiceServer(grpcServer, layoutService)

	ts := &TestServer{
		t:          t,
		server:     srv,
		grpcServer: grpcServer,
		listener:   listener,
		address:    address,
		stopCh:     make(chan struct{}),
	}

	// Start tview server
	if err := srv.Start(); err != nil {
		listener.Close()
		t.Fatalf("Failed to start server: %v", err)
	}

	// Run tview event loop in background (required for QueueUpdateDraw)
	go func() {
		_ = srv.Run()
	}()

	// Start gRPC server in background
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			select {
			case <-ts.stopCh:
				// Server stopped intentionally
			default:
				t.Logf("Server error: %v", err)
			}
		}
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	return ts
}

// Address returns the server address.
func (ts *TestServer) Address() string {
	return ts.address
}

// Stop stops the test server.
func (ts *TestServer) Stop() {
	ts.t.Helper()
	close(ts.stopCh)
	ts.grpcServer.GracefulStop()
	ts.listener.Close()
	ts.server.Stop()
}

// ConnectToTestServer creates a client connected to the test server.
func ConnectToTestServer(t *testing.T, ts *TestServer) *client.Client {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, ts.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}

	c, err := client.ConnectWithConn(conn)
	if err != nil {
		conn.Close()
		t.Fatalf("Failed to create client: %v", err)
	}

	return c
}

// WithTestServer runs a test function with a test server.
func WithTestServer(t *testing.T, fn func(*testing.T, *TestServer)) {
	t.Helper()
	ts := StartTestServer(t)
	defer ts.Stop()
	fn(t, ts)
}

// WithTestClient runs a test function with a test server and connected client.
func WithTestClient(t *testing.T, fn func(*testing.T, *client.Client)) {
	t.Helper()
	WithTestServer(t, func(t *testing.T, ts *TestServer) {
		c := ConnectToTestServer(t, ts)
		defer c.Close()
		fn(t, c)
	})
}

// AssertServerRunning asserts that the test server is running.
func AssertServerRunning(t *testing.T, ts *TestServer) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, ts.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Errorf("Server not running: %v", err)
		return
	}
	conn.Close()
}

// CreateTestWidget creates a widget on the test server for testing.
func CreateTestWidget(t *testing.T, c *client.Client, widgetType string) string {
	t.Helper()

	var widgetID string
	var err error

	switch widgetType {
	case "list":
		list, e := c.NewList().Title("Test List").Build()
		err = e
		if list != nil {
			widgetID = list.ID()
		}
	case "table":
		table, e := c.NewTable().Title("Test Table").Build()
		err = e
		if table != nil {
			widgetID = table.ID()
		}
	case "form":
		form, e := c.NewForm().Title("Test Form").Build()
		err = e
		if form != nil {
			widgetID = form.ID()
		}
	default:
		t.Fatalf("Unknown widget type: %s", widgetType)
	}

	if err != nil {
		t.Fatalf("Failed to create %s widget: %v", widgetType, err)
	}

	return widgetID
}

