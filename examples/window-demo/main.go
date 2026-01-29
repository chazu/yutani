// Package main demonstrates the window management system.
// Creates multiple draggable, resizable windows within a window manager.
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

	// Create a window manager as the root widget
	wm, err := c.NewWindowManager().
		Title("Desktop").
		Build()
	if err != nil {
		log.Fatalf("Failed to create window manager: %v", err)
	}

	// Set window manager as root
	if err := c.SetRoot(wm); err != nil {
		log.Fatalf("Failed to set root: %v", err)
	}

	fmt.Println("Window manager set as root")

	// Create Window 1: a text view with info
	win1, err := c.NewWindow().
		Title("System Info").
		Rect(2, 2, 35, 12).
		Resizable(true).
		Movable(true).
		MinSize(15, 5).
		Build()
	if err != nil {
		log.Fatalf("Failed to create window 1: %v", err)
	}

	// Create Window 2: another text view
	win2, err := c.NewWindow().
		Title("Notes").
		Rect(15, 6, 30, 10).
		Resizable(true).
		Movable(true).
		MinSize(10, 4).
		Build()
	if err != nil {
		log.Fatalf("Failed to create window 2: %v", err)
	}

	// Create Window 3: a smaller window
	win3, err := c.NewWindow().
		Title("Status").
		Rect(40, 3, 25, 8).
		Resizable(true).
		Movable(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create window 3: %v", err)
	}

	// Add windows to the manager
	if err := wm.AddWindow(win1); err != nil {
		log.Fatalf("Failed to add window 1: %v", err)
	}
	if err := wm.AddWindow(win2); err != nil {
		log.Fatalf("Failed to add window 2: %v", err)
	}
	if err := wm.AddWindow(win3); err != nil {
		log.Fatalf("Failed to add window 3: %v", err)
	}

	fmt.Println("Created 3 windows")

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	// Handle keyboard events
	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		switch event.Key.Rune {
		case 'q':
			fmt.Println("Exiting...")
			os.Exit(0)
		case '1':
			wm.BringToFront(win1)
			fmt.Println("Brought Window 1 to front")
		case '2':
			wm.BringToFront(win2)
			fmt.Println("Brought Window 2 to front")
		case '3':
			wm.BringToFront(win3)
			fmt.Println("Brought Window 3 to front")
		case 'm':
			win1.Minimize()
			fmt.Println("Minimized Window 1")
		case 'r':
			win1.Restore()
			fmt.Println("Restored Window 1")
		case 'x':
			win1.Maximize()
			fmt.Println("Maximized Window 1")
		case 'h':
			fmt.Println("\nKeyboard shortcuts:")
			fmt.Println("  1/2/3 - Bring window to front")
			fmt.Println("  m     - Minimize Window 1")
			fmt.Println("  r     - Restore Window 1")
			fmt.Println("  x     - Maximize Window 1")
			fmt.Println("  h     - Show this help")
			fmt.Println("  q     - Quit")
		}
	})

	// Handle window events
	c.OnEventType(client.EventTypeWidget, func(event *client.Event) {
		if event.Widget == nil {
			return
		}
		switch event.Widget.Type {
		case client.WidgetEventWindowMoved:
			fmt.Printf("Window %s moved\n", event.Widget.WidgetID)
		case client.WidgetEventWindowResized:
			fmt.Printf("Window %s resized\n", event.Widget.WidgetID)
		case client.WidgetEventWindowStateChanged:
			fmt.Printf("Window %s state changed\n", event.Widget.WidgetID)
		case client.WidgetEventWindowClosed:
			fmt.Printf("Window %s closed\n", event.Widget.WidgetID)
		case client.WidgetEventWindowActivated:
			fmt.Printf("Window %s activated\n", event.Widget.WidgetID)
		}
	})

	fmt.Println("\nWindow demo running!")
	fmt.Println("Drag title bars to move, drag edges/corners to resize")
	fmt.Println("Click [_] to minimize, [^] to maximize, [X] to close")
	fmt.Println("Press 'h' for keyboard shortcuts, 'q' to quit")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}
