package client

import (
	pb "industries/loosh/yutani/pkg/proto/industries/loosh/yutani/v1"
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

// Visible sets whether the list is visible.
func (b *ListBuilder) Visible(visible bool) *ListBuilder {
	b.props.Visible = &visible
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
	resp, err := l.client.listClient.AddItem(l.client.ctx, &pb.AddItemRequest{
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
	_, err := l.client.listClient.RemoveItem(l.client.ctx, &pb.RemoveItemRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
		Index:     int32(index),
	})
	return err
}

// Clear removes all items from the list.
func (l *List) Clear() error {
	_, err := l.client.listClient.Clear(l.client.ctx, &pb.ClearListRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
	})
	return err
}

// GetItemCount returns the number of items in the list.
func (l *List) GetItemCount() (int, error) {
	resp, err := l.client.listClient.GetItemCount(l.client.ctx, &pb.GetItemCountRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

// GetSelected returns the index of the currently selected item.
func (l *List) GetSelected() (int, error) {
	resp, err := l.client.listClient.GetSelected(l.client.ctx, &pb.GetSelectedRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
	})
	if err != nil {
		return -1, err
	}
	return int(resp.Index), nil
}

// SetSelected sets the selected item by index.
func (l *List) SetSelected(index int) error {
	_, err := l.client.listClient.SetSelected(l.client.ctx, &pb.SetSelectedRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
		Index:     int32(index),
	})
	return err
}

// GetItem returns an item's text by index.
func (l *List) GetItem(index int) (mainText, secondaryText, shortcut string, err error) {
	resp, err := l.client.listClient.GetItem(l.client.ctx, &pb.GetItemRequest{
		SessionId: l.client.sessionID,
		WidgetId:  l.widgetID,
		Index:     int32(index),
	})
	if err != nil {
		return "", "", "", err
	}
	return resp.MainText, resp.SecondaryText, resp.Shortcut, nil
}

