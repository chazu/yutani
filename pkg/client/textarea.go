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

// SuppressedKeys sets keys that are suppressed from widget processing but still
// emitted as events. Format: "ctrl+d", "ctrl+p", "ctrl+i", "ctrl+space", etc.
func (b *TextAreaBuilder) SuppressedKeys(keys ...string) *TextAreaBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{},
		}
	}
	b.props.GetTextArea().SuppressedKeys = keys
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

// CursorPosition holds cursor and optional selection endpoint coordinates.
type CursorPosition struct {
	Row          int
	Column       int
	ToRow        int
	ToColumn     int
	HasSelection bool
}

// GetCursorPosition returns the current cursor position.
func (ta *TextArea) GetCursorPosition() (*CursorPosition, error) {
	resp, err := ta.client.textAreaClient.GetCursorPosition(ta.client.ctx, &pb.TextAreaGetCursorRequest{
		SessionId: ta.client.sessionID,
		WidgetId:  ta.widgetID,
	})
	if err != nil {
		return nil, err
	}
	return &CursorPosition{
		Row:          int(resp.Row),
		Column:       int(resp.Column),
		ToRow:        int(resp.ToRow),
		ToColumn:     int(resp.ToColumn),
		HasSelection: resp.HasSelection,
	}, nil
}

// SetCursorPosition moves the cursor to the specified row and column.
func (ta *TextArea) SetCursorPosition(row, column int) error {
	_, err := ta.client.textAreaClient.SetCursorPosition(ta.client.ctx, &pb.TextAreaSetCursorRequest{
		SessionId: ta.client.sessionID,
		WidgetId:  ta.widgetID,
		Row:       int32(row),
		Column:    int32(column),
	})
	return err
}

// Selection holds selection state including text, byte offsets, and row/column coordinates.
type Selection struct {
	HasSelection bool
	SelectedText string
	Start        int
	End          int
	FromRow      int
	FromColumn   int
	ToRow        int
	ToColumn     int
}

// GetSelection returns the current selection range and text.
func (ta *TextArea) GetSelection() (*Selection, error) {
	resp, err := ta.client.textAreaClient.GetSelection(ta.client.ctx, &pb.TextAreaGetSelectionRequest{
		SessionId: ta.client.sessionID,
		WidgetId:  ta.widgetID,
	})
	if err != nil {
		return nil, err
	}
	return &Selection{
		HasSelection: resp.HasSelection,
		SelectedText: resp.SelectedText,
		Start:        int(resp.Start),
		End:          int(resp.End),
		FromRow:      int(resp.FromRow),
		FromColumn:   int(resp.FromColumn),
		ToRow:        int(resp.ToRow),
		ToColumn:     int(resp.ToColumn),
	}, nil
}

// SetSelection sets the selection range by byte offsets.
func (ta *TextArea) SetSelection(start, end int) error {
	_, err := ta.client.textAreaClient.SetSelection(ta.client.ctx, &pb.TextAreaSetSelectionRequest{
		SessionId: ta.client.sessionID,
		WidgetId:  ta.widgetID,
		Start:     int32(start),
		End:       int32(end),
	})
	return err
}

// HasSelection checks if text is currently selected.
func (ta *TextArea) HasSelection() (bool, error) {
	resp, err := ta.client.textAreaClient.HasSelection(ta.client.ctx, &pb.TextAreaHasSelectionRequest{
		SessionId: ta.client.sessionID,
		WidgetId:  ta.widgetID,
	})
	if err != nil {
		return false, err
	}
	return resp.HasSelection, nil
}

// GetSelectedText returns the currently selected text.
func (ta *TextArea) GetSelectedText() (string, error) {
	resp, err := ta.client.textAreaClient.GetSelectedText(ta.client.ctx, &pb.TextAreaGetSelectedTextRequest{
		SessionId: ta.client.sessionID,
		WidgetId:  ta.widgetID,
	})
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

// InsertTextAtCursor inserts text at the current cursor position, replacing any selection.
func (ta *TextArea) InsertTextAtCursor(text string) error {
	_, err := ta.client.textAreaClient.InsertTextAtCursor(ta.client.ctx, &pb.TextAreaInsertTextRequest{
		SessionId: ta.client.sessionID,
		WidgetId:  ta.widgetID,
		Text:      text,
	})
	return err
}

// SetSuppressedKeys configures keys that are suppressed from widget processing
// but still emitted as events. Format: "ctrl+d", "ctrl+p", "ctrl+i", etc.
func (ta *TextArea) SetSuppressedKeys(keys []string) error {
	return ta.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{
				SuppressedKeys: keys,
			},
		},
	})
}
