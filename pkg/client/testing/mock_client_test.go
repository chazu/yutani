package testing

import (
	"fmt"
	"testing"

	"github.com/chazu/yutani/pkg/client"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestMockClient_RecordCall(t *testing.T) {
	mc := NewMockClient()

	mc.RecordCall("TestMethod", "arg1", 42)
	mc.RecordCall("TestMethod", "arg2", 43)
	mc.RecordCall("OtherMethod")

	if count := mc.GetCallCount("TestMethod"); count != 2 {
		t.Errorf("Expected TestMethod to be called 2 times, got %d", count)
	}

	if count := mc.GetCallCount("OtherMethod"); count != 1 {
		t.Errorf("Expected OtherMethod to be called 1 time, got %d", count)
	}

	calls := mc.GetCalls("TestMethod")
	if len(calls) != 2 {
		t.Errorf("Expected 2 calls to TestMethod, got %d", len(calls))
	}
}

func TestMockClient_ClearCalls(t *testing.T) {
	mc := NewMockClient()

	mc.RecordCall("Method1")
	mc.RecordCall("Method2")

	if count := mc.GetCallCount("Method1"); count != 1 {
		t.Errorf("Expected Method1 to be called 1 time, got %d", count)
	}

	mc.ClearCalls()

	if count := mc.GetCallCount("Method1"); count != 0 {
		t.Errorf("Expected Method1 call count to be 0 after clear, got %d", count)
	}
}

func TestMockClient_SetError(t *testing.T) {
	mc := NewMockClient()

	testErr := fmt.Errorf("test error")
	mc.SetError("TestMethod", testErr)

	if err := mc.checkError("TestMethod"); err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}

	mc.ClearError("TestMethod")

	if err := mc.checkError("TestMethod"); err != nil {
		t.Errorf("Expected no error after clear, got %v", err)
	}
}

func TestMockClient_RegisterWidget(t *testing.T) {
	mc := NewMockClient()

	mc.RegisterWidget("widget-1", pb.WidgetType_WIDGET_LIST)

	widget, ok := mc.GetWidget("widget-1")
	if !ok {
		t.Fatal("Expected widget to exist")
	}

	if widget.ID != "widget-1" {
		t.Errorf("Expected widget ID 'widget-1', got %s", widget.ID)
	}

	if widget.Type != pb.WidgetType_WIDGET_LIST {
		t.Errorf("Expected widget type LIST, got %v", widget.Type)
	}
}

func TestMockClient_SimulateKeyPress(t *testing.T) {
	mc := NewMockClient()

	var receivedEvent *client.Event
	mc.OnEvent(func(event *client.Event) {
		receivedEvent = event
	})

	mc.SimulateKeyPress('q')

	if receivedEvent == nil {
		t.Fatal("Expected to receive event")
	}

	if receivedEvent.Type != client.EventTypeKey {
		t.Errorf("Expected key event, got %v", receivedEvent.Type)
	}

	if receivedEvent.Key.Rune != 'q' {
		t.Errorf("Expected rune 'q', got %c", receivedEvent.Key.Rune)
	}

	if count := mc.GetCallCount("SimulateKeyPress"); count != 1 {
		t.Errorf("Expected SimulateKeyPress to be called 1 time, got %d", count)
	}
}

func TestMockClient_SimulateMouseClick(t *testing.T) {
	mc := NewMockClient()

	var receivedEvent *client.Event
	mc.OnEvent(func(event *client.Event) {
		receivedEvent = event
	})

	mc.SimulateMouseClick(10, 20, 1) // 1 = left button

	if receivedEvent == nil {
		t.Fatal("Expected to receive event")
	}

	if receivedEvent.Type != client.EventTypeMouse {
		t.Errorf("Expected mouse event, got %v", receivedEvent.Type)
	}

	if receivedEvent.Mouse.X != 10 || receivedEvent.Mouse.Y != 20 {
		t.Errorf("Expected position (10, 20), got (%d, %d)", receivedEvent.Mouse.X, receivedEvent.Mouse.Y)
	}

	if receivedEvent.Mouse.Button != 1 {
		t.Errorf("Expected button 1, got %d", receivedEvent.Mouse.Button)
	}
}

func TestMockClient_SimulateResize(t *testing.T) {
	mc := NewMockClient()

	var receivedEvent *client.Event
	mc.OnEvent(func(event *client.Event) {
		receivedEvent = event
	})

	mc.SimulateResize(80, 24)

	if receivedEvent == nil {
		t.Fatal("Expected to receive event")
	}

	if receivedEvent.Type != client.EventTypeResize {
		t.Errorf("Expected resize event, got %v", receivedEvent.Type)
	}

	if receivedEvent.Resize.Width != 80 || receivedEvent.Resize.Height != 24 {
		t.Errorf("Expected size (80, 24), got (%d, %d)", receivedEvent.Resize.Width, receivedEvent.Resize.Height)
	}
}

