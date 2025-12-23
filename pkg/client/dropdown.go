package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Dropdown represents a dropdown widget.
type Dropdown struct {
	baseWidget
}

// DropdownBuilder provides a fluent interface for creating dropdown widgets.
type DropdownBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewDropdown creates a new dropdown builder.
func (c *Client) NewDropdown() *DropdownBuilder {
	return &DropdownBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the dropdown's title (border title).
func (b *DropdownBuilder) Title(title string) *DropdownBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the dropdown has a border.
func (b *DropdownBuilder) Border(border bool) *DropdownBuilder {
	b.props.Border = &border
	return b
}

// Label sets the dropdown label text.
func (b *DropdownBuilder) Label(label string) *DropdownBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{},
		}
	}
	b.props.GetDropdown().Label = &label
	return b
}

// Options sets the dropdown options.
func (b *DropdownBuilder) Options(options []string) *DropdownBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{},
		}
	}
	b.props.GetDropdown().Options = options
	return b
}

// SelectedIndex sets the initially selected index.
func (b *DropdownBuilder) SelectedIndex(index int) *DropdownBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{},
		}
	}
	idx := int32(index)
	b.props.GetDropdown().SelectedIndex = &idx
	return b
}

// FieldWidth sets the field width.
func (b *DropdownBuilder) FieldWidth(width int) *DropdownBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{},
		}
	}
	w := int32(width)
	b.props.GetDropdown().FieldWidth = &w
	return b
}

// LabelColor sets the label color.
func (b *DropdownBuilder) LabelColor(color *pb.Color) *DropdownBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{},
		}
	}
	b.props.GetDropdown().LabelColor = color
	return b
}

// FieldTextColor sets the field text color.
func (b *DropdownBuilder) FieldTextColor(color *pb.Color) *DropdownBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{},
		}
	}
	b.props.GetDropdown().FieldTextColor = color
	return b
}

// FieldBackgroundColor sets the field background color.
func (b *DropdownBuilder) FieldBackgroundColor(color *pb.Color) *DropdownBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{},
		}
	}
	b.props.GetDropdown().FieldBackgroundColor = color
	return b
}

// Build creates the dropdown widget on the server.
func (b *DropdownBuilder) Build() (*Dropdown, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_DROPDOWN,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	dropdown := &Dropdown{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(dropdown)
	return dropdown, nil
}

// SetSelectedIndex sets the selected index.
func (dd *Dropdown) SetSelectedIndex(index int) error {
	idx := int32(index)
	return dd.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{
				SelectedIndex: &idx,
			},
		},
	})
}

// SetOptions sets the dropdown options.
func (dd *Dropdown) SetOptions(options []string) error {
	return dd.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{
				Options: options,
			},
		},
	})
}

// SetLabel sets the dropdown label.
func (dd *Dropdown) SetLabel(label string) error {
	return dd.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Dropdown{
			Dropdown: &pb.DropdownProperties{
				Label: &label,
			},
		},
	})
}

// GetSelectedIndex returns the currently selected index.
func (dd *Dropdown) GetSelectedIndex() (int, error) {
	props, err := dd.GetProperties()
	if err != nil {
		return -1, err
	}
	if ddProps := props.GetDropdown(); ddProps != nil {
		if ddProps.SelectedIndex != nil {
			return int(*ddProps.SelectedIndex), nil
		}
	}
	return -1, nil
}

// GetSelectedText returns the currently selected text.
func (dd *Dropdown) GetSelectedText() (string, error) {
	props, err := dd.GetProperties()
	if err != nil {
		return "", err
	}
	if ddProps := props.GetDropdown(); ddProps != nil {
		if ddProps.SelectedText != nil {
			return *ddProps.SelectedText, nil
		}
		// Fallback: get text from options by index
		if ddProps.SelectedIndex != nil && len(ddProps.Options) > int(*ddProps.SelectedIndex) {
			return ddProps.Options[*ddProps.SelectedIndex], nil
		}
	}
	return "", nil
}

// GetOptions returns the dropdown options.
func (dd *Dropdown) GetOptions() ([]string, error) {
	props, err := dd.GetProperties()
	if err != nil {
		return nil, err
	}
	if ddProps := props.GetDropdown(); ddProps != nil {
		return ddProps.Options, nil
	}
	return nil, nil
}

// GetLabel returns the current label of the dropdown.
func (dd *Dropdown) GetLabel() (string, error) {
	props, err := dd.GetProperties()
	if err != nil {
		return "", err
	}
	if ddProps := props.GetDropdown(); ddProps != nil {
		if ddProps.Label != nil {
			return *ddProps.Label, nil
		}
	}
	return "", nil
}
