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

var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Inject keys and text into a test server",
	Long: `Inject keys and text into a Yutani test server for headless UI testing.

These commands only work when the server is running in test mode.
They inject input directly into the tview event queue via SimulationScreen.`,
}

var injectKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Inject a single keystroke",
	Long: `Inject a single keystroke into the test server.

Available keys: enter, tab, backtab, backspace, escape, up, down, left, right,
home, end, pgup, pgdn, delete, insert, f1-f12, or any single character.`,
	Example: `  # Inject Enter key
  yutani inject key -s <session-id> --key enter

  # Inject Ctrl+C
  yutani inject key -s <session-id> --key c --mod ctrl

  # Inject a letter
  yutani inject key -s <session-id> --key a`,
	RunE: runInjectKey,
}

var injectTextCmd = &cobra.Command{
	Use:   "text [string]",
	Short: "Inject a string as keystrokes",
	Long:  `Inject a string as a sequence of KEY_RUNE events (character by character).`,
	Example: `  # Inject text
  yutani inject text -s <session-id> "Hello World"

  # Inject code
  yutani inject text -s <session-id> "x := 42"`,
	Args: cobra.ExactArgs(1),
	RunE: runInjectText,
}

var injectWaitIdleCmd = &cobra.Command{
	Use:   "wait-idle",
	Short: "Wait until tview has processed all events",
	Long: `Block until tview has processed all pending events.

This is useful after injecting keys/text to ensure the UI has updated
before taking screenshots or checking state.`,
	Example: `  # Wait for idle with default timeout (5s)
  yutani inject wait-idle -s <session-id>

  # Wait with custom timeout
  yutani inject wait-idle -s <session-id> --timeout 10000`,
	RunE: runInjectWaitIdle,
}

var injectMouseCmd = &cobra.Command{
	Use:   "mouse",
	Short: "Inject a mouse event",
	Long: `Inject a mouse event (click or scroll) into the test server.

Available buttons: primary (left), secondary (right), middle,
wheel-up, wheel-down, wheel-left, wheel-right, none.`,
	Example: `  # Click at coordinates
  yutani inject mouse -s <session-id> --x 10 --y 5 --button primary

  # Right-click with modifier
  yutani inject mouse -s <session-id> --x 10 --y 5 --button secondary --mod ctrl

  # Scroll up
  yutani inject mouse -s <session-id> --x 10 --y 5 --button wheel-up`,
	RunE: runInjectMouse,
}

var injectTestModeCmd = &cobra.Command{
	Use:   "test-mode",
	Short: "Check if server is in test mode",
	Long:  `Check if the connected server is running in test mode (supports injection).`,
	Example: `  # Check test mode
  yutani inject test-mode -a localhost:7756`,
	RunE: runInjectTestMode,
}

var (
	injectAddress   string
	injectSessionID string
	injectKey       string
	injectMod       string
	injectTimeout   int64
	injectFormat    string
	injectMouseX    int32
	injectMouseY    int32
	injectButton    string
)

func init() {
	injectCmd.AddCommand(injectKeyCmd)
	injectCmd.AddCommand(injectTextCmd)
	injectCmd.AddCommand(injectMouseCmd)
	injectCmd.AddCommand(injectWaitIdleCmd)
	injectCmd.AddCommand(injectTestModeCmd)

	// Key command flags
	injectKeyCmd.Flags().StringVarP(&injectAddress, "address", "a", "localhost:7755", "Server address")
	injectKeyCmd.Flags().StringVarP(&injectSessionID, "session", "s", "", "Session ID (required)")
	injectKeyCmd.Flags().StringVarP(&injectKey, "key", "k", "", "Key to inject (required)")
	injectKeyCmd.Flags().StringVarP(&injectMod, "mod", "m", "", "Modifier (ctrl, alt, shift, meta)")
	injectKeyCmd.Flags().StringVarP(&injectFormat, "format", "f", "text", "Output format (text, json)")

	// Text command flags
	injectTextCmd.Flags().StringVarP(&injectAddress, "address", "a", "localhost:7755", "Server address")
	injectTextCmd.Flags().StringVarP(&injectSessionID, "session", "s", "", "Session ID (required)")
	injectTextCmd.Flags().StringVarP(&injectFormat, "format", "f", "text", "Output format (text, json)")

	// Mouse command flags
	injectMouseCmd.Flags().StringVarP(&injectAddress, "address", "a", "localhost:7755", "Server address")
	injectMouseCmd.Flags().StringVarP(&injectSessionID, "session", "s", "", "Session ID (required)")
	injectMouseCmd.Flags().Int32VarP(&injectMouseX, "x", "x", 0, "X coordinate (column)")
	injectMouseCmd.Flags().Int32VarP(&injectMouseY, "y", "y", 0, "Y coordinate (row)")
	injectMouseCmd.Flags().StringVarP(&injectButton, "button", "b", "primary", "Button (primary, secondary, middle, wheel-up, wheel-down)")
	injectMouseCmd.Flags().StringVarP(&injectMod, "mod", "m", "", "Modifier (ctrl, alt, shift, meta)")
	injectMouseCmd.Flags().StringVarP(&injectFormat, "format", "f", "text", "Output format (text, json)")

	// Wait-idle command flags
	injectWaitIdleCmd.Flags().StringVarP(&injectAddress, "address", "a", "localhost:7755", "Server address")
	injectWaitIdleCmd.Flags().StringVarP(&injectSessionID, "session", "s", "", "Session ID (required)")
	injectWaitIdleCmd.Flags().Int64VarP(&injectTimeout, "timeout", "t", 5000, "Timeout in milliseconds")
	injectWaitIdleCmd.Flags().StringVarP(&injectFormat, "format", "f", "text", "Output format (text, json)")

	// Test-mode command flags
	injectTestModeCmd.Flags().StringVarP(&injectAddress, "address", "a", "localhost:7755", "Server address")
	injectTestModeCmd.Flags().StringVarP(&injectFormat, "format", "f", "text", "Output format (text, json)")
}

