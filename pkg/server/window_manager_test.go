package server

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestNewWindowManagerPrimitive(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	if len(wm.GetWindows()) != 0 {
		t.Error("expected empty window list")
	}
}

func TestWindowManager_AddRemove(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	win1 := NewWindowPrimitive(tview.NewBox())
	win2 := NewWindowPrimitive(tview.NewBox())

	wm.AddWindow(win1)
	if len(wm.GetWindows()) != 1 {
		t.Errorf("expected 1 window, got %d", len(wm.GetWindows()))
	}
	if !win1.IsActive() {
		t.Error("expected win1 to be active")
	}

	wm.AddWindow(win2)
	if len(wm.GetWindows()) != 2 {
		t.Errorf("expected 2 windows, got %d", len(wm.GetWindows()))
	}
	if win1.IsActive() {
		t.Error("expected win1 to be inactive after adding win2")
	}
	if !win2.IsActive() {
		t.Error("expected win2 to be active")
	}

	wm.RemoveWindow(win1)
	if len(wm.GetWindows()) != 1 {
		t.Errorf("expected 1 window after remove, got %d", len(wm.GetWindows()))
	}
}

func TestWindowManager_ZOrder(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	win1 := NewWindowPrimitive(tview.NewBox())
	win1.SetTitle("win1")
	win2 := NewWindowPrimitive(tview.NewBox())
	win2.SetTitle("win2")
	win3 := NewWindowPrimitive(tview.NewBox())
	win3.SetTitle("win3")

	wm.AddWindow(win1)
	wm.AddWindow(win2)
	wm.AddWindow(win3)

	// Z-order: front-to-back should be win3, win2, win1
	zorder := wm.GetZOrder()
	if len(zorder) != 3 {
		t.Fatalf("expected 3 windows in z-order, got %d", len(zorder))
	}
	if zorder[0] != win3 {
		t.Error("expected win3 at front")
	}
	if zorder[1] != win2 {
		t.Error("expected win2 in middle")
	}
	if zorder[2] != win1 {
		t.Error("expected win1 at back")
	}
}

func TestWindowManager_BringToFront(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	win1 := NewWindowPrimitive(tview.NewBox())
	win2 := NewWindowPrimitive(tview.NewBox())
	win3 := NewWindowPrimitive(tview.NewBox())

	wm.AddWindow(win1)
	wm.AddWindow(win2)
	wm.AddWindow(win3)

	// Bring win1 to front
	wm.BringToFront(win1)

	zorder := wm.GetZOrder()
	if zorder[0] != win1 {
		t.Error("expected win1 at front after BringToFront")
	}
	if !win1.IsActive() {
		t.Error("expected win1 to be active")
	}
	if win3.IsActive() {
		t.Error("expected win3 to be inactive")
	}
}

func TestWindowManager_SendToBack(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	win1 := NewWindowPrimitive(tview.NewBox())
	win2 := NewWindowPrimitive(tview.NewBox())
	win3 := NewWindowPrimitive(tview.NewBox())

	wm.AddWindow(win1)
	wm.AddWindow(win2)
	wm.AddWindow(win3)

	// Send win3 to back
	wm.SendToBack(win3)

	zorder := wm.GetZOrder()
	if zorder[len(zorder)-1] != win3 {
		t.Error("expected win3 at back after SendToBack")
	}
	if win3.IsActive() {
		t.Error("expected win3 to be inactive after send to back")
	}
	// win2 should now be the active (frontmost)
	if !zorder[0].IsActive() {
		t.Error("expected frontmost window to be active")
	}
}

func TestWindowManager_DragState(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	wm.SetRect(0, 0, 80, 24)

	win := NewWindowPrimitive(tview.NewBox())
	win.SetTitle("Draggable")
	win.SetRect(10, 5, 30, 15)

	wm.AddWindow(win)

	// Simulate mouse down on title bar (y=5 is the top row where title is drawn)
	consumed, _ := wm.handleMouseDown(15, 5, func(p tview.Primitive) {})
	if !consumed {
		t.Error("expected mouse down on title bar to be consumed")
	}
	if !wm.dragState.Active {
		t.Error("expected drag state to be active")
	}
	if wm.dragState.Mode != DragModeMove {
		t.Errorf("expected move mode, got %d", wm.dragState.Mode)
	}

	// Simulate drag motion
	consumed, _ = wm.handleDrag(tview.MouseMove, 20, 10)
	if !consumed {
		t.Error("expected drag motion to be consumed")
	}
	x, y, _, _ := win.GetRect()
	if x != 15 || y != 10 {
		t.Errorf("expected moved window at (15,10), got (%d,%d)", x, y)
	}

	// Simulate release
	consumed, _ = wm.handleDrag(tview.MouseLeftUp, 20, 10)
	if !consumed {
		t.Error("expected drag release to be consumed")
	}
	if wm.dragState.Active {
		t.Error("expected drag state to be cleared after release")
	}
}

