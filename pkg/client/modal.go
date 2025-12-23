package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Modal represents a modal dialog widget.
type Modal struct {
	baseWidget
}

// ModalBuilder provides a fluent interface for creating modal widgets.
type ModalBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewModal creates a new modal builder.
func (c *Client) NewModal() *ModalBuilder {
	return &ModalBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the modal's title (border title).
func (b *ModalBuilder) Title(title string) *ModalBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the modal has a border.
func (b *ModalBuilder) Border(border bool) *ModalBuilder {
	b.props.Border = &border
	return b
}

// Text sets the modal message text.
func (b *ModalBuilder) Text(text string) *ModalBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{},
		}
	}
	b.props.GetModal().Text = &text
	return b
}

// Buttons sets the modal buttons.
func (b *ModalBuilder) Buttons(buttons ...string) *ModalBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{},
		}
	}
	b.props.GetModal().Buttons = buttons
	return b
}

// TextColor sets the modal text color.
func (b *ModalBuilder) TextColor(color *pb.Color) *ModalBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{},
		}
	}
	b.props.GetModal().TextColor = color
	return b
}

// BackgroundColor sets the modal background color.
func (b *ModalBuilder) BackgroundColor(color *pb.Color) *ModalBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{},
		}
	}
	b.props.GetModal().BackgroundColor = color
	return b
}

// ButtonBackgroundColor sets the button background color.
func (b *ModalBuilder) ButtonBackgroundColor(color *pb.Color) *ModalBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{},
		}
	}
	b.props.GetModal().ButtonBackgroundColor = color
	return b
}

// ButtonTextColor sets the button text color.
func (b *ModalBuilder) ButtonTextColor(color *pb.Color) *ModalBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{},
		}
	}
	b.props.GetModal().ButtonTextColor = color
	return b
}

// Build creates the modal widget on the server.
func (b *ModalBuilder) Build() (*Modal, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_MODAL,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	modal := &Modal{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(modal)
	return modal, nil
}

// SetText sets the modal message text.
func (m *Modal) SetText(text string) error {
	return m.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{
				Text: &text,
			},
		},
	})
}

// GetText returns the current modal message text.
func (m *Modal) GetText() (string, error) {
	props, err := m.GetProperties()
	if err != nil {
		return "", err
	}
	if modalProps := props.GetModal(); modalProps != nil {
		if modalProps.Text != nil {
			return *modalProps.Text, nil
		}
	}
	return "", nil
}

// SetButtons sets the modal buttons.
func (m *Modal) SetButtons(buttons ...string) error {
	return m.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{
				Buttons: buttons,
			},
		},
	})
}

// GetButtons returns the current modal buttons.
func (m *Modal) GetButtons() ([]string, error) {
	props, err := m.GetProperties()
	if err != nil {
		return nil, err
	}
	if modalProps := props.GetModal(); modalProps != nil {
		return modalProps.Buttons, nil
	}
	return nil, nil
}
