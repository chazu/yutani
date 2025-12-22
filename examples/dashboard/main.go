// Package main demonstrates a system dashboard using Grid layout.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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

	// Create a grid layout
	grid, err := c.NewGrid().
		Title("System Dashboard").
		Border(true).
		Rows(2).
		Columns(2).
		Build()
	if err != nil {
		log.Fatalf("Failed to create grid: %v", err)
	}

	fmt.Printf("Created grid widget: %s\n", grid.ID())

	// Create CPU widget (top-left)
	cpuWidget, err := c.NewTextView().
		Title("CPU Usage").
		Border(true).
		WordWrap(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create CPU widget: %v", err)
	}

	// Create Memory widget (top-right)
	memWidget, err := c.NewTextView().
		Title("Memory Usage").
		Border(true).
		WordWrap(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create memory widget: %v", err)
	}

	// Create Goroutines widget (bottom-left)
	goroutineWidget, err := c.NewTextView().
		Title("Goroutines").
		Border(true).
		WordWrap(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create goroutine widget: %v", err)
	}

	// Create System Info widget (bottom-right)
	sysInfoWidget, err := c.NewTextView().
		Title("System Info").
		Border(true).
		WordWrap(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create system info widget: %v", err)
	}

	// Add widgets to grid
	// AddItem(widget, row, column, rowSpan, columnSpan, minWidth, minHeight, focus)
	if err := grid.AddItem(cpuWidget, 0, 0, 1, 1, 0, 0, false); err != nil {
		log.Fatalf("Failed to add CPU widget: %v", err)
	}

	if err := grid.AddItem(memWidget, 0, 1, 1, 1, 0, 0, false); err != nil {
		log.Fatalf("Failed to add memory widget: %v", err)
	}

	if err := grid.AddItem(goroutineWidget, 1, 0, 1, 1, 0, 0, false); err != nil {
		log.Fatalf("Failed to add goroutine widget: %v", err)
	}

	if err := grid.AddItem(sysInfoWidget, 1, 1, 1, 1, 0, 0, false); err != nil {
		log.Fatalf("Failed to add system info widget: %v", err)
	}

	fmt.Println("All widgets added to grid")

	// Set the grid as the root widget to display it on the server
	if err := c.SetRoot(grid); err != nil {
		log.Fatalf("Failed to set root widget: %v", err)
	}

	fmt.Println("Dashboard displayed on server")

	// Set static system info
	updateSystemInfo(sysInfoWidget)

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle keyboard shortcuts
	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		switch event.Key.Rune {
		case 'q':
			fmt.Println("Exiting...")
			os.Exit(0)
		case 'r':
			// Force refresh
			updateCPU(cpuWidget)
			updateMemory(memWidget)
			updateGoroutines(goroutineWidget)
			fmt.Println("Dashboard refreshed")
		case 'h':
			fmt.Println("\nKeyboard shortcuts:")
			fmt.Println("  r - Refresh dashboard")
			fmt.Println("  h - Show this help")
			fmt.Println("  q - Quit")
		}
	})

	// Start update ticker
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			updateCPU(cpuWidget)
			updateMemory(memWidget)
			updateGoroutines(goroutineWidget)
		}
	}()

	fmt.Println("\nDashboard created successfully!")
	fmt.Println("Stats update every 2 seconds")
	fmt.Println("Press 'r' to refresh, 'h' for help, 'q' to quit")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

// updateCPU updates the CPU usage widget
func updateCPU(widget *client.TextView) {
	numCPU := runtime.NumCPU()

	text := fmt.Sprintf("Number of CPUs: %d\n\n", numCPU)
	text += "Go Runtime Stats:\n"
	text += fmt.Sprintf("  GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	// Get memory stats for CPU-related info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	text += fmt.Sprintf("  GC Runs: %d\n", m.NumGC)
	text += fmt.Sprintf("  Last GC: %v ago\n", time.Since(time.Unix(0, int64(m.LastGC))))

	widget.SetText(text)
}

// updateMemory updates the memory usage widget
func updateMemory(widget *client.TextView) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	text := "Memory Statistics:\n\n"
	text += fmt.Sprintf("Allocated: %s\n", formatBytes(m.Alloc))
	text += fmt.Sprintf("Total Allocated: %s\n", formatBytes(m.TotalAlloc))
	text += fmt.Sprintf("System: %s\n", formatBytes(m.Sys))
	text += fmt.Sprintf("Heap Allocated: %s\n", formatBytes(m.HeapAlloc))
	text += fmt.Sprintf("Heap System: %s\n", formatBytes(m.HeapSys))
	text += fmt.Sprintf("Heap Objects: %d\n", m.HeapObjects)

	widget.SetText(text)
}

// updateGoroutines updates the goroutines widget
func updateGoroutines(widget *client.TextView) {
	numGoroutines := runtime.NumGoroutine()

	text := fmt.Sprintf("Active Goroutines: %d\n\n", numGoroutines)
	text += "Goroutine Info:\n"

	// Add some visual representation
	bars := numGoroutines / 5
	if bars > 20 {
		bars = 20
	}
	text += "Activity: "
	for i := 0; i < bars; i++ {
		text += "█"
	}
	text += "\n\n"

	text += "Status: "
	if numGoroutines < 10 {
		text += "Low activity"
	} else if numGoroutines < 50 {
		text += "Normal activity"
	} else {
		text += "High activity"
	}

	widget.SetText(text)
}

// updateSystemInfo updates the system info widget
func updateSystemInfo(widget *client.TextView) {
	text := "System Information:\n\n"
	text += fmt.Sprintf("OS: %s\n", runtime.GOOS)
	text += fmt.Sprintf("Architecture: %s\n", runtime.GOARCH)
	text += fmt.Sprintf("Go Version: %s\n", runtime.Version())
	text += fmt.Sprintf("Compiler: %s\n\n", runtime.Compiler)

	// Get hostname
	hostname, err := os.Hostname()
	if err == nil {
		text += fmt.Sprintf("Hostname: %s\n", hostname)
	}

	// Get working directory
	wd, err := os.Getwd()
	if err == nil {
		text += fmt.Sprintf("Working Dir: %s\n", wd)
	}

	widget.SetText(text)
}

// formatBytes formats bytes in human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

