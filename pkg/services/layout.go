package services

import (
	"context"
	"log/slog"
	"sync"

	"github.com/chazu/yutani/pkg/server"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/rivo/tview"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LayoutService implements the LayoutService gRPC service
type LayoutService struct {
	pb.UnimplementedLayoutServiceServer
	server *server.Server
	mu     sync.RWMutex
}

// NewLayoutService creates a new LayoutService
func NewLayoutService(srv *server.Server) *LayoutService {
	return &LayoutService{
		server: srv,
	}
}

// FlexAddItem adds an item to a Flex container
func (s *LayoutService) FlexAddItem(ctx context.Context, req *pb.FlexAddItemRequest) (*pb.FlexAddItemResponse, error) {
	if req.SessionId == nil || req.FlexId == nil || req.ItemId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, flex_id, and item_id are required")
	}

	sessionID := req.SessionId.Id
	flexID := req.FlexId.Id
	itemID := req.ItemId.Id

	// Verify ownership of flex container
	owner, ok := s.server.Widgets().GetOwner(flexID)
	if !ok {
		return nil, status.Error(codes.NotFound, "flex container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "flex container not owned by session")
	}

	// Verify ownership of item
	itemOwner, ok := s.server.Widgets().GetOwner(itemID)
	if !ok {
		return nil, status.Error(codes.NotFound, "item widget not found")
	}
	if itemOwner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "item widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the flex container
		primitive, ok := s.server.Widgets().Get(flexID)
		if !ok {
			err = status.Error(codes.NotFound, "flex container not found")
			return
		}

		flex, ok := primitive.(*tview.Flex)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a flex container")
			return
		}

		// Get the item widget
		itemPrimitive, ok := s.server.Widgets().Get(itemID)
		if !ok {
			err = status.Error(codes.NotFound, "item widget not found")
			return
		}

		// Add the item to the flex container
		if req.Proportion > 0 {
			flex.AddItem(itemPrimitive, 0, int(req.Proportion), req.Focus)
		} else {
			flex.AddItem(itemPrimitive, int(req.FixedSize), 0, req.Focus)
		}

		// Add child relationship
		if !s.server.Widgets().AddChild(flexID, itemID) {
			err = status.Error(codes.Internal, "failed to add child relationship")
			return
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Flex item added",
		"flex_id", flexID,
		"item_id", itemID,
		"proportion", req.Proportion,
		"fixed_size", req.FixedSize,
	)

	return &pb.FlexAddItemResponse{Success: true}, nil
}

// FlexRemoveItem removes an item from a Flex container
func (s *LayoutService) FlexRemoveItem(ctx context.Context, req *pb.FlexRemoveItemRequest) (*pb.FlexRemoveItemResponse, error) {
	if req.SessionId == nil || req.FlexId == nil || req.ItemId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, flex_id, and item_id are required")
	}

	sessionID := req.SessionId.Id
	flexID := req.FlexId.Id
	itemID := req.ItemId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(flexID)
	if !ok {
		return nil, status.Error(codes.NotFound, "flex container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "flex container not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the flex container
		primitive, ok := s.server.Widgets().Get(flexID)
		if !ok {
			err = status.Error(codes.NotFound, "flex container not found")
			return
		}

		flex, ok := primitive.(*tview.Flex)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a flex container")
			return
		}

		// Get the item widget
		itemPrimitive, ok := s.server.Widgets().Get(itemID)
		if !ok {
			err = status.Error(codes.NotFound, "item widget not found")
			return
		}

		// Remove the item from the flex container
		flex.RemoveItem(itemPrimitive)

		// Remove child relationship
		if !s.server.Widgets().RemoveChild(flexID, itemID) {
			err = status.Error(codes.Internal, "failed to remove child relationship")
			return
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Flex item removed",
		"flex_id", flexID,
		"item_id", itemID,
	)

	return &pb.FlexRemoveItemResponse{Success: true}, nil
}