func TestWindowManager_ResizeDrag(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	wm.SetRect(0, 0, 80, 24)

	win := NewWindowPrimitive(tview.NewBox())
	win.SetRect(10, 5, 30, 15)

	wm.AddWindow(win)

	// Start resize from bottom-right corner
	wm.startResize(win, DragModeResizeBottomRight, 39, 19)
	if !wm.dragState.Active {
		t.Error("expected resize drag to be active")
	}

	// Drag to expand
	wm.handleDrag(tview.MouseMove, 44, 21)
	_, _, w, h := win.GetRect()
	if w != 35 {
		t.Errorf("expected width 35 after resize, got %d", w)
	}
	if h != 17 {
		t.Errorf("expected height 17 after resize, got %d", h)
	}

	// Release
	wm.handleDrag(tview.MouseLeftUp, 44, 21)
	if wm.dragState.Active {
		t.Error("expected drag state cleared after release")
	}
}

func TestWindowManager_ClampConstraints(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	win := NewWindowPrimitive(tview.NewBox())
	win.SetMinWidth(15)
	win.SetMinHeight(8)
	win.SetMaxWidth(50)
	win.SetMaxHeight(30)

	// Test clamp width
	if wm.clampWidth(win, 5) != 15 {
		t.Error("expected width clamped to min 15")
	}
	if wm.clampWidth(win, 60) != 50 {
		t.Error("expected width clamped to max 50")
	}
	if wm.clampWidth(win, 30) != 30 {
		t.Error("expected width 30 unchanged")
	}

	// Test clamp height
	if wm.clampHeight(win, 3) != 8 {
		t.Error("expected height clamped to min 8")
	}
	if wm.clampHeight(win, 40) != 30 {
		t.Error("expected height clamped to max 30")
	}
}

func TestWindowManager_Draw(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)

	wm := NewWindowManagerPrimitive()
	wm.SetRect(0, 0, 80, 24)

	win := NewWindowPrimitive(tview.NewTextView())
	win.SetTitle("TestWin")
	win.SetRect(5, 5, 30, 10)
	wm.AddWindow(win)

	// Should not panic
	wm.Draw(screen)
	screen.Show()
}

func TestWindowManager_FindWindow(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	win1 := NewWindowPrimitive(tview.NewBox())
	win2 := NewWindowPrimitive(tview.NewBox())

	wm.AddWindow(win1)
	wm.AddWindow(win2)

	if wm.FindWindow(win1) != 0 {
		t.Error("expected win1 at index 0")
	}
	if wm.FindWindow(win2) != 1 {
		t.Error("expected win2 at index 1")
	}

	unknown := NewWindowPrimitive(tview.NewBox())
	if wm.FindWindow(unknown) != -1 {
		t.Error("expected -1 for unknown window")
	}
}

func TestWindowManager_RemoveNonExistent(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	win := NewWindowPrimitive(tview.NewBox())

	if wm.RemoveWindow(win) {
		t.Error("should return false for non-existent window")
	}
}

func TestWindowManager_MoveCallback(t *testing.T) {
	wm := NewWindowManagerPrimitive()
	wm.SetRect(0, 0, 80, 24)

	win := NewWindowPrimitive(tview.NewBox())
	win.SetRect(10, 5, 30, 15)

	movedCalled := false
	wm.SetOnWindowMoved(func(w *WindowPrimitive) {
		movedCalled = true
	})

	wm.AddWindow(win)

	// Start move drag on title bar
	wm.handleMouseDown(15, 5, func(p tview.Primitive) {})
	wm.handleDrag(tview.MouseMove, 20, 10)
	wm.handleDrag(tview.MouseLeftUp, 20, 10)

	if !movedCalled {
		t.Error("expected onWindowMoved callback to be called")
	}
}
