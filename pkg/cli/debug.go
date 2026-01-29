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
  - Screen dump (ASCII representation)
  - Widget state inspection
  - Widget tree visualization
  - Event stream monitoring
  - Connection debugging`,
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

var debugScreenCmd = &cobra.Command{
	Use:   "screen",
	Short: "Dump current screen contents",
	Long: `Get an ASCII art representation of the current screen contents.

This is useful for LLM agents to understand what is currently displayed
in the terminal UI without needing visual access.`,
	Example: `  # Dump screen contents
  yutani debug screen -s <session-id>

  # Dump with widget bounds overlay
  yutani debug screen -s <session-id> --bounds

  # Dump with legend showing widget info
  yutani debug screen -s <session-id> --bounds --legend

  # Output as JSON
  yutani debug screen -s <session-id> --format json`,
	RunE: runDebugScreen,
}

var debugWidgetCmd = &cobra.Command{
	Use:   "widget",
	Short: "Inspect widget state",
	Long: `Get detailed state information about a specific widget.

Shows:
  - Widget type and ID
  - Screen bounds (position and size)
  - Focus state
  - Visibility
  - Properties
  - Parent/child relationships
  - Recent events`,
	Example: `  # Inspect a specific widget
  yutani debug widget -s <session-id> -w <widget-id>

  # Output as JSON
  yutani debug widget -s <session-id> -w <widget-id> --format json`,
	RunE: runDebugWidget,
}

var debugBoundsCmd = &cobra.Command{
	Use:   "bounds",
	Short: "List all widget bounds",
	Long:  `Get bounds information for all widgets in a session.`,
	Example: `  # List all widget bounds
  yutani debug bounds -s <session-id>

  # Output as JSON
  yutani debug bounds -s <session-id> --format json`,
	RunE: runDebugBounds,
}

var (
	debugAddress    string
	debugVerbose    bool
	debugTypes      []string
	debugSessionID  string
	debugFormat     string
	debugWidgetID   string
	debugShowBounds bool
	debugShowLegend bool
)

func init() {
	debugCmd.AddCommand(debugTreeCmd)
	debugCmd.AddCommand(debugEventsCmd)
	debugCmd.AddCommand(debugScreenCmd)
	debugCmd.AddCommand(debugWidgetCmd)
	debugCmd.AddCommand(debugBoundsCmd)

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

	// Screen command flags
	debugScreenCmd.Flags().StringVarP(&debugAddress, "address", "a", "localhost:7755", "Server address")
	debugScreenCmd.Flags().StringVarP(&debugSessionID, "session", "s", "", "Session ID (required)")
	debugScreenCmd.Flags().BoolVar(&debugShowBounds, "bounds", false, "Overlay widget bounds on screen")
	debugScreenCmd.Flags().BoolVar(&debugShowLegend, "legend", false, "Include widget bounds legend")
	debugScreenCmd.Flags().StringVarP(&debugFormat, "format", "f", "text", "Output format (text, json)")

	// Widget command flags
	debugWidgetCmd.Flags().StringVarP(&debugAddress, "address", "a", "localhost:7755", "Server address")
	debugWidgetCmd.Flags().StringVarP(&debugSessionID, "session", "s", "", "Session ID (required)")
	debugWidgetCmd.Flags().StringVarP(&debugWidgetID, "widget", "w", "", "Widget ID (required)")
	debugWidgetCmd.Flags().StringVarP(&debugFormat, "format", "f", "text", "Output format (text, json)")

	// Bounds command flags
	debugBoundsCmd.Flags().StringVarP(&debugAddress, "address", "a", "localhost:7755", "Server address")
	debugBoundsCmd.Flags().StringVarP(&debugSessionID, "session", "s", "", "Session ID (required)")
	debugBoundsCmd.Flags().StringVarP(&debugFormat, "format", "f", "text", "Output format (text, json)")
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

func runDebugScreen(cmd *cobra.Command, args []string) error {
	if debugSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := debugConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewDebugServiceClient(conn)

	resp, err := client.GetScreenDump(context.Background(), &pb.GetScreenDumpRequest{
		SessionId:        &pb.SessionId{Id: debugSessionID},
		ShowWidgetBounds: debugShowBounds,
		IncludeLegend:    debugShowLegend,
	})
	if err != nil {
		return fmt.Errorf("failed to get screen dump: %w", err)
	}

	if debugFormat == "json" {
		type jsonOutput struct {
			ScreenDump    string                 `json:"screen_dump"`
			ScreenSize    map[string]int32       `json:"screen_size"`
			WidgetBounds  []*pb.WidgetBoundsInfo `json:"widget_bounds,omitempty"`
			FocusedWidget string                 `json:"focused_widget,omitempty"`
		}

		output := jsonOutput{
			ScreenDump: resp.ScreenDump,
			ScreenSize: map[string]int32{
				"width":  resp.ScreenSize.Width,
				"height": resp.ScreenSize.Height,
			},
			WidgetBounds: resp.WidgetBounds,
		}
		if resp.FocusedWidget != nil {
			output.FocusedWidget = resp.FocusedWidget.Id
		}

		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Text output
	fmt.Println("Screen Dump")
	fmt.Println("===========")
	fmt.Printf("Session: %s\n", debugSessionID)
	fmt.Printf("Size: %dx%d\n", resp.ScreenSize.Width, resp.ScreenSize.Height)
	if resp.FocusedWidget != nil {
		focusedID := resp.FocusedWidget.Id
		if len(focusedID) > 12 {
			focusedID = focusedID[:12] + "..."
		}
		fmt.Printf("Focused: %s\n", focusedID)
	}
	fmt.Println()
	fmt.Println(resp.ScreenDump)

	// Show legend if requested
	if debugShowLegend && len(resp.WidgetBounds) > 0 {
		fmt.Println()
		fmt.Println("Widget Bounds Legend")
		fmt.Println("--------------------")
		for _, b := range resp.WidgetBounds {
			widgetType := strings.TrimPrefix(b.WidgetType.String(), "WIDGET_")
			focusMarker := ""
			if b.HasFocus {
				focusMarker = " [FOCUSED]"
			}
			shortID := b.WidgetId.Id
			if len(shortID) > 12 {
				shortID = shortID[:12] + "..."
			}
			fmt.Printf("  %s: %s (%s) at (%d,%d) %dx%d%s\n",
				b.ShortId, shortID, widgetType,
				b.Bounds.X, b.Bounds.Y,
				b.Bounds.Width, b.Bounds.Height,
				focusMarker)
		}
	}

	return nil
}

func runDebugWidget(cmd *cobra.Command, args []string) error {
	if debugSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}
	if debugWidgetID == "" {
		return fmt.Errorf("widget ID required (--widget)")
	}

	conn, err := debugConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewDebugServiceClient(conn)

	resp, err := client.GetWidgetState(context.Background(), &pb.GetWidgetStateRequest{
		SessionId: &pb.SessionId{Id: debugSessionID},
		WidgetId:  &pb.WidgetId{Id: debugWidgetID},
	})
	if err != nil {
		return fmt.Errorf("failed to get widget state: %w", err)
	}

	if debugFormat == "json" {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Text output
	widgetType := strings.TrimPrefix(resp.WidgetType.String(), "WIDGET_")

	fmt.Println("Widget State")
	fmt.Println("============")
	fmt.Printf("ID:       %s\n", resp.WidgetId.Id)
	fmt.Printf("Type:     %s\n", widgetType)
	fmt.Printf("Bounds:   (%d,%d) %dx%d\n",
		resp.Bounds.X, resp.Bounds.Y,
		resp.Bounds.Width, resp.Bounds.Height)
	fmt.Printf("Focus:    %v\n", resp.HasFocus)
	fmt.Printf("Visible:  %v\n", resp.IsVisible)

	if resp.ParentId != nil {
		parentID := resp.ParentId.Id
		if len(parentID) > 16 {
			parentID = parentID[:16] + "..."
		}
		fmt.Printf("Parent:   %s\n", parentID)
	}

	if len(resp.Children) > 0 {
		fmt.Printf("Children: %d\n", len(resp.Children))
		for i, child := range resp.Children {
			childID := child.Id
			if len(childID) > 16 {
				childID = childID[:16] + "..."
			}
			fmt.Printf("  [%d] %s\n", i, childID)
		}
	}

	if len(resp.RecentEvents) > 0 {
		fmt.Println()
		fmt.Println("Recent Events:")
		for _, event := range resp.RecentEvents {
			timestamp := time.Unix(0, event.Timestamp).Format("15:04:05.000")
			fmt.Printf("  %s: %s\n", timestamp, event.EventType)
			for k, v := range event.Details {
				fmt.Printf("           %s: %s\n", k, v)
			}
		}
	}

	return nil
}

func runDebugBounds(cmd *cobra.Command, args []string) error {
	if debugSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := debugConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewDebugServiceClient(conn)

	resp, err := client.GetAllWidgetBounds(context.Background(), &pb.GetAllWidgetBoundsRequest{
		SessionId: &pb.SessionId{Id: debugSessionID},
	})
	if err != nil {
		return fmt.Errorf("failed to get widget bounds: %w", err)
	}

	if debugFormat == "json" {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Text output
	fmt.Println("Widget Bounds")
	fmt.Println("=============")
	fmt.Printf("Session: %s\n", debugSessionID)
	fmt.Printf("Screen:  %dx%d\n", resp.ScreenSize.Width, resp.ScreenSize.Height)
	if resp.FocusedWidget != nil {
		focusedID := resp.FocusedWidget.Id
		if len(focusedID) > 12 {
			focusedID = focusedID[:12] + "..."
		}
		fmt.Printf("Focused: %s\n", focusedID)
	}
	fmt.Println()

	if len(resp.Widgets) == 0 {
		fmt.Println("(no widgets)")
		return nil
	}

	// Table header
	fmt.Printf("%-4s %-16s %-12s %-20s %s\n", "ID", "Widget ID", "Type", "Bounds", "Focus")
	fmt.Println(strings.Repeat("-", 70))

	for _, w := range resp.Widgets {
		widgetType := strings.TrimPrefix(w.WidgetType.String(), "WIDGET_")
		widgetID := w.WidgetId.Id
		if len(widgetID) > 14 {
			widgetID = widgetID[:14] + ".."
		}

		bounds := fmt.Sprintf("(%d,%d) %dx%d",
			w.Bounds.X, w.Bounds.Y,
			w.Bounds.Width, w.Bounds.Height)

		focusStr := ""
		if w.HasFocus {
			focusStr = "*"
		}

		fmt.Printf("%-4s %-16s %-12s %-20s %s\n",
			w.ShortId, widgetID, widgetType, bounds, focusStr)
	}

	return nil
}

