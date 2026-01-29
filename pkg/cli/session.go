package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	sessionAddress string
	sessionFormat  string
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage server sessions",
	Long:  `Commands for managing Yutani server sessions.`,
}

var sessionCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new session",
	Long:  `Create a new client session on the server.`,
	Example: `  # Create session with default name
  yutani session create

  # Create named session
  yutani session create my-client`,
	RunE: runSessionCreate,
}

var sessionDestroyCmd = &cobra.Command{
	Use:   "destroy <session-id>",
	Short: "Destroy a session",
	Long:  `Destroy an existing session and cleanup its resources.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionDestroy,
}

var sessionInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show server session info",
	Long:  `Display information about the server and active sessions.`,
	RunE:  runSessionInfo,
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active sessions",
	Long:  `List all active sessions on the server with their IDs and client names.`,
	Example: `  # List all sessions
  yutani session list

  # List sessions as JSON
  yutani session list --format json`,
	RunE: runSessionList,
}

func init() {
	sessionCmd.AddCommand(sessionCreateCmd)
	sessionCmd.AddCommand(sessionDestroyCmd)
	sessionCmd.AddCommand(sessionInfoCmd)
	sessionCmd.AddCommand(sessionListCmd)

	// Common flags
	sessionCmd.PersistentFlags().StringVarP(&sessionAddress, "address", "a", "localhost:7755", "Server address")
	sessionCmd.PersistentFlags().StringVarP(&sessionFormat, "format", "f", "text", "Output format (text, json)")
}

func connectToServer(address string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

func runSessionCreate(cmd *cobra.Command, args []string) error {
	conn, err := connectToServer(sessionAddress, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewSessionServiceClient(conn)

	name := "cli-session"
	if len(args) > 0 {
		name = args[0]
	}

	resp, err := client.CreateSession(context.Background(), &pb.CreateSessionRequest{
		ClientName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	if sessionFormat == "json" {
		data, _ := json.MarshalIndent(map[string]string{
			"session_id": resp.SessionId.Id,
			"status":     "created",
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Session created!\n")
		fmt.Printf("  ID: %s\n", resp.SessionId.Id)
		fmt.Println("\nUse this ID to interact with widgets or destroy the session.")
	}

	return nil
}

func runSessionDestroy(cmd *cobra.Command, args []string) error {
	conn, err := connectToServer(sessionAddress, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewSessionServiceClient(conn)

	sessionID := args[0]

	_, err = client.DestroySession(context.Background(), &pb.DestroySessionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		return fmt.Errorf("failed to destroy session: %w", err)
	}

	if sessionFormat == "json" {
		data, _ := json.MarshalIndent(map[string]string{
			"session_id": sessionID,
			"status":     "destroyed",
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Session %s destroyed.\n", sessionID)
	}

	return nil
}

func runSessionInfo(cmd *cobra.Command, args []string) error {
	conn, err := connectToServer(sessionAddress, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewSessionServiceClient(conn)

	resp, err := client.GetServerInfo(context.Background(), &pb.GetServerInfoRequest{})
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	if sessionFormat == "json" {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("Server Information")
		fmt.Println("==================")
		if resp.Version != "" {
			fmt.Printf("Version:        %s\n", resp.Version)
		}
		fmt.Printf("Max Sessions:   %d\n", resp.MaxSessions)
		fmt.Printf("Active:         %d\n", resp.ActiveSessions)
		if resp.ScreenSize != nil {
			fmt.Printf("Screen Size:    %dx%d\n", resp.ScreenSize.Width, resp.ScreenSize.Height)
		}
	}

	return nil
}

func runSessionList(cmd *cobra.Command, args []string) error {
	conn, err := connectToServer(sessionAddress, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewSessionServiceClient(conn)

	resp, err := client.ListSessions(context.Background(), &pb.ListSessionsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if sessionFormat == "json" {
		type jsonSession struct {
			SessionID  string `json:"session_id"`
			ClientName string `json:"client_name"`
			CreatedAt  string `json:"created_at"`
		}
		sessions := make([]jsonSession, 0, len(resp.Sessions))
		for _, s := range resp.Sessions {
			sessions = append(sessions, jsonSession{
				SessionID:  s.SessionId.Id,
				ClientName: s.ClientName,
				CreatedAt:  time.Unix(0, s.CreatedAt).Format(time.RFC3339),
			})
		}
		data, _ := json.MarshalIndent(sessions, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("Active Sessions")
		fmt.Println("===============")

		if len(resp.Sessions) == 0 {
			fmt.Println("(no active sessions)")
			return nil
		}

		fmt.Printf("\n%-38s %-20s %s\n", "SESSION ID", "CLIENT NAME", "CREATED")
		fmt.Println(strings.Repeat("-", 80))

		for _, s := range resp.Sessions {
			createdAt := time.Unix(0, s.CreatedAt).Format("2006-01-02 15:04:05")
			fmt.Printf("%-38s %-20s %s\n", s.SessionId.Id, s.ClientName, createdAt)
		}
	}

	return nil
}
