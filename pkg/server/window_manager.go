package server

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DragMode identifies the type of drag operation.
type DragMode int

const (
	DragModeNone DragMode = iota
	DragModeMove
	DragModeResizeLeft
	DragModeResizeRight
	DragModeResizeTop
	DragModeResizeBottom
	DragModeResizeTopLeft
	DragModeResizeTopRight
	DragModeResizeBottomLeft
	DragModeResizeBottomRight
)

// DragState tracks an active drag operation.
type DragState struct {
	Active    bool
	Window    *WindowPrimitive
	Mode      DragMode
	StartX    int // Mouse position at drag start
	StartY    int
	StartRect Rect // Window rect at drag start
}

// WindowManagerPrimitive manages child windows with absolute positioning,
// z-ordering, and drag-to-move/resize.
type WindowManagerPrimitive struct {
	*tview.Box

	windows   []*WindowPrimitive // z-ordered: last = front
	dragState DragState
	bgColor   tcell.Color

	// Callback for window events (wired by the service layer)
	onWindowMoved        func(win *WindowPrimitive)
	onWindowResized      func(win *WindowPrimitive)
	onWindowStateChanged func(win *WindowPrimitive, state WindowState)
	onWindowClosed       func(win *WindowPrimitive)
	onWindowActivated    func(win *WindowPrimitive)
}

// NewWindowManagerPrimitive creates a new window manager.
func NewWindowManagerPrimitive() *WindowManagerPrimitive {
	wm := &WindowManagerPrimitive{
		Box:     tview.NewBox(),
		windows: make([]*WindowPrimitive, 0),
		bgColor: tcell.ColorDarkBlue,
	}
	return wm
}

// SetBgColor sets the background fill color.
func (wm *WindowManagerPrimitive) SetBgColor(c tcell.Color) {
	wm.bgColor = c
}

// AddWindow adds a window to the manager (on top).
func (wm *WindowManagerPrimitive) AddWindow(win *WindowPrimitive) {
	// Deactivate all existing windows
	for _, w := range wm.windows {
		w.SetActive(false)
	}
	win.SetActive(true)
	wm.windows = append(wm.windows, win)

	// Wire callbacks from window to manager
	wm.wireWindowCallbacks(win)
}

// RemoveWindow removes a window from the manager.
func (wm *WindowManagerPrimitive) RemoveWindow(win *WindowPrimitive) bool {
	for i, w := range wm.windows {
		if w == win {
			wm.windows = append(wm.windows[:i], wm.windows[i+1:]...)
			// Activate new front window
			if len(wm.windows) > 0 {
				wm.windows[len(wm.windows)-1].SetActive(true)
			}
			return true
		}
	}
	return false
}

// BringToFront moves a window to the top of the z-order.
func (wm *WindowManagerPrimitive) BringToFront(win *WindowPrimitive) {
	for i, w := range wm.windows {
		if w == win {
			// Remove from current position
			wm.windows = append(wm.windows[:i], wm.windows[i+1:]...)
			// Deactivate all
			for _, w2 := range wm.windows {
				w2.SetActive(false)
			}
			// Add to end (front)
			win.SetActive(true)
			wm.windows = append(wm.windows, win)
			return
		}
	}
}

// SendToBack moves a window to the bottom of the z-order.
func (wm *WindowManagerPrimitive) SendToBack(win *WindowPrimitive) {
	for i, w := range wm.windows {
		if w == win {
			wm.windows = append(wm.windows[:i], wm.windows[i+1:]...)
			win.SetActive(false)
			wm.windows = append([]*WindowPrimitive{win}, wm.windows...)
			// Activate new front
			if len(wm.windows) > 0 {
				wm.windows[len(wm.windows)-1].SetActive(true)
			}
			return
		}
	}
}

// GetZOrder returns window list from front to back.
func (wm *WindowManagerPrimitive) GetZOrder() []*WindowPrimitive {
	result := make([]*WindowPrimitive, len(wm.windows))
	for i, w := range wm.windows {
		result[len(wm.windows)-1-i] = w // reverse: front first
	}
	return result
}

// GetWindows returns the internal window list (back-to-front).
func (wm *WindowManagerPrimitive) GetWindows() []*WindowPrimitive {
	return wm.windows
}

// FindWindow returns the window at a given pointer to WindowPrimitive.
func (wm *WindowManagerPrimitive) FindWindow(win *WindowPrimitive) int {
	for i, w := range wm.windows {
		if w == win {
			return i
		}
	}
	return -1
}

// Event callback setters

func (wm *WindowManagerPrimitive) SetOnWindowMoved(f func(win *WindowPrimitive))                          { wm.onWindowMoved = f }
func (wm *WindowManagerPrimitive) SetOnWindowResized(f func(win *WindowPrimitive))                        { wm.onWindowResized = f }
func (wm *WindowManagerPrimitive) SetOnWindowStateChanged(f func(win *WindowPrimitive, state WindowState)) { wm.onWindowStateChanged = f }
func (wm *WindowManagerPrimitive) SetOnWindowClosed(f func(win *WindowPrimitive))                          { wm.onWindowClosed = f }
func (wm *WindowManagerPrimitive) SetOnWindowActivated(f func(win *WindowPrimitive))                       { wm.onWindowActivated = f }

