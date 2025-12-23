package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Button represents a button widget.
type Button struct {
	baseWidget
}

// ButtonBuilder provides a fluent interface for creating button widgets.
type ButtonBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewButton creates a new button builder.
func (c *Client) NewButton() *ButtonBuilder {
	return &ButtonBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the button's title (border title).
func (b *ButtonBuilder) Title(title string) *ButtonBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the button has a border.
func (b *ButtonBuilder) Border(border bool) *ButtonBuilder {
	b.props.Border = &border
	return b
}

// Label sets the button label text.
func (b *ButtonBuilder) Label(label string) *ButtonBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Button{
			Button: &pb.ButtonProperties{},
		}
	}
	b.props.GetButton().Label = &label
	return b
}

// LabelColor sets the button label color.
func (b *ButtonBuilder) LabelColor(color *pb.Color) *ButtonBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Button{
			Button: &pb.ButtonProperties{},
		}
	}
	b.props.GetButton().LabelColor = color
	return b
}

// BackgroundColor sets the button background color.
func (b *ButtonBuilder) BackgroundColor(color *pb.Color) *ButtonBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Button{
			Button: &pb.ButtonProperties{},
		}
	}
	b.props.GetButton().BackgroundColor = color
	return b
}

// ActivatedColor sets the button color when activated.
func (b *ButtonBuilder) ActivatedColor(color *pb.Color) *ButtonBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Button{
			Button: &pb.ButtonProperties{},
		}
	}
	b.props.GetButton().ActivatedColor = color
	return b
}

// Build creates the button widget on the server.
func (b *ButtonBuilder) Build() (*Button, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_BUTTON,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	button := &Button{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(button)
	return button, nil
}

// SetLabel sets the button label.
func (btn *Button) SetLabel(label string) error {
	return btn.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Button{
			Button: &pb.ButtonProperties{
				Label: &label,
			},
		},
	})
}

// GetLabel returns the current label of the button.
func (btn *Button) GetLabel() (string, error) {
	props, err := btn.GetProperties()
	if err != nil {
		return "", err
	}
	if buttonProps := props.GetButton(); buttonProps != nil {
		if buttonProps.Label != nil {
			return *buttonProps.Label, nil
		}
	}
	return "", nil
}

