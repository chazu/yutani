package server

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// WindowState represents the display state of a window.
type WindowState int

const (
	WindowStateNormal    WindowState = 0
	WindowStateMaximized WindowState = 1
	WindowStateMinimized WindowState = 2
)

// HitZone identifies which part of a window was clicked.
type HitZone int

const (
	HitZoneNone HitZone = iota
	HitZoneTitleBar
	HitZoneCloseButton
	HitZoneMaximizeButton
	HitZoneMinimizeButton
	HitZoneBorderLeft
	HitZoneBorderRight
	HitZoneBorderTop
	HitZoneBorderBottom
	HitZoneCornerTopLeft
	HitZoneCornerTopRight
	HitZoneCornerBottomLeft
	HitZoneCornerBottomRight
	HitZoneInterior
)

// WindowPrimitive is a custom tview.Primitive that wraps a child widget with
// a decorated frame including a title bar with [_][^][X] buttons and
// resizable borders.
type WindowPrimitive struct {
	*tview.Box

	child tview.Primitive

	// State
	state        WindowState
	restoredRect Rect // Saved position/size for restore from maximized/minimized

	// Constraints
	resizable   bool
	movable     bool
	closable    bool
	minimizable bool
	maximizable bool
	minWidth    int
	minHeight   int
	maxWidth    int
	maxHeight   int

	// Appearance
	titleBarColor         tcell.Color
	titleBarTextColor     tcell.Color
	activeBorderColor     tcell.Color
	inactiveBorderColor   tcell.Color
	active                bool

	// Callbacks (set by WindowManager)
	onMoved        func(x, y int)
	onResized      func(w, h int)
	onStateChanged func(state WindowState)
	onClosed       func()
	onActivated    func()
}

// Rect is a simple rectangle.
type Rect struct {
	X, Y, Width, Height int
}

// NewWindowPrimitive creates a new window wrapping the given child.
func NewWindowPrimitive(child tview.Primitive) *WindowPrimitive {
	w := &WindowPrimitive{
		Box:                 tview.NewBox(),
		child:               child,
		state:               WindowStateNormal,
		resizable:           true,
		movable:             true,
		closable:            true,
		minimizable:         true,
		maximizable:         true,
		minWidth:            10,
		minHeight:           3,
		maxWidth:            0, // 0 = no limit
		maxHeight:           0,
		titleBarColor:       tcell.ColorDarkCyan,
		titleBarTextColor:   tcell.ColorWhite,
		activeBorderColor:   tcell.ColorWhite,
		inactiveBorderColor: tcell.ColorGray,
		active:              false,
	}
	return w
}

// SetChild sets the child primitive.
func (w *WindowPrimitive) SetChild(child tview.Primitive) {
	w.child = child
}

// GetChild returns the child primitive.
func (w *WindowPrimitive) GetChild() tview.Primitive {
	return w.child
}

// GetWindowState returns the current window state.
func (w *WindowPrimitive) GetWindowState() WindowState {
	return w.state
}

// SetWindowState sets the window state.
func (w *WindowPrimitive) SetWindowState(state WindowState) {
	oldState := w.state
	w.state = state
	if oldState != state && w.onStateChanged != nil {
		w.onStateChanged(state)
	}
}

// GetRestoredRect returns the saved rect for restore from maximized/minimized.
func (w *WindowPrimitive) GetRestoredRect() Rect {
	return w.restoredRect
}

// SetRestoredRect saves the rect for later restore.
func (w *WindowPrimitive) SetRestoredRect(r Rect) {
	w.restoredRect = r
}

// SetActive marks the window as active or inactive.
func (w *WindowPrimitive) SetActive(active bool) {
	if !w.active && active && w.onActivated != nil {
		w.onActivated()
	}
	w.active = active
}

// IsActive returns whether the window is active.
func (w *WindowPrimitive) IsActive() bool {
	return w.active
}

// Constraint getters/setters

func (w *WindowPrimitive) IsResizable() bool   { return w.resizable }
func (w *WindowPrimitive) IsMovable() bool      { return w.movable }
func (w *WindowPrimitive) IsClosable() bool     { return w.closable }
func (w *WindowPrimitive) IsMinimizable() bool  { return w.minimizable }
func (w *WindowPrimitive) IsMaximizable() bool  { return w.maximizable }
func (w *WindowPrimitive) GetMinWidth() int     { return w.minWidth }
func (w *WindowPrimitive) GetMinHeight() int    { return w.minHeight }
func (w *WindowPrimitive) GetMaxWidth() int     { return w.maxWidth }
func (w *WindowPrimitive) GetMaxHeight() int    { return w.maxHeight }

