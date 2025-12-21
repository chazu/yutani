# Yutani Phase 4 Demo

This demo showcases all the Phase 4 complex widgets implemented for the Yutani Terminal Display Server.

## What's Demonstrated

The demo creates a complex layout hierarchy using all Phase 4 services:

### Layout Structure
```
Main Flex (column direction)
├── Title (TextView with colored text)
├── Top Row Flex (row direction)
│   ├── List Widget - Shows project phases
│   └── Table Widget - Shows service statistics
└── Bottom Row Flex (row direction)
    ├── Form Widget - User settings form
    └── Tree Widget - Project structure
```

### Services Used

1. **LayoutService** - Creates Flex containers for layout composition
2. **ListService** - Displays a list of project phases with secondary text
3. **TableService** - Shows a table of service statistics with colored headers
4. **FormService** - Creates a form with input, password, checkbox, and dropdown fields
5. **TreeService** - Displays a hierarchical tree of the project structure

## Running the Demo

### Prerequisites

1. Build the server and demo:
   ```bash
   make build
   go build -o bin/phase4-demo ./cmd/phase4-demo
   ```

### Steps

1. Start the Yutani server in one terminal:
   ```bash
   ./bin/yutani-server
   ```

2. In another terminal, run the demo:
   ```bash
   ./bin/phase4-demo
   ```

3. The demo will:
   - Create a session
   - Build the layout hierarchy
   - Populate all widgets with data
   - Display the complete UI
   - Run for 60 seconds (or until you press Ctrl+C)

## Expected Output

You should see a terminal UI with:

- **Title**: Yellow bold text at the top
- **List** (top-left): 5 project phases with status and descriptions
- **Table** (top-right): Service statistics with 3 columns (Service, RPCs, Status)
- **Form** (bottom-left): User settings form with 4 fields and 2 buttons
- **Tree** (bottom-right): Project directory structure with expandable nodes

## Implementation Details

The demo demonstrates:

- **Flex Layout**: Using both column and row directions for responsive layout
- **Proportional Sizing**: Widgets sized proportionally within their containers
- **Fixed Sizing**: Title uses fixed size of 1 line
- **Widget Properties**: Borders, titles, colors, and other visual properties
- **Data Population**: Adding items to lists, cells to tables, fields to forms, and nodes to trees
- **Proper API Usage**: Correct protobuf structures and field types for all services

## Code Structure

- `main()` - Creates session and builds layout hierarchy
- `demoList()` - Creates and populates the list widget
- `demoTable()` - Creates and populates the table widget
- `demoForm()` - Creates and populates the form widget
- `demoTree()` - Creates and populates the tree widget

Each function demonstrates the proper API usage for its respective service.

