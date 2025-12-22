package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Flex represents a flex layout widget.
type Flex struct {
	baseWidget
}

// FlexBuilder provides a fluent interface for creating flex widgets.
type FlexBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewFlex creates a new flex builder.
func (c *Client) NewFlex() *FlexBuilder {
	return &FlexBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the flex title.
func (b *FlexBuilder) Title(title string) *FlexBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the flex has a border.
func (b *FlexBuilder) Border(border bool) *FlexBuilder {
	b.props.Border = &border
	return b
}

// Direction sets the flex direction (row or column).
func (b *FlexBuilder) Direction(direction pb.FlexDirection) *FlexBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Flex{
			Flex: &pb.FlexProperties{},
		}
	}
	b.props.GetFlex().Direction = &direction
	return b
}

// FullScreen sets whether the flex is full screen.
func (b *FlexBuilder) FullScreen(fullScreen bool) *FlexBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Flex{
			Flex: &pb.FlexProperties{},
		}
	}
	b.props.GetFlex().FullScreen = &fullScreen
	return b
}

// Build creates the flex widget on the server.
func (b *FlexBuilder) Build() (*Flex, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_FLEX,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	flex := &Flex{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(flex)
	return flex, nil
}

// AddItem adds a widget to the flex layout.
func (f *Flex) AddItem(widget Widget, proportion, fixedSize int, focus bool) error {
	_, err := f.client.layoutClient.FlexAddItem(f.client.ctx, &pb.FlexAddItemRequest{
		SessionId: f.client.sessionID,
		FlexId:    f.widgetID,
		ItemId:    &pb.WidgetId{Id: widget.ID()},
		Proportion: int32(proportion),
		FixedSize:  int32(fixedSize),
		Focus:      focus,
	})
	return err
}

// RemoveItem removes a widget from the flex layout.
func (f *Flex) RemoveItem(widget Widget) error {
	_, err := f.client.layoutClient.FlexRemoveItem(f.client.ctx, &pb.FlexRemoveItemRequest{
		SessionId: f.client.sessionID,
		FlexId:    f.widgetID,
		ItemId:    &pb.WidgetId{Id: widget.ID()},
	})
	return err
}

// SetDirection sets the flex direction.
func (f *Flex) SetDirection(direction pb.FlexDirection) error {
	_, err := f.client.layoutClient.FlexSetDirection(f.client.ctx, &pb.FlexSetDirectionRequest{
		SessionId: f.client.sessionID,
		FlexId:    f.widgetID,
		Direction: direction,
	})
	return err
}

// Grid represents a grid layout widget.
type Grid struct {
	baseWidget
}

// GridBuilder provides a fluent interface for creating grid widgets.
type GridBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewGrid creates a new grid builder.
func (c *Client) NewGrid() *GridBuilder {
	return &GridBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the grid title.
func (b *GridBuilder) Title(title string) *GridBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the grid has a border.
func (b *GridBuilder) Border(border bool) *GridBuilder {
	b.props.Border = &border
	return b
}

// Rows sets the number of rows.
func (b *GridBuilder) Rows(rows int) *GridBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Grid{
			Grid: &pb.GridProperties{},
		}
	}
	r := int32(rows)
	b.props.GetGrid().Rows = &r
	return b
}

// Columns sets the number of columns.
func (b *GridBuilder) Columns(columns int) *GridBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Grid{
			Grid: &pb.GridProperties{},
		}
	}
	c := int32(columns)
	b.props.GetGrid().Columns = &c
	return b
}

// MinWidth sets the minimum width.
func (b *GridBuilder) MinWidth(width int) *GridBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Grid{
			Grid: &pb.GridProperties{},
		}
	}
	w := int32(width)
	b.props.GetGrid().MinWidth = &w
	return b
}

// MinHeight sets the minimum height.
func (b *GridBuilder) MinHeight(height int) *GridBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Grid{
			Grid: &pb.GridProperties{},
		}
	}
	h := int32(height)
	b.props.GetGrid().MinHeight = &h
	return b
}

