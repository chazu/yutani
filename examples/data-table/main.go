// Package main demonstrates a data table widget using the Yutani client library.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func main() {
	// Connect to the Yutani server
	c, err := client.Connect("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	fmt.Println("Connected to Yutani server")

	// Create a table widget
	table, err := c.NewTable().
		Title("Employee Directory").
		Border(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Printf("Created table widget: %s\n", table.ID())

	// Set header row with yellow background
	headers := []string{"Name", "Department", "Email", "Phone"}
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

	// Add employee data
	employees := [][]string{
		{"Alice Johnson", "Engineering", "alice@example.com", "555-0101"},
		{"Bob Smith", "Marketing", "bob@example.com", "555-0102"},
		{"Carol White", "Sales", "carol@example.com", "555-0103"},
		{"David Brown", "Engineering", "david@example.com", "555-0104"},
		{"Eve Davis", "HR", "eve@example.com", "555-0105"},
		{"Frank Miller", "Engineering", "frank@example.com", "555-0106"},
		{"Grace Wilson", "Marketing", "grace@example.com", "555-0107"},
		{"Henry Moore", "Sales", "henry@example.com", "555-0108"},
	}

	// Batch set all employee cells
	var cells []*pb.TableCellUpdate
	for row, employee := range employees {
		for col, value := range employee {
			cells = append(cells, &pb.TableCellUpdate{
				Row:    int32(row + 1), // +1 for header row
				Column: int32(col),
				Cell:   client.NewTableCell(value),
			})
		}
	}

	cellsSet, err := table.SetCells(cells)
	if err != nil {
		log.Fatalf("Failed to set cells: %v", err)
	}

	fmt.Printf("Set %d cells in table\n", cellsSet)

	// Set initial selection
	if err := table.SetSelection(1, 0); err != nil {
		log.Printf("Failed to set selection: %v", err)
	}

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start event stream: %v", err)
	}

	fmt.Println("Event stream started")

	// Handle events
	c.OnEvent(func(event *client.Event) {
		if event.IsWidget() && event.Widget.WidgetID == table.ID() {
			fmt.Printf("Table event: %s\n", event.Widget.Type)

			// Get current selection
			if row, col, err := table.GetSelection(); err == nil {
				if cell, err := table.GetCell(row, col); err == nil {
					fmt.Printf("  Selected cell [%d,%d]: %s\n", row, col, cell.Text)
				}
			}
		}

		if event.IsKey() {
			// Exit on 'q'
			if event.Key.Rune == 'q' {
				fmt.Println("Exiting...")
				os.Exit(0)
			}

			// Show dimensions on 'd'
			if event.Key.Rune == 'd' {
				if rows, cols, err := table.GetDimensions(); err == nil {
					fmt.Printf("Table dimensions: %d rows x %d columns\n", rows, cols)
				}
			}
		}
	})

	fmt.Println("\nTable widget created successfully!")
	fmt.Println("Use arrow keys to navigate, 'd' to show dimensions, 'q' to quit")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