func injectConnect() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return grpc.DialContext(ctx, injectAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

// parseKey converts a key name to a protobuf Key enum
func parseKey(keyName string) (pb.Key, string) {
	keyLower := strings.ToLower(keyName)

	switch keyLower {
	case "enter", "return":
		return pb.Key_KEY_ENTER, ""
	case "tab":
		return pb.Key_KEY_TAB, ""
	case "backtab":
		return pb.Key_KEY_BACKTAB, ""
	case "backspace", "bs":
		return pb.Key_KEY_BACKSPACE, ""
	case "escape", "esc":
		return pb.Key_KEY_ESCAPE, ""
	case "up":
		return pb.Key_KEY_UP, ""
	case "down":
		return pb.Key_KEY_DOWN, ""
	case "left":
		return pb.Key_KEY_LEFT, ""
	case "right":
		return pb.Key_KEY_RIGHT, ""
	case "home":
		return pb.Key_KEY_HOME, ""
	case "end":
		return pb.Key_KEY_END, ""
	case "pgup", "pageup":
		return pb.Key_KEY_PGUP, ""
	case "pgdn", "pagedown":
		return pb.Key_KEY_PGDN, ""
	case "delete", "del":
		return pb.Key_KEY_DELETE, ""
	case "insert", "ins":
		return pb.Key_KEY_INSERT, ""
	case "f1":
		return pb.Key_KEY_F1, ""
	case "f2":
		return pb.Key_KEY_F2, ""
	case "f3":
		return pb.Key_KEY_F3, ""
	case "f4":
		return pb.Key_KEY_F4, ""
	case "f5":
		return pb.Key_KEY_F5, ""
	case "f6":
		return pb.Key_KEY_F6, ""
	case "f7":
		return pb.Key_KEY_F7, ""
	case "f8":
		return pb.Key_KEY_F8, ""
	case "f9":
		return pb.Key_KEY_F9, ""
	case "f10":
		return pb.Key_KEY_F10, ""
	case "f11":
		return pb.Key_KEY_F11, ""
	case "f12":
		return pb.Key_KEY_F12, ""
	default:
		// Single character - treat as KEY_RUNE
		if len(keyName) == 1 {
			return pb.Key_KEY_RUNE, keyName
		}
		// Unknown key
		return pb.Key_KEY_RUNE, keyName
	}
}

// parseModifiers converts a modifier string to protobuf Modifiers
func parseModifiers(modStr string) []pb.Modifier {
	if modStr == "" {
		return nil
	}

	var mods []pb.Modifier
	modLower := strings.ToLower(modStr)

	// Split by comma or plus in case multiple modifiers
	parts := strings.FieldsFunc(modLower, func(c rune) bool {
		return c == ',' || c == '+'
	})

	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch part {
		case "ctrl", "control":
			mods = append(mods, pb.Modifier_MOD_CTRL)
		case "alt":
			mods = append(mods, pb.Modifier_MOD_ALT)
		case "shift":
			mods = append(mods, pb.Modifier_MOD_SHIFT)
		case "meta", "cmd", "command", "super":
			mods = append(mods, pb.Modifier_MOD_META)
		}
	}

	return mods
}

func runInjectKey(cmd *cobra.Command, args []string) error {
	if injectSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}
	if injectKey == "" {
		return fmt.Errorf("key required (--key)")
	}

	conn, err := injectConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewTestServiceClient(conn)

	// Parse key and modifiers
	key, runeStr := parseKey(injectKey)
	mods := parseModifiers(injectMod)

	resp, err := client.InjectKey(context.Background(), &pb.InjectKeyRequest{
		SessionId: &pb.SessionId{Id: injectSessionID},
		Key:       key,
		Rune:      runeStr,
		Modifiers: mods,
	})
	if err != nil {
		return fmt.Errorf("failed to inject key: %w", err)
	}

	if injectFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"success": resp.Success,
			"key":     injectKey,
			"mod":     injectMod,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if resp.Success {
		keyDesc := injectKey
		if injectMod != "" {
			keyDesc = injectMod + "+" + keyDesc
		}
		fmt.Printf("Injected key: %s\n", keyDesc)
	} else {
		fmt.Println("Failed to inject key")
	}

	return nil
}

