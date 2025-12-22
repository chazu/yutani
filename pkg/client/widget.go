package client

import (
	pb "industries/loosh/yutani/pkg/proto/industries/loosh/yutani/v1"
)

// Widget is the interface that all widgets implement.
type Widget interface {
	// ID returns the widget's unique identifier.
	ID() string

	// Delete removes the widget from the server.
	Delete() error

	// SetTitle sets the widget's title.
	SetTitle(title string) error

	// SetBorder sets whether the widget has a border.
	SetBorder(border bool) error

	// SetVisible sets whether the widget is visible.
	SetVisible(visible bool) error

	// SetFocus sets focus to this widget.
	SetFocus() error

	// GetProperties returns the widget's current properties.
	GetProperties() (*pb.WidgetProperties, error)
}

// baseWidget provides common functionality for all widgets.
type baseWidget struct {
	client   *Client
	widgetID *pb.WidgetID
}

// ID returns the widget's unique identifier.
func (w *baseWidget) ID() string {
	return w.widgetID.Id
}

// Delete removes the widget from the server.
func (w *baseWidget) Delete() error {
	_, err := w.client.widgetClient.DeleteWidget(w.client.ctx, &pb.DeleteWidgetRequest{
		SessionId: w.client.sessionID,
		WidgetId:  w.widgetID,
	})
	if err == nil {
		w.client.unregisterWidget(w.widgetID.Id)
	}
	return err
}

// SetTitle sets the widget's title.
func (w *baseWidget) SetTitle(title string) error {
	return w.setProperty(&pb.WidgetProperties{
		Title: &title,
	})
}

// SetBorder sets whether the widget has a border.
func (w *baseWidget) SetBorder(border bool) error {
	return w.setProperty(&pb.WidgetProperties{
		Border: &border,
	})
}

// SetVisible sets whether the widget is visible.
func (w *baseWidget) SetVisible(visible bool) error {
	return w.setProperty(&pb.WidgetProperties{
		Visible: &visible,
	})
}

// SetFocus sets focus to this widget.
func (w *baseWidget) SetFocus() error {
	_, err := w.client.widgetClient.SetFocus(w.client.ctx, &pb.SetFocusRequest{
		SessionId: w.client.sessionID,
		WidgetId:  w.widgetID,
	})
	return err
}

// GetProperties returns the widget's current properties.
func (w *baseWidget) GetProperties() (*pb.WidgetProperties, error) {
	resp, err := w.client.widgetClient.GetProperties(w.client.ctx, &pb.GetPropertiesRequest{
		SessionId: w.client.sessionID,
		WidgetId:  w.widgetID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Properties, nil
}

// setProperty sets widget properties.
func (w *baseWidget) setProperty(props *pb.WidgetProperties) error {
	_, err := w.client.widgetClient.SetProperties(w.client.ctx, &pb.SetPropertiesRequest{
		SessionId:  w.client.sessionID,
		WidgetId:   w.widgetID,
		Properties: props,
	})
	return err
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func uint32Ptr(u uint32) *uint32 {
	return &u
}

// Color creates a named color.
func Color(name string) *pb.Color {
	return &pb.Color{
		Color: &pb.Color_Name{Name: name},
	}
}

// ColorRGB creates an RGB color.
func ColorRGB(r, g, b uint32) *pb.Color {
	return &pb.Color{
		Color: &pb.Color_Rgb{
			Rgb: &pb.RGBColor{
				R: r,
				G: g,
				B: b,
			},
		},
	}
}

// ColorHex creates a color from a hex string.
func ColorHex(hex string) *pb.Color {
	return &pb.Color{
		Color: &pb.Color_Hex{Hex: hex},
	}
}

