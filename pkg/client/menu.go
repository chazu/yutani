package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// MenuBar represents a menu bar widget.
type MenuBar struct {
	baseWidget
}

// MenuBarBuilder provides a fluent interface for creating menu bar widgets.
type MenuBarBuilder struct {
	client *Client
	props  *pb.WidgetProperties
	mbP    *pb.MenuBarProperties
}

// NewMenuBar creates a new menu bar builder.
func (c *Client) NewMenuBar() *MenuBarBuilder {
	return &MenuBarBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
		mbP:    &pb.MenuBarProperties{},
	}
}

// Title sets the menu bar title.
func (b *MenuBarBuilder) Title(title string) *MenuBarBuilder {
	b.props.Title = &title
	return b
}

// BackgroundColor sets the menu bar background color.
func (b *MenuBarBuilder) BackgroundColor(color *pb.Color) *MenuBarBuilder {
	b.mbP.BackgroundColor = color
	return b
}

// TextColor sets the menu bar text color.
func (b *MenuBarBuilder) TextColor(color *pb.Color) *MenuBarBuilder {
	b.mbP.TextColor = color
	return b
}

// ActiveBackgroundColor sets the active menu title background.
func (b *MenuBarBuilder) ActiveBackgroundColor(color *pb.Color) *MenuBarBuilder {
	b.mbP.ActiveBackgroundColor = color
	return b
}

// ActiveTextColor sets the active menu title text color.
func (b *MenuBarBuilder) ActiveTextColor(color *pb.Color) *MenuBarBuilder {
	b.mbP.ActiveTextColor = color
	return b
}

// Build creates the menu bar widget on the server.
func (b *MenuBarBuilder) Build() (*MenuBar, error) {
	b.props.TypeProperties = &pb.WidgetProperties_MenuBar{MenuBar: b.mbP}

	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_MENU_BAR,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	mb := &MenuBar{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}
	b.client.registerWidget(mb)
	return mb, nil
}

// AddMenu adds a menu to the menu bar.
func (mb *MenuBar) AddMenu(menu *Menu) error {
	_, err := mb.client.menuClient.AddMenu(mb.client.ctx, &pb.MenuAddMenuRequest{
		SessionId: mb.client.sessionID,
		MenuBarId: mb.widgetID,
		MenuId:    menu.widgetID,
	})
	return err
}

// Menu represents a menu widget (container for menu items).
type Menu struct {
	baseWidget
}

// MenuBuilder provides a fluent interface for creating menu widgets.
type MenuBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewMenu creates a new menu builder.
func (c *Client) NewMenu() *MenuBuilder {
	return &MenuBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the menu title (displayed in the menu bar).
func (b *MenuBuilder) Title(title string) *MenuBuilder {
	b.props.TypeProperties = &pb.WidgetProperties_Menu{
		Menu: &pb.MenuProperties{Title: &title},
	}
	// Also set common title for fallback
	b.props.Title = &title
	return b
}

// Build creates the menu widget on the server.
func (b *MenuBuilder) Build() (*Menu, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_MENU,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	menu := &Menu{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}
	b.client.registerWidget(menu)
	return menu, nil
}

// AddItem adds a menu item to the menu.
func (m *Menu) AddItem(item *MenuItem) error {
	_, err := m.client.menuClient.AddMenuItem(m.client.ctx, &pb.MenuAddMenuItemRequest{
		SessionId:  m.client.sessionID,
		MenuId:     m.widgetID,
		MenuItemId: item.widgetID,
	})
	return err
}

// RemoveItem removes a menu item from the menu.
func (m *Menu) RemoveItem(item *MenuItem) error {
	_, err := m.client.menuClient.RemoveMenuItem(m.client.ctx, &pb.MenuRemoveMenuItemRequest{
		SessionId:  m.client.sessionID,
		MenuId:     m.widgetID,
		MenuItemId: item.widgetID,
	})
	return err
}

// MenuItem represents a menu item widget.
type MenuItem struct {
	baseWidget
}

// MenuItemBuilder provides a fluent interface for creating menu item widgets.
type MenuItemBuilder struct {
	client *Client
	props  *pb.WidgetProperties
	miP    *pb.MenuItemProperties
}

// NewMenuItem creates a new menu item builder.
func (c *Client) NewMenuItem() *MenuItemBuilder {
	return &MenuItemBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
		miP:    &pb.MenuItemProperties{},
	}
}

// Label sets the menu item label.
func (b *MenuItemBuilder) Label(label string) *MenuItemBuilder {
	b.miP.Label = &label
	return b
}

// ShortcutDisplay sets the shortcut display text (cosmetic only).
func (b *MenuItemBuilder) ShortcutDisplay(shortcut string) *MenuItemBuilder {
	b.miP.ShortcutDisplay = &shortcut
	return b
}

// Separator marks this item as a separator.
func (b *MenuItemBuilder) Separator() *MenuItemBuilder {
	sep := true
	b.miP.Separator = &sep
	return b
}

// Disabled marks this item as disabled.
func (b *MenuItemBuilder) Disabled() *MenuItemBuilder {
	dis := true
	b.miP.Disabled = &dis
	return b
}

// Build creates the menu item widget on the server.
func (b *MenuItemBuilder) Build() (*MenuItem, error) {
	b.props.TypeProperties = &pb.WidgetProperties_MenuItem{MenuItem: b.miP}

	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_MENU_ITEM,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	item := &MenuItem{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}
	b.client.registerWidget(item)
	return item, nil
}