// FlexSetDirection sets the direction of a Flex container
func (s *LayoutService) FlexSetDirection(ctx context.Context, req *pb.FlexSetDirectionRequest) (*pb.FlexSetDirectionResponse, error) {
	if req.SessionId == nil || req.FlexId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and flex_id are required")
	}

	sessionID := req.SessionId.Id
	flexID := req.FlexId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(flexID)
	if !ok {
		return nil, status.Error(codes.NotFound, "flex container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "flex container not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the flex container
		primitive, ok := s.server.Widgets().Get(flexID)
		if !ok {
			err = status.Error(codes.NotFound, "flex container not found")
			return
		}

		flex, ok := primitive.(*tview.Flex)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a flex container")
			return
		}

		// Set the direction
		// NOTE: tview's naming is confusing and opposite of CSS flexbox:
		// - tview.FlexColumn = items arranged HORIZONTALLY (side by side)
		// - tview.FlexRow = items arranged VERTICALLY (stacked)
		// Our proto uses CSS-style semantics where FLEX_COLUMN = vertical stacking
		// So we SWAP the mapping here.
		if req.Direction == pb.FlexDirection_FLEX_COLUMN {
			// User wants vertical stacking (CSS column) -> use tview.FlexRow
			flex.SetDirection(tview.FlexRow)
		} else {
			// User wants horizontal arrangement (CSS row) -> use tview.FlexColumn
			flex.SetDirection(tview.FlexColumn)
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Flex direction set",
		"flex_id", flexID,
		"direction", req.Direction,
	)

	return &pb.FlexSetDirectionResponse{Success: true}, nil
}

// GridAddItem adds an item to a Grid container
func (s *LayoutService) GridAddItem(ctx context.Context, req *pb.GridAddItemRequest) (*pb.GridAddItemResponse, error) {
	if req.SessionId == nil || req.GridId == nil || req.ItemId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, grid_id, and item_id are required")
	}

	sessionID := req.SessionId.Id
	gridID := req.GridId.Id
	itemID := req.ItemId.Id

	// Verify ownership of grid container
	owner, ok := s.server.Widgets().GetOwner(gridID)
	if !ok {
		return nil, status.Error(codes.NotFound, "grid container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "grid container not owned by session")
	}

	// Verify ownership of item
	itemOwner, ok := s.server.Widgets().GetOwner(itemID)
	if !ok {
		return nil, status.Error(codes.NotFound, "item widget not found")
	}
	if itemOwner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "item widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the grid container
		primitive, ok := s.server.Widgets().Get(gridID)
		if !ok {
			err = status.Error(codes.NotFound, "grid container not found")
			return
		}

		grid, ok := primitive.(*tview.Grid)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a grid container")
			return
		}

		// Get the item widget
		itemPrimitive, ok := s.server.Widgets().Get(itemID)
		if !ok {
			err = status.Error(codes.NotFound, "item widget not found")
			return
		}

		// Add the item to the grid container
		grid.AddItem(itemPrimitive,
			int(req.Row), int(req.Column),
			int(req.RowSpan), int(req.ColumnSpan),
			int(req.MinWidth), int(req.MinHeight),
			req.Focus)

		// Add child relationship
		if !s.server.Widgets().AddChild(gridID, itemID) {
			err = status.Error(codes.Internal, "failed to add child relationship")
			return
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Grid item added",
		"grid_id", gridID,
		"item_id", itemID,
		"row", req.Row,
		"column", req.Column,
	)

	return &pb.GridAddItemResponse{Success: true}, nil
}

// GridRemoveItem removes an item from a Grid container
func (s *LayoutService) GridRemoveItem(ctx context.Context, req *pb.GridRemoveItemRequest) (*pb.GridRemoveItemResponse, error) {
	if req.SessionId == nil || req.GridId == nil || req.ItemId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, grid_id, and item_id are required")
	}

	sessionID := req.SessionId.Id
	gridID := req.GridId.Id
	itemID := req.ItemId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(gridID)
	if !ok {
		return nil, status.Error(codes.NotFound, "grid container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "grid container not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the grid container
		primitive, ok := s.server.Widgets().Get(gridID)
		if !ok {
			err = status.Error(codes.NotFound, "grid container not found")
			return
		}

		grid, ok := primitive.(*tview.Grid)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a grid container")
			return
		}

		// Get the item widget
		itemPrimitive, ok := s.server.Widgets().Get(itemID)
		if !ok {
			err = status.Error(codes.NotFound, "item widget not found")
			return
		}

		// Remove the item from the grid container
		grid.RemoveItem(itemPrimitive)

		// Remove child relationship
		if !s.server.Widgets().RemoveChild(gridID, itemID) {
			err = status.Error(codes.Internal, "failed to remove child relationship")
			return
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Grid item removed",
		"grid_id", gridID,
		"item_id", itemID,
	)

	return &pb.GridRemoveItemResponse{Success: true}, nil
}

