package client

import (
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// TestButtonBuilder tests the Button widget builder pattern
func TestButtonBuilder(t *testing.T) {
	// Create a builder directly (we're just testing the builder pattern)
	builder := &ButtonBuilder{
		client: nil, // Not needed for property tests
		props:  &pb.WidgetProperties{},
	}
	if builder == nil {
		t.Fatal("NewButton() returned nil")
	}

	// Test fluent API
	builder = builder.
		Title("Test Button").
		Border(true).
		Label("Click Me").
		LabelColor(Color("blue")).
		BackgroundColor(Color("white")).
		ActivatedColor(Color("green"))

	if builder.props.Title == nil || *builder.props.Title != "Test Button" {
		t.Error("Title not set correctly")
	}
	if builder.props.Border == nil || !*builder.props.Border {
		t.Error("Border not set correctly")
	}
	if builder.props.GetButton() == nil {
		t.Error("Button properties not initialized")
	}
	if builder.props.GetButton().Label == nil || *builder.props.GetButton().Label != "Click Me" {
		t.Error("Label not set correctly")
	}
}

// TestCheckboxBuilder tests the Checkbox widget builder pattern
func TestCheckboxBuilder(t *testing.T) {
	builder := &CheckboxBuilder{
		client: nil,
		props:  &pb.WidgetProperties{},
	}
	if builder == nil {
		t.Fatal("NewCheckbox() returned nil")
	}

	// Test fluent API
	builder = builder.
		Title("Test Checkbox").
		Border(true).
		Label("Enable feature").
		Checked(true).
		LabelColor(Color("yellow")).
		CheckedColor(Color("green"))

	if builder.props.Title == nil || *builder.props.Title != "Test Checkbox" {
		t.Error("Title not set correctly")
	}
	if builder.props.GetCheckbox() == nil {
		t.Error("Checkbox properties not initialized")
	}
	if builder.props.GetCheckbox().Label == nil || *builder.props.GetCheckbox().Label != "Enable feature" {
		t.Error("Label not set correctly")
	}
	if builder.props.GetCheckbox().Checked == nil || !*builder.props.GetCheckbox().Checked {
		t.Error("Checked state not set correctly")
	}
}

// TestInputFieldBuilder tests the InputField widget builder pattern
func TestInputFieldBuilder(t *testing.T) {
	builder := &InputFieldBuilder{
		client: nil,
		props:  &pb.WidgetProperties{},
	}
	if builder == nil {
		t.Fatal("NewInputField() returned nil")
	}

	// Test fluent API
	builder = builder.
		Title("Test Input").
		Border(true).
		Label("Name:").
		Placeholder("Enter your name").
		Text("John Doe").
		FieldWidth(30).
		Masked(false).
		LabelColor(Color("cyan")).
		FieldTextColor(Color("white")).
		FieldBackgroundColor(Color("black"))

	if builder.props.Title == nil || *builder.props.Title != "Test Input" {
		t.Error("Title not set correctly")
	}
	if builder.props.GetInputField() == nil {
		t.Error("InputField properties not initialized")
	}
	if builder.props.GetInputField().Label == nil || *builder.props.GetInputField().Label != "Name:" {
		t.Error("Label not set correctly")
	}
	if builder.props.GetInputField().Placeholder == nil || *builder.props.GetInputField().Placeholder != "Enter your name" {
		t.Error("Placeholder not set correctly")
	}
	if builder.props.GetInputField().FieldWidth == nil || *builder.props.GetInputField().FieldWidth != 30 {
		t.Error("FieldWidth not set correctly")
	}
}

// TestTreeViewBuilder tests the TreeView widget builder pattern
func TestTreeViewBuilder(t *testing.T) {
	builder := &TreeViewBuilder{
		client: nil,
		props:  &pb.WidgetProperties{},
	}
	if builder == nil {
		t.Fatal("NewTreeView() returned nil")
	}

	// Test fluent API
	builder = builder.
		Title("Test Tree").
		Border(true).
		NodeTextColor(Color("white")).
		SelectedTextColor(Color("black")).
		SelectedBackgroundColor(Color("blue")).
		TopLevelPrefix("* ").
		ShowGraphics(true)

	if builder.props.Title == nil || *builder.props.Title != "Test Tree" {
		t.Error("Title not set correctly")
	}
	if builder.props.GetTree() == nil {
		t.Error("Tree properties not initialized")
	}
	if builder.props.GetTree().TopLevelPrefix == nil || *builder.props.GetTree().TopLevelPrefix != "* " {
		t.Error("TopLevelPrefix not set correctly")
	}
	if builder.props.GetTree().ShowGraphics == nil || !*builder.props.GetTree().ShowGraphics {
		t.Error("ShowGraphics not set correctly")
	}
}

// TestFlexBuilder tests the Flex widget builder pattern
func TestFlexBuilder(t *testing.T) {
	builder := &FlexBuilder{
		client: nil,
		props:  &pb.WidgetProperties{},
	}
	if builder == nil {
		t.Fatal("NewFlex() returned nil")
	}

	// Test fluent API
	builder = builder.
		Title("Test Flex").
		Border(true).
		Direction(pb.FlexDirection_FLEX_COLUMN).
		FullScreen(true)

	if builder.props.Title == nil || *builder.props.Title != "Test Flex" {
		t.Error("Title not set correctly")
	}
	if builder.props.GetFlex() == nil {
		t.Error("Flex properties not initialized")
	}
	if builder.props.GetFlex().Direction == nil || *builder.props.GetFlex().Direction != pb.FlexDirection_FLEX_COLUMN {
		t.Error("Direction not set correctly")
	}
	if builder.props.GetFlex().FullScreen == nil || !*builder.props.GetFlex().FullScreen {
		t.Error("FullScreen not set correctly")
	}
}

// TestGridBuilder tests the Grid widget builder pattern
func TestGridBuilder(t *testing.T) {
	builder := &GridBuilder{
		client: nil,
		props:  &pb.WidgetProperties{},
	}
	if builder == nil {
		t.Fatal("NewGrid() returned nil")
	}

	// Test fluent API
	builder = builder.
		Title("Test Grid").
		Border(true).
		Rows(3).
		Columns(4).
		MinWidth(80).
		MinHeight(24)

	if builder.props.Title == nil || *builder.props.Title != "Test Grid" {
		t.Error("Title not set correctly")
	}
	if builder.props.GetGrid() == nil {
		t.Error("Grid properties not initialized")
	}
	if builder.props.GetGrid().Rows == nil || *builder.props.GetGrid().Rows != 3 {
		t.Error("Rows not set correctly")
	}
	if builder.props.GetGrid().Columns == nil || *builder.props.GetGrid().Columns != 4 {
		t.Error("Columns not set correctly")
	}
	if builder.props.GetGrid().MinWidth == nil || *builder.props.GetGrid().MinWidth != 80 {
		t.Error("MinWidth not set correctly")
	}
	if builder.props.GetGrid().MinHeight == nil || *builder.props.GetGrid().MinHeight != 24 {
		t.Error("MinHeight not set correctly")
	}
}

// TestPagesBuilder tests the Pages widget builder pattern
func TestPagesBuilder(t *testing.T) {
	builder := &PagesBuilder{
		client: nil,
		props:  &pb.WidgetProperties{},
	}
	if builder == nil {
		t.Fatal("NewPages() returned nil")
	}

	// Test fluent API
	builder = builder.
		Title("Test Pages").
		Border(true).
		ShowPageNames(true).
		PageNameColor(Color("cyan"))

	if builder.props.Title == nil || *builder.props.Title != "Test Pages" {
		t.Error("Title not set correctly")
	}
	if builder.props.GetPages() == nil {
		t.Error("Pages properties not initialized")
	}
	if builder.props.GetPages().ShowPageNames == nil || !*builder.props.GetPages().ShowPageNames {
		t.Error("ShowPageNames not set correctly")
	}
}

// TestTreeNodeHelpers tests the tree node helper functions
func TestTreeNodeHelpers(t *testing.T) {
	// Test NewTreeNode
	node := NewTreeNode("Test Node")
	if node == nil {
		t.Fatal("NewTreeNode returned nil")
	}
	if node.Text != "Test Node" {
		t.Error("NewTreeNode text not set correctly")
	}

	// Test TreeNodeWithColor
	node = TreeNodeWithColor("Colored Node", Color("red"))
	if node == nil {
		t.Fatal("TreeNodeWithColor returned nil")
	}
	if node.Text != "Colored Node" {
		t.Error("TreeNodeWithColor text not set correctly")
	}
	if node.Color == nil {
		t.Error("TreeNodeWithColor color not set")
	}

	// Test TreeNodeWithOptions
	node = TreeNodeWithOptions("Full Node", Color("blue"), true, false, "ref123")
	if node == nil {
		t.Fatal("TreeNodeWithOptions returned nil")
	}
	if node.Text != "Full Node" {
		t.Error("TreeNodeWithOptions text not set correctly")
	}
	if node.Selectable == nil || !*node.Selectable {
		t.Error("TreeNodeWithOptions selectable not set correctly")
	}
	if node.Expanded == nil || *node.Expanded {
		t.Error("TreeNodeWithOptions expanded not set correctly")
	}
	if node.Reference == nil || *node.Reference != "ref123" {
		t.Error("TreeNodeWithOptions reference not set correctly")
	}
}

