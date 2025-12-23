package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// TextArea represents a multi-line text editor widget.
type TextArea struct {
	baseWidget
}

// TextAreaBuilder provides a fluent interface for creating text area widgets.
type TextAreaBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewTextArea creates a new text area builder.
func (c *Client) NewTextArea() *TextAreaBuilder {
	return &TextAreaBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the text area's title (border title).
func (b *TextAreaBuilder) Title(title string) *TextAreaBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the text area has a border.
func (b *TextAreaBuilder) Border(border bool) *TextAreaBuilder {
	b.props.Border = &border
	return b
}

// Text sets the initial text content.
func (b *TextAreaBuilder) Text(text string) *TextAreaBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{},
		}
	}
	b.props.GetTextArea().Text = &text
	return b
}

// Placeholder sets the placeholder text.
func (b *TextAreaBuilder) Placeholder(placeholder string) *TextAreaBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{},
		}
	}
	b.props.GetTextArea().Placeholder = &placeholder
	return b
}

// MaxLength sets the maximum text length.
func (b *TextAreaBuilder) MaxLength(maxLength int) *TextAreaBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{},
		}
	}
	ml := int32(maxLength)
	b.props.GetTextArea().MaxLength = &ml
	return b
}

// WordWrap sets whether text wraps at word boundaries.
func (b *TextAreaBuilder) WordWrap(wrap bool) *TextAreaBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{},
		}
	}
	b.props.GetTextArea().WordWrap = &wrap
	return b
}

// TextColor sets the text color.
func (b *TextAreaBuilder) TextColor(color *pb.Color) *TextAreaBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{},
		}
	}
	b.props.GetTextArea().TextColor = color
	return b
}

// PlaceholderColor sets the placeholder text color.
func (b *TextAreaBuilder) PlaceholderColor(color *pb.Color) *TextAreaBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{},
		}
	}
	b.props.GetTextArea().PlaceholderColor = color
	return b
}

// Build creates the text area widget on the server.
func (b *TextAreaBuilder) Build() (*TextArea, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_TEXT_AREA,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	textArea := &TextArea{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(textArea)
	return textArea, nil
}

// SetText sets the text area content.
func (ta *TextArea) SetText(text string) error {
	return ta.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{
				Text: &text,
			},
		},
	})
}

// GetText returns the current text content.
func (ta *TextArea) GetText() (string, error) {
	props, err := ta.GetProperties()
	if err != nil {
		return "", err
	}
	if taProps := props.GetTextArea(); taProps != nil {
		if taProps.Text != nil {
			return *taProps.Text, nil
		}
	}
	return "", nil
}

// SetPlaceholder sets the placeholder text.
func (ta *TextArea) SetPlaceholder(placeholder string) error {
	return ta.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{
				Placeholder: &placeholder,
			},
		},
	})
}

// GetPlaceholder returns the current placeholder text.
func (ta *TextArea) GetPlaceholder() (string, error) {
	props, err := ta.GetProperties()
	if err != nil {
		return "", err
	}
	if taProps := props.GetTextArea(); taProps != nil {
		if taProps.Placeholder != nil {
			return *taProps.Placeholder, nil
		}
	}
	return "", nil
}
