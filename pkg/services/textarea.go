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

// TextAreaService implements the TextAreaService gRPC service
type TextAreaService struct {
	pb.UnimplementedTextAreaServiceServer
	server *server.Server
}

// NewTextAreaService creates a new TextAreaService
func NewTextAreaService(srv *server.Server) *TextAreaService {
	return &TextAreaService{
		server: srv,
	}
}

// getTextArea is a helper that validates the request and returns the tview.TextArea primitive.
// Must be called from within QueueUpdateDraw.
func (s *TextAreaService) getTextArea(sessionID, widgetID string) (*tview.TextArea, error) {
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	primitive, ok := s.server.Widgets().Get(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}

	ta, ok := primitive.(*tview.TextArea)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "widget is not a text area")
	}

	return ta, nil
}

func validateSessionWidget(sessionID *pb.SessionId, widgetID *pb.WidgetId) (string, string, error) {
	if sessionID == nil || widgetID == nil {
		return "", "", status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}
	return sessionID.Id, widgetID.Id, nil
}

// GetCursorPosition returns the current cursor position
func (s *TextAreaService) GetCursorPosition(ctx context.Context, req *pb.TextAreaGetCursorRequest) (*pb.TextAreaGetCursorResponse, error) {
	sid, wid, err := validateSessionWidget(req.SessionId, req.WidgetId)
	if err != nil {
		return nil, err
	}

	var resp pb.TextAreaGetCursorResponse
	var rpcErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		ta, err := s.getTextArea(sid, wid)
		if err != nil {
			rpcErr = err
			return
		}

		fromRow, fromCol, toRow, toCol := ta.GetCursor()
		hasSel := ta.HasSelection()

		resp.Row = int32(fromRow)
		resp.Column = int32(fromCol)
		resp.ToRow = int32(toRow)
		resp.ToColumn = int32(toCol)
		resp.HasSelection = hasSel
	})
	<-done

	if rpcErr != nil {
		return nil, rpcErr
	}
	return &resp, nil
}

// SetCursorPosition sets the cursor position by selecting a zero-length range
func (s *TextAreaService) SetCursorPosition(ctx context.Context, req *pb.TextAreaSetCursorRequest) (*pb.TextAreaSetCursorResponse, error) {
	sid, wid, err := validateSessionWidget(req.SessionId, req.WidgetId)
	if err != nil {
		return nil, err
	}

	var rpcErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		ta, err := s.getTextArea(sid, wid)
		if err != nil {
			rpcErr = err
			return
		}

		// Convert row/column to byte offset in the text
		offset := rowColToOffset(ta.GetText(), int(req.Row), int(req.Column))
		// Select a zero-length range at that offset to move cursor there
		ta.Select(offset, offset)

		slog.Info("TextArea cursor set", "widget_id", wid, "row", req.Row, "col", req.Column)
	})
	<-done

	if rpcErr != nil {
		return nil, rpcErr
	}
	return &pb.TextAreaSetCursorResponse{Success: true}, nil
}

// GetSelection returns the current selection range and text
func (s *TextAreaService) GetSelection(ctx context.Context, req *pb.TextAreaGetSelectionRequest) (*pb.TextAreaGetSelectionResponse, error) {
	sid, wid, err := validateSessionWidget(req.SessionId, req.WidgetId)
	if err != nil {
		return nil, err
	}

	var resp pb.TextAreaGetSelectionResponse
	var rpcErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		ta, err := s.getTextArea(sid, wid)
		if err != nil {
			rpcErr = err
			return
		}

		resp.HasSelection = ta.HasSelection()
		if resp.HasSelection {
			text, start, end := ta.GetSelection()
			resp.SelectedText = text
			resp.Start = int32(start)
			resp.End = int32(end)
		}

		fromRow, fromCol, toRow, toCol := ta.GetCursor()
		resp.FromRow = int32(fromRow)
		resp.FromColumn = int32(fromCol)
		resp.ToRow = int32(toRow)
		resp.ToColumn = int32(toCol)
	})
	<-done

	if rpcErr != nil {
		return nil, rpcErr
	}
	return &resp, nil
}

