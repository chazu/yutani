package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	newTemplate string
	newModule   string
)

var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: "Create a new Yutani project",
	Long: `Create a new Yutani project with scaffolding.

Available templates:
  - basic: Basic application with a single widget
  - list: List-based application
  - table: Table-based application
  - dashboard: Dashboard with multiple widgets
  - full: Full-featured application with all widgets`,
	Example: `  # Create basic project
  yutani new my-app

  # Create project with specific template
  yutani new my-app --template dashboard

  # Create project with custom module name
  yutani new my-app --module github.com/user/my-app`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	newCmd.Flags().StringVarP(&newTemplate, "template", "t", "basic", "Project template")
	newCmd.Flags().StringVarP(&newModule, "module", "m", "", "Go module name (default: project-name)")
}

func runNew(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Set default module name
	if newModule == "" {
		newModule = projectName
	}

	fmt.Printf("Creating new Yutani project: %s\n", projectName)
	fmt.Printf("Template: %s\n", newTemplate)
	fmt.Printf("Module: %s\n", newModule)
	fmt.Println()

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create project structure
	if err := createProjectStructure(projectName, newModule, newTemplate); err != nil {
		return err
	}

	fmt.Printf("\n✓ Project created successfully!\n\n")
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  go run main.go")
	fmt.Println()

	return nil
}

func createProjectStructure(projectName, moduleName, template string) error {
	// Create go.mod
	if err := createGoMod(projectName, moduleName); err != nil {
		return err
	}

	// Create main.go based on template
	if err := createMainGo(projectName, template); err != nil {
		return err
	}

	// Create README.md
	if err := createReadme(projectName); err != nil {
		return err
	}

	return nil
}

func createGoMod(projectName, moduleName string) error {
	content := fmt.Sprintf(`module %s

go 1.21

require github.com/chazu/yutani v0.1.0
`, moduleName)

	path := filepath.Join(projectName, "go.mod")
	return os.WriteFile(path, []byte(content), 0644)
}

func createMainGo(projectName, template string) error {
	var content string

	switch template {
	case "basic":
		content = getBasicTemplate()
	case "list":
		content = getListTemplate()
	case "table":
		content = getTableTemplate()
	case "dashboard":
		content = getDashboardTemplate()
	case "full":
		content = getFullTemplate()
	default:
		return fmt.Errorf("unknown template: %s", template)
	}

	path := filepath.Join(projectName, "main.go")
	return os.WriteFile(path, []byte(content), 0644)
}

func createReadme(projectName string) error {
	content := fmt.Sprintf(`# %s

A Yutani TUI application.

## Running

Start the Yutani server:
`+"```"+`bash
yutani server
`+"```"+`

In another terminal, run the application:
`+"```"+`bash
go run main.go
`+"```"+`

## Building

`+"```"+`bash
go build
`+"```"+`
`, strings.Title(projectName))

	path := filepath.Join(projectName, "README.md")
	return os.WriteFile(path, []byte(content), 0644)
}

func getBasicTemplate() string {
	return `package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
)

func main() {
	// Connect to server
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Create a simple text view
	tv, err := c.NewTextView().
		Title("My Application").
		Border(true).
		Text("Hello, Yutani!").
		Build()
	if err != nil {
		log.Fatalf("Failed to create widget: %v", err)
	}

	// Set as root widget
	if err := c.SetRoot(tv); err != nil {
		log.Fatalf("Failed to set root: %v", err)
	}

	// Start event stream
	if err := c.StartEventStream(); err != nil {
		log.Fatalf("Failed to start events: %v", err)
	}

	// Handle events
	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		if event.Key.Rune == 'q' {
			os.Exit(0)
		}
	})

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
`
}

func getListTemplate() string {
	return `package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
)

func main() {
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Create list
	list, err := c.NewList().
		Title("My List").
		Border(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create list: %v", err)
	}

	// Add items
	list.AddItem("Item 1", "Description 1", nil)
	list.AddItem("Item 2", "Description 2", nil)
	list.AddItem("Item 3", "Description 3", nil)

	c.SetRoot(list)
	c.StartEventStream()

	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		if event.Key.Rune == 'q' {
			os.Exit(0)
		}
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
`
}

func getTableTemplate() string {
	return `package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
)

func main() {
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Create table
	table, err := c.NewTable().
		Title("My Table").
		Border(true).
		Build()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Set headers
	table.SetCell(0, 0, "Name", nil)
	table.SetCell(0, 1, "Value", nil)
	table.SetCell(0, 2, "Status", nil)

	// Add data
	table.SetCell(1, 0, "Item 1", nil)
	table.SetCell(1, 1, "100", nil)
	table.SetCell(1, 2, "Active", nil)

	c.SetRoot(table)
	c.StartEventStream()

	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		if event.Key.Rune == 'q' {
			os.Exit(0)
		}
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
`
}

func getDashboardTemplate() string {
	return `package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chazu/yutani/pkg/client"
)

func main() {
	c, err := client.Connect("localhost:7755")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Create grid
	grid, err := c.NewGrid().
		Rows(2).
		Columns(2).
		Build()
	if err != nil {
		log.Fatalf("Failed to create grid: %v", err)
	}

	// Create widgets
	tv1, _ := c.NewTextView().Title("Panel 1").Border(true).Text("Content 1").Build()
	tv2, _ := c.NewTextView().Title("Panel 2").Border(true).Text("Content 2").Build()
	tv3, _ := c.NewTextView().Title("Panel 3").Border(true).Text("Content 3").Build()
	tv4, _ := c.NewTextView().Title("Panel 4").Border(true).Text("Content 4").Build()

	// Add to grid
	grid.AddItem(tv1, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(tv2, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(tv3, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(tv4, 1, 1, 1, 1, 0, 0, false)

	c.SetRoot(grid)
	c.StartEventStream()

	c.OnEventType(client.EventTypeKey, func(event *client.Event) {
		if event.Key.Rune == 'q' {
			os.Exit(0)
		}
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
`
}

func getFullTemplate() string {
	return getBasicTemplate() // For now, use basic template
}

