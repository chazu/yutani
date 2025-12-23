package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug utilities for Yutani applications",
	Long: `Debug utilities for Yutani applications.

Provides tools for:
  - Widget tree visualization
  - Event stream monitoring
  - Connection debugging
  - State inspection`,
}

var debugTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display widget tree",
	Long:  `Display the widget hierarchy tree for a running application.`,
	Example: `  # Display widget tree
  yutani debug tree

  # Display tree with properties
  yutani debug tree --verbose`,
	RunE: runDebugTree,
}

var debugEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Monitor event stream",
	Long:  `Monitor the event stream from a running application.`,
	Example: `  # Monitor all events
  yutani debug events

  # Monitor specific event types
  yutani debug events --type key,mouse`,
	RunE: runDebugEvents,
}

var (
	debugAddress   string
	debugVerbose   bool
	debugTypes     []string
	debugSessionID string
	debugFormat    string
)

func init() {
	debugCmd.AddCommand(debugTreeCmd)
	debugCmd.AddCommand(debugEventsCmd)

	// Tree command flags
	debugTreeCmd.Flags().StringVarP(&debugAddress, "address", "a", "localhost:7755", "Server address")
	debugTreeCmd.Flags().StringVarP(&debugSessionID, "session", "s", "", "Session ID")
	debugTreeCmd.Flags().BoolVarP(&debugVerbose, "verbose", "v", false, "Show widget properties")
	debugTreeCmd.Flags().StringVarP(&debugFormat, "format", "f", "text", "Output format (text, json)")

	// Events command flags
	debugEventsCmd.Flags().StringVarP(&debugAddress, "address", "a", "localhost:7755", "Server address")
	debugEventsCmd.Flags().StringVarP(&debugSessionID, "session", "s", "", "Session ID (required)")
	debugEventsCmd.Flags().StringSliceVarP(&debugTypes, "type", "t", []string{}, "Event types to monitor (key, mouse, resize, focus, widget)")
	debugEventsCmd.Flags().StringVarP(&debugFormat, "format", "f", "text", "Output format (text, json)")
}

func debugConnect() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return grpc.DialContext(ctx, debugAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

func runDebugTree(cmd *cobra.Command, args []string) error {
	if debugSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := debugConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewWidgetServiceClient(conn)

	resp, err := client.ListWidgets(context.Background(), &pb.ListWidgetsRequest{
		SessionId: &pb.SessionId{Id: debugSessionID},
	})
	if err != nil {
		return fmt.Errorf("failed to list widgets: %w", err)
	}

	if debugFormat == "json" {
		data, _ := json.MarshalIndent(resp.Widgets, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("Widget Tree")
	fmt.Println("===========")
	fmt.Printf("Session: %s\n\n", debugSessionID)

	if len(resp.Widgets) == 0 {
		fmt.Println("(no widgets)")
		return nil
	}

	// Build a simple tree display
	// Group widgets by parent to show hierarchy
	for i, w := range resp.Widgets {
		prefix := "├──"
		if i == len(resp.Widgets)-1 {
			prefix = "└──"
		}

		widgetType := strings.TrimPrefix(w.Type.String(), "WIDGET_")
		widgetID := w.WidgetId.Id
		if len(widgetID) > 8 {
			widgetID = widgetID[:8]
		}

		parentInfo := ""
		if w.ParentId != nil {
			parentID := w.ParentId.Id
			if len(parentID) > 8 {
				parentID = parentID[:8]
			}
			parentInfo = fmt.Sprintf(" (parent: %s)", parentID)
		}

		fmt.Printf("%s %s (%s)%s\n", prefix, widgetID, widgetType, parentInfo)

		if debugVerbose && len(w.Children) > 0 {
			indent := "│   "
			if i == len(resp.Widgets)-1 {
				indent = "    "
			}
			fmt.Printf("%s  children: %d\n", indent, len(w.Children))
		}
	}

	return nil
}

func runDebugEvents(cmd *cobra.Command, args []string) error {
	if debugSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := debugConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewEventServiceClient(conn)

	fmt.Println("Event Stream Monitor")
	fmt.Println("====================")
	fmt.Printf("Server:  %s\n", debugAddress)
	fmt.Printf("Session: %s\n", debugSessionID)

	if len(debugTypes) > 0 {
		fmt.Printf("Filter:  %v\n", debugTypes)
	} else {
		fmt.Println("Filter:  all events")
	}

	fmt.Println("\nPress Ctrl+C to stop")
	fmt.Println(strings.Repeat("-", 60))

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nStopping event monitor...")
		cancel()
	}()

	// Subscribe to events
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		SessionId: &pb.SessionId{Id: debugSessionID},
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	eventCount := 0
	for {
		event, err := stream.Recv()
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, normal shutdown
				break
			}
			return fmt.Errorf("stream error: %w", err)
		}

		// Filter by event type if specified
		if len(debugTypes) > 0 && !matchesEventType(event, debugTypes) {
			continue
		}

		eventCount++
		displayEvent(event, eventCount)
	}

	fmt.Printf("\nTotal events received: %d\n", eventCount)
	return nil
}

