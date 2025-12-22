package testing

import (
	"fmt"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// TestHelper provides convenience methods for testing.
type TestHelper struct {
	t      *testing.T
	client *MockClient
}

// NewTestHelper creates a new test helper.
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:      t,
		client: NewMockClient(),
	}
}

// Client returns the mock client.
func (h *TestHelper) Client() *MockClient {
	return h.client
}

// AssertCallCount asserts that a method was called a specific number of times.
func (h *TestHelper) AssertCallCount(method string, expected int) {
	h.t.Helper()
	actual := h.client.GetCallCount(method)
	if actual != expected {
		h.t.Errorf("Expected %s to be called %d times, but was called %d times", method, expected, actual)
	}
}

// AssertCalled asserts that a method was called at least once.
func (h *TestHelper) AssertCalled(method string) {
	h.t.Helper()
	if h.client.GetCallCount(method) == 0 {
		h.t.Errorf("Expected %s to be called, but it was not", method)
	}
}

// AssertNotCalled asserts that a method was not called.
func (h *TestHelper) AssertNotCalled(method string) {
	h.t.Helper()
	if h.client.GetCallCount(method) > 0 {
		h.t.Errorf("Expected %s not to be called, but it was called %d times", method, h.client.GetCallCount(method))
	}
}

// AssertWidgetExists asserts that a widget exists.
func (h *TestHelper) AssertWidgetExists(widgetID string) {
	h.t.Helper()
	if _, ok := h.client.GetWidget(widgetID); !ok {
		h.t.Errorf("Expected widget %s to exist, but it does not", widgetID)
	}
}

// AssertWidgetType asserts that a widget has a specific type.
func (h *TestHelper) AssertWidgetType(widgetID string, expectedType pb.WidgetType) {
	h.t.Helper()
	widget, ok := h.client.GetWidget(widgetID)
	if !ok {
		h.t.Errorf("Widget %s does not exist", widgetID)
		return
	}
	if widget.Type != expectedType {
		h.t.Errorf("Expected widget %s to be type %v, but was %v", widgetID, expectedType, widget.Type)
	}
}

// AssertWidgetProperty asserts that a widget has a specific property value.
func (h *TestHelper) AssertWidgetProperty(widgetID, property string, expected interface{}) {
	h.t.Helper()
	widget, ok := h.client.GetWidget(widgetID)
	if !ok {
		h.t.Errorf("Widget %s does not exist", widgetID)
		return
	}
	actual, ok := widget.Properties[property]
	if !ok {
		h.t.Errorf("Widget %s does not have property %s", widgetID, property)
		return
	}
	if actual != expected {
		h.t.Errorf("Expected widget %s property %s to be %v, but was %v", widgetID, property, expected, actual)
	}
}

// AssertNoError asserts that an error is nil.
func (h *TestHelper) AssertNoError(err error) {
	h.t.Helper()
	if err != nil {
		h.t.Errorf("Expected no error, but got: %v", err)
	}
}

// AssertError asserts that an error is not nil.
func (h *TestHelper) AssertError(err error) {
	h.t.Helper()
	if err == nil {
		h.t.Error("Expected an error, but got nil")
	}
}

// AssertEqual asserts that two values are equal.
func (h *TestHelper) AssertEqual(actual, expected interface{}, message string) {
	h.t.Helper()
	if actual != expected {
		h.t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertTrue asserts that a condition is true.
func (h *TestHelper) AssertTrue(condition bool, message string) {
	h.t.Helper()
	if !condition {
		h.t.Errorf("Expected condition to be true: %s", message)
	}
}

// AssertFalse asserts that a condition is false.
func (h *TestHelper) AssertFalse(condition bool, message string) {
	h.t.Helper()
	if condition {
		h.t.Errorf("Expected condition to be false: %s", message)
	}
}

// Convenience Functions

// CreateMockList creates a mock list widget for testing.
func CreateMockList(c *MockClient, title string) string {
	widgetID := fmt.Sprintf("list-%d", c.GetCallCount("CreateWidget")+1)
	c.RegisterWidget(widgetID, pb.WidgetType_WIDGET_LIST)
	c.RecordCall("CreateWidget", "list", title)
	
	widget, _ := c.GetWidget(widgetID)
	widget.Properties["title"] = title
	widget.Properties["border"] = true
	
	return widgetID
}

// CreateMockTable creates a mock table widget for testing.
func CreateMockTable(c *MockClient, title string) string {
	widgetID := fmt.Sprintf("table-%d", c.GetCallCount("CreateWidget")+1)
	c.RegisterWidget(widgetID, pb.WidgetType_WIDGET_TABLE)
	c.RecordCall("CreateWidget", "table", title)
	
	widget, _ := c.GetWidget(widgetID)
	widget.Properties["title"] = title
	widget.Properties["border"] = true
	
	return widgetID
}