// SetSelection sets the selection range by byte offsets
func (s *TextAreaService) SetSelection(ctx context.Context, req *pb.TextAreaSetSelectionRequest) (*pb.TextAreaSetSelectionResponse, error) {
	sid, wid, err := validateSessionWidget(req.SessionId, req.WidgetId)
	if err != nil {
		return nil, err
	}

	var rpcErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		ta, err := s.getTextArea(sid, wid)
		if err != nil {
			rpcErr = err
			return
		}

		ta.Select(int(req.Start), int(req.End))

		slog.Info("TextArea selection set", "widget_id", wid, "start", req.Start, "end", req.End)
	})
	<-done

	if rpcErr != nil {
		return nil, rpcErr
	}
	return &pb.TextAreaSetSelectionResponse{Success: true}, nil
}

// HasSelection checks if text is currently selected
func (s *TextAreaService) HasSelection(ctx context.Context, req *pb.TextAreaHasSelectionRequest) (*pb.TextAreaHasSelectionResponse, error) {
	sid, wid, err := validateSessionWidget(req.SessionId, req.WidgetId)
	if err != nil {
		return nil, err
	}

	var hasSel bool
	var rpcErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		ta, err := s.getTextArea(sid, wid)
		if err != nil {
			rpcErr = err
			return
		}

		hasSel = ta.HasSelection()
	})
	<-done

	if rpcErr != nil {
		return nil, rpcErr
	}
	return &pb.TextAreaHasSelectionResponse{HasSelection: hasSel}, nil
}

// GetSelectedText returns the currently selected text
func (s *TextAreaService) GetSelectedText(ctx context.Context, req *pb.TextAreaGetSelectedTextRequest) (*pb.TextAreaGetSelectedTextResponse, error) {
	sid, wid, err := validateSessionWidget(req.SessionId, req.WidgetId)
	if err != nil {
		return nil, err
	}

	var text string
	var hasSel bool
	var rpcErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		ta, err := s.getTextArea(sid, wid)
		if err != nil {
			rpcErr = err
			return
		}

		hasSel = ta.HasSelection()
		if hasSel {
			text, _, _ = ta.GetSelection()
		}
	})
	<-done

	if rpcErr != nil {
		return nil, rpcErr
	}
	return &pb.TextAreaGetSelectedTextResponse{
		Text:         text,
		HasSelection: hasSel,
	}, nil
}

// InsertTextAtCursor inserts text at the current cursor position,
// replacing any selected text.
func (s *TextAreaService) InsertTextAtCursor(ctx context.Context, req *pb.TextAreaInsertTextRequest) (*pb.TextAreaInsertTextResponse, error) {
	sid, wid, err := validateSessionWidget(req.SessionId, req.WidgetId)
	if err != nil {
		return nil, err
	}

	var rpcErr error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		ta, err := s.getTextArea(sid, wid)
		if err != nil {
			rpcErr = err
			return
		}

		if ta.HasSelection() {
			// Replace selected text
			_, start, end := ta.GetSelection()
			ta.Replace(start, end, req.Text)
		} else {
			// Insert at cursor position - get cursor byte offset
			_, _, toRow, toCol := ta.GetCursor()
			offset := rowColToOffset(ta.GetText(), toRow, toCol)
			ta.Replace(offset, offset, req.Text)
		}

		slog.Info("TextArea text inserted", "widget_id", wid, "text_len", len(req.Text))
	})
	<-done

	if rpcErr != nil {
		return nil, rpcErr
	}
	return &pb.TextAreaInsertTextResponse{Success: true}, nil
}

// rowColToOffset converts a row/column position to a byte offset in the text.
func rowColToOffset(text string, row, col int) int {
	currentRow := 0
	currentCol := 0
	for i, r := range text {
		if currentRow == row && currentCol == col {
			return i
		}
		if r == '\n' {
			if currentRow == row {
				// Column is past end of line, return end of this line
				return i
			}
			currentRow++
			currentCol = 0
		} else {
			currentCol++
		}
	}
	// If we reached the end, return the length (cursor at end of text)
	return len(text)
}
