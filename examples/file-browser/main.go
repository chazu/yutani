// Package main demonstrates a file browser using the TreeView widget.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func main() {
	// Connect to the Yutani server
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	fmt.Println("Connected to Yutani server")

	// Create a tree view widget
	tree, err := c.NewTreeView().
		Title("File Browser").
		Border(true).
		NodeTextColor(client.Color("white")).
		SelectedTextColor(client.Color("black")).
		SelectedBackgroundColor(client.Color("cyan")).
		ShowGraphics(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create tree view: %v", err)
	}

	fmt.Printf("Created tree view widget: %s\n", tree.ID())

	// Get starting directory (current directory or home)
	startDir, err := os.Getwd()
	if err != nil {
		startDir = os.Getenv("HOME")
		if startDir == "" {
			startDir = "/"
		}
	}

	// Create root node
	rootNode := client.TreeNodeWithOptions(
		filepath.Base(startDir),
		client.Color("yellow"),
		true,  // selectable
		true,  // expanded
		startDir, // reference (full path)
	)

	rootID, err := tree.SetRoot(rootNode)
	if err != nil {
		log.Fatalf("Failed to set root: %v", err)
	}

	fmt.Printf("Root node created: %s\n", startDir)

	// Load initial directory contents
	if err := loadDirectory(tree, rootID, startDir); err != nil {
		log.Printf("Failed to load directory: %v", err)
	}

	// Set the tree as the root widget to display it on the server
	if err := c.SetRoot(tree); err != nil {
		log.Fatalf("Failed to set root widget: %v", err)
	}

	fmt.Println("File browser displayed on server")

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle events
	c.OnWidgetEvent(tree.ID(), func(event *client.Event) {
		fmt.Printf("Tree event: %s\n", event.Widget.Type)

		// Handle selection changes
		if event.Widget.Type == "WIDGET_SELECTED" {
			nodeID, text, ref, err := tree.GetSelection()
			if err != nil {
				log.Printf("Failed to get selection: %v", err)
				return
			}

			fmt.Printf("Selected: %s (path: %s)\n", text, ref)

			// If it's a directory, load its contents
			if ref != "" {
				info, err := os.Stat(ref)
				if err == nil && info.IsDir() {
					// Check if already loaded (has children)
					children, err := tree.GetChildren(nodeID)
					if err == nil && len(children) == 0 {
						// Not loaded yet, load it
						if err := loadDirectory(tree, nodeID, ref); err != nil {
							log.Printf("Failed to load directory: %v", err)
						} else {
							// Expand the node
							tree.SetExpanded(nodeID, true)
						}
					}
				}
			}
		}
	})

	// Handle keyboard shortcuts
	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		switch event.Key.Rune {
		case 'q':
			fmt.Println("Exiting...")
			os.Exit(0)
		case 'r':
			// Refresh current directory
			fmt.Println("Refreshing...")
		case 'h':
			// Show help
			fmt.Println("\nKeyboard shortcuts:")
			fmt.Println("  Arrow keys - Navigate tree")
			fmt.Println("  Enter      - Expand/collapse directory")
			fmt.Println("  r          - Refresh current directory")
			fmt.Println("  h          - Show this help")
			fmt.Println("  q          - Quit")
		}
	})

	fmt.Println("\nFile Browser created successfully!")
	fmt.Println("Use arrow keys to navigate, Enter to expand/collapse")
	fmt.Println("Press 'h' for help, 'q' to quit")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

// loadDirectory loads the contents of a directory into the tree
func loadDirectory(tree *client.TreeView, parentID *pb.TreeNodeId, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Sort: directories first, then files
	var dirs, files []os.DirEntry
	for _, entry := range entries {
		// Skip hidden files
		if entry.Name()[0] == '.' {
			continue
		}

		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	// Add directories
	for _, dir := range dirs {
		fullPath := filepath.Join(dirPath, dir.Name())
		node := client.TreeNodeWithOptions(
			"📁 "+dir.Name(),
			client.Color("cyan"),
			true,  // selectable
			false, // not expanded initially
			fullPath, // reference
		)

		_, err := tree.AddChild(parentID, node)
		if err != nil {
			log.Printf("Failed to add directory node: %v", err)
		}
	}

	// Add files
	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())

		// Get file info for size
		info, err := file.Info()
		var displayName string
		if err == nil {
			size := formatSize(info.Size())
			displayName = fmt.Sprintf("📄 %s (%s)", file.Name(), size)
		} else {
			displayName = "📄 " + file.Name()
		}

		node := client.TreeNodeWithOptions(
			displayName,
			client.Color("white"),
			true,  // selectable
			false, // files can't be expanded
			fullPath, // reference
		)

		_, err = tree.AddChild(parentID, node)
		if err != nil {
			log.Printf("Failed to add file node: %v", err)
		}
	}

	return nil
}

// formatSize formats a file size in human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