func runInjectText(cmd *cobra.Command, args []string) error {
	if injectSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	text := args[0]

	conn, err := injectConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewTestServiceClient(conn)

	resp, err := client.InjectText(context.Background(), &pb.InjectTextRequest{
		SessionId: &pb.SessionId{Id: injectSessionID},
		Text:      text,
	})
	if err != nil {
		return fmt.Errorf("failed to inject text: %w", err)
	}

	if injectFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"success":             resp.Success,
			"text":                text,
			"characters_injected": resp.CharactersInjected,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if resp.Success {
		fmt.Printf("Injected %d characters: %q\n", resp.CharactersInjected, text)
	} else {
		fmt.Println("Failed to inject text")
	}

	return nil
}

func runInjectWaitIdle(cmd *cobra.Command, args []string) error {
	if injectSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := injectConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewTestServiceClient(conn)

	resp, err := client.WaitForIdle(context.Background(), &pb.WaitForIdleRequest{
		SessionId: &pb.SessionId{Id: injectSessionID},
		TimeoutMs: injectTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to wait for idle: %w", err)
	}

	if injectFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"success":   resp.Success,
			"timed_out": resp.TimedOut,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if resp.TimedOut {
		fmt.Printf("Timed out waiting for idle (timeout: %dms)\n", injectTimeout)
	} else if resp.Success {
		fmt.Println("Server is idle")
	} else {
		fmt.Println("Wait for idle failed")
	}

	return nil
}

func runInjectTestMode(cmd *cobra.Command, args []string) error {
	conn, err := injectConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewTestServiceClient(conn)

	resp, err := client.IsTestMode(context.Background(), &pb.IsTestModeRequest{})
	if err != nil {
		return fmt.Errorf("failed to check test mode: %w", err)
	}

	if injectFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"test_mode": resp.TestMode,
			"address":   injectAddress,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if resp.TestMode {
		fmt.Printf("Server at %s is in test mode (injection supported)\n", injectAddress)
	} else {
		fmt.Printf("Server at %s is NOT in test mode (injection disabled)\n", injectAddress)
	}

	return nil
}

func runInjectMouse(cmd *cobra.Command, args []string) error {
	if injectSessionID == "" {
		return fmt.Errorf("session ID required (--session)")
	}

	conn, err := injectConnect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewTestServiceClient(conn)

	// Parse button and modifiers
	button := parseButton(injectButton)
	mods := parseModifiers(injectMod)

	resp, err := client.InjectMouse(context.Background(), &pb.InjectMouseRequest{
		SessionId: &pb.SessionId{Id: injectSessionID},
		X:         injectMouseX,
		Y:         injectMouseY,
		Button:    button,
		Modifiers: mods,
	})
	if err != nil {
		return fmt.Errorf("failed to inject mouse: %w", err)
	}

	if injectFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"success": resp.Success,
			"x":       injectMouseX,
			"y":       injectMouseY,
			"button":  injectButton,
			"mod":     injectMod,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if resp.Success {
		buttonDesc := injectButton
		if injectMod != "" {
			buttonDesc = injectMod + "+" + buttonDesc
		}
		fmt.Printf("Injected mouse: %s at (%d, %d)\n", buttonDesc, injectMouseX, injectMouseY)
	} else {
		fmt.Println("Failed to inject mouse event")
	}

	return nil
}

// parseButton converts a button name to a protobuf MouseButton enum
func parseButton(buttonName string) pb.MouseButton {
	buttonLower := strings.ToLower(buttonName)

	switch buttonLower {
	case "primary", "left", "button1":
		return pb.MouseButton_MOUSE_PRIMARY
	case "secondary", "right", "button2":
		return pb.MouseButton_MOUSE_SECONDARY
	case "middle", "button3":
		return pb.MouseButton_MOUSE_MIDDLE
	case "wheel-up", "wheelup", "scrollup":
		return pb.MouseButton_MOUSE_WHEEL_UP
	case "wheel-down", "wheeldown", "scrolldown":
		return pb.MouseButton_MOUSE_WHEEL_DOWN
	case "wheel-left", "wheelleft", "scrollleft":
		return pb.MouseButton_MOUSE_WHEEL_LEFT
	case "wheel-right", "wheelright", "scrollright":
		return pb.MouseButton_MOUSE_WHEEL_RIGHT
	case "none", "":
		return pb.MouseButton_MOUSE_NONE
	default:
		return pb.MouseButton_MOUSE_PRIMARY
	}
}
