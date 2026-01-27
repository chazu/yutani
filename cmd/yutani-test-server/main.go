// yutani-test-server - A headless Yutani server for testing
// This server runs in test mode (no TUI) and is useful for automated testing
// and LLM debugging.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/config"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/chazu/yutani/pkg/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	address := flag.String("address", ":7755", "gRPC server listen address")
	flag.Parse()

	fmt.Printf("Starting Yutani test server (headless mode) on %s...\n", *address)

	// Create listener
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create server in test mode (uses simulation screen)
	cfg := &config.Config{
		Address:     *address,
		MaxSessions: 100,
		Mouse:       true,
		Paste:       true,
		LogLevel:    "info",
	}

	srv, err := server.New(cfg, server.WithTestMode(true))
	if err != nil {
		listener.Close()
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server (initializes simulation screen)
	if err := srv.Start(); err != nil {
		listener.Close()
		log.Fatalf("Failed to start server: %v", err)
	}

	// Run the tview application loop in a background goroutine
	// This is needed to process QueueUpdateDraw calls from gRPC handlers
	go func() {
		if err := srv.Run(); err != nil {
			log.Printf("Server run error: %v", err)
		}
	}()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterSessionServiceServer(grpcServer, services.NewSessionService(srv))
	pb.RegisterWidgetServiceServer(grpcServer, services.NewWidgetService(srv))
	pb.RegisterScreenServiceServer(grpcServer, services.NewScreenService(srv))
	pb.RegisterEventServiceServer(grpcServer, services.NewEventService(srv))
	pb.RegisterListServiceServer(grpcServer, services.NewListService(srv))
	pb.RegisterTableServiceServer(grpcServer, services.NewTableService(srv))
	pb.RegisterFormServiceServer(grpcServer, services.NewFormService(srv))
	pb.RegisterTreeServiceServer(grpcServer, services.NewTreeService(srv))
	pb.RegisterLayoutServiceServer(grpcServer, services.NewLayoutService(srv))
	pb.RegisterDebugServiceServer(grpcServer, services.NewDebugService(srv))
	pb.RegisterTestServiceServer(grpcServer, services.NewTestService(srv))

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		grpcServer.GracefulStop()
	}()

	fmt.Println("Test server ready. Press Ctrl+C to stop.")

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
