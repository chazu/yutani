package server

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestNewWindowPrimitive(t *testing.T) {
	child := tview.NewTextView()
	win := NewWindowPrimitive(child)

	if win.GetChild() != child {
		t.Error("expected child to be set")
	}
	if win.GetWindowState() != WindowStateNormal {
		t.Errorf("expected normal state, got %d", win.GetWindowState())
	}
	if !win.IsResizable() {
		t.Error("expected resizable by default")
	}
	if !win.IsMovable() {
		t.Error("expected movable by default")
	}
	if !win.IsClosable() {
		t.Error("expected closable by default")
	}
	if !win.IsMinimizable() {
		t.Error("expected minimizable by default")
	}
	if !win.IsMaximizable() {
		t.Error("expected maximizable by default")
	}
}

func TestWindowPrimitive_StateTransitions(t *testing.T) {
	win := NewWindowPrimitive(tview.NewBox())
	win.SetRect(10, 10, 40, 20)

	// Minimize
	win.Minimize()
	if win.GetWindowState() != WindowStateMinimized {
		t.Errorf("expected minimized, got %d", win.GetWindowState())
	}
	rr := win.GetRestoredRect()
	if rr.X != 10 || rr.Y != 10 || rr.Width != 40 || rr.Height != 20 {
		t.Errorf("expected restored rect (10,10,40,20), got (%d,%d,%d,%d)", rr.X, rr.Y, rr.Width, rr.Height)
	}

	// Restore from minimized
	win.Restore()
	if win.GetWindowState() != WindowStateNormal {
		t.Errorf("expected normal after restore, got %d", win.GetWindowState())
	}
	x, y, w, h := win.GetRect()
	if x != 10 || y != 10 || w != 40 || h != 20 {
		t.Errorf("expected rect restored to (10,10,40,20), got (%d,%d,%d,%d)", x, y, w, h)
	}

	// Maximize
	parentRect := Rect{X: 0, Y: 0, Width: 80, Height: 24}
	win.Maximize(parentRect)
	if win.GetWindowState() != WindowStateMaximized {
		t.Errorf("expected maximized, got %d", win.GetWindowState())
	}
	x, y, w, h = win.GetRect()
	if x != 0 || y != 0 || w != 80 || h != 24 {
		t.Errorf("expected maximized rect (0,0,80,24), got (%d,%d,%d,%d)", x, y, w, h)
	}

	// Restore from maximized
	win.Restore()
	if win.GetWindowState() != WindowStateNormal {
		t.Errorf("expected normal after restore, got %d", win.GetWindowState())
	}
	x, y, w, h = win.GetRect()
	if x != 10 || y != 10 || w != 40 || h != 20 {
		t.Errorf("expected rect restored to (10,10,40,20), got (%d,%d,%d,%d)", x, y, w, h)
	}
}

func TestWindowPrimitive_MinimizeRestoreMinimized(t *testing.T) {
	win := NewWindowPrimitive(tview.NewBox())
	win.SetMinimizable(false)

	win.Minimize()
	if win.GetWindowState() != WindowStateNormal {
		t.Error("should not minimize when not minimizable")
	}
}

func TestWindowPrimitive_MaximizeNotAllowed(t *testing.T) {
	win := NewWindowPrimitive(tview.NewBox())
	win.SetMaximizable(false)

	win.Maximize(Rect{X: 0, Y: 0, Width: 80, Height: 24})
	if win.GetWindowState() != WindowStateNormal {
		t.Error("should not maximize when not maximizable")
	}
}

func TestWindowPrimitive_HitTest(t *testing.T) {
	win := NewWindowPrimitive(tview.NewTextView())
	win.SetTitle("Test")
	win.SetRect(5, 5, 30, 15)

	tests := []struct {
		name     string
		x, y     int
		expected HitZone
	}{
		{"outside", 0, 0, HitZoneNone},
		{"top-left corner", 5, 5, HitZoneCornerTopLeft},
		{"top-right corner", 34, 5, HitZoneCornerTopRight},
		{"bottom-left corner", 5, 19, HitZoneCornerBottomLeft},
		{"bottom-right corner", 34, 19, HitZoneCornerBottomRight},
		{"left border", 5, 10, HitZoneBorderLeft},
		{"right border", 34, 10, HitZoneBorderRight},
		{"bottom border", 15, 19, HitZoneBorderBottom},
		{"interior", 15, 10, HitZoneInterior},
		{"title bar", 10, 5, HitZoneTitleBar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zone := win.HitTest(tt.x, tt.y)
			if zone != tt.expected {
				t.Errorf("HitTest(%d,%d) = %d, want %d", tt.x, tt.y, zone, tt.expected)
			}
		})
	}
}