func (w *WindowPrimitive) SetResizable(v bool)   { w.resizable = v }
func (w *WindowPrimitive) SetMovable(v bool)      { w.movable = v }
func (w *WindowPrimitive) SetClosable(v bool)     { w.closable = v }
func (w *WindowPrimitive) SetMinimizable(v bool)  { w.minimizable = v }
func (w *WindowPrimitive) SetMaximizable(v bool)  { w.maximizable = v }
func (w *WindowPrimitive) SetMinWidth(v int)      { w.minWidth = v }
func (w *WindowPrimitive) SetMinHeight(v int)     { w.minHeight = v }
func (w *WindowPrimitive) SetMaxWidth(v int)      { w.maxWidth = v }
func (w *WindowPrimitive) SetMaxHeight(v int)     { w.maxHeight = v }

// Appearance setters

func (w *WindowPrimitive) SetTitleBarColor(c tcell.Color)       { w.titleBarColor = c }
func (w *WindowPrimitive) SetTitleBarTextColor(c tcell.Color)   { w.titleBarTextColor = c }
func (w *WindowPrimitive) SetActiveBorderColor(c tcell.Color)   { w.activeBorderColor = c }
func (w *WindowPrimitive) SetInactiveBorderColor(c tcell.Color) { w.inactiveBorderColor = c }

// Callback setters

func (w *WindowPrimitive) SetOnMoved(f func(x, y int))                 { w.onMoved = f }
func (w *WindowPrimitive) SetOnResized(f func(w, h int))               { w.onResized = f }
func (w *WindowPrimitive) SetOnStateChanged(f func(state WindowState)) { w.onStateChanged = f }
func (w *WindowPrimitive) SetOnClosed(f func())                        { w.onClosed = f }
func (w *WindowPrimitive) SetOnActivated(f func())                     { w.onActivated = f }

// Draw renders the window to the screen.
func (w *WindowPrimitive) Draw(screen tcell.Screen) {
	x, y, width, height := w.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	borderColor := w.inactiveBorderColor
	if w.active {
		borderColor = w.activeBorderColor
	}

	if w.state == WindowStateMinimized {
		// Minimized: draw only title bar row (height=1)
		w.drawTitleBar(screen, x, y, width, borderColor)
		return
	}

	// Draw border
	w.drawBorder(screen, x, y, width, height, borderColor)

	// Draw title bar (first row inside border)
	w.drawTitleBar(screen, x+1, y, width-2, borderColor)

	// Draw child in interior (below title bar, inside borders)
	if w.child != nil && height > 2 {
		innerX := x + 1
		innerY := y + 1
		innerW := width - 2
		innerH := height - 2
		if innerW > 0 && innerH > 0 {
			w.child.SetRect(innerX, innerY, innerW, innerH)
			w.child.Draw(screen)
		}
	}
}

// drawBorder draws the window border.
func (w *WindowPrimitive) drawBorder(screen tcell.Screen, x, y, width, height int, borderColor tcell.Color) {
	style := tcell.StyleDefault.Foreground(borderColor).Background(tcell.ColorDefault)

	// Corners
	screen.SetContent(x, y, tview.Borders.TopLeft, nil, style)
	screen.SetContent(x+width-1, y, tview.Borders.TopRight, nil, style)
	screen.SetContent(x, y+height-1, tview.Borders.BottomLeft, nil, style)
	screen.SetContent(x+width-1, y+height-1, tview.Borders.BottomRight, nil, style)

	// Top and bottom edges
	for i := x + 1; i < x+width-1; i++ {
		screen.SetContent(i, y+height-1, tview.Borders.Horizontal, nil, style)
	}

	// Left and right edges
	for j := y + 1; j < y+height-1; j++ {
		screen.SetContent(x, j, tview.Borders.Vertical, nil, style)
		screen.SetContent(x+width-1, j, tview.Borders.Vertical, nil, style)
	}
}

// drawTitleBar draws the title bar with [_][^][X] buttons.
func (w *WindowPrimitive) drawTitleBar(screen tcell.Screen, x, y, width int, borderColor tcell.Color) {
	if width <= 0 {
		return
	}

	barStyle := tcell.StyleDefault.
		Foreground(w.titleBarTextColor).
		Background(w.titleBarColor)

	// Fill title bar background
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, ' ', nil, barStyle)
	}

	// Draw title text (left-aligned)
	title := w.GetTitle()
	if title != "" {
		title = " " + title + " "
		for i, ch := range title {
			if i >= width-9 { // Leave space for buttons
				break
			}
			screen.SetContent(x+i, y, ch, nil, barStyle)
		}
	}

	// Draw buttons on the right: [_][^][X]
	buttonStyle := tcell.StyleDefault.
		Foreground(w.titleBarTextColor).
		Background(w.titleBarColor)

	buttons := w.getButtonString()
	btnStart := x + width - len(buttons)
	if btnStart < x {
		btnStart = x
	}
	for i, ch := range buttons {
		pos := btnStart + i
		if pos < x+width {
			screen.SetContent(pos, y, ch, nil, buttonStyle)
		}
	}
}

// getButtonString returns the button string based on available features.
func (w *WindowPrimitive) getButtonString() string {
	s := ""
	if w.minimizable {
		s += "[_]"
	}
	if w.maximizable {
		if w.state == WindowStateMaximized {
			s += "[v]"
		} else {
			s += "[^]"
		}
	}
	if w.closable {
		s += "[X]"
	}
	return s
}

