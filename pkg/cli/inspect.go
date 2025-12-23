package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	inspectAddress string
	inspectFormat  string
	inspectWatch   bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect a running Yutani server",
	Long: `Inspect a running Yutani server and display information about sessions and widgets.

This command connects to a running server and displays:
  - Active sessions
  - Widget hierarchy
  - Widget properties
  - Connection statistics`,
	Example: `  # Inspect server on default port
  yutani inspect

  # Inspect server on custom port
  yutani inspect --address localhost:8080

  # Watch for changes
  yutani inspect --watch

  # Output as JSON
  yutani inspect --format json`,
	RunE: runInspect,
}

func init() {
	inspectCmd.Flags().StringVarP(&inspectAddress, "address", "a", "localhost:7755", "Server address to inspect")
	inspectCmd.Flags().StringVarP(&inspectFormat, "format", "f", "text", "Output format (text, json)")
	inspectCmd.Flags().BoolVarP(&inspectWatch, "watch", "w", false, "Watch for changes")
}

func runInspect(cmd *cobra.Command, args []string) error {
	// Connect to server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, inspectAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)

	// Get server info
	if inspectWatch {
		return watchServer(sessionClient)
	}

	return inspectServer(sessionClient)
}

func inspectServer(sessionClient pb.SessionServiceClient) error {
	ctx := context.Background()

	// Get server info
	info, err := sessionClient.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	if inspectFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"address":         inspectAddress,
			"version":         info.Version,
			"max_sessions":    info.MaxSessions,
			"active_sessions": info.ActiveSessions,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("Yutani Server Inspection")
	fmt.Println("========================")
	fmt.Printf("\nAddress: %s\n", inspectAddress)

	fmt.Println("\nServer Info:")
	if info.Version != "" {
		fmt.Printf("  Version:        %s\n", info.Version)
	}
	fmt.Printf("  Max Sessions:   %d\n", info.MaxSessions)
	fmt.Printf("  Active:         %d\n", info.ActiveSessions)
	if info.ScreenSize != nil {
		fmt.Printf("  Screen Size:    %dx%d\n", info.ScreenSize.Width, info.ScreenSize.Height)
	}
	if info.Capabilities != nil {
		fmt.Printf("  Mouse Support:  %v\n", info.Capabilities.MouseSupport)
		fmt.Printf("  Paste Support:  %v\n", info.Capabilities.PasteSupport)
	}

	return nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func watchServer(sessionClient pb.SessionServiceClient) error {
	fmt.Println("Watching server (press Ctrl+C to stop)...")
	fmt.Println()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		if err := inspectServer(sessionClient); err != nil {
			return err
		}
	}

	return nil
}

