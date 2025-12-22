package client

import (
	pb "industries/loosh/yutani/pkg/proto/industries/loosh/yutani/v1"
)

// EventType represents the type of event.
type EventType int

const (
	EventTypeKey EventType = iota
	EventTypeMouse
	EventTypeResize
	EventTypeFocus
	EventTypeWidget
)

// Event represents an event from the server.
type Event struct {
	Type   EventType
	Key    *KeyEvent
	Mouse  *MouseEvent
	Resize *ResizeEvent
	Focus  *FocusEvent
	Widget *WidgetEvent
}

// KeyEvent represents a keyboard event.
type KeyEvent struct {
	Key  string
	Rune rune
	Mod  int
}

// MouseEvent represents a mouse event.
type MouseEvent struct {
	X       int
	Y       int
	Button  int
	Action  string
	Buttons int
	Mod     int
}

// ResizeEvent represents a screen resize event.
type ResizeEvent struct {
	Width  int
	Height int
}

// FocusEvent represents a focus change event.
type FocusEvent struct {
	Focused bool
}

// WidgetEvent represents a widget-specific event.
type WidgetEvent struct {
	WidgetID string
	Type     string
	Data     map[string]string
}

// convertEvent converts a protobuf event to a client event.
func convertEvent(pbEvent *pb.Event) *Event {
	event := &Event{}

	switch e := pbEvent.Event.(type) {
	case *pb.Event_Key:
		event.Type = EventTypeKey
		event.Key = &KeyEvent{
			Key:  e.Key.Key,
			Rune: rune(e.Key.Rune),
			Mod:  int(e.Key.Mod),
		}

	case *pb.Event_Mouse:
		event.Type = EventTypeMouse
		event.Mouse = &MouseEvent{
			X:       int(e.Mouse.X),
			Y:       int(e.Mouse.Y),
			Button:  int(e.Mouse.Button),
			Action:  e.Mouse.Action,
			Buttons: int(e.Mouse.Buttons),
			Mod:     int(e.Mouse.Mod),
		}

	case *pb.Event_Resize:
		event.Type = EventTypeResize
		event.Resize = &ResizeEvent{
			Width:  int(e.Resize.Width),
			Height: int(e.Resize.Height),
		}

	case *pb.Event_Focus:
		event.Type = EventTypeFocus
		event.Focus = &FocusEvent{
			Focused: e.Focus.Focused,
		}

	case *pb.Event_Widget:
		event.Type = EventTypeWidget
		event.Widget = &WidgetEvent{
			WidgetID: e.Widget.WidgetId.Id,
			Type:     e.Widget.Type,
			Data:     e.Widget.Data,
		}
	}

	return event
}

// IsKey returns true if this is a key event.
func (e *Event) IsKey() bool {
	return e.Type == EventTypeKey
}

// IsMouse returns true if this is a mouse event.
func (e *Event) IsMouse() bool {
	return e.Type == EventTypeMouse
}

// IsResize returns true if this is a resize event.
func (e *Event) IsResize() bool {
	return e.Type == EventTypeResize
}

// IsFocus returns true if this is a focus event.
func (e *Event) IsFocus() bool {
	return e.Type == EventTypeFocus
}

// IsWidget returns true if this is a widget event.
func (e *Event) IsWidget() bool {
	return e.Type == EventTypeWidget
}

// IsWidgetEvent returns true if this is a widget event for the specified widget.
func (e *Event) IsWidgetEvent(widgetID string) bool {
	return e.Type == EventTypeWidget && e.Widget != nil && e.Widget.WidgetID == widgetID
}

// IsWidgetEventType returns true if this is a widget event of the specified type.
func (e *Event) IsWidgetEventType(eventType string) bool {
	return e.Type == EventTypeWidget && e.Widget != nil && e.Widget.Type == eventType
}

