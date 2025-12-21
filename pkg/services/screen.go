package services

import (
	"context"
	"log/slog"

	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ScreenService implements the ScreenService gRPC service
type ScreenService struct {
	pb.UnimplementedScreenServiceServer
	server *server.Server
}

// NewScreenService creates a new ScreenService
func NewScreenService(srv *server.Server) *ScreenService {
	return &ScreenService{
		server: srv,
	}
}

// GetSize returns the screen dimensions
func (s *ScreenService) GetSize(ctx context.Context, req *pb.GetSizeRequest) (*pb.GetSizeResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	// Get screen size
	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	width, height := screen.Size()

	return &pb.GetSizeResponse{
		Size: &pb.Size{
			Width:  int32(width),
			Height: int32(height),
		},
	}, nil
}

// Clear clears the screen or a region
func (s *ScreenService) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the clear operation
	done := make(chan error, 1)
	s.server.App().QueueUpdateDraw(func() {
		// Convert style if provided
		style := tcell.StyleDefault
		if req.Style != nil {
			style = convertStyle(req.Style)
		}

		if req.Region != nil {
			// Clear specific region
			r := req.Region
			for y := r.Y; y < r.Y+r.Height; y++ {
				for x := r.X; x < r.X+r.Width; x++ {
					screen.SetContent(int(x), int(y), ' ', nil, style)
				}
			}
		} else {
			// Clear entire screen
			width, height := screen.Size()
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					screen.SetContent(x, y, ' ', nil, style)
				}
			}
		}

		done <- nil
	})

	if err := <-done; err != nil {
		slog.Error("Failed to clear screen", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to clear screen: %v", err)
	}

	return &pb.ClearResponse{Success: true}, nil
}

// Sync forces a screen refresh
func (s *ScreenService) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the sync operation
	done := make(chan error, 1)
	s.server.App().QueueUpdateDraw(func() {
		screen.Sync()
		done <- nil
	})

	if err := <-done; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to sync screen: %v", err)
	}

	return &pb.SyncResponse{Success: true}, nil
}

// convertStyle converts a protobuf Style to tcell.Style
func convertStyle(pbStyle *pb.Style) tcell.Style {
	style := tcell.StyleDefault

	// Convert foreground color
	if pbStyle.Foreground != nil {
		style = style.Foreground(convertColor(pbStyle.Foreground))
	}

	// Convert background color
	if pbStyle.Background != nil {
		style = style.Background(convertColor(pbStyle.Background))
	}

	// Convert attributes
	for _, attr := range pbStyle.Attributes {
		style = applyAttribute(style, attr)
	}

	return style
}

// SetCell sets a single cell at the specified position
func (s *ScreenService) SetCell(ctx context.Context, req *pb.SetCellRequest) (*pb.SetCellResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the operation
	done := make(chan error, 1)
	s.server.App().QueueUpdateDraw(func() {
		style := tcell.StyleDefault
		if req.Cell != nil && req.Cell.Style != nil {
			style = convertStyle(req.Cell.Style)
		}

		ch := ' '
		if req.Cell != nil && len(req.Cell.Rune) > 0 {
			runes := []rune(req.Cell.Rune)
			ch = runes[0]
		}

		screen.SetContent(int(req.Position.X), int(req.Position.Y), ch, nil, style)
		done <- nil
	})

	if err := <-done; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set cell: %v", err)
	}

	return &pb.SetCellResponse{Success: true}, nil
}

// SetCells sets multiple cells in a batch operation
func (s *ScreenService) SetCells(ctx context.Context, req *pb.SetCellsRequest) (*pb.SetCellsResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the operation
	done := make(chan error, 1)
	var count int32
	s.server.App().QueueUpdateDraw(func() {
		for _, cellPos := range req.Cells {
			style := tcell.StyleDefault
			if cellPos.Cell != nil && cellPos.Cell.Style != nil {
				style = convertStyle(cellPos.Cell.Style)
			}

			ch := ' '
			if cellPos.Cell != nil && len(cellPos.Cell.Rune) > 0 {
				runes := []rune(cellPos.Cell.Rune)
				ch = runes[0]
			}

			screen.SetContent(int(cellPos.Position.X), int(cellPos.Position.Y), ch, nil, style)
			count++
		}
		done <- nil
	})

	if err := <-done; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set cells: %v", err)
	}

	return &pb.SetCellsResponse{Success: true, CellsSet: count}, nil
}