// wireWindowCallbacks connects a window's callbacks to the manager.
func (wm *WindowManagerPrimitive) wireWindowCallbacks(win *WindowPrimitive) {
	win.SetOnMoved(func(x, y int) {
		if wm.onWindowMoved != nil {
			wm.onWindowMoved(win)
		}
	})
	win.SetOnResized(func(w, h int) {
		if wm.onWindowResized != nil {
			wm.onWindowResized(win)
		}
	})
	win.SetOnStateChanged(func(state WindowState) {
		if wm.onWindowStateChanged != nil {
			wm.onWindowStateChanged(win, state)
		}
	})
	win.SetOnClosed(func() {
		if wm.onWindowClosed != nil {
			wm.onWindowClosed(win)
		}
	})
	win.SetOnActivated(func() {
		if wm.onWindowActivated != nil {
			wm.onWindowActivated(win)
		}
	})
}

// Draw renders the window manager and all child windows.
func (wm *WindowManagerPrimitive) Draw(screen tcell.Screen) {
	x, y, width, height := wm.GetRect()

	// Fill background
	bgStyle := tcell.StyleDefault.Background(wm.bgColor)
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw windows back-to-front
	for _, win := range wm.windows {
		wx, wy, ww, wh := win.GetRect()
		if win.GetWindowState() == WindowStateMinimized {
			// Minimized windows are only 1 row tall (title bar)
			win.SetRect(wx, wy, ww, 1)
		} else if win.GetWindowState() == WindowStateMaximized {
			// Maximized windows fill the manager area
			win.SetRect(x, y, width, height)
		} else {
			// Normal: keep current rect
			win.SetRect(wx, wy, ww, wh)
		}
		win.Draw(screen)
	}
}

// InputHandler delegates keyboard input to the frontmost window.
func (wm *WindowManagerPrimitive) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return wm.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if len(wm.windows) > 0 {
			front := wm.windows[len(wm.windows)-1]
			if handler := front.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

// MouseHandler is the central mouse dispatcher for the window manager.
func (wm *WindowManagerPrimitive) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return wm.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		mx, my := event.Position()

		// Handle active drag
		if wm.dragState.Active {
			return wm.handleDrag(action, mx, my)
		}

		// Button1 press: hit-test windows front-to-back
		if action == tview.MouseLeftDown {
			return wm.handleMouseDown(mx, my, setFocus)
		}

		return false, nil
	})
}

// handleMouseDown processes a left mouse button press.
func (wm *WindowManagerPrimitive) handleMouseDown(mx, my int, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	// Hit-test from front (end) to back (start)
	for i := len(wm.windows) - 1; i >= 0; i-- {
		win := wm.windows[i]
		zone := win.HitTest(mx, my)
		if zone == HitZoneNone {
			continue
		}

		// Bring to front
		wm.BringToFront(win)

		managerRect := wm.getManagerRect()

		switch zone {
		case HitZoneCloseButton:
			if win.IsClosable() {
				win.Close()
				wm.RemoveWindow(win)
			}
			return true, nil

		case HitZoneMinimizeButton:
			if win.GetWindowState() == WindowStateMinimized {
				win.Restore()
			} else {
				win.Minimize()
			}
			return true, nil

		case HitZoneMaximizeButton:
			if win.GetWindowState() == WindowStateMaximized {
				win.Restore()
			} else {
				win.Maximize(managerRect)
			}
			return true, nil

		case HitZoneTitleBar:
			if win.IsMovable() && win.GetWindowState() != WindowStateMaximized {
				wx, wy, ww, wh := win.GetRect()
				wm.dragState = DragState{
					Active:    true,
					Window:    win,
					Mode:      DragModeMove,
					StartX:    mx,
					StartY:    my,
					StartRect: Rect{X: wx, Y: wy, Width: ww, Height: wh},
				}
			}
			return true, nil

		case HitZoneBorderLeft:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeLeft, mx, my)
			}
			return true, nil
		case HitZoneBorderRight:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeRight, mx, my)
			}
			return true, nil
		case HitZoneBorderTop:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeTop, mx, my)
			}
			return true, nil
		case HitZoneBorderBottom:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeBottom, mx, my)
			}
			return true, nil
		case HitZoneCornerTopLeft:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeTopLeft, mx, my)
			}
			return true, nil
		case HitZoneCornerTopRight:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeTopRight, mx, my)
			}
			return true, nil
		case HitZoneCornerBottomLeft:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeBottomLeft, mx, my)
			}
			return true, nil
		case HitZoneCornerBottomRight:
			if win.IsResizable() && win.GetWindowState() == WindowStateNormal {
				wm.startResize(win, DragModeResizeBottomRight, mx, my)
			}
			return true, nil

		case HitZoneInterior:
			// Delegate to child widget
			if handler := win.MouseHandler(); handler != nil {
				return handler(tview.MouseLeftDown, tcell.NewEventMouse(mx, my, tcell.Button1, tcell.ModNone), setFocus)
			}
			return true, nil
		}

		return true, nil
	}

	return false, nil
}

