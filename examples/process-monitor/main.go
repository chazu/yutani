// Package main demonstrates a process monitor using the Table widget.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/chazu/yutani/pkg/client"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

type ProcessInfo struct {
	PID     string
	Name    string
	CPU     string
	Memory  string
	Status  string
}

func main() {
	// Connect to the Yutani server
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	fmt.Println("Connected to Yutani server")

	// Create a table widget
	table, err := c.NewTable().
		Title("Process Monitor").
		Border(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Printf("Created table widget: %s\n", table.ID())

	// Set header row
	headers := []string{"PID", "Name", "CPU %", "Memory", "Status"}
	for i, header := range headers {
		err := table.SetCell(0, i, &pb.TableCell{
			Text:  header,
			Color: client.Color("yellow"),
		})
		if err != nil {
			log.Printf("Failed to set header cell: %v", err)
		}
	}

	// Set fixed header row
	if err := table.SetFixed(1, 0); err != nil {
		log.Printf("Failed to set fixed header: %v", err)
	}

	fmt.Println("Table headers set")

	// Initial update
	updateProcessTable(table)

	// Set the table as the root widget to display it on the server
	if err := c.SetRoot(table); err != nil {
		log.Fatalf("Failed to set root widget: %v", err)
	}

	fmt.Println("Process monitor displayed on server")

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle events
	c.OnWidgetEvent(table.ID(), func(event *client.Event) {
		if event.Widget.Type == "WIDGET_SELECTED" {
			if row, col, err := table.GetSelection(); err == nil {
				if cell, err := table.GetCell(row, col); err == nil {
					fmt.Printf("Selected cell [%d,%d]: %s\n", row, col, cell.Text)
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
			// Force refresh
			updateProcessTable(table)
			fmt.Println("Process list refreshed")
		case 'k':
			// Kill selected process (demo - doesn't actually kill)
			if row, _, err := table.GetSelection(); err == nil && row > 0 {
				if cell, err := table.GetCell(row, 0); err == nil {
					fmt.Printf("Would kill process: %s\n", cell.Text)
				}
			}
		case 'h':
			fmt.Println("\nKeyboard shortcuts:")
			fmt.Println("  Arrow keys - Navigate table")
			fmt.Println("  r          - Refresh process list")
			fmt.Println("  k          - Kill selected process (demo)")
			fmt.Println("  h          - Show this help")
			fmt.Println("  q          - Quit")
		}
	})

	// Start update ticker
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			updateProcessTable(table)
		}
	}()

	fmt.Println("\nProcess Monitor created successfully!")
	fmt.Println("Process list updates every 3 seconds")
	fmt.Println("Press 'r' to refresh, 'k' to kill, 'h' for help, 'q' to quit")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

// updateProcessTable updates the process table with current processes
func updateProcessTable(table *client.Table) {
	processes := getProcessList()

	// Batch update all cells
	var cells []*pb.TableCellUpdate

	for i, proc := range processes {
		row := int32(i + 1) // +1 for header row

		// PID
		cells = append(cells, &pb.TableCellUpdate{
			Row:    row,
			Column: 0,
			Cell:   client.NewTableCell(proc.PID),
		})

		// Name
		cells = append(cells, &pb.TableCellUpdate{
			Row:    row,
			Column: 1,
			Cell:   client.NewTableCell(proc.Name),
		})

		// CPU
		cells = append(cells, &pb.TableCellUpdate{
			Row:    row,
			Column: 2,
			Cell:   client.NewTableCell(proc.CPU),
		})

		// Memory
		cells = append(cells, &pb.TableCellUpdate{
			Row:    row,
			Column: 3,
			Cell:   client.NewTableCell(proc.Memory),
		})

		// Status
		statusColor := client.Color("green")
		if proc.Status == "sleeping" {
			statusColor = client.Color("yellow")
		} else if proc.Status == "stopped" {
			statusColor = client.Color("red")
		}

		cells = append(cells, &pb.TableCellUpdate{
			Row:    row,
			Column: 4,
			Cell: &pb.TableCell{
				Text:  proc.Status,
				Color: statusColor,
			},
		})
	}

	if len(cells) > 0 {
		table.SetCells(cells)
	}
}

// getProcessList gets a list of running processes
func getProcessList() []ProcessInfo {
	var processes []ProcessInfo

	// Platform-specific process listing
	switch runtime.GOOS {
	case "darwin", "linux":
		processes = getUnixProcesses()
	default:
		// Fallback: show Go runtime info
		processes = getGoProcesses()
	}

	// Limit to top 20 processes
	if len(processes) > 20 {
		processes = processes[:20]
	}

	return processes
}

// getUnixProcesses gets processes on Unix-like systems
func getUnixProcesses() []ProcessInfo {
	var processes []ProcessInfo

	// Use ps command
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return getGoProcesses() // Fallback
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		processes = append(processes, ProcessInfo{
			PID:    fields[1],
			Name:   fields[10],
			CPU:    fields[2] + "%",
			Memory: fields[3] + "%",
			Status: fields[7],
		})
	}

	return processes
}

// getGoProcesses gets Go runtime information as fallback
func getGoProcesses() []ProcessInfo {
	var processes []ProcessInfo

	// Add current process
	pid := os.Getpid()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	processes = append(processes, ProcessInfo{
		PID:    strconv.Itoa(pid),
		Name:   "yutani-client",
		CPU:    "0.0%",
		Memory: fmt.Sprintf("%.1f MB", float64(m.Alloc)/1024/1024),
		Status: "running",
	})

	// Add some dummy processes for demo
	dummyProcesses := []ProcessInfo{
		{"1", "systemd", "0.1%", "0.5%", "sleeping"},
		{"100", "sshd", "0.0%", "0.2%", "sleeping"},
		{"200", "nginx", "0.5%", "1.2%", "running"},
		{"300", "postgres", "2.1%", "5.3%", "running"},
		{"400", "redis", "0.3%", "0.8%", "sleeping"},
	}

	processes = append(processes, dummyProcesses...)
	return processes
}

