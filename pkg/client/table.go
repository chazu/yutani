package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Table represents a table widget.
type Table struct {
	baseWidget
}

// TableBuilder provides a fluent interface for creating table widgets.
type TableBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewTable creates a new table builder.
func (c *Client) NewTable() *TableBuilder {
	return &TableBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the table title.
func (b *TableBuilder) Title(title string) *TableBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the table has a border.
func (b *TableBuilder) Border(border bool) *TableBuilder {
	b.props.Border = &border
	return b
}

// Build creates the table widget on the server.
func (b *TableBuilder) Build() (*Table, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_TABLE,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	table := &Table{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(table)
	return table, nil
}

// SetCell sets a single cell's content.
func (t *Table) SetCell(row, column int, cell *pb.TableCell) error {
	_, err := t.client.tableClient.SetCell(t.client.ctx, &pb.SetTableCellRequest{
		SessionId: t.client.sessionID,
		WidgetId:  t.widgetID,
		Row:       int32(row),
		Column:    int32(column),
		Cell:      cell,
	})
	return err
}

// GetCell gets a single cell's content.
func (t *Table) GetCell(row, column int) (*pb.TableCell, error) {
	resp, err := t.client.tableClient.GetCell(t.client.ctx, &pb.GetTableCellRequest{
		SessionId: t.client.sessionID,
		WidgetId:  t.widgetID,
		Row:       int32(row),
		Column:    int32(column),
	})
	if err != nil {
		return nil, err
	}
	return resp.Cell, nil
}

// SetCells sets multiple cells at once.
func (t *Table) SetCells(cells []*pb.TableCellUpdate) (int, error) {
	resp, err := t.client.tableClient.SetCells(t.client.ctx, &pb.SetTableCellsRequest{
		SessionId: t.client.sessionID,
		WidgetId:  t.widgetID,
		Cells:     cells,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.CellsSet), nil
}

// Clear clears the entire table.
func (t *Table) Clear() error {
	_, err := t.client.tableClient.Clear(t.client.ctx, &pb.TableClearRequest{
		SessionId: t.client.sessionID,
		WidgetId:  t.widgetID,
	})
	return err
}

// GetDimensions returns the table dimensions.
func (t *Table) GetDimensions() (rows, columns int, err error) {
	resp, err := t.client.tableClient.GetDimensions(t.client.ctx, &pb.TableGetDimensionsRequest{
		SessionId: t.client.sessionID,
		WidgetId:  t.widgetID,
	})
	if err != nil {
		return 0, 0, err
	}
	return int(resp.Rows), int(resp.Columns), nil
}

// GetSelection returns the current cell selection.
func (t *Table) GetSelection() (row, column int, err error) {
	resp, err := t.client.tableClient.GetSelection(t.client.ctx, &pb.TableGetSelectionRequest{
		SessionId: t.client.sessionID,
		WidgetId:  t.widgetID,
	})
	if err != nil {
		return 0, 0, err
	}
	return int(resp.Row), int(resp.Column), nil
}

// SetSelection sets the current cell selection.
func (t *Table) SetSelection(row, column int) error {
	_, err := t.client.tableClient.SetSelection(t.client.ctx, &pb.TableSetSelectionRequest{
		SessionId: t.client.sessionID,
		WidgetId:  t.widgetID,
		Row:       int32(row),
		Column:    int32(column),
	})
	return err
}

// SetFixed sets the number of fixed rows and columns.
func (t *Table) SetFixed(fixedRows, fixedColumns int) error {
	_, err := t.client.tableClient.SetFixed(t.client.ctx, &pb.TableSetFixedRequest{
		SessionId:    t.client.sessionID,
		WidgetId:     t.widgetID,
		FixedRows:    int32(fixedRows),
		FixedColumns: int32(fixedColumns),
	})
	return err
}

// NewTableCell creates a new table cell with text.
func NewTableCell(text string) *pb.TableCell {
	return &pb.TableCell{Text: text}
}

// NewTableCellWithColor creates a new table cell with text and color.
func NewTableCellWithColor(text string, color *pb.Color) *pb.TableCell {
	return &pb.TableCell{
		Text:  text,
		Color: color,
	}
}

