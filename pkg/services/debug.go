package services

import (
	"context"
	"strings"
	"sync"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/rivo/tview"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxRecentEvents = 10

// DebugService implements the DebugService gRPC service
type DebugService struct {
	pb.UnimplementedDebugServiceServer
	server *server.Server

	// Event tracking per widget
	eventsMu     sync.RWMutex
	widgetEvents map[string][]*pb.WidgetEventRecord
}

// NewDebugService creates a new DebugService
func NewDebugService(srv *server.Server) *DebugService {
	ds := &DebugService{
		server:       srv,
		widgetEvents: make(map[string][]*pb.WidgetEventRecord),
	}
	return ds
}

// RecordWidgetEvent records an event for debugging purposes
// This should be called by other services when events occur
func (s *DebugService) RecordWidgetEvent(widgetID, eventType string, details map[string]string) {
	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	record := &pb.WidgetEventRecord{
		Timestamp: time.Now().UnixNano(),
		EventType: eventType,
		Details:   details,
	}

	events := s.widgetEvents[widgetID]
	events = append(events, record)

	// Keep only the most recent events
	if len(events) > maxRecentEvents {
		events = events[len(events)-maxRecentEvents:]
	}

	s.widgetEvents[widgetID] = events
}

// GetScreenDump returns an ASCII art representation of the current screen
func (s *DebugService) GetScreenDump(ctx context.Context, req *pb.GetScreenDumpRequest) (*pb.GetScreenDumpResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	type result struct {
		dump          string
		width, height int
		bounds        []*pb.WidgetBoundsInfo
		focusedID     string
	}

	done := make(chan result, 1)

	s.server.App().QueueUpdate(func() {
		width, height := screen.Size()

		// Build screen dump
		var lines []string
		for y := 0; y < height; y++ {
			var line strings.Builder
			for x := 0; x < width; x++ {
				mainc, _, _, _ := screen.GetContent(x, y)
				if mainc == 0 {
					line.WriteRune(' ')
				} else {
					line.WriteRune(mainc)
				}
			}
			lines = append(lines, line.String())
		}

		// Get widget bounds
		var bounds []*pb.WidgetBoundsInfo
		shortIDCounter := 0
		infos := s.server.Widgets().ListBySession(sessionID)

		// Get focused widget
		focused := s.server.App().GetFocus()
		var focusedWidgetID string

		for _, info := range infos {
			// Get widget bounds using GetRect
			var x, y, w, h int
			if box, ok := info.Primitive.(interface {
				GetRect() (int, int, int, int)
			}); ok {
				x, y, w, h = box.GetRect()
			}

			isFocused := info.Primitive == focused
			if isFocused {
				focusedWidgetID = info.ID
			}

			shortID := generateShortID(shortIDCounter)
			shortIDCounter++

			bounds = append(bounds, &pb.WidgetBoundsInfo{
				WidgetId:   &pb.WidgetId{Id: info.ID},
				WidgetType: info.Type,
				Bounds: &pb.Rect{
					X:      int32(x),
					Y:      int32(y),
					Width:  int32(w),
					Height: int32(h),
				},
				HasFocus: isFocused,
				ShortId:  shortID,
			})
		}

		// Optionally overlay widget bounds markers
		if req.ShowWidgetBounds && len(bounds) > 0 {
			lines = overlayWidgetBounds(lines, bounds)
		}

		dump := strings.Join(lines, "\n")

		done <- result{
			dump:      dump,
			width:     width,
			height:    height,
			bounds:    bounds,
			focusedID: focusedWidgetID,
		}
	})

	res := <-done

	resp := &pb.GetScreenDumpResponse{
		ScreenDump: res.dump,
		ScreenSize: &pb.Size{
			Width:  int32(res.width),
			Height: int32(res.height),
		},
	}

	if req.IncludeLegend {
		resp.WidgetBounds = res.bounds
	}

	if res.focusedID != "" {
		resp.FocusedWidget = &pb.WidgetId{Id: res.focusedID}
	}

	return resp, nil
}

// GetWidgetState returns detailed state information for a specific widget
func (s *DebugService) GetWidgetState(ctx context.Context, req *pb.GetWidgetStateRequest) (*pb.GetWidgetStateResponse, error) {
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

	info, ok := s.server.Widgets().GetInfo(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}

	type result struct {
		x, y, w, h int
		hasFocus   bool
		isVisible  bool
	}

	done := make(chan result, 1)

	s.server.App().QueueUpdate(func() {
		var x, y, w, h int
		if box, ok := info.Primitive.(interface {
			GetRect() (int, int, int, int)
		}); ok {
			x, y, w, h = box.GetRect()
		}

		focused := s.server.App().GetFocus()
		hasFocus := info.Primitive == focused

		// Widget is visible if it has non-zero bounds
		isVisible := w > 0 && h > 0

		done <- result{
			x:         x,
			y:         y,
			w:         w,
			h:         h,
			hasFocus:  hasFocus,
			isVisible: isVisible,
		}
	})

	res := <-done

	// Build response
	resp := &pb.GetWidgetStateResponse{
		WidgetId:   &pb.WidgetId{Id: info.ID},
		WidgetType: info.Type,
		Bounds: &pb.Rect{
			X:      int32(res.x),
			Y:      int32(res.y),
			Width:  int32(res.w),
			Height: int32(res.h),
		},
		HasFocus:   res.hasFocus,
		IsVisible:  res.isVisible,
		Properties: info.Properties,
	}

	if info.ParentID != "" {
		resp.ParentId = &pb.WidgetId{Id: info.ParentID}
	}

	for _, childID := range info.ChildIDs {
		resp.Children = append(resp.Children, &pb.WidgetId{Id: childID})
	}

	// Add recent events
	s.eventsMu.RLock()
	if events, ok := s.widgetEvents[widgetID]; ok {
		resp.RecentEvents = events
	}
	s.eventsMu.RUnlock()

	return resp, nil
}

// GetAllWidgetBounds returns bounds for all widgets on screen
func (s *DebugService) GetAllWidgetBounds(ctx context.Context, req *pb.GetAllWidgetBoundsRequest) (*pb.GetAllWidgetBoundsResponse, error) {
	if req.SessionId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID := req.SessionId.Id
	if !s.server.Sessions().Exists(sessionID) {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	screen := s.server.Screen()
	if screen == nil {
		return nil, status.Error(codes.Unavailable, "screen not initialized")
	}

	type result struct {
		bounds      []*pb.WidgetBoundsInfo
		focusedID   string
		screenWidth int
		screenHeight int
	}

	done := make(chan result, 1)

	s.server.App().QueueUpdate(func() {
		width, height := screen.Size()
		infos := s.server.Widgets().ListBySession(sessionID)

		focused := s.server.App().GetFocus()
		var focusedWidgetID string

		var bounds []*pb.WidgetBoundsInfo
		shortIDCounter := 0

		for _, info := range infos {
			var x, y, w, h int
			if box, ok := info.Primitive.(interface {
				GetRect() (int, int, int, int)
			}); ok {
				x, y, w, h = box.GetRect()
			}

			isFocused := info.Primitive == focused
			if isFocused {
				focusedWidgetID = info.ID
			}

			shortID := generateShortID(shortIDCounter)
			shortIDCounter++

			bounds = append(bounds, &pb.WidgetBoundsInfo{
				WidgetId:   &pb.WidgetId{Id: info.ID},
				WidgetType: info.Type,
				Bounds: &pb.Rect{
					X:      int32(x),
					Y:      int32(y),
					Width:  int32(w),
					Height: int32(h),
				},
				HasFocus: isFocused,
				ShortId:  shortID,
			})
		}

		done <- result{
			bounds:       bounds,
			focusedID:    focusedWidgetID,
			screenWidth:  width,
			screenHeight: height,
		}
	})

	res := <-done

	resp := &pb.GetAllWidgetBoundsResponse{
		Widgets: res.bounds,
		ScreenSize: &pb.Size{
			Width:  int32(res.screenWidth),
			Height: int32(res.screenHeight),
		},
	}

	if res.focusedID != "" {
		resp.FocusedWidget = &pb.WidgetId{Id: res.focusedID}
	}

	return resp, nil
}

// generateShortID generates a short alphabetic ID (A, B, C, ..., Z, AA, AB, ...)
func generateShortID(index int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if index < 26 {
		return string(alphabet[index])
	}
	// For indices >= 26, use two letters
	return string(alphabet[index/26-1]) + string(alphabet[index%26])
}

// overlayWidgetBounds adds visual markers to show widget bounds on the screen dump
func overlayWidgetBounds(lines []string, bounds []*pb.WidgetBoundsInfo) []string {
	// Convert to rune slices for easier manipulation
	runeLines := make([][]rune, len(lines))
	for i, line := range lines {
		runeLines[i] = []rune(line)
	}

	// Draw bounds markers for each widget
	for _, b := range bounds {
		x := int(b.Bounds.X)
		y := int(b.Bounds.Y)
		w := int(b.Bounds.Width)
		h := int(b.Bounds.Height)

		if w == 0 || h == 0 {
			continue
		}

		// Draw corner markers
		if y >= 0 && y < len(runeLines) {
			if x >= 0 && x < len(runeLines[y]) {
				// Top-left corner with short ID
				for i, r := range b.ShortId {
					if x+i < len(runeLines[y]) {
						runeLines[y][x+i] = r
					}
				}
			}
		}

		// Mark focused widget with asterisk
		if b.HasFocus && y >= 0 && y < len(runeLines) {
			marker := "*" + b.ShortId
			if x >= 0 {
				for i, r := range marker {
					if x+i < len(runeLines[y]) {
						runeLines[y][x+i] = r
					}
				}
			}
		}
	}

	// Convert back to strings
	result := make([]string, len(runeLines))
	for i, runeLine := range runeLines {
		result[i] = string(runeLine)
	}

	return result
}

// Helper to check if a primitive is visible (has been drawn)
func isPrimitiveVisible(p tview.Primitive) bool {
	if box, ok := p.(interface {
		GetRect() (int, int, int, int)
	}); ok {
		_, _, w, h := box.GetRect()
		return w > 0 && h > 0
	}
	return false
}
