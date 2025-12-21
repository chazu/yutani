package services

import (
	"testing"

	"github.com/rivo/tview"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestConvertAlignment(t *testing.T) {
	tests := []struct {
		name     string
		align    pb.Alignment
		expected int
	}{
		{"left", pb.Alignment_ALIGN_LEFT, tview.AlignLeft},
		{"center", pb.Alignment_ALIGN_CENTER, tview.AlignCenter},
		{"right", pb.Alignment_ALIGN_RIGHT, tview.AlignRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAlignment(tt.align)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCreateBox(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with border and title",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				Title:  strPtr("Test Box"),
			},
		},
		{
			name: "with colors",
			props: &pb.WidgetProperties{
				BackgroundColor: &pb.Color{Color: &pb.Color_Name{Name: "blue"}},
				BorderColor:     &pb.Color{Color: &pb.Color_Name{Name: "white"}},
				TitleColor:      &pb.Color{Color: &pb.Color_Name{Name: "yellow"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := ws.createBox(tt.props)
			if box == nil {
				t.Fatal("Expected box to be created")
			}
		})
	}
}

func TestCreateTextView(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with text",
			props: &pb.WidgetProperties{
				TypeProperties: &pb.WidgetProperties_TextView{
					TextView: &pb.TextViewProperties{
						Text: strPtr("Hello, World!"),
					},
				},
			},
		},
		{
			name: "with dynamic colors",
			props: &pb.WidgetProperties{
				TypeProperties: &pb.WidgetProperties_TextView{
					TextView: &pb.TextViewProperties{
						Text:          strPtr("[red]Red text[white] white text"),
						DynamicColors: boolPtr(true),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv := ws.createTextView(tt.props)
			if tv == nil {
				t.Fatal("Expected TextView to be created")
			}
		})
	}
}

func TestCreateInputField(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with label and placeholder",
			props: &pb.WidgetProperties{
				TypeProperties: &pb.WidgetProperties_InputField{
					InputField: &pb.InputFieldProperties{
						Label:       strPtr("Name: "),
						Placeholder: strPtr("Enter your name"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ws.createInputField(tt.props)
			if input == nil {
				t.Fatal("Expected InputField to be created")
			}
		})
	}
}

func TestCreateButton(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with label",
			props: &pb.WidgetProperties{
				TypeProperties: &pb.WidgetProperties_Button{
					Button: &pb.ButtonProperties{
						Label: strPtr("Click Me!"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := ws.createButton(tt.props)
			if btn == nil {
				t.Fatal("Expected Button to be created")
			}
		})
	}
}

func TestCreateCheckbox(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with label checked",
			props: &pb.WidgetProperties{
				TypeProperties: &pb.WidgetProperties_Checkbox{
					Checkbox: &pb.CheckboxProperties{
						Label:   strPtr("Enable feature"),
						Checked: boolPtr(true),
					},
				},
			},
		},
		{
			name: "with label unchecked",
			props: &pb.WidgetProperties{
				TypeProperties: &pb.WidgetProperties_Checkbox{
					Checkbox: &pb.CheckboxProperties{
						Label:   strPtr("Disabled"),
						Checked: boolPtr(false),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := ws.createCheckbox(tt.props)
			if cb == nil {
				t.Fatal("Expected Checkbox to be created")
			}
		})
	}
}

func TestApplyBoxProperties(t *testing.T) {
	ws := &WidgetService{}
	box := tview.NewBox()

	props := &pb.WidgetProperties{
		Border:          boolPtr(true),
		Title:           strPtr("Test Title"),
		TitleAlign:      pb.Alignment_ALIGN_CENTER.Enum(),
		BackgroundColor: &pb.Color{Color: &pb.Color_Name{Name: "blue"}},
		BorderColor:     &pb.Color{Color: &pb.Color_Name{Name: "white"}},
		TitleColor:      &pb.Color{Color: &pb.Color_Name{Name: "yellow"}},
		Padding: &pb.Padding{
			Top:    1,
			Bottom: 1,
			Left:   2,
			Right:  2,
		},
	}

	ws.applyBoxProperties(box, props)

	// Verify the function doesn't panic
	if box == nil {
		t.Error("Box should not be nil after applying properties")
	}
}

func TestApplyTextViewProperties(t *testing.T) {
	ws := &WidgetService{}
	tv := tview.NewTextView()

	props := &pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_TextView{
			TextView: &pb.TextViewProperties{
				Text:          strPtr("Test text"),
				DynamicColors: boolPtr(true),
				WordWrap:      boolPtr(true),
				Scrollable:    boolPtr(true),
				TextColor:     &pb.Color{Color: &pb.Color_Name{Name: "green"}},
			},
		},
	}

	err := ws.applyTextViewProperties(tv, props)
	if err != nil {
		t.Errorf("applyTextViewProperties failed: %v", err)
	}

	// Verify the function doesn't panic
	if tv == nil {
		t.Error("TextView should not be nil after applying properties")
	}
}

func TestApplyInputFieldProperties(t *testing.T) {
	ws := &WidgetService{}
	input := tview.NewInputField()

	props := &pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{
				Label:          strPtr("Name: "),
				Placeholder:    strPtr("Enter name"),
				Text:           strPtr("Initial text"),
				FieldWidth:     int32Ptr(20),
				LabelColor:     &pb.Color{Color: &pb.Color_Name{Name: "white"}},
				FieldTextColor: &pb.Color{Color: &pb.Color_Name{Name: "yellow"}},
			},
		},
	}

	err := ws.applyInputFieldProperties(input, props)
	if err != nil {
		t.Errorf("applyInputFieldProperties failed: %v", err)
	}

	// Verify the function doesn't panic
	if input == nil {
		t.Error("InputField should not be nil after applying properties")
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

// Tests for complex widgets

func TestCreateList(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with list properties",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				Title:  strPtr("Test List"),
				TypeProperties: &pb.WidgetProperties_List{
					List: &pb.ListProperties{
						ShowSecondaryText: boolPtr(true),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := ws.createList(tt.props)
			if list == nil {
				t.Error("Expected non-nil list")
			}
		})
	}
}

func TestCreateTable(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with table properties",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				Title:  strPtr("Test Table"),
				TypeProperties: &pb.WidgetProperties_Table{
					Table: &pb.TableProperties{
						Borders:    boolPtr(true),
						Selectable: boolPtr(true),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := ws.createTable(tt.props)
			if table == nil {
				t.Error("Expected non-nil table")
			}
		})
	}
}

func TestCreateTreeView(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with tree properties",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				Title:  strPtr("Test Tree"),
				TypeProperties: &pb.WidgetProperties_Tree{
					Tree: &pb.TreeProperties{
						ShowGraphics: boolPtr(true),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := ws.createTreeView(tt.props)
			if tree == nil {
				t.Error("Expected non-nil tree")
			}
		})
	}
}

func TestCreateForm(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with form properties",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				Title:  strPtr("Test Form"),
				TypeProperties: &pb.WidgetProperties_Form{
					Form: &pb.FormProperties{
						Horizontal: boolPtr(false),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := ws.createForm(tt.props)
			if form == nil {
				t.Error("Expected non-nil form")
			}
		})
	}
}