// startResize begins a resize drag operation.
func (wm *WindowManagerPrimitive) startResize(win *WindowPrimitive, mode DragMode, mx, my int) {
	wx, wy, ww, wh := win.GetRect()
	wm.dragState = DragState{
		Active:    true,
		Window:    win,
		Mode:      mode,
		StartX:    mx,
		StartY:    my,
		StartRect: Rect{X: wx, Y: wy, Width: ww, Height: wh},
	}
}

// handleDrag processes mouse motion/release during an active drag.
func (wm *WindowManagerPrimitive) handleDrag(action tview.MouseAction, mx, my int) (bool, tview.Primitive) {
	win := wm.dragState.Window
	dx := mx - wm.dragState.StartX
	dy := my - wm.dragState.StartY
	sr := wm.dragState.StartRect

	switch action {
	case tview.MouseMove:
		switch wm.dragState.Mode {
		case DragModeMove:
			win.SetRect(sr.X+dx, sr.Y+dy, sr.Width, sr.Height)

		case DragModeResizeRight:
			newW := wm.clampWidth(win, sr.Width+dx)
			win.SetRect(sr.X, sr.Y, newW, sr.Height)

		case DragModeResizeBottom:
			newH := wm.clampHeight(win, sr.Height+dy)
			win.SetRect(sr.X, sr.Y, sr.Width, newH)

		case DragModeResizeLeft:
			newW := wm.clampWidth(win, sr.Width-dx)
			newX := sr.X + sr.Width - newW
			win.SetRect(newX, sr.Y, newW, sr.Height)

		case DragModeResizeTop:
			newH := wm.clampHeight(win, sr.Height-dy)
			newY := sr.Y + sr.Height - newH
			win.SetRect(sr.X, newY, sr.Width, newH)

		case DragModeResizeBottomRight:
			newW := wm.clampWidth(win, sr.Width+dx)
			newH := wm.clampHeight(win, sr.Height+dy)
			win.SetRect(sr.X, sr.Y, newW, newH)

		case DragModeResizeBottomLeft:
			newW := wm.clampWidth(win, sr.Width-dx)
			newH := wm.clampHeight(win, sr.Height+dy)
			newX := sr.X + sr.Width - newW
			win.SetRect(newX, sr.Y, newW, newH)

		case DragModeResizeTopRight:
			newW := wm.clampWidth(win, sr.Width+dx)
			newH := wm.clampHeight(win, sr.Height-dy)
			newY := sr.Y + sr.Height - newH
			win.SetRect(sr.X, newY, newW, newH)

		case DragModeResizeTopLeft:
			newW := wm.clampWidth(win, sr.Width-dx)
			newH := wm.clampHeight(win, sr.Height-dy)
			newX := sr.X + sr.Width - newW
			newY := sr.Y + sr.Height - newH
			win.SetRect(newX, newY, newW, newH)
		}
		return true, nil

	case tview.MouseLeftUp:
		// Finalize drag
		wm.finalizeDrag()
		return true, nil

	default:
		// Continue drag on any other event while active
		return true, nil
	}
}

// finalizeDrag ends the drag and fires appropriate callbacks.
func (wm *WindowManagerPrimitive) finalizeDrag() {
	win := wm.dragState.Window
	mode := wm.dragState.Mode
	wm.dragState = DragState{}

	if win == nil {
		return
	}

	x, y, w, h := win.GetRect()

	if mode == DragModeMove {
		if win.onMoved != nil {
			win.onMoved(x, y)
		}
	} else {
		if win.onResized != nil {
			win.onResized(w, h)
		}
	}
}

// clampWidth enforces min/max width constraints.
func (wm *WindowManagerPrimitive) clampWidth(win *WindowPrimitive, w int) int {
	if w < win.GetMinWidth() {
		w = win.GetMinWidth()
	}
	if win.GetMaxWidth() > 0 && w > win.GetMaxWidth() {
		w = win.GetMaxWidth()
	}
	return w
}

// clampHeight enforces min/max height constraints.
func (wm *WindowManagerPrimitive) clampHeight(win *WindowPrimitive, h int) int {
	if h < win.GetMinHeight() {
		h = win.GetMinHeight()
	}
	if win.GetMaxHeight() > 0 && h > win.GetMaxHeight() {
		h = win.GetMaxHeight()
	}
	return h
}

// getManagerRect returns the manager's rect as a Rect.
func (wm *WindowManagerPrimitive) getManagerRect() Rect {
	x, y, w, h := wm.GetRect()
	return Rect{X: x, Y: y, Width: w, Height: h}
}

// Focus delegates to the frontmost window.
func (wm *WindowManagerPrimitive) Focus(delegate func(p tview.Primitive)) {
	if len(wm.windows) > 0 {
		delegate(wm.windows[len(wm.windows)-1])
	}
}

// HasFocus returns whether any window has focus.
func (wm *WindowManagerPrimitive) HasFocus() bool {
	for _, win := range wm.windows {
		if win.HasFocus() {
			return true
		}
	}
	return false
}
