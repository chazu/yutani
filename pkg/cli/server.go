package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/config"
	"github.com/chazu/yutani/pkg/server"
	"github.com/spf13/cobra"
)

var (
	serverAddress    string
	serverMaxSessions int
	serverMouse      bool
	serverPaste      bool
	serverLogLevel   string
	serverDebug      bool
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a Yutani server",
	Long: `Start a Yutani TUI server.

The server listens for client connections and manages the terminal UI.
Multiple clients can connect to a single server instance.`,
	Example: `  # Start server on default port
  yutani server

  # Start server on custom port
  yutani server --address :8080

  # Start server with debug logging
  yutani server --debug

  # Start server with custom settings
  yutani server --address :7755 --max-sessions 20 --mouse --paste`,
	RunE: runServer,
}

func init() {
	serverCmd.Flags().StringVarP(&serverAddress, "address", "a", ":7755", "Server address to listen on")
	serverCmd.Flags().IntVarP(&serverMaxSessions, "max-sessions", "m", 10, "Maximum number of concurrent sessions")
	serverCmd.Flags().BoolVar(&serverMouse, "mouse", true, "Enable mouse support")
	serverCmd.Flags().BoolVar(&serverPaste, "paste", true, "Enable paste support")
	serverCmd.Flags().StringVarP(&serverLogLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	serverCmd.Flags().BoolVarP(&serverDebug, "debug", "d", false, "Enable debug mode")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Override log level if debug is enabled
	if serverDebug {
		serverLogLevel = "debug"
	}

	// Create configuration
	cfg := &config.Config{
		Address:     serverAddress,
		MaxSessions: serverMaxSessions,
		Mouse:       serverMouse,
		Paste:       serverPaste,
		LogLevel:    serverLogLevel,
	}

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Print startup message
	fmt.Printf("Starting Yutani server on %s\n", serverAddress)
	fmt.Printf("Max sessions: %d\n", serverMaxSessions)
	fmt.Printf("Mouse support: %v\n", serverMouse)
	fmt.Printf("Paste support: %v\n", serverPaste)
	fmt.Printf("Log level: %s\n", serverLogLevel)
	fmt.Println("\nPress Ctrl+C to stop the server")

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	// Wait for interrupt or error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-sigCh:
		fmt.Println("\nShutting down server...")
		srv.Stop()
		fmt.Println("Server stopped")
	}

	return nil
}

