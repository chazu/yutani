package testing

import (
	"fmt"
	"testing"
)

// WidgetAssertion provides widget-specific assertions.
type WidgetAssertion struct {
	t        *testing.T
	client   *MockClient
	widgetID string
}

// NewWidgetAssertion creates a new widget assertion helper.
func NewWidgetAssertion(t *testing.T, client *MockClient, widgetID string) *WidgetAssertion {
	return &WidgetAssertion{
		t:        t,
		client:   client,
		widgetID: widgetID,
	}
}

// Exists asserts that the widget exists.
func (a *WidgetAssertion) Exists() *WidgetAssertion {
	a.t.Helper()
	if _, ok := a.client.GetWidget(a.widgetID); !ok {
		a.t.Errorf("Widget %s does not exist", a.widgetID)
	}
	return a
}

// HasProperty asserts that the widget has a specific property.
func (a *WidgetAssertion) HasProperty(key string, value interface{}) *WidgetAssertion {
	a.t.Helper()
	widget, ok := a.client.GetWidget(a.widgetID)
	if !ok {
		a.t.Errorf("Widget %s does not exist", a.widgetID)
		return a
	}
	
	actual, ok := widget.Properties[key]
	if !ok {
		a.t.Errorf("Widget %s does not have property %s", a.widgetID, key)
		return a
	}
	
	if actual != value {
		a.t.Errorf("Widget %s property %s: expected %v, got %v", a.widgetID, key, value, actual)
	}
	return a
}

// HasTitle asserts that the widget has a specific title.
func (a *WidgetAssertion) HasTitle(title string) *WidgetAssertion {
	return a.HasProperty("title", title)
}

// HasBorder asserts that the widget has a border.
func (a *WidgetAssertion) HasBorder(hasBorder bool) *WidgetAssertion {
	return a.HasProperty("border", hasBorder)
}

// HasText asserts that the widget has specific text content.
func (a *WidgetAssertion) HasText(text string) *WidgetAssertion {
	return a.HasProperty("text", text)
}

// ListAssertion provides list-specific assertions.
type ListAssertion struct {
	*WidgetAssertion
	items []ListItem
}

// ListItem represents a list item for testing.
type ListItem struct {
	Main      string
	Secondary string
	Shortcut  string
}

// NewListAssertion creates a new list assertion helper.
func NewListAssertion(t *testing.T, client *MockClient, widgetID string) *ListAssertion {
	return &ListAssertion{
		WidgetAssertion: NewWidgetAssertion(t, client, widgetID),
		items:           make([]ListItem, 0),
	}
}

// AddItem records a list item (for testing).
func (a *ListAssertion) AddItem(main, secondary, shortcut string) *ListAssertion {
	a.items = append(a.items, ListItem{
		Main:      main,
		Secondary: secondary,
		Shortcut:  shortcut,
	})
	return a
}

// HasItemCount asserts the number of items in the list.
func (a *ListAssertion) HasItemCount(count int) *ListAssertion {
	a.t.Helper()
	if len(a.items) != count {
		a.t.Errorf("List %s: expected %d items, got %d", a.widgetID, count, len(a.items))
	}
	return a
}

// HasItem asserts that the list contains a specific item.
func (a *ListAssertion) HasItem(index int, main string) *ListAssertion {
	a.t.Helper()
	if index >= len(a.items) {
		a.t.Errorf("List %s: item index %d out of range (have %d items)", a.widgetID, index, len(a.items))
		return a
	}
	
	if a.items[index].Main != main {
		a.t.Errorf("List %s item %d: expected main text %q, got %q", a.widgetID, index, main, a.items[index].Main)
	}
	return a
}

// TableAssertion provides table-specific assertions.
type TableAssertion struct {
	*WidgetAssertion
	cells map[string]string // "row,col" -> value
}

// NewTableAssertion creates a new table assertion helper.
func NewTableAssertion(t *testing.T, client *MockClient, widgetID string) *TableAssertion {
	return &TableAssertion{
		WidgetAssertion: NewWidgetAssertion(t, client, widgetID),
		cells:           make(map[string]string),
	}
}

// SetCell records a cell value (for testing).
func (a *TableAssertion) SetCell(row, col int, value string) *TableAssertion {
	key := fmt.Sprintf("%d,%d", row, col)
	a.cells[key] = value
	return a
}

// HasCell asserts that a cell has a specific value.
func (a *TableAssertion) HasCell(row, col int, expected string) *TableAssertion {
	a.t.Helper()
	key := fmt.Sprintf("%d,%d", row, col)
	actual, ok := a.cells[key]
	if !ok {
		a.t.Errorf("Table %s: cell [%d,%d] not set", a.widgetID, row, col)
		return a
	}
	
	if actual != expected {
		a.t.Errorf("Table %s cell [%d,%d]: expected %q, got %q", a.widgetID, row, col, expected, actual)
	}
	return a
}

// HasDimensions asserts the table dimensions.
func (a *TableAssertion) HasDimensions(rows, cols int) *TableAssertion {
	a.t.Helper()
	// Count unique rows and columns
	maxRow, maxCol := 0, 0
	for key := range a.cells {
		var r, c int
		fmt.Sscanf(key, "%d,%d", &r, &c)
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}
	
	actualRows := maxRow + 1
	actualCols := maxCol + 1
	
	if actualRows != rows || actualCols != cols {
		a.t.Errorf("Table %s: expected dimensions %dx%d, got %dx%d", a.widgetID, rows, cols, actualRows, actualCols)
	}
	return a
}

