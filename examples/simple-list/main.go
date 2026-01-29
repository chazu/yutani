// Package main demonstrates a simple list widget using the Yutani client library.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
)

func main() {
	// Connect to the Yutani server
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	fmt.Println("Connected to Yutani server")

	// Create a list widget
	list, err := c.NewList().
		Title("File Menu").
		Border(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create list: %v", err)
	}

	fmt.Printf("Created list widget: %s\n", list.ID())

	// Add menu items
	items := []struct {
		main      string
		secondary string
		shortcut  string
	}{
		{"New File", "Create a new file", "n"},
		{"Open File", "Open an existing file", "o"},
		{"Save File", "Save the current file", "s"},
		{"Save As", "Save with a new name", "a"},
		{"Close File", "Close the current file", "c"},
		{"Exit", "Exit the application", "q"},
	}

	for _, item := range items {
		shortcut := item.shortcut
		_, err := list.AddItem(item.main, item.secondary, &shortcut)
		if err != nil {
			log.Printf("Failed to add item: %v", err)
		}
	}

	fmt.Printf("Added %d items to list\n", len(items))

	// Set initial selection
	if err := list.SetSelection(0); err != nil {
		log.Printf("Failed to set selection: %v", err)
	}

	// Set the list as the root widget to display it on the server
	if err := c.SetRoot(list); err != nil {
		log.Fatalf("Failed to set root widget: %v", err)
	}

	fmt.Println("List displayed on server")

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle events
	c.OnEvent(func(event *client.Event) {
		if event.IsWidget() && event.Widget.WidgetID == list.ID() {
			fmt.Printf("Widget event: %s (type: %s)\n", event.Widget.WidgetID, event.Widget.Type)

			// Get current selection
			if selected, err := list.GetSelection(); err == nil {
				if main, secondary, shortcut, err := list.GetItem(selected); err == nil {
					fmt.Printf("  Selected item %d: %s (%s) [%s]\n", selected, main, secondary, shortcut)
				}
			}
		}

		if event.IsKey() {
			fmt.Printf("Key event: %s (rune: %c)\n", event.Key.Key, event.Key.Rune)

			// Exit on 'q'
			if event.Key.Rune == 'q' {
				fmt.Println("Exiting...")
				os.Exit(0)
			}
		}
	})

	fmt.Println("\nList widget created successfully!")
	fmt.Println("Use arrow keys to navigate, Enter to select, 'q' to quit")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