// GridSetRows sets the row sizes for a Grid container
func (s *LayoutService) GridSetRows(ctx context.Context, req *pb.GridSetRowsRequest) (*pb.GridSetRowsResponse, error) {
	if req.SessionId == nil || req.GridId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and grid_id are required")
	}

	sessionID := req.SessionId.Id
	gridID := req.GridId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(gridID)
	if !ok {
		return nil, status.Error(codes.NotFound, "grid container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "grid container not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the grid container
		primitive, ok := s.server.Widgets().Get(gridID)
		if !ok {
			err = status.Error(codes.NotFound, "grid container not found")
			return
		}

		grid, ok := primitive.(*tview.Grid)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a grid container")
			return
		}

		// Convert int32 to int
		rowSizes := make([]int, len(req.RowSizes))
		for i, size := range req.RowSizes {
			rowSizes[i] = int(size)
		}

		// Set the row sizes
		grid.SetRows(rowSizes...)
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Grid rows set",
		"grid_id", gridID,
		"row_count", len(req.RowSizes),
	)

	return &pb.GridSetRowsResponse{Success: true}, nil
}

// GridSetColumns sets the column sizes for a Grid container
func (s *LayoutService) GridSetColumns(ctx context.Context, req *pb.GridSetColumnsRequest) (*pb.GridSetColumnsResponse, error) {
	if req.SessionId == nil || req.GridId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and grid_id are required")
	}

	sessionID := req.SessionId.Id
	gridID := req.GridId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(gridID)
	if !ok {
		return nil, status.Error(codes.NotFound, "grid container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "grid container not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the grid container
		primitive, ok := s.server.Widgets().Get(gridID)
		if !ok {
			err = status.Error(codes.NotFound, "grid container not found")
			return
		}

		grid, ok := primitive.(*tview.Grid)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a grid container")
			return
		}

		// Convert int32 to int
		columnSizes := make([]int, len(req.ColumnSizes))
		for i, size := range req.ColumnSizes {
			columnSizes[i] = int(size)
		}

		// Set the column sizes
		grid.SetColumns(columnSizes...)
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Grid columns set",
		"grid_id", gridID,
		"column_count", len(req.ColumnSizes),
	)

	return &pb.GridSetColumnsResponse{Success: true}, nil
}

// PagesAddPage adds a page to a Pages container
func (s *LayoutService) PagesAddPage(ctx context.Context, req *pb.PagesAddPageRequest) (*pb.PagesAddPageResponse, error) {
	if req.SessionId == nil || req.PagesId == nil || req.ItemId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, pages_id, and item_id are required")
	}
	if req.PageName == "" {
		return nil, status.Error(codes.InvalidArgument, "page_name is required")
	}

	sessionID := req.SessionId.Id
	pagesID := req.PagesId.Id
	itemID := req.ItemId.Id

	// Verify ownership of pages container
	owner, ok := s.server.Widgets().GetOwner(pagesID)
	if !ok {
		return nil, status.Error(codes.NotFound, "pages container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "pages container not owned by session")
	}

	// Verify ownership of item
	itemOwner, ok := s.server.Widgets().GetOwner(itemID)
	if !ok {
		return nil, status.Error(codes.NotFound, "item widget not found")
	}
	if itemOwner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "item widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the pages container
		primitive, ok := s.server.Widgets().Get(pagesID)
		if !ok {
			err = status.Error(codes.NotFound, "pages container not found")
			return
		}

		pages, ok := primitive.(*tview.Pages)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a pages container")
			return
		}

		// Get the item widget
		itemPrimitive, ok := s.server.Widgets().Get(itemID)
		if !ok {
			err = status.Error(codes.NotFound, "item widget not found")
			return
		}

		// Add the page
		pages.AddPage(req.PageName, itemPrimitive, req.Resize, req.Visible)

		// Add child relationship
		if !s.server.Widgets().AddChild(pagesID, itemID) {
			err = status.Error(codes.Internal, "failed to add child relationship")
			return
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Page added",
		"pages_id", pagesID,
		"page_name", req.PageName,
		"item_id", itemID,
	)

	return &pb.PagesAddPageResponse{Success: true}, nil
}

