// Package main demonstrates a text editor using TextView with syntax highlighting.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

var (
	currentFile    = ""
	fileContent    = ""
	modified       = false
	cursorLine     = 0
	cursorCol      = 0
	syntaxLanguage = "go"
)

func main() {
	// Connect to the Yutani server
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	fmt.Println("Connected to Yutani server")

	// Create main flex layout (vertical)
	mainFlex, err := c.NewFlex().
		Title("Text Editor").
		Border(true).
		Direction(pb.FlexDirection_FLEX_COLUMN).
		Build()
	if err != nil {
		log.Fatalf("Failed to create main flex: %v", err)
	}

	// Create menu bar
	menuBar, err := c.NewTextView().
		Text("Press: h=Help | q=Quit | Ctrl+C=Exit").
		Build()
	if err != nil {
		log.Fatalf("Failed to create menu bar: %v", err)
	}

	// Create editor area
	editor, err := c.NewTextView().
		Title("Untitled").
		Border(true).
		WordWrap(false).
		DynamicColors(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create editor: %v", err)
	}

	// Create status bar
	statusBar, err := c.NewTextView().
		Text(getStatusText()).
		Build()
	if err != nil {
		log.Fatalf("Failed to create status bar: %v", err)
	}

	// Add widgets to main flex
	mainFlex.AddItem(menuBar, 0, 1, false)    // Menu bar (fixed 1 line)
	mainFlex.AddItem(editor, 1, 0, true)      // Editor (proportional)
	mainFlex.AddItem(statusBar, 0, 1, false)  // Status bar (fixed 1 line)

	// Set the main flex as the root widget to display it on the server
	if err := c.SetRoot(mainFlex); err != nil {
		log.Fatalf("Failed to set root widget: %v", err)
	}

	fmt.Println("Text editor UI created and displayed on server")

	// Load initial content
	if len(os.Args) > 1 {
		filename := os.Args[1]
		if err := loadFile(filename, editor, statusBar); err != nil {
			log.Printf("Failed to load file: %v", err)
		}
	} else {
		// Show welcome message
		welcomeText := `Welcome to Yutani Text Editor Demo!

This is a demonstration of the TextView widget with a Flex layout.

Keyboard Shortcuts:
  h - Show this help
  q - Quit
  Ctrl+C - Exit

This demo shows how to:
- Create a multi-panel layout with Flex
- Display text in a TextView widget
- Handle keyboard events
- Set a root widget to display on the server

The text editor is read-only in this demo.
Press 'h' for help or 'q' to quit.
`
		fileContent = welcomeText
		editor.SetText(applySyntaxHighlighting(fileContent, "text"))
	}

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle keyboard shortcuts
	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		// Log the key event for debugging
		fmt.Printf("Key event: Key=%s, Rune=%c\n", event.Key.Key, event.Key.Rune)

		// Check for 'q' to quit (simple and reliable)
		if event.Key.Rune == 'q' {
			fmt.Println("Exiting...")
			os.Exit(0)
		}

		// Check for Ctrl+C
		if event.Key.Key == "Ctrl+C" {
			fmt.Println("Exiting...")
			os.Exit(0)
		}

		// Check for other shortcuts
		switch event.Key.Rune {
		case 'h':
			// Show help
			showHelp(editor)
		case 'o':
			// Open file (simplified - just show message)
			fmt.Println("Open file: Not implemented in this demo")
		case 's':
			// Save file (simplified - just show message)
			fmt.Println("Save file: Not implemented in this demo")
		case 'f':
			// Find text (simplified - just show message)
			fmt.Println("Find text: Not implemented in this demo")
		}
	})

	fmt.Println("\nText Editor created successfully!")
	fmt.Println("Look at the server terminal to see the UI")
	fmt.Println("Press 'h' for help, 'q' to quit, or Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

// loadFile loads a file into the editor
func loadFile(filename string, editor *client.TextView, statusBar *client.TextView) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	currentFile = filename
	fileContent = string(content)
	modified = false

	// Detect language from extension
	ext := filepath.Ext(filename)
	syntaxLanguage = detectLanguage(ext)

	// Apply syntax highlighting
	highlighted := applySyntaxHighlighting(fileContent, syntaxLanguage)
	editor.SetText(highlighted)
	editor.SetTitle(filepath.Base(filename))
	statusBar.SetText(getStatusText())

	fmt.Printf("Loaded file: %s (%d bytes)\n", filename, len(content))
	return nil
}