// HitTest determines which zone of the window was clicked.
func (w *WindowPrimitive) HitTest(screenX, screenY int) HitZone {
	x, y, width, height := w.GetRect()

	// Outside the window entirely
	if screenX < x || screenX >= x+width || screenY < y || screenY >= y+height {
		return HitZoneNone
	}

	// Minimized: only title bar
	if w.state == WindowStateMinimized {
		return w.hitTestTitleBar(screenX, screenY, x, y, width)
	}

	// Title bar (top row inside border, or the border row itself since we draw
	// the title bar at y for the border case)
	if screenY == y {
		// Check if it's a corner
		if screenX == x {
			return HitZoneCornerTopLeft
		}
		if screenX == x+width-1 {
			return HitZoneCornerTopRight
		}
		// It's on the title bar row
		return w.hitTestTitleBar(screenX, screenY, x+1, y, width-2)
	}

	// Bottom edge
	if screenY == y+height-1 {
		if screenX == x {
			return HitZoneCornerBottomLeft
		}
		if screenX == x+width-1 {
			return HitZoneCornerBottomRight
		}
		return HitZoneBorderBottom
	}

	// Left edge
	if screenX == x {
		return HitZoneBorderLeft
	}

	// Right edge
	if screenX == x+width-1 {
		return HitZoneBorderRight
	}

	// Interior
	return HitZoneInterior
}

// hitTestTitleBar checks if a click on the title bar row hits a button.
func (w *WindowPrimitive) hitTestTitleBar(screenX, screenY, barX, barY, barWidth int) HitZone {
	if screenY != barY {
		return HitZoneNone
	}

	buttons := w.getButtonString()
	btnStart := barX + barWidth - len(buttons)

	if screenX >= btnStart && screenX < barX+barWidth {
		// Which button?
		relX := screenX - btnStart
		idx := 0

		if w.minimizable {
			if relX >= idx && relX < idx+3 {
				return HitZoneMinimizeButton
			}
			idx += 3
		}
		if w.maximizable {
			if relX >= idx && relX < idx+3 {
				return HitZoneMaximizeButton
			}
			idx += 3
		}
		if w.closable {
			if relX >= idx && relX < idx+3 {
				return HitZoneCloseButton
			}
		}
	}

	return HitZoneTitleBar
}

// Minimize collapses the window to just the title bar.
func (w *WindowPrimitive) Minimize() {
	if w.state == WindowStateMinimized || !w.minimizable {
		return
	}
	x, y, width, height := w.GetRect()
	w.restoredRect = Rect{X: x, Y: y, Width: width, Height: height}
	w.SetWindowState(WindowStateMinimized)
}

// Maximize expands the window to fill the parent area.
func (w *WindowPrimitive) Maximize(parentRect Rect) {
	if w.state == WindowStateMaximized || !w.maximizable {
		return
	}
	x, y, width, height := w.GetRect()
	w.restoredRect = Rect{X: x, Y: y, Width: width, Height: height}
	w.Box.SetRect(parentRect.X, parentRect.Y, parentRect.Width, parentRect.Height)
	w.SetWindowState(WindowStateMaximized)
}

// Restore returns the window to its saved position/size.
func (w *WindowPrimitive) Restore() {
	if w.state == WindowStateNormal {
		return
	}
	r := w.restoredRect
	w.Box.SetRect(r.X, r.Y, r.Width, r.Height)
	w.SetWindowState(WindowStateNormal)
}

// Close fires the close callback.
func (w *WindowPrimitive) Close() {
	if w.onClosed != nil {
		w.onClosed()
	}
}

// InputHandler returns the input handler for keyboard events.
func (w *WindowPrimitive) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return w.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if w.child != nil {
			if handler := w.child.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

// MouseHandler returns the mouse handler.
// The actual drag logic is handled by the WindowManager.
func (w *WindowPrimitive) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return w.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Delegate interior clicks to the child
		if w.child != nil {
			x, y := event.Position()
			zone := w.HitTest(x, y)
			if zone == HitZoneInterior {
				if handler := w.child.MouseHandler(); handler != nil {
					return handler(action, event, setFocus)
				}
			}
		}
		return false, nil
	})
}

// Focus delegates focus to the child.
func (w *WindowPrimitive) Focus(delegate func(p tview.Primitive)) {
	if w.child != nil {
		delegate(w.child)
	}
}

// HasFocus returns whether the child has focus.
func (w *WindowPrimitive) HasFocus() bool {
	if w.child != nil {
		return w.child.HasFocus()
	}
	return false
}

// String returns a debug representation.
func (w *WindowPrimitive) String() string {
	x, y, width, height := w.GetRect()
	return fmt.Sprintf("Window[%q state=%d rect=(%d,%d,%d,%d)]",
		w.GetTitle(), w.state, x, y, width, height)
}
