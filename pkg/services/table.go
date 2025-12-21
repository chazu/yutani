package services

import (
	"context"
	"log/slog"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/rivo/tview"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TableService implements the TableService gRPC service
type TableService struct {
	pb.UnimplementedTableServiceServer
	server *server.Server
}

// NewTableService creates a new TableService
func NewTableService(srv *server.Server) *TableService {
	return &TableService{
		server: srv,
	}
}

// SetCell sets a single cell's content
func (s *TableService) SetCell(ctx context.Context, req *pb.SetTableCellRequest) (*pb.SetTableCellResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		// Apply cell properties
		cell := tview.NewTableCell(req.Cell.Text)
		if req.Cell.Color != nil {
			cell.SetTextColor(convertColor(req.Cell.Color))
		}
		if req.Cell.Align != nil {
			cell.SetAlign(convertAlignment(*req.Cell.Align))
		}
		if req.Cell.Selectable != nil {
			cell.SetSelectable(*req.Cell.Selectable)
		}
		if req.Cell.Expansion != nil {
			cell.SetExpansion(int(*req.Cell.Expansion))
		}

		table.SetCell(int(req.Row), int(req.Column), cell)

		slog.Info("Table cell set", "widget_id", widgetID, "row", req.Row, "column", req.Column)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetTableCellResponse{
		Success: true,
	}, nil
}

// GetCell gets a single cell's content
func (s *TableService) GetCell(ctx context.Context, req *pb.GetTableCellRequest) (*pb.GetTableCellResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var cell *pb.TableCell
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		tviewCell := table.GetCell(int(req.Row), int(req.Column))
		if tviewCell == nil {
			err = status.Error(codes.NotFound, "cell not found")
			return
		}

		// Convert tview cell to protobuf
		cell = &pb.TableCell{
			Text: tviewCell.Text,
		}
		// Note: tview doesn't expose all cell properties for reading
		// We can only get the text reliably
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetTableCellResponse{
		Cell: cell,
	}, nil
}

// SetCells sets multiple cells (batch operation)
func (s *TableService) SetCells(ctx context.Context, req *pb.SetTableCellsRequest) (*pb.SetTableCellsResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var count int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		// Set all cells
		for _, cellUpdate := range req.Cells {
			cell := tview.NewTableCell(cellUpdate.Cell.Text)
			if cellUpdate.Cell.Color != nil {
				cell.SetTextColor(convertColor(cellUpdate.Cell.Color))
			}
			if cellUpdate.Cell.Align != nil {
				cell.SetAlign(convertAlignment(*cellUpdate.Cell.Align))
			}
			if cellUpdate.Cell.Selectable != nil {
				cell.SetSelectable(*cellUpdate.Cell.Selectable)
			}
			if cellUpdate.Cell.Expansion != nil {
				cell.SetExpansion(int(*cellUpdate.Cell.Expansion))
			}

			table.SetCell(int(cellUpdate.Row), int(cellUpdate.Column), cell)
			count++
		}

		slog.Info("Table cells set", "widget_id", widgetID, "count", count)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetTableCellsResponse{
		Success:  true,
		CellsSet: count,
	}, nil
}

// Clear clears the table
func (s *TableService) Clear(ctx context.Context, req *pb.ClearTableRequest) (*pb.ClearTableResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		table.Clear()

		slog.Info("Table cleared", "widget_id", widgetID)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.ClearTableResponse{
		Success: true,
	}, nil
}

// GetDimensions gets table dimensions (rows, columns)
func (s *TableService) GetDimensions(ctx context.Context, req *pb.GetDimensionsRequest) (*pb.GetDimensionsResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var rows, columns int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		rows = int32(table.GetRowCount())
		columns = int32(table.GetColumnCount())
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetDimensionsResponse{
		Rows:    rows,
		Columns: columns,
	}, nil
}

// GetSelection gets current selection
func (s *TableService) GetSelection(ctx context.Context, req *pb.GetSelectionRequest) (*pb.GetSelectionResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var row, column int32
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		r, c := table.GetSelection()
		row = int32(r)
		column = int32(c)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetSelectionResponse{
		Row:    row,
		Column: column,
	}, nil
}

// SetSelection sets current selection
func (s *TableService) SetSelection(ctx context.Context, req *pb.SetSelectionRequest) (*pb.SetSelectionResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		table.Select(int(req.Row), int(req.Column))

		slog.Info("Table selection set", "widget_id", widgetID, "row", req.Row, "column", req.Column)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetSelectionResponse{
		Success: true,
	}, nil
}

// SetFixed sets fixed rows/columns (headers)
func (s *TableService) SetFixed(ctx context.Context, req *pb.SetFixedRequest) (*pb.SetFixedResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		table, ok := primitive.(*tview.Table)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a table")
			return
		}

		table.SetFixed(int(req.FixedRows), int(req.FixedColumns))

		slog.Info("Table fixed rows/columns set", "widget_id", widgetID, "rows", req.FixedRows, "columns", req.FixedColumns)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetFixedResponse{
		Success: true,
	}, nil
}

