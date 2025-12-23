// Widgets Demo - Demonstrates TextArea, Modal, ProgressBar, and Theme system
//
// This example showcases the new widgets added to Yutani:
// - TextArea: Multi-line text editor
// - Modal: Dialog boxes with buttons
// - ProgressBar: Visual progress indicator
// - Theme: Consistent color schemes
//
// Run with: go run ./examples/widgets-demo
// (Requires yutani-server running on localhost:7755)

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

func main() {
	// Connect to the Yutani server
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Use the Ocean theme for consistent styling
	theme := client.OceanTheme

	// Create a flex container for the layout
	flex, err := c.NewFlex().
		Direction(pb.FlexDirection_FLEX_COLUMN).
		Build()
	if err != nil {
		log.Fatalf("Failed to create flex: %v", err)
	}

	// Create header
	header, err := c.NewTextView().
		Title("Yutani Widget Demo").
		Border(true).
		Text("[::b]New Widgets Showcase[::-]\n\nThis demo shows TextArea, ProgressBar, and Modal widgets with theming.").
		DynamicColors(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create header: %v", err)
	}

	// Create a TextArea for multi-line editing
	textArea, err := c.NewTextArea().
		Title("TextArea - Multi-line Editor").
		Border(true).
		Placeholder("Type your notes here...").
		WordWrap(true).
		TextColor(theme.TextColor).
		Build()
	if err != nil {
		log.Fatalf("Failed to create text area: %v", err)
	}

	// Create a ProgressBar
	progressBar, err := c.NewProgressBar().
		Title("Progress").
		Border(true).
		Progress(0).
		Max(100).
		Label("Loading").
		ShowPercentage(true).
		FilledColor(theme.SuccessColor).
		EmptyColor(theme.DividerColor).
		Build()
	if err != nil {
		log.Fatalf("Failed to create progress bar: %v", err)
	}

	// Create a button to show modal
	showModalBtn, err := c.NewButton().
		Label("Show Modal Dialog").
		Build()
	if err != nil {
		log.Fatalf("Failed to create button: %v", err)
	}

	// Create inner flex for bottom row
	bottomFlex, err := c.NewFlex().
		Direction(pb.FlexDirection_FLEX_ROW).
		Build()
	if err != nil {
		log.Fatalf("Failed to create bottom flex: %v", err)
	}

	// Add widgets to bottom flex
	if err := bottomFlex.AddItem(progressBar, 0, 1, false); err != nil {
		log.Fatalf("Failed to add progress bar: %v", err)
	}
	if err := bottomFlex.AddItem(showModalBtn, 20, 0, true); err != nil {
		log.Fatalf("Failed to add button: %v", err)
	}

	// Add widgets to main flex
	if err := flex.AddItem(header, 5, 0, false); err != nil {
		log.Fatalf("Failed to add header: %v", err)
	}
	if err := flex.AddItem(textArea, 0, 1, true); err != nil {
		log.Fatalf("Failed to add text area: %v", err)
	}
	if err := flex.AddItem(bottomFlex, 3, 0, false); err != nil {
		log.Fatalf("Failed to add bottom flex: %v", err)
	}

	// Set as root
	if err := c.SetRoot(flex); err != nil {
		log.Fatalf("Failed to set root: %v", err)
	}

	// Handle button click to show modal
	c.OnWidgetEvent(showModalBtn.ID(), func(e *client.Event) {
		// Create and show modal
		modal, err := c.NewModal().
			Title("Confirmation").
			Text("Do you want to save your changes?").
			Buttons("Save", "Discard", "Cancel").
			TextColor(theme.TextColor).
			BackgroundColor(theme.SurfaceColor).
			ButtonBackgroundColor(theme.ButtonColor).
			ButtonTextColor(theme.ButtonTextColor).
			Build()
		if err != nil {
			log.Printf("Failed to create modal: %v", err)
			return
		}

		// Handle modal button clicks
		c.OnWidgetEvent(modal.ID(), func(me *client.Event) {
			if me.Widget != nil {
				buttonLabel := me.Widget.Data["button_label"]
				log.Printf("Modal button clicked: %s", buttonLabel)
			}

			// Delete modal after button click
			modal.Delete()

			// Restore focus to main flex
			c.SetRoot(flex)
		})

		// Show modal
		c.SetRoot(modal)
	})

	// Animate progress bar in background
	go func() {
		progress := 0
		for {
			time.Sleep(100 * time.Millisecond)
			progress = (progress + 1) % 101
			if err := progressBar.SetProgress(progress); err != nil {
				log.Printf("Failed to update progress: %v", err)
			}

			// Update label based on progress
			var label string
			switch {
			case progress < 30:
				label = "Starting"
			case progress < 60:
				label = "Processing"
			case progress < 90:
				label = "Finishing"
			default:
				label = "Complete"
			}
			if err := progressBar.SetLabel(label); err != nil {
				log.Printf("Failed to update label: %v", err)
			}
		}
	}()

	fmt.Println("Widgets Demo running. Press Ctrl+C to exit.")
	fmt.Println("Try:")
	fmt.Println("  - Type in the TextArea")
	fmt.Println("  - Watch the ProgressBar animate")
	fmt.Println("  - Click 'Show Modal Dialog' button")

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nShutting down...")
}
