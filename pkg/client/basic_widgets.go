package client

import (
	pb "industries/loosh/yutani/pkg/proto/industries/loosh/yutani/v1"
)

// Box represents a box widget.
type Box struct {
	baseWidget
}

// BoxBuilder provides a fluent interface for creating box widgets.
type BoxBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewBox creates a new box builder.
func (c *Client) NewBox() *BoxBuilder {
	return &BoxBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the box title.
func (b *BoxBuilder) Title(title string) *BoxBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the box has a border.
func (b *BoxBuilder) Border(border bool) *BoxBuilder {
	b.props.Border = &border
	return b
}

// Visible sets whether the box is visible.
func (b *BoxBuilder) Visible(visible bool) *BoxBuilder {
	b.props.Visible = &visible
	return b
}

// BackgroundColor sets the background color.
func (b *BoxBuilder) BackgroundColor(color *pb.Color) *BoxBuilder {
	b.props.BackgroundColor = color
	return b
}

// Build creates the box widget on the server.
func (b *BoxBuilder) Build() (*Box, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_BOX,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	box := &Box{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(box)
	return box, nil
}

// TextView represents a text view widget.
type TextView struct {
	baseWidget
}

// TextViewBuilder provides a fluent interface for creating text view widgets.
type TextViewBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewTextView creates a new text view builder.
func (c *Client) NewTextView() *TextViewBuilder {
	return &TextViewBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the text view title.
func (b *TextViewBuilder) Title(title string) *TextViewBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the text view has a border.
func (b *TextViewBuilder) Border(border bool) *TextViewBuilder {
	b.props.Border = &border
	return b
}

// Text sets the initial text content.
func (b *TextViewBuilder) Text(text string) *TextViewBuilder {
	b.props.TypeSpecific = &pb.WidgetProperties_TextView{
		TextView: &pb.TextViewProperties{
			Text: &text,
		},
	}
	return b
}

// WordWrap sets whether text wraps.
func (b *TextViewBuilder) WordWrap(wrap bool) *TextViewBuilder {
	if b.props.TypeSpecific == nil {
		b.props.TypeSpecific = &pb.WidgetProperties_TextView{
			TextView: &pb.TextViewProperties{},
		}
	}
	b.props.GetTextView().WordWrap = &wrap
	return b
}

// DynamicColors enables dynamic color tags.
func (b *TextViewBuilder) DynamicColors(enabled bool) *TextViewBuilder {
	if b.props.TypeSpecific == nil {
		b.props.TypeSpecific = &pb.WidgetProperties_TextView{
			TextView: &pb.TextViewProperties{},
		}
	}
	b.props.GetTextView().DynamicColors = &enabled
	return b
}

// Build creates the text view widget on the server.
func (b *TextViewBuilder) Build() (*TextView, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_TEXT_VIEW,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	textView := &TextView{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(textView)
	return textView, nil
}

// SetText sets the text content.
func (tv *TextView) SetText(text string) error {
	return tv.setProperty(&pb.WidgetProperties{
		TypeSpecific: &pb.WidgetProperties_TextView{
			TextView: &pb.TextViewProperties{
				Text: &text,
			},
		},
	})
}