func TestCreateFlex(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with flex properties",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				TypeProperties: &pb.WidgetProperties_Flex{
					Flex: &pb.FlexProperties{
						Direction: func() *pb.FlexDirection {
							d := pb.FlexDirection_FLEX_COLUMN
							return &d
						}(),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := ws.createFlex(tt.props)
			if flex == nil {
				t.Error("Expected non-nil flex")
			}
		})
	}
}

func TestCreateGrid(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with grid properties",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				TypeProperties: &pb.WidgetProperties_Grid{
					Grid: &pb.GridProperties{
						Rows:    int32Ptr(3),
						Columns: int32Ptr(3),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := ws.createGrid(tt.props)
			if grid == nil {
				t.Error("Expected non-nil grid")
			}
		})
	}
}

func TestCreatePages(t *testing.T) {
	ws := &WidgetService{}

	tests := []struct {
		name  string
		props *pb.WidgetProperties
	}{
		{
			name:  "nil properties",
			props: nil,
		},
		{
			name: "with pages properties",
			props: &pb.WidgetProperties{
				Border: boolPtr(true),
				TypeProperties: &pb.WidgetProperties_Pages{
					Pages: &pb.PagesProperties{
						ShowPageNames: boolPtr(true),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages := ws.createPages(tt.props)
			if pages == nil {
				t.Error("Expected non-nil pages")
			}
		})
	}
}
