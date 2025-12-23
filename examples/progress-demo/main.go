// Progress Bar Demo - Shows various progress bar configurations
//
// This example demonstrates:
// - Basic progress bar with percentage
// - Labeled progress bars
// - Custom colors and characters
// - Multiple progress bars with different styles
//
// Run with: go run ./examples/progress-demo
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

	// Create a flex container
	flex, err := c.NewFlex().
		Direction(pb.FlexDirection_FLEX_COLUMN).
		Build()
	if err != nil {
		log.Fatalf("Failed to create flex: %v", err)
	}

	// Create header
	header, err := c.NewTextView().
		Title("Progress Bar Demo").
		Border(true).
		Text("[::b]Progress Bar Styles[::-]\n\nDifferent progress bar configurations:").
		DynamicColors(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create header: %v", err)
	}

	// Create various progress bars with different styles

	// 1. Default style
	defaultBar, err := c.NewProgressBar().
		Title("Default Style").
		Border(true).
		Progress(0).
		Label("Download").
		ShowPercentage(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create default bar: %v", err)
	}

	// 2. Custom colors
	colorBar, err := c.NewProgressBar().
		Title("Custom Colors").
		Border(true).
		Progress(0).
		Label("Upload").
		ShowPercentage(true).
		FilledColor(client.NamedColor("cyan")).
		EmptyColor(client.NamedColor("darkgray")).
		Build()
	if err != nil {
		log.Fatalf("Failed to create color bar: %v", err)
	}

	// 3. Custom characters
	charBar, err := c.NewProgressBar().
		Title("Custom Characters").
		Border(true).
		Progress(0).
		Label("Processing").
		ShowPercentage(true).
		FilledChar("=").
		EmptyChar("-").
		FilledColor(client.NamedColor("yellow")).
		Build()
	if err != nil {
		log.Fatalf("Failed to create char bar: %v", err)
	}

	// 4. No percentage
	simpleBar, err := c.NewProgressBar().
		Title("Simple (no percentage)").
		Border(true).
		Progress(0).
		Label("Installing").
		ShowPercentage(false).
		FilledColor(client.NamedColor("green")).
		Build()
	if err != nil {
		log.Fatalf("Failed to create simple bar: %v", err)
	}

	// 5. RGB colors
	rgbBar, err := c.NewProgressBar().
		Title("RGB Colors").
		Border(true).
		Progress(0).
		Label("Compiling").
		ShowPercentage(true).
		FilledColor(client.RGBColor(255, 100, 50)).
		EmptyColor(client.RGBColor(50, 50, 50)).
		Build()
	if err != nil {
		log.Fatalf("Failed to create RGB bar: %v", err)
	}

	// Add all progress bars to the flex
	if err := flex.AddItem(header, 5, 0, false); err != nil {
		log.Fatalf("Failed to add header: %v", err)
	}
	if err := flex.AddItem(defaultBar, 3, 0, false); err != nil {
		log.Fatalf("Failed to add default bar: %v", err)
	}
	if err := flex.AddItem(colorBar, 3, 0, false); err != nil {
		log.Fatalf("Failed to add color bar: %v", err)
	}
	if err := flex.AddItem(charBar, 3, 0, false); err != nil {
		log.Fatalf("Failed to add char bar: %v", err)
	}
	if err := flex.AddItem(simpleBar, 3, 0, false); err != nil {
		log.Fatalf("Failed to add simple bar: %v", err)
	}
	if err := flex.AddItem(rgbBar, 3, 0, false); err != nil {
		log.Fatalf("Failed to add RGB bar: %v", err)
	}

	// Set as root
	if err := c.SetRoot(flex); err != nil {
		log.Fatalf("Failed to set root: %v", err)
	}

	// Animate all progress bars at different speeds
	bars := []*client.ProgressBar{defaultBar, colorBar, charBar, simpleBar, rgbBar}
	speeds := []int{1, 2, 3, 2, 1}
	progress := []int{0, 20, 40, 60, 80}

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			for i, bar := range bars {
				progress[i] = (progress[i] + speeds[i]) % 101
				if err := bar.SetProgress(progress[i]); err != nil {
					log.Printf("Failed to update progress: %v", err)
				}
			}
		}
	}()

	fmt.Println("Progress Bar Demo running. Press Ctrl+C to exit.")
	fmt.Println("Watch the different progress bar styles animate!")

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nShutting down...")
}