// saveFile saves the editor content to a file
func saveFile(filename string, editor *client.TextView, statusBar *client.TextView) error {
	if filename == "" {
		return fmt.Errorf("no filename specified")
	}

	err := os.WriteFile(filename, []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	currentFile = filename
	modified = false
	statusBar.SetText(getStatusText())

	fmt.Printf("Saved file: %s (%d bytes)\n", filename, len(fileContent))
	return nil
}

// getStatusText returns the status bar text
func getStatusText() string {
	modifiedStr := ""
	if modified {
		modifiedStr = " [Modified]"
	}

	filename := currentFile
	if filename == "" {
		filename = "Untitled"
	}

	return fmt.Sprintf("File: %s%s | Language: %s | Lines: %d | Ln %d, Col %d",
		filename, modifiedStr, syntaxLanguage, strings.Count(fileContent, "\n")+1, cursorLine+1, cursorCol+1)
}

// detectLanguage detects the programming language from file extension
func detectLanguage(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".ts":
		return "javascript"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc":
		return "cpp"
	case ".rs":
		return "rust"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "text"
	}
}

// applySyntaxHighlighting applies basic syntax highlighting
func applySyntaxHighlighting(content, language string) string {
	// For now, just return the content as-is
	// In a real implementation, you would parse the code and add color tags
	// Example: [blue]keyword[white] text

	if language == "go" {
		// Simple Go keyword highlighting
		keywords := []string{"package", "import", "func", "var", "const", "type", "struct", "interface", "return", "if", "else", "for", "range", "switch", "case", "default"}
		highlighted := content
		for _, kw := range keywords {
			// This is a simplified version - real syntax highlighting is more complex
			highlighted = strings.ReplaceAll(highlighted, " "+kw+" ", " [blue]"+kw+"[white] ")
			highlighted = strings.ReplaceAll(highlighted, "\n"+kw+" ", "\n[blue]"+kw+"[white] ")
		}
		return highlighted
	}

	return content
}

// highlightSearch highlights search results in the editor
func highlightSearch(editor *client.TextView, searchText string) {
	if searchText == "" {
		return
	}

	// Count occurrences
	count := strings.Count(fileContent, searchText)
	fmt.Printf("Found %d occurrences of '%s'\n", count, searchText)

	// Highlight all occurrences
	highlighted := strings.ReplaceAll(fileContent, searchText, "[yellow]"+searchText+"[white]")
	editor.SetText(highlighted)
}

// showHelp shows the help text
func showHelp(editor *client.TextView) {
	helpText := `Yutani Text Editor Demo - Help

KEYBOARD SHORTCUTS:
  h - Show this help
  q - Quit the application
  Ctrl+C - Exit

ABOUT THIS DEMO:
This example demonstrates:
  - Flex layout widget (vertical column layout)
  - TextView widgets (menu bar, editor, status bar)
  - Keyboard event handling
  - Setting a root widget to display on the server

ARCHITECTURE:
  - Main Flex container (vertical)
    - Menu bar (TextView, fixed 1 line)
    - Editor area (TextView, proportional)
    - Status bar (TextView, fixed 1 line)

This is a simplified demo. A full text editor would include:
  - Actual text editing capabilities
  - File I/O operations
  - Syntax highlighting
  - Find and replace
  - Multiple file support

Press 'q' to quit...
`
	editor.SetText(helpText)
	fmt.Println("Help displayed on server")
}

