// Package main demonstrates a chat application using Pages, InputField, and List widgets.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chazu/yutani/pkg/client"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

type Message struct {
	User      string
	Text      string
	Timestamp time.Time
}

var (
	messages    []Message
	currentUser = "User1"
	currentRoom = "general"
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
		Title("Chat Application").
		Border(true).
		Direction(pb.FlexDirection_FLEX_COLUMN).
		Build()
	if err != nil {
		log.Fatalf("Failed to create main flex: %v", err)
	}

	// Create pages for different chat rooms
	pages, err := c.NewPages().
		ShowPageNames(true).
		PageNameColor(client.Color("cyan")).
		Build()
	if err != nil {
		log.Fatalf("Failed to create pages: %v", err)
	}

	// Create chat rooms
	generalRoom := createChatRoom(c, "general")
	randomRoom := createChatRoom(c, "random")
	helpRoom := createChatRoom(c, "help")

	// Add pages
	pages.AddPage("general", generalRoom, true, true)
	pages.AddPage("random", randomRoom, true, false)
	pages.AddPage("help", helpRoom, true, false)

	// Create input field for messages
	inputField, err := c.NewInputField().
		Label("Message: ").
		Placeholder("Type your message...").
		FieldWidth(60).
		Build()
	if err != nil {
		log.Fatalf("Failed to create input field: %v", err)
	}

	// Create status bar
	statusBar, err := c.NewTextView().
		Text(fmt.Sprintf("User: %s | Room: %s | Press Tab to switch rooms, Ctrl+U to change user, q to quit", currentUser, currentRoom)).
		Build()
	if err != nil {
		log.Fatalf("Failed to create status bar: %v", err)
	}

	// Add widgets to main flex
	mainFlex.AddItem(pages, 1, 0, true)      // Chat area (proportional)
	mainFlex.AddItem(inputField, 0, 3, false) // Input field (fixed 3 lines)
	mainFlex.AddItem(statusBar, 0, 1, false)  // Status bar (fixed 1 line)

	// Set the main flex as the root widget to display it on the server
	if err := c.SetRoot(mainFlex); err != nil {
		log.Fatalf("Failed to set root widget: %v", err)
	}

	fmt.Println("Chat UI created and displayed on server")

	// Add some initial messages
	addMessage("System", "Welcome to the chat!", "general")
	addMessage("System", "Type your message and press Enter to send", "general")
	updateChatRoom(generalRoom, "general")

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle input field events
	c.OnWidgetEvent(inputField.ID(), func(event *client.Event) {
		if event.Widget.Type == "WIDGET_CHANGED" {
			// Message submitted (Enter pressed)
			// In a real app, we'd get the text from the input field
			// For now, we'll simulate it
			messageText := "Hello from " + currentUser
			addMessage(currentUser, messageText, currentRoom)
			
			// Update the current room
			switch currentRoom {
			case "general":
				updateChatRoom(generalRoom, "general")
			case "random":
				updateChatRoom(randomRoom, "random")
			case "help":
				updateChatRoom(helpRoom, "help")
			}
			
			// Clear input field
			inputField.SetText("")
		}
	})

	// Handle keyboard shortcuts
	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		switch event.Key.Key {
		case "KEY_TAB":
			// Switch rooms
			rooms := []string{"general", "random", "help"}
			for i, room := range rooms {
				if room == currentRoom {
					currentRoom = rooms[(i+1)%len(rooms)]
					pages.ShowPage(currentRoom)
					statusBar.SetText(fmt.Sprintf("User: %s | Room: %s | Press Tab to switch rooms, Ctrl+U to change user, q to quit", currentUser, currentRoom))
					fmt.Printf("Switched to room: %s\n", currentRoom)
					break
				}
			}
		}

		switch event.Key.Rune {
		case 'q':
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	})

	fmt.Println("\nChat Application created successfully!")
	fmt.Println("Type messages and press Enter to send")
	fmt.Println("Press Tab to switch rooms, 'q' to quit")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

// createChatRoom creates a list widget for a chat room
func createChatRoom(c *client.Client, roomName string) *client.List {
	list, err := c.NewList().
		Title(fmt.Sprintf("#%s", roomName)).
		Border(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create chat room list: %v", err)
	}
	return list
}

// addMessage adds a message to the message history
func addMessage(user, text, room string) {
	msg := Message{
		User:      user,
		Text:      text,
		Timestamp: time.Now(),
	}
	messages = append(messages, msg)

	// Keep only last 100 messages
	if len(messages) > 100 {
		messages = messages[len(messages)-100:]
	}
}

// updateChatRoom updates a chat room list with messages
func updateChatRoom(list *client.List, room string) {
	// Clear existing items
	list.Clear()

	// Add messages for this room
	for _, msg := range messages {
		// For simplicity, show all messages in all rooms
		// In a real app, you'd filter by room
		timestamp := msg.Timestamp.Format("15:04:05")
		text := fmt.Sprintf("[%s] %s: %s", timestamp, msg.User, msg.Text)

		list.AddItem(text, "", nil)
	}
}

