// Package main demonstrates a login form using the Yutani client library.
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

	// Create a form widget
	form, err := c.NewForm().
		Title("Login").
		Border(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create form: %v", err)
	}

	fmt.Printf("Created form widget: %s\n", form.ID())

	// Add username field
	usernameIdx, err := form.AddInputField("Username", 30, "")
	if err != nil {
		log.Fatalf("Failed to add username field: %v", err)
	}

	// Add password field
	passwordIdx, err := form.AddPasswordField("Password", 30)
	if err != nil {
		log.Fatalf("Failed to add password field: %v", err)
	}

	// Add remember me checkbox
	rememberIdx, err := form.AddCheckbox("Remember me", false)
	if err != nil {
		log.Fatalf("Failed to add checkbox: %v", err)
	}

	// Add role dropdown
	roleIdx, err := form.AddDropdown("Role", []string{"User", "Admin", "Guest"}, "User")
	if err != nil {
		log.Fatalf("Failed to add dropdown: %v", err)
	}

	// Add login button
	if err := form.AddButton("Login"); err != nil {
		log.Fatalf("Failed to add button: %v", err)
	}

	// Add cancel button
	if err := form.AddButton("Cancel"); err != nil {
		log.Fatalf("Failed to add button: %v", err)
	}

	fmt.Println("Form fields added successfully")

	// Set the form as the root widget to display it on the server
	if err := c.SetRoot(form); err != nil {
		log.Fatalf("Failed to set root widget: %v", err)
	}

	fmt.Println("Form displayed on server")

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle events
	c.OnEvent(func(event *client.Event) {
		if event.IsWidget() && event.Widget.WidgetID == form.ID() {
			fmt.Printf("Form event: %s\n", event.Widget.Type)

			// Handle form submission
			if event.Widget.Type == "SUBMITTED" {
				fmt.Println("\n=== Form Submitted ===")

				// Get all field values
				if username, err := form.GetFieldValue(usernameIdx); err == nil {
					fmt.Printf("Username: %s\n", username)
				}

				if password, err := form.GetFieldValue(passwordIdx); err == nil {
					fmt.Printf("Password: %s\n", password)
				}

				if remember, err := form.GetFieldValue(rememberIdx); err == nil {
					fmt.Printf("Remember me: %s\n", remember)
				}

				if role, err := form.GetFieldValue(roleIdx); err == nil {
					fmt.Printf("Role: %s\n", role)
				}

				fmt.Println("=====================\n")
			}
		}

		if event.IsKey() {
			// Exit on Ctrl+C or 'q'
			if event.Key.Rune == 'q' || event.Key.Key == "Ctrl+C" {
				fmt.Println("Exiting...")
				os.Exit(0)
			}

			// Clear form on 'c'
			if event.Key.Rune == 'c' {
				if err := form.Clear(); err == nil {
					fmt.Println("Form cleared")
				}
			}

			// Show field count on 'i'
			if event.Key.Rune == 'i' {
				if count, err := form.GetItemCount(); err == nil {
					fmt.Printf("Form has %d fields\n", count)
				}
			}
		}
	})

	fmt.Println("\nLogin form created successfully!")
	fmt.Println("Fill in the form and press Enter to submit")
	fmt.Println("Press 'c' to clear, 'i' for info, 'q' to quit")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