func TestWindowPrimitive_HitTestButtons(t *testing.T) {
	win := NewWindowPrimitive(tview.NewTextView())
	win.SetTitle("Test")
	win.SetRect(5, 5, 30, 15)

	// Buttons are: [_][^][X] = 9 chars at the right end of title bar
	// Title bar goes from x=6 to x=33 (width=28 inside border)
	// Button string = "[_][^][X]" = 9 chars
	// btnStart = 6 + 28 - 9 = 25
	// [_] = 25,26,27  [^] = 28,29,30  [X] = 31,32,33

	zone := win.HitTest(25, 5)
	if zone != HitZoneMinimizeButton {
		t.Errorf("expected minimize button at (25,5), got %d", zone)
	}

	zone = win.HitTest(28, 5)
	if zone != HitZoneMaximizeButton {
		t.Errorf("expected maximize button at (28,5), got %d", zone)
	}

	zone = win.HitTest(31, 5)
	if zone != HitZoneCloseButton {
		t.Errorf("expected close button at (31,5), got %d", zone)
	}
}

func TestWindowPrimitive_Active(t *testing.T) {
	win := NewWindowPrimitive(tview.NewBox())

	if win.IsActive() {
		t.Error("should not be active initially")
	}

	activated := false
	win.SetOnActivated(func() { activated = true })

	win.SetActive(true)
	if !win.IsActive() {
		t.Error("should be active after SetActive(true)")
	}
	if !activated {
		t.Error("onActivated should have been called")
	}
}

func TestWindowPrimitive_CloseCallback(t *testing.T) {
	win := NewWindowPrimitive(tview.NewBox())

	closed := false
	win.SetOnClosed(func() { closed = true })

	win.Close()
	if !closed {
		t.Error("close callback should have been called")
	}
}

func TestWindowPrimitive_StateChangeCallback(t *testing.T) {
	win := NewWindowPrimitive(tview.NewBox())
	win.SetRect(10, 10, 40, 20)

	var lastState WindowState
	win.SetOnStateChanged(func(state WindowState) { lastState = state })

	win.Minimize()
	if lastState != WindowStateMinimized {
		t.Errorf("expected minimized callback, got %d", lastState)
	}

	win.Restore()
	if lastState != WindowStateNormal {
		t.Errorf("expected normal callback, got %d", lastState)
	}
}

func TestWindowPrimitive_Draw(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)

	child := tview.NewTextView()
	child.SetText("Hello")
	win := NewWindowPrimitive(child)
	win.SetTitle("MyWindow")
	win.SetRect(5, 5, 30, 10)

	win.Draw(screen)
	screen.Show()

	// Verify title bar area has content (check for 'M' of MyWindow)
	r, _, _, _ := screen.GetContent(7, 5)
	if r != 'M' {
		t.Errorf("expected 'M' at (7,5), got %c", r)
	}
}

func TestWindowPrimitive_DrawMinimized(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)

	win := NewWindowPrimitive(tview.NewTextView())
	win.SetTitle("Min")
	win.SetRect(5, 5, 30, 1)
	win.SetWindowState(WindowStateMinimized)

	win.Draw(screen)
	screen.Show()

	// Should draw only the title bar row
	r, _, _, _ := screen.GetContent(6, 5)
	if r != 'M' {
		t.Errorf("expected 'M' at (6,5) for minimized title, got %c", r)
	}
}

func TestWindowPrimitive_Constraints(t *testing.T) {
	win := NewWindowPrimitive(tview.NewBox())

	win.SetMinWidth(20)
	win.SetMinHeight(10)
	win.SetMaxWidth(100)
	win.SetMaxHeight(50)

	if win.GetMinWidth() != 20 {
		t.Errorf("expected min width 20, got %d", win.GetMinWidth())
	}
	if win.GetMinHeight() != 10 {
		t.Errorf("expected min height 10, got %d", win.GetMinHeight())
	}
	if win.GetMaxWidth() != 100 {
		t.Errorf("expected max width 100, got %d", win.GetMaxWidth())
	}
	if win.GetMaxHeight() != 50 {
		t.Errorf("expected max height 50, got %d", win.GetMaxHeight())
	}
}
