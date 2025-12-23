package cli

import (
	"context"
	"fmt"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	pingAddress string
	pingTimeout int
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check if a Yutani server is running",
	Long: `Check if a Yutani server is running and responding.

Returns server information and latency.`,
	Example: `  # Ping default server
  yutani ping

  # Ping custom address
  yutani ping --address localhost:8080

  # Set timeout
  yutani ping --timeout 10`,
	RunE: runPing,
}

func init() {
	pingCmd.Flags().StringVarP(&pingAddress, "address", "a", "localhost:7755", "Server address")
	pingCmd.Flags().IntVarP(&pingTimeout, "timeout", "t", 5, "Timeout in seconds")
}

func runPing(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pingTimeout)*time.Second)
	defer cancel()

	fmt.Printf("Pinging %s...\n", pingAddress)

	start := time.Now()

	conn, err := grpc.DialContext(ctx, pingAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	connectTime := time.Since(start)

	sessionClient := pb.NewSessionServiceClient(conn)

	// Ping the server
	pingStart := time.Now()
	resp, err := sessionClient.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	pingTime := time.Since(pingStart)

	fmt.Printf("\nServer is alive!\n")
	fmt.Printf("  Connect time: %v\n", connectTime)
	fmt.Printf("  Ping latency: %v\n", pingTime)
	if resp.ServerTimestamp != 0 {
		serverTime := time.Unix(0, resp.ServerTimestamp)
		fmt.Printf("  Server time:  %s\n", serverTime.Format(time.RFC3339))
	}

	// Get server info
	infoResp, err := sessionClient.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
	if err == nil && infoResp != nil {
		fmt.Printf("\nServer Info:\n")
		if infoResp.Version != "" {
			fmt.Printf("  Version:      %s\n", infoResp.Version)
		}
		if infoResp.MaxSessions != 0 {
			fmt.Printf("  Max sessions: %d\n", infoResp.MaxSessions)
		}
		if infoResp.ActiveSessions != 0 {
			fmt.Printf("  Active:       %d\n", infoResp.ActiveSessions)
		}
	}

	return nil
}
