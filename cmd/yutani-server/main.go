package main

import (
	"log/slog"
	"net"
	"os"

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
	// IMPORTANT: Logging is disabled when TUI is running to avoid interfering with the display
	// To enable logging, set YUTANI_LOG_FILE environment variable to a file path
	logLevel := parseLogLevel(cfg.LogLevel)

	var logWriter *os.File
	logFile := os.Getenv("YUTANI_LOG_FILE")
	if logFile != "" {
		var err error
		logWriter, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.Error("Failed to open log file", "error", err)
			os.Exit(1)
		}
		defer logWriter.Close()
	} else {
		// Discard logs when no log file is specified (TUI mode)
		logWriter, _ = os.Open(os.DevNull)
		defer logWriter.Close()
	}

	logger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{
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

	// Start the TUI (initializes screen and widgets)
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

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			// Server stopped
		}
	}()

	// Run the TUI application (blocks until Ctrl+C or app.Stop())
	// This MUST run in the main goroutine
	if err := yutaniServer.Run(); err != nil {
		slog.Error("TUI error", "error", err)
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

