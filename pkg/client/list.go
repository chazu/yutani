package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// List represents a list widget.
type List struct {
	baseWidget
}

// ListBuilder provides a fluent interface for creating list widgets.
type ListBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewList creates a new list builder.
func (c *Client) NewList() *ListBuilder {
	return &ListBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the list title.
func (b *ListBuilder) Title(title string) *ListBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the list has a border.
func (b *ListBuilder) Border(border bool) *ListBuilder {
	b.props.Border = &border
	return b
}

// BackgroundColor sets the background color.
func (b *ListBuilder) BackgroundColor(color *pb.Color) *ListBuilder {
	b.props.BackgroundColor = color
	return b
}

// BorderColor sets the border color.
func (b *ListBuilder) BorderColor(color *pb.Color) *ListBuilder {
	b.props.BorderColor = color
	return b
}

// TitleColor sets the title color.
func (b *ListBuilder) TitleColor(color *pb.Color) *ListBuilder {
	b.props.TitleColor = color
	return b
}

// Build creates the list widget on the server.
func (b *ListBuilder) Build() (*List, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_LIST,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	list := &List{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(list)
	return list, nil
}

// AddItem adds an item to the list.
func (l *List) AddItem(mainText, secondaryText string, shortcut *string) (int, error) {
	resp, err := l.client.listClient.AddItem(l.client.ctx, &pb.ListAddItemRequest{
		SessionId:     l.client.sessionID,
		WidgetId:      l.widgetID,
		MainText:      mainText,
		SecondaryText: secondaryText,
		Shortcut:      shortcut,
	})
	if err != nil {
		return -1, err
	}
	return int(resp.Index), nil
}

// RemoveItem removes an item from the list by index.
func (l *List) RemoveItem(index int) error {
	_, err := l.client.listClient.RemoveItem(l.client.ctx, &pb.ListRemoveItemRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
		Index:     int32(index),
	})
	return err
}

// Clear removes all items from the list.
func (l *List) Clear() error {
	_, err := l.client.listClient.Clear(l.client.ctx, &pb.ListClearRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
	})
	return err
}

// GetItemCount returns the number of items in the list.
func (l *List) GetItemCount() (int, error) {
	resp, err := l.client.listClient.GetItemCount(l.client.ctx, &pb.ListGetItemCountRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

// GetSelection returns the index of the currently selected item.
func (l *List) GetSelection() (int, error) {
	resp, err := l.client.listClient.GetSelection(l.client.ctx, &pb.ListGetSelectionRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
	})
	if err != nil {
		return -1, err
	}
	return int(resp.Index), nil
}

// SetSelection sets the selected item by index.
func (l *List) SetSelection(index int) error {
	_, err := l.client.listClient.SetSelection(l.client.ctx, &pb.ListSetSelectionRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
		Index:     int32(index),
	})
	return err
}

// GetItem returns an item's text by index.
func (l *List) GetItem(index int) (mainText, secondaryText, shortcut string, err error) {
	resp, err := l.client.listClient.GetItem(l.client.ctx, &pb.ListGetItemRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
		Index:     int32(index),
	})
	if err != nil {
		return "", "", "", err
	}
	return resp.MainText, resp.SecondaryText, resp.Shortcut, nil
}

