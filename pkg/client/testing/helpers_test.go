package testing

import (
	"fmt"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestTestHelper_AssertCallCount(t *testing.T) {
	h := NewTestHelper(t)

	h.Client().RecordCall("TestMethod")
	h.Client().RecordCall("TestMethod")

	// This should pass
	h.AssertCallCount("TestMethod", 2)

	// This should fail (but we can't test the failure easily)
	// h.AssertCallCount("TestMethod", 3)
}

func TestTestHelper_AssertCalled(t *testing.T) {
	h := NewTestHelper(t)

	h.Client().RecordCall("TestMethod")

	// This should pass
	h.AssertCalled("TestMethod")
}

func TestTestHelper_AssertNotCalled(t *testing.T) {
	h := NewTestHelper(t)

	// This should pass
	h.AssertNotCalled("TestMethod")
}

func TestTestHelper_AssertWidgetExists(t *testing.T) {
	h := NewTestHelper(t)

	h.Client().RegisterWidget("widget-1", pb.WidgetType_WIDGET_LIST)

	// This should pass
	h.AssertWidgetExists("widget-1")
}

func TestTestHelper_AssertWidgetType(t *testing.T) {
	h := NewTestHelper(t)

	h.Client().RegisterWidget("widget-1", pb.WidgetType_WIDGET_LIST)

	// This should pass
	h.AssertWidgetType("widget-1", pb.WidgetType_WIDGET_LIST)
}

func TestTestHelper_AssertWidgetProperty(t *testing.T) {
	h := NewTestHelper(t)

	h.Client().RegisterWidget("widget-1", pb.WidgetType_WIDGET_LIST)
	widget, _ := h.Client().GetWidget("widget-1")
	widget.Properties["title"] = "Test Title"

	// This should pass
	h.AssertWidgetProperty("widget-1", "title", "Test Title")
}

func TestTestHelper_AssertEqual(t *testing.T) {
	h := NewTestHelper(t)

	// This should pass
	h.AssertEqual(42, 42, "values should be equal")
}

func TestTestHelper_AssertTrue(t *testing.T) {
	h := NewTestHelper(t)

	// This should pass
	h.AssertTrue(true, "condition should be true")
}

func TestTestHelper_AssertFalse(t *testing.T) {
	h := NewTestHelper(t)

	// This should pass
	h.AssertFalse(false, "condition should be false")
}

func TestCreateMockList(t *testing.T) {
	mc := NewMockClient()

	widgetID := CreateMockList(mc, "Test List")

	if widgetID == "" {
		t.Fatal("Expected widget ID to be non-empty")
	}

	widget, ok := mc.GetWidget(widgetID)
	if !ok {
		t.Fatal("Expected widget to exist")
	}

	if widget.Type != pb.WidgetType_WIDGET_LIST {
		t.Errorf("Expected widget type LIST, got %v", widget.Type)
	}

	if widget.Properties["title"] != "Test List" {
		t.Errorf("Expected title 'Test List', got %v", widget.Properties["title"])
	}

	if count := mc.GetCallCount("CreateWidget"); count != 1 {
		t.Errorf("Expected CreateWidget to be called 1 time, got %d", count)
	}
}

func TestCreateMockTable(t *testing.T) {
	mc := NewMockClient()

	widgetID := CreateMockTable(mc, "Test Table")

	if widgetID == "" {
		t.Fatal("Expected widget ID to be non-empty")
	}

	widget, ok := mc.GetWidget(widgetID)
	if !ok {
		t.Fatal("Expected widget to exist")
	}

	if widget.Type != pb.WidgetType_WIDGET_TABLE {
		t.Errorf("Expected widget type TABLE, got %v", widget.Type)
	}

	if widget.Properties["title"] != "Test Table" {
		t.Errorf("Expected title 'Test Table', got %v", widget.Properties["title"])
	}
}

func TestTestHelper_AssertNoError(t *testing.T) {
	h := NewTestHelper(t)

	// This should pass
	h.AssertNoError(nil)
}

func TestTestHelper_AssertError(t *testing.T) {
	h := NewTestHelper(t)

	// This should pass
	h.AssertError(fmt.Errorf("test error"))
}

