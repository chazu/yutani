package main

import (
	"log/slog"
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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Configure logging
	logLevel := parseLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Yutani Display Server",
		"version", server.Version,
		"address", cfg.Address,
		"max_sessions", cfg.MaxSessions,
	)

	// Create the Yutani server
	yutaniServer, err := server.NewServer(cfg.MaxSessions, cfg.Mouse, cfg.Paste)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	// Start the TUI event loop
	if err := yutaniServer.Start(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	defer yutaniServer.Stop()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	sessionService := services.NewSessionService(yutaniServer)
	screenService := services.NewScreenService(yutaniServer)
	eventService := services.NewEventService(yutaniServer)
	widgetService := services.NewWidgetService(yutaniServer)
	listService := services.NewListService(yutaniServer)
	tableService := services.NewTableService(yutaniServer)
	formService := services.NewFormService(yutaniServer)
	treeService := services.NewTreeService(yutaniServer)
	layoutService := services.NewLayoutService(yutaniServer)

	pb.RegisterSessionServiceServer(grpcServer, sessionService)
	pb.RegisterScreenServiceServer(grpcServer, screenService)
	pb.RegisterEventServiceServer(grpcServer, eventService)
	pb.RegisterWidgetServiceServer(grpcServer, widgetService)
	pb.RegisterListServiceServer(grpcServer, listService)
	pb.RegisterTableServiceServer(grpcServer, tableService)
	pb.RegisterFormServiceServer(grpcServer, formService)
	pb.RegisterTreeServiceServer(grpcServer, treeService)
	pb.RegisterLayoutServiceServer(grpcServer, layoutService)

	// Enable gRPC reflection
	reflection.Register(grpcServer)
	slog.Info("gRPC reflection enabled")

	// Start listening
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		slog.Error("Failed to listen", "error", err, "address", cfg.Address)
		os.Exit(1)
	}

	slog.Info("gRPC server listening", "address", cfg.Address)

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start gRPC server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigCh:
		slog.Info("Received shutdown signal")
	case err := <-errCh:
		slog.Error("gRPC server error", "error", err)
	}

	// Graceful shutdown
	slog.Info("Shutting down gracefully...")
	grpcServer.GracefulStop()
	slog.Info("Shutdown complete")
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