func matchesEventType(event *pb.Event, types []string) bool {
	eventType := ""
	switch {
	case event.GetKey() != nil:
		eventType = "key"
	case event.GetMouse() != nil:
		eventType = "mouse"
	case event.GetResize() != nil:
		eventType = "resize"
	case event.GetFocus() != nil:
		eventType = "focus"
	case event.GetWidget() != nil:
		eventType = "widget"
	}

	for _, t := range types {
		if strings.EqualFold(t, eventType) {
			return true
		}
	}
	return false
}

func displayEvent(event *pb.Event, count int) {
	timestamp := time.Unix(0, event.Timestamp).Format("15:04:05.000")

	if debugFormat == "json" {
		data, _ := json.Marshal(event)
		fmt.Printf("%s\n", string(data))
		return
	}

	switch {
	case event.GetKey() != nil:
		key := event.GetKey()
		keyName := key.Key.String()
		if key.Rune != "" {
			keyName = fmt.Sprintf("'%s'", key.Rune)
		}
		mods := formatModifiers(key.Modifiers)
		fmt.Printf("[%d] %s KEY      %s%s\n", count, timestamp, mods, keyName)

	case event.GetMouse() != nil:
		mouse := event.GetMouse()
		action := strings.TrimPrefix(mouse.Action.String(), "MOUSE_")
		pos := fmt.Sprintf("(%d,%d)", mouse.Position.X, mouse.Position.Y)
		fmt.Printf("[%d] %s MOUSE    %-12s %s\n", count, timestamp, action, pos)

	case event.GetResize() != nil:
		resize := event.GetResize()
		fmt.Printf("[%d] %s RESIZE   %dx%d\n", count, timestamp, resize.NewSize.Width, resize.NewSize.Height)

	case event.GetFocus() != nil:
		focus := event.GetFocus()
		oldID := "(none)"
		newID := "(none)"
		if focus.OldWidget != nil {
			oldID = focus.OldWidget.Id[:8]
		}
		if focus.NewWidget != nil {
			newID = focus.NewWidget.Id[:8]
		}
		fmt.Printf("[%d] %s FOCUS    %s -> %s\n", count, timestamp, oldID, newID)

	case event.GetWidget() != nil:
		widget := event.GetWidget()
		eventType := strings.TrimPrefix(widget.Type.String(), "WIDGET_")
		widgetID := "(none)"
		if widget.WidgetId != nil {
			widgetID = widget.WidgetId.Id[:8]
		}
		fmt.Printf("[%d] %s WIDGET   %-10s %s\n", count, timestamp, eventType, widgetID)
		if len(widget.Data) > 0 {
			for k, v := range widget.Data {
				fmt.Printf("                           %s: %s\n", k, v)
			}
		}

	default:
		fmt.Printf("[%d] %s UNKNOWN\n", count, timestamp)
	}
}

func formatModifiers(mods []pb.Modifier) string {
	if len(mods) == 0 {
		return ""
	}

	parts := make([]string, 0, len(mods))
	for _, m := range mods {
		switch m {
		case pb.Modifier_MOD_CTRL:
			parts = append(parts, "Ctrl+")
		case pb.Modifier_MOD_ALT:
			parts = append(parts, "Alt+")
		case pb.Modifier_MOD_SHIFT:
			parts = append(parts, "Shift+")
		case pb.Modifier_MOD_META:
			parts = append(parts, "Meta+")
		}
	}
	return strings.Join(parts, "")
}

