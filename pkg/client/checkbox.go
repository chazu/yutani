package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Checkbox represents a checkbox widget.
type Checkbox struct {
	baseWidget
}

// CheckboxBuilder provides a fluent interface for creating checkbox widgets.
type CheckboxBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewCheckbox creates a new checkbox builder.
func (c *Client) NewCheckbox() *CheckboxBuilder {
	return &CheckboxBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the checkbox's title (border title).
func (b *CheckboxBuilder) Title(title string) *CheckboxBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the checkbox has a border.
func (b *CheckboxBuilder) Border(border bool) *CheckboxBuilder {
	b.props.Border = &border
	return b
}

// Label sets the checkbox label text.
func (b *CheckboxBuilder) Label(label string) *CheckboxBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Checkbox{
			Checkbox: &pb.CheckboxProperties{},
		}
	}
	b.props.GetCheckbox().Label = &label
	return b
}

// Checked sets the initial checked state.
func (b *CheckboxBuilder) Checked(checked bool) *CheckboxBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Checkbox{
			Checkbox: &pb.CheckboxProperties{},
		}
	}
	b.props.GetCheckbox().Checked = &checked
	return b
}

// LabelColor sets the checkbox label color.
func (b *CheckboxBuilder) LabelColor(color *pb.Color) *CheckboxBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Checkbox{
			Checkbox: &pb.CheckboxProperties{},
		}
	}
	b.props.GetCheckbox().LabelColor = color
	return b
}

// CheckedColor sets the checkbox color when checked.
func (b *CheckboxBuilder) CheckedColor(color *pb.Color) *CheckboxBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Checkbox{
			Checkbox: &pb.CheckboxProperties{},
		}
	}
	b.props.GetCheckbox().CheckedColor = color
	return b
}

// Build creates the checkbox widget on the server.
func (b *CheckboxBuilder) Build() (*Checkbox, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_CHECKBOX,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	checkbox := &Checkbox{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(checkbox)
	return checkbox, nil
}

// SetChecked sets the checkbox checked state.
func (cb *Checkbox) SetChecked(checked bool) error {
	return cb.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Checkbox{
			Checkbox: &pb.CheckboxProperties{
				Checked: &checked,
			},
		},
	})
}

// SetLabel sets the checkbox label.
func (cb *Checkbox) SetLabel(label string) error {
	return cb.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Checkbox{
			Checkbox: &pb.CheckboxProperties{
				Label: &label,
			},
		},
	})
}

// IsChecked returns the current checked state of the checkbox.
func (cb *Checkbox) IsChecked() (bool, error) {
	props, err := cb.GetProperties()
	if err != nil {
		return false, err
	}
	if checkboxProps := props.GetCheckbox(); checkboxProps != nil {
		if checkboxProps.Checked != nil {
			return *checkboxProps.Checked, nil
		}
	}
	return false, nil
}

// GetLabel returns the current label of the checkbox.
func (cb *Checkbox) GetLabel() (string, error) {
	props, err := cb.GetProperties()
	if err != nil {
		return "", err
	}
	if checkboxProps := props.GetCheckbox(); checkboxProps != nil {
		if checkboxProps.Label != nil {
			return *checkboxProps.Label, nil
		}
	}
	return "", nil
}

