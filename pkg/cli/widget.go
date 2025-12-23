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
	widgetAddress   string
	widgetSessionID string
	widgetFormat    string
)

var widgetCmd = &cobra.Command{
	Use:   "widget",
	Short: "Manage widgets",
	Long:  `Commands for creating, listing, and managing widgets.`,
}

var widgetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all widgets",
	Long:  `List all widgets for a session.`,
	Example: `  # List widgets for a session
  yutani widget list --session abc123

  # List as JSON
  yutani widget list --session abc123 --format json`,
	RunE: runWidgetList,
}

var widgetCreateCmd = &cobra.Command{
	Use:   "create <type>",
	Short: "Create a widget",
	Long: `Create a new widget of the specified type.

Available types:
  box, textview, inputfield, textarea, button, checkbox,
  dropdown, list, table, treeview, form, flex, grid, pages,
  modal, image, progressbar`,
	Example: `  # Create a text view
  yutani widget create textview --session abc123 --title "Hello"

  # Create a button
  yutani widget create button --session abc123 --label "Click Me"`,
	Args: cobra.ExactArgs(1),
	RunE: runWidgetCreate,
}

var widgetDeleteCmd = &cobra.Command{
	Use:   "delete <widget-id>",
	Short: "Delete a widget",
	Long:  `Delete a widget by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWidgetDelete,
}

var widgetGetCmd = &cobra.Command{
	Use:   "get <widget-id>",
	Short: "Get widget properties",
	Long:  `Get properties of a widget by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWidgetGet,
}

var widgetFocusCmd = &cobra.Command{
	Use:   "focus [widget-id]",
	Short: "Get or set focused widget",
	Long:  `Get the currently focused widget, or set focus to a specific widget.`,
	RunE:  runWidgetFocus,
}