// PagesRemovePage removes a page from a Pages container
func (s *LayoutService) PagesRemovePage(ctx context.Context, req *pb.PagesRemovePageRequest) (*pb.PagesRemovePageResponse, error) {
	if req.SessionId == nil || req.PagesId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and pages_id are required")
	}
	if req.PageName == "" {
		return nil, status.Error(codes.InvalidArgument, "page_name is required")
	}

	sessionID := req.SessionId.Id
	pagesID := req.PagesId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(pagesID)
	if !ok {
		return nil, status.Error(codes.NotFound, "pages container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "pages container not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the pages container
		primitive, ok := s.server.Widgets().Get(pagesID)
		if !ok {
			err = status.Error(codes.NotFound, "pages container not found")
			return
		}

		pages, ok := primitive.(*tview.Pages)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a pages container")
			return
		}

		// Remove the page
		pages.RemovePage(req.PageName)

		// Note: We can't easily determine the item_id from the page name,
		// so we don't remove the child relationship here.
		// The client should call DeleteWidget on the item if needed.
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Page removed",
		"pages_id", pagesID,
		"page_name", req.PageName,
	)

	return &pb.PagesRemovePageResponse{Success: true}, nil
}

// PagesShowPage shows a specific page in a Pages container
func (s *LayoutService) PagesShowPage(ctx context.Context, req *pb.PagesShowPageRequest) (*pb.PagesShowPageResponse, error) {
	if req.SessionId == nil || req.PagesId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and pages_id are required")
	}
	if req.PageName == "" {
		return nil, status.Error(codes.InvalidArgument, "page_name is required")
	}

	sessionID := req.SessionId.Id
	pagesID := req.PagesId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(pagesID)
	if !ok {
		return nil, status.Error(codes.NotFound, "pages container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "pages container not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the pages container
		primitive, ok := s.server.Widgets().Get(pagesID)
		if !ok {
			err = status.Error(codes.NotFound, "pages container not found")
			return
		}

		pages, ok := primitive.(*tview.Pages)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a pages container")
			return
		}

		// Show the page
		pages.SwitchToPage(req.PageName)
	})
	<-done

	if err != nil {
		return nil, err
	}

	slog.Info("Page shown",
		"pages_id", pagesID,
		"page_name", req.PageName,
	)

	return &pb.PagesShowPageResponse{Success: true}, nil
}

// PagesGetCurrentPage gets the name of the currently visible page
func (s *LayoutService) PagesGetCurrentPage(ctx context.Context, req *pb.PagesGetCurrentPageRequest) (*pb.PagesGetCurrentPageResponse, error) {
	if req.SessionId == nil || req.PagesId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and pages_id are required")
	}

	sessionID := req.SessionId.Id
	pagesID := req.PagesId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(pagesID)
	if !ok {
		return nil, status.Error(codes.NotFound, "pages container not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "pages container not owned by session")
	}

	var err error
	var pageName string
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		// Get the pages container
		primitive, ok := s.server.Widgets().Get(pagesID)
		if !ok {
			err = status.Error(codes.NotFound, "pages container not found")
			return
		}

		pages, ok := primitive.(*tview.Pages)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a pages container")
			return
		}

		// Get the current page name
		pageName, _ = pages.GetFrontPage()
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.PagesGetCurrentPageResponse{PageName: pageName}, nil
}

