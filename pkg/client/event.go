package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
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
		var runeChar rune
		if len(e.Key.Rune) > 0 {
			runeChar = rune(e.Key.Rune[0])
		}
		// Combine modifiers into a single int (simplified)
		mod := 0
		for i, m := range e.Key.Modifiers {
			mod |= (int(m) << (i * 8))
		}
		event.Key = &KeyEvent{
			Key:  e.Key.Key.String(),
			Rune: runeChar,
			Mod:  mod,
		}

	case *pb.Event_Mouse:
		event.Type = EventTypeMouse
		var x, y int
		if e.Mouse.Position != nil {
			x = int(e.Mouse.Position.X)
			y = int(e.Mouse.Position.Y)
		}
		// Combine modifiers
		mod := 0
		for i, m := range e.Mouse.Modifiers {
			mod |= (int(m) << (i * 8))
		}
		event.Mouse = &MouseEvent{
			X:       x,
			Y:       y,
			Button:  0, // Not directly available in new proto
			Action:  e.Mouse.Action.String(),
			Buttons: 0, // Not directly available
			Mod:     mod,
		}

	case *pb.Event_Resize:
		event.Type = EventTypeResize
		var width, height int
		if e.Resize.NewSize != nil {
			width = int(e.Resize.NewSize.Width)
			height = int(e.Resize.NewSize.Height)
		}
		event.Resize = &ResizeEvent{
			Width:  width,
			Height: height,
		}

	case *pb.Event_Focus:
		event.Type = EventTypeFocus
		// Focus event now has old/new widgets, not a simple boolean
		focused := e.Focus.NewWidget != nil
		event.Focus = &FocusEvent{
			Focused: focused,
		}

	case *pb.Event_Widget:
		event.Type = EventTypeWidget
		event.Widget = &WidgetEvent{
			WidgetID: e.Widget.WidgetId.Id,
			Type:     e.Widget.Type.String(),
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