var widgetSetRootCmd = &cobra.Command{
	Use:   "setroot <widget-id>",
	Short: "Set a widget as the root display",
	Long: `Set a widget as the root display widget.

Only the root widget (and its children) are visible on screen.
You must set a root widget for anything to appear.`,
	Example: `  # Create a button and display it
  yutani widget create button --session abc123 --label "Hello"
  yutani widget setroot <widget-id> --session abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runWidgetSetRoot,
}

// Widget property flags
var (
	widgetTitle   string
	widgetBorder  bool
	widgetText    string
	widgetLabel   string
	widgetSetRoot bool
)

func init() {
	widgetCmd.AddCommand(widgetListCmd)
	widgetCmd.AddCommand(widgetCreateCmd)
	widgetCmd.AddCommand(widgetDeleteCmd)
	widgetCmd.AddCommand(widgetGetCmd)
	widgetCmd.AddCommand(widgetFocusCmd)
	widgetCmd.AddCommand(widgetSetRootCmd)

	// Common flags
	widgetCmd.PersistentFlags().StringVarP(&widgetAddress, "address", "a", "localhost:7755", "Server address")
	widgetCmd.PersistentFlags().StringVarP(&widgetSessionID, "session", "s", "", "Session ID (required for most operations)")
	widgetCmd.PersistentFlags().StringVarP(&widgetFormat, "format", "f", "text", "Output format (text, json)")

	// Create command flags
	widgetCreateCmd.Flags().StringVar(&widgetTitle, "title", "", "Widget title")
	widgetCreateCmd.Flags().BoolVar(&widgetBorder, "border", true, "Show border")
	widgetCreateCmd.Flags().StringVar(&widgetText, "text", "", "Text content (for textview, textarea)")
	widgetCreateCmd.Flags().StringVar(&widgetLabel, "label", "", "Label (for button, checkbox)")
	widgetCreateCmd.Flags().BoolVar(&widgetSetRoot, "root", false, "Set as root widget (display immediately)")
}

func widgetConnect() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return grpc.DialContext(ctx, widgetAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

func runWidgetList(cmd *cobra.Command, args []string) error {
	if widgetSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := widgetConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewWidgetServiceClient(conn)

	resp, err := client.ListWidgets(context.Background(), &pb.ListWidgetsRequest{
		SessionId: &pb.SessionId{Id: widgetSessionID},
	})
	if err != nil {
		return fmt.Errorf("failed to list widgets: %w", err)
	}

	if widgetFormat == "json" {
		data, _ := json.MarshalIndent(resp.Widgets, "", "  ")
		fmt.Println(string(data))
	} else {
		if len(resp.Widgets) == 0 {
			fmt.Println("No widgets found.")
			return nil
		}

		fmt.Printf("Widgets for session %s:\n", widgetSessionID)
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("%-36s  %-15s  %s\n", "ID", "TYPE", "PARENT")
		fmt.Println(strings.Repeat("-", 60))

		for _, w := range resp.Widgets {
			widgetType := strings.TrimPrefix(w.Type.String(), "WIDGET_")
			parent := "-"
			if w.ParentId != nil {
				parent = w.ParentId.Id[:8]
			}
			fmt.Printf("%-36s  %-15s  %s\n", w.WidgetId.Id, widgetType, parent)
		}
	}

	return nil
}

func runWidgetCreate(cmd *cobra.Command, args []string) error {
	if widgetSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	widgetType := strings.ToUpper(args[0])
	if !strings.HasPrefix(widgetType, "WIDGET_") {
		widgetType = "WIDGET_" + widgetType
	}

	// Parse widget type
	pbType, ok := pb.WidgetType_value[widgetType]
	if !ok {
		return fmt.Errorf("unknown widget type: %s", args[0])
	}

	conn, err := widgetConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewWidgetServiceClient(conn)

	// Build properties
	props := &pb.WidgetProperties{
		Border: &widgetBorder,
	}
	if widgetTitle != "" {
		props.Title = &widgetTitle
	}

	// Add type-specific properties
	switch pb.WidgetType(pbType) {
	case pb.WidgetType_WIDGET_TEXT_VIEW:
		if widgetText != "" {
			props.TypeProperties = &pb.WidgetProperties_TextView{
				TextView: &pb.TextViewProperties{
					Text: &widgetText,
				},
			}
		}
	case pb.WidgetType_WIDGET_BUTTON:
		if widgetLabel != "" {
			props.TypeProperties = &pb.WidgetProperties_Button{
				Button: &pb.ButtonProperties{
					Label: &widgetLabel,
				},
			}
		}
	case pb.WidgetType_WIDGET_CHECKBOX:
		if widgetLabel != "" {
			props.TypeProperties = &pb.WidgetProperties_Checkbox{
				Checkbox: &pb.CheckboxProperties{
					Label: &widgetLabel,
				},
			}
		}
	}

	resp, err := client.CreateWidget(context.Background(), &pb.CreateWidgetRequest{
		SessionId:  &pb.SessionId{Id: widgetSessionID},
		Type:       pb.WidgetType(pbType),
		Properties: props,
	})
	if err != nil {
		return fmt.Errorf("failed to create widget: %w", err)
	}

	// Set as root if requested
	if widgetSetRoot {
		_, err = client.SetRoot(context.Background(), &pb.SetRootRequest{
			SessionId: &pb.SessionId{Id: widgetSessionID},
			WidgetId:  resp.WidgetId,
		})
		if err != nil {
			return fmt.Errorf("widget created but failed to set as root: %w", err)
		}
	}

	if widgetFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"widget_id": resp.WidgetId.Id,
			"type":      args[0],
			"success":   resp.Success,
			"is_root":   widgetSetRoot,
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Widget created!\n")
		fmt.Printf("  ID:   %s\n", resp.WidgetId.Id)
		fmt.Printf("  Type: %s\n", args[0])
		if widgetSetRoot {
			fmt.Println("  Set as root display")
		}
	}

	return nil
}

func runWidgetDelete(cmd *cobra.Command, args []string) error {
	if widgetSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := widgetConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewWidgetServiceClient(conn)

	_, err = client.DeleteWidget(context.Background(), &pb.DeleteWidgetRequest{
		SessionId: &pb.SessionId{Id: widgetSessionID},
		WidgetId:  &pb.WidgetId{Id: args[0]},
	})
	if err != nil {
		return fmt.Errorf("failed to delete widget: %w", err)
	}

	if widgetFormat == "json" {
		data, _ := json.MarshalIndent(map[string]string{
			"widget_id": args[0],
			"status":    "deleted",
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Widget %s deleted.\n", args[0])
	}

	return nil
}

func runWidgetGet(cmd *cobra.Command, args []string) error {
	if widgetSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := widgetConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewWidgetServiceClient(conn)

	resp, err := client.GetProperties(context.Background(), &pb.GetPropertiesRequest{
		SessionId: &pb.SessionId{Id: widgetSessionID},
		WidgetId:  &pb.WidgetId{Id: args[0]},
	})
	if err != nil {
		return fmt.Errorf("failed to get widget: %w", err)
	}

	if widgetFormat == "json" {
		data, _ := json.MarshalIndent(resp.Properties, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Widget Properties\n")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf("ID: %s\n", args[0])
		if resp.Properties != nil {
			if resp.Properties.Title != nil {
				fmt.Printf("Title: %s\n", *resp.Properties.Title)
			}
			if resp.Properties.Border != nil {
				fmt.Printf("Border: %v\n", *resp.Properties.Border)
			}
		}
	}

	return nil
}

func runWidgetFocus(cmd *cobra.Command, args []string) error {
	if widgetSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := widgetConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewWidgetServiceClient(conn)

	if len(args) > 0 {
		// Set focus
		_, err = client.SetFocus(context.Background(), &pb.SetFocusRequest{
			SessionId: &pb.SessionId{Id: widgetSessionID},
			WidgetId:  &pb.WidgetId{Id: args[0]},
		})
		if err != nil {
			return fmt.Errorf("failed to set focus: %w", err)
		}
		fmt.Printf("Focus set to widget %s\n", args[0])
	} else {
		// Get focus
		resp, err := client.GetFocus(context.Background(), &pb.GetFocusRequest{
			SessionId: &pb.SessionId{Id: widgetSessionID},
		})
		if err != nil {
			return fmt.Errorf("failed to get focus: %w", err)
		}
		if resp.WidgetId != nil && resp.WidgetId.Id != "" {
			fmt.Printf("Focused widget: %s\n", resp.WidgetId.Id)
		} else {
			fmt.Println("No widget is focused.")
		}
	}

	return nil
}

func runWidgetSetRoot(cmd *cobra.Command, args []string) error {
	if widgetSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := widgetConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewWidgetServiceClient(conn)

	_, err = client.SetRoot(context.Background(), &pb.SetRootRequest{
		SessionId: &pb.SessionId{Id: widgetSessionID},
		WidgetId:  &pb.WidgetId{Id: args[0]},
	})
	if err != nil {
		return fmt.Errorf("failed to set root: %w", err)
	}

	if widgetFormat == "json" {
		data, _ := json.MarshalIndent(map[string]string{
			"widget_id": args[0],
			"status":    "set_as_root",
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Widget %s is now the root display.\n", args[0])
	}

	return nil
}