// Fill fills a region with the specified cell
func (s *ScreenService) Fill(ctx context.Context, req *pb.FillRequest) (*pb.FillResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the operation
	done := make(chan error, 1)
	s.server.App().QueueUpdateDraw(func() {
		style := tcell.StyleDefault
		if req.Cell != nil && req.Cell.Style != nil {
			style = convertStyle(req.Cell.Style)
		}

		ch := ' '
		if req.Cell != nil && len(req.Cell.Rune) > 0 {
			runes := []rune(req.Cell.Rune)
			ch = runes[0]
		}

		r := req.Region
		for y := r.Y; y < r.Y+r.Height; y++ {
			for x := r.X; x < r.X+r.Width; x++ {
				screen.SetContent(int(x), int(y), ch, nil, style)
			}
		}
		done <- nil
	})

	if err := <-done; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fill region: %v", err)
	}

	return &pb.FillResponse{Success: true}, nil
}

// DrawText draws text at the specified position
func (s *ScreenService) DrawText(ctx context.Context, req *pb.DrawTextRequest) (*pb.DrawTextResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the operation
	done := make(chan error, 1)
	var count int32
	s.server.App().QueueUpdateDraw(func() {
		style := tcell.StyleDefault
		if req.Style != nil {
			style = convertStyle(req.Style)
		}

		x := int(req.Position.X)
		y := int(req.Position.Y)

		for _, r := range req.Text {
			screen.SetContent(x, y, r, nil, style)
			x++
			count++
		}
		done <- nil
	})

	if err := <-done; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to draw text: %v", err)
	}

	return &pb.DrawTextResponse{Success: true, CharsDrawn: count}, nil
}

// DrawBox draws a box/border at the specified rectangle
func (s *ScreenService) DrawBox(ctx context.Context, req *pb.DrawBoxRequest) (*pb.DrawBoxResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the operation
	done := make(chan error, 1)
	s.server.App().QueueUpdateDraw(func() {
		style := tcell.StyleDefault
		if req.Style != nil {
			style = convertStyle(req.Style)
		}

		r := req.Rect
		x1, y1 := int(r.X), int(r.Y)
		x2, y2 := int(r.X+r.Width-1), int(r.Y+r.Height-1)

		// Box drawing characters
		const (
			topLeft     = '┌'
			topRight    = '┐'
			bottomLeft  = '└'
			bottomRight = '┘'
			horizontal  = '─'
			vertical    = '│'
		)

		// Draw corners
		screen.SetContent(x1, y1, topLeft, nil, style)
		screen.SetContent(x2, y1, topRight, nil, style)
		screen.SetContent(x1, y2, bottomLeft, nil, style)
		screen.SetContent(x2, y2, bottomRight, nil, style)

		// Draw horizontal lines
		for x := x1 + 1; x < x2; x++ {
			screen.SetContent(x, y1, horizontal, nil, style)
			screen.SetContent(x, y2, horizontal, nil, style)
		}

		// Draw vertical lines
		for y := y1 + 1; y < y2; y++ {
			screen.SetContent(x1, y, vertical, nil, style)
			screen.SetContent(x2, y, vertical, nil, style)
		}

		// Fill interior if requested
		if req.Filled != nil && *req.Filled {
			for y := y1 + 1; y < y2; y++ {
				for x := x1 + 1; x < x2; x++ {
					screen.SetContent(x, y, ' ', nil, style)
				}
			}
		}

		done <- nil
	})

	if err := <-done; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to draw box: %v", err)
	}

	return &pb.DrawBoxResponse{Success: true}, nil
}

// GetCell retrieves the cell content at the specified position
func (s *ScreenService) GetCell(ctx context.Context, req *pb.GetCellRequest) (*pb.GetCellResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Verify session exists
	sessionID := req.SessionId.Id
	if _, ok := s.server.Sessions().Get(sessionID); !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	// Queue the operation
	done := make(chan *pb.Cell, 1)
	s.server.App().QueueUpdate(func() {
		mainc, _, style, _ := screen.GetContent(int(req.Position.X), int(req.Position.Y))

		cell := &pb.Cell{
			Rune:  string(mainc),
			Style: convertStyleToProto(style),
		}
		done <- cell
	})

	cell := <-done
	return &pb.GetCellResponse{Cell: cell}, nil
}
