package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// InputField represents an input field widget.
type InputField struct {
	baseWidget
}

// InputFieldBuilder provides a fluent interface for creating input field widgets.
type InputFieldBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewInputField creates a new input field builder.
func (c *Client) NewInputField() *InputFieldBuilder {
	return &InputFieldBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the input field's title (border title).
func (b *InputFieldBuilder) Title(title string) *InputFieldBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the input field has a border.
func (b *InputFieldBuilder) Border(border bool) *InputFieldBuilder {
	b.props.Border = &border
	return b
}

// Label sets the input field label text.
func (b *InputFieldBuilder) Label(label string) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	b.props.GetInputField().Label = &label
	return b
}

// Placeholder sets the placeholder text.
func (b *InputFieldBuilder) Placeholder(placeholder string) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	b.props.GetInputField().Placeholder = &placeholder
	return b
}

// Text sets the initial text content.
func (b *InputFieldBuilder) Text(text string) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	b.props.GetInputField().Text = &text
	return b
}

// FieldWidth sets the field width.
func (b *InputFieldBuilder) FieldWidth(width int) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	w := int32(width)
	b.props.GetInputField().FieldWidth = &w
	return b
}

// Masked sets whether the input is masked (password field).
func (b *InputFieldBuilder) Masked(masked bool) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	b.props.GetInputField().Masked = &masked
	return b
}

// LabelColor sets the label color.
func (b *InputFieldBuilder) LabelColor(color *pb.Color) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	b.props.GetInputField().LabelColor = color
	return b
}

// FieldTextColor sets the field text color.
func (b *InputFieldBuilder) FieldTextColor(color *pb.Color) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	b.props.GetInputField().FieldTextColor = color
	return b
}

// FieldBackgroundColor sets the field background color.
func (b *InputFieldBuilder) FieldBackgroundColor(color *pb.Color) *InputFieldBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{},
		}
	}
	b.props.GetInputField().FieldBackgroundColor = color
	return b
}

// Build creates the input field widget on the server.
func (b *InputFieldBuilder) Build() (*InputField, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_INPUT_FIELD,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	inputField := &InputField{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(inputField)
	return inputField, nil
}

// SetText sets the input field text.
func (i *InputField) SetText(text string) error {
	return i.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{
				Text: &text,
			},
		},
	})
}

// SetLabel sets the input field label.
func (i *InputField) SetLabel(label string) error {
	return i.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_InputField{
			InputField: &pb.InputFieldProperties{
				Label: &label,
			},
		},
	})
}

// GetText returns the current text content of the input field.
func (i *InputField) GetText() (string, error) {
	props, err := i.GetProperties()
	if err != nil {
		return "", err
	}
	if inputProps := props.GetInputField(); inputProps != nil {
		if inputProps.Text != nil {
			return *inputProps.Text, nil
		}
	}
	return "", nil
}

// GetLabel returns the current label of the input field.
func (i *InputField) GetLabel() (string, error) {
	props, err := i.GetProperties()
	if err != nil {
		return "", err
	}
	if inputProps := props.GetInputField(); inputProps != nil {
		if inputProps.Label != nil {
			return *inputProps.Label, nil
		}
	}
	return "", nil
}

