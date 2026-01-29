package cli

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/config"
	"github.com/chazu/yutani/pkg/server"
	"github.com/chazu/yutani/pkg/services"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	serverAddress     string
	serverMaxSessions int
	serverMouse       bool
	serverPaste       bool
	serverLogLevel    string
	serverDebug       bool
	serverHeadless    bool
	serverLogFile     string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a Yutani server",
	Long: `Start a Yutani display server.

The server listens for gRPC client connections and manages the terminal UI.
Multiple clients can connect to a single server instance.

In normal (TUI) mode the server renders widgets on screen.
In headless mode (--headless) the server uses a simulated screen,
useful for automated testing and CI environments.`,
	Example: `  # Start server on default port
  yutani server

  # Start server on custom port
  yutani server --address :8080

  # Start headless server for testing
  yutani server --headless

  # Start with debug logging to file
  yutani server --log-file server.log --debug`,
	RunE: runServer,
}

func init() {
	serverCmd.Flags().StringVarP(&serverAddress, "address", "a", "", "Server address to listen on")
	serverCmd.Flags().IntVarP(&serverMaxSessions, "max-sessions", "m", 0, "Maximum number of concurrent sessions")
	serverCmd.Flags().BoolVar(&serverMouse, "mouse", true, "Enable mouse support")
	serverCmd.Flags().BoolVar(&serverPaste, "paste", true, "Enable paste support")
	serverCmd.Flags().StringVarP(&serverLogLevel, "log-level", "l", "", "Log level (debug, info, warn, error)")
	serverCmd.Flags().BoolVarP(&serverDebug, "debug", "d", false, "Enable debug mode")
	serverCmd.Flags().BoolVar(&serverHeadless, "headless", false, "Run in headless mode (no TUI, simulated screen)")
	serverCmd.Flags().StringVar(&serverLogFile, "log-file", "", "Log to file instead of stderr (TUI mode discards logs by default)")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load config from file + env (skip flag.Parse — Cobra handles flags)
	cfg, err := config.LoadWithoutFlags()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with explicitly-set Cobra flags
	if cmd.Flags().Changed("address") {
		cfg.Address = serverAddress
	}
	if cmd.Flags().Changed("max-sessions") {
		cfg.MaxSessions = serverMaxSessions
	}
	if cmd.Flags().Changed("mouse") {
		cfg.Mouse = serverMouse
	}
	if cmd.Flags().Changed("paste") {
		cfg.Paste = serverPaste
	}
	if cmd.Flags().Changed("log-level") {
		cfg.LogLevel = serverLogLevel
	}
	if serverDebug {
		cfg.LogLevel = "debug"
	}

	// Configure logging
	logWriter := configureLogging(cfg.LogLevel)
	if c, ok := logWriter.(io.Closer); ok && c != os.Stderr {
		defer c.Close()
	}

	slog.Info("Starting Yutani Display Server",
		"version", server.Version,
		"address", cfg.Address,
		"max_sessions", cfg.MaxSessions,
		"headless", serverHeadless,
	)

	// Create the Yutani server
	var serverOpts []server.ServerOption
	if serverHeadless {
		serverOpts = append(serverOpts, server.WithTestMode(true))
	}

	srv, err := server.New(cfg, serverOpts...)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Start the server (initializes screen and widgets)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer srv.Stop()

	// Create gRPC server and register all services
	grpcServer := grpc.NewServer()
	services.RegisterAllServices(grpcServer, srv)
	reflection.Register(grpcServer)
	slog.Info("gRPC reflection enabled")

	// Create listener
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.Address, err)
	}
	slog.Info("gRPC server listening", "address", cfg.Address)

	if serverHeadless {
		return runHeadless(srv, grpcServer, listener)
	}
	return runTUI(srv, grpcServer, listener)
}

// runTUI runs the server in normal TUI mode.
// gRPC serves in a background goroutine; tview blocks the main goroutine.
func runTUI(srv *server.Server, grpcServer *grpc.Server, listener net.Listener) error {
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			// Server stopped
		}
	}()

	// Run the TUI application (blocks until Ctrl+C or app.Stop())
	if err := srv.Run(); err != nil {
		slog.Error("TUI error", "error", err)
	}

	slog.Info("Shutting down gracefully...")
	grpcServer.GracefulStop()
	slog.Info("Shutdown complete")
	return nil
}

// runHeadless runs the server without a TUI.
// tview runs in a background goroutine (for QueueUpdateDraw);
// gRPC serves on the main goroutine. Ctrl+C triggers graceful shutdown.
func runHeadless(srv *server.Server, grpcServer *grpc.Server, listener net.Listener) error {
	// Run tview event loop in background (required for QueueUpdateDraw)
	go func() {
		if err := srv.Run(); err != nil {
			slog.Error("Server run error", "error", err)
		}
	}()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		grpcServer.GracefulStop()
	}()

	fmt.Printf("Yutani server (headless) listening on %s\n", listener.Addr())
	fmt.Println("Press Ctrl+C to stop.")

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("gRPC serve error: %w", err)
	}
	return nil
}

// configureLogging sets up slog based on mode and flags.
func configureLogging(levelStr string) io.Writer {
	level := parseLogLevel(levelStr)

	// Determine log destination
	var w io.Writer
	if serverLogFile != "" {
		f, err := os.OpenFile(serverLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.Error("Failed to open log file, falling back to stderr", "error", err)
			w = os.Stderr
		} else {
			w = f
		}
	} else if serverHeadless {
		// Headless mode: log to stderr
		w = os.Stderr
	} else {
		// TUI mode: discard logs (they corrupt the display)
		w = io.Discard
	}

	logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	return w
}

// parseLogLevel converts a string log level to slog.Level.
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