// Build creates the grid widget on the server.
func (b *GridBuilder) Build() (*Grid, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_GRID,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	grid := &Grid{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(grid)
	return grid, nil
}

// AddItem adds a widget to the grid layout.
func (g *Grid) AddItem(widget Widget, row, column, rowSpan, columnSpan, minWidth, minHeight int, focus bool) error {
	_, err := g.client.layoutClient.GridAddItem(g.client.ctx, &pb.GridAddItemRequest{
		SessionId:  g.client.sessionID,
		GridId:     g.widgetID,
		ItemId:     &pb.WidgetId{Id: widget.ID()},
		Row:        int32(row),
		Column:     int32(column),
		RowSpan:    int32(rowSpan),
		ColumnSpan: int32(columnSpan),
		MinWidth:   int32(minWidth),
		MinHeight:  int32(minHeight),
		Focus:      focus,
	})
	return err
}

// RemoveItem removes a widget from the grid layout.
func (g *Grid) RemoveItem(widget Widget) error {
	_, err := g.client.layoutClient.GridRemoveItem(g.client.ctx, &pb.GridRemoveItemRequest{
		SessionId: g.client.sessionID,
		GridId:    g.widgetID,
		ItemId:    &pb.WidgetId{Id: widget.ID()},
	})
	return err
}

// SetRows sets the row sizes (negative for proportional, positive for fixed).
func (g *Grid) SetRows(rowSizes []int32) error {
	_, err := g.client.layoutClient.GridSetRows(g.client.ctx, &pb.GridSetRowsRequest{
		SessionId: g.client.sessionID,
		GridId:    g.widgetID,
		RowSizes:  rowSizes,
	})
	return err
}

// SetColumns sets the column sizes (negative for proportional, positive for fixed).
func (g *Grid) SetColumns(columnSizes []int32) error {
	_, err := g.client.layoutClient.GridSetColumns(g.client.ctx, &pb.GridSetColumnsRequest{
		SessionId: g.client.sessionID,
		GridId:    g.widgetID,
		ColumnSizes: columnSizes,
	})
	return err
}

// Pages represents a pages layout widget.
type Pages struct {
	baseWidget
}

// PagesBuilder provides a fluent interface for creating pages widgets.
type PagesBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewPages creates a new pages builder.
func (c *Client) NewPages() *PagesBuilder {
	return &PagesBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the pages title.
func (b *PagesBuilder) Title(title string) *PagesBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the pages has a border.
func (b *PagesBuilder) Border(border bool) *PagesBuilder {
	b.props.Border = &border
	return b
}

// ShowPageNames sets whether to show page names.
func (b *PagesBuilder) ShowPageNames(show bool) *PagesBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Pages{
			Pages: &pb.PagesProperties{},
		}
	}
	b.props.GetPages().ShowPageNames = &show
	return b
}

// PageNameColor sets the page name color.
func (b *PagesBuilder) PageNameColor(color *pb.Color) *PagesBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Pages{
			Pages: &pb.PagesProperties{},
		}
	}
	b.props.GetPages().PageNameColor = color
	return b
}

// Build creates the pages widget on the server.
func (b *PagesBuilder) Build() (*Pages, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_PAGES,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	pages := &Pages{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(pages)
	return pages, nil
}

// AddPage adds a page to the pages widget.
func (p *Pages) AddPage(pageName string, widget Widget, resize, visible bool) error {
	_, err := p.client.layoutClient.PagesAddPage(p.client.ctx, &pb.PagesAddPageRequest{
		SessionId: p.client.sessionID,
		PagesId:   p.widgetID,
		PageName:  pageName,
		ItemId:    &pb.WidgetId{Id: widget.ID()},
		Resize:    resize,
		Visible:   visible,
	})
	return err
}

// RemovePage removes a page from the pages widget.
func (p *Pages) RemovePage(pageName string) error {
	_, err := p.client.layoutClient.PagesRemovePage(p.client.ctx, &pb.PagesRemovePageRequest{
		SessionId: p.client.sessionID,
		PagesId:   p.widgetID,
		PageName:  pageName,
	})
	return err
}

// ShowPage shows a specific page.
func (p *Pages) ShowPage(pageName string) error {
	_, err := p.client.layoutClient.PagesShowPage(p.client.ctx, &pb.PagesShowPageRequest{
		SessionId: p.client.sessionID,
		PagesId:   p.widgetID,
		PageName:  pageName,
	})
	return err
}

// GetCurrentPage gets the currently visible page name.
func (p *Pages) GetCurrentPage() (string, error) {
	resp, err := p.client.layoutClient.PagesGetCurrentPage(p.client.ctx, &pb.PagesGetCurrentPageRequest{
		SessionId: p.client.sessionID,
		PagesId:   p.widgetID,
	})
	if err != nil {
		return "", err
	}
	return resp.PageName, nil
}

