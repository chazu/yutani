package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Window represents a window widget.
type Window struct {
	baseWidget
}

// WindowBuilder provides a fluent interface for creating window widgets.
type WindowBuilder struct {
	client *Client
	props  *pb.WidgetProperties
	winP   *pb.WindowProperties
}

// NewWindow creates a new window builder.
func (c *Client) NewWindow() *WindowBuilder {
	return &WindowBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
		winP:   &pb.WindowProperties{},
	}
}

// Title sets the window title.
func (b *WindowBuilder) Title(title string) *WindowBuilder {
	b.props.Title = &title
	return b
}

// Rect sets the initial window position and size.
func (b *WindowBuilder) Rect(x, y, width, height int32) *WindowBuilder {
	b.winP.InitialRect = &pb.Rect{X: x, Y: y, Width: width, Height: height}
	return b
}

// Resizable sets whether the window can be resized.
func (b *WindowBuilder) Resizable(v bool) *WindowBuilder {
	b.winP.Resizable = &v
	return b
}

// Movable sets whether the window can be moved.
func (b *WindowBuilder) Movable(v bool) *WindowBuilder {
	b.winP.Movable = &v
	return b
}

// Closable sets whether the window has a close button.
func (b *WindowBuilder) Closable(v bool) *WindowBuilder {
	b.winP.Closable = &v
	return b
}

// Minimizable sets whether the window has a minimize button.
func (b *WindowBuilder) Minimizable(v bool) *WindowBuilder {
	b.winP.Minimizable = &v
	return b
}

// Maximizable sets whether the window has a maximize button.
func (b *WindowBuilder) Maximizable(v bool) *WindowBuilder {
	b.winP.Maximizable = &v
	return b
}

// MinSize sets the minimum window size.
func (b *WindowBuilder) MinSize(width, height int32) *WindowBuilder {
	b.winP.MinWidth = &width
	b.winP.MinHeight = &height
	return b
}

// MaxSize sets the maximum window size.
func (b *WindowBuilder) MaxSize(width, height int32) *WindowBuilder {
	b.winP.MaxWidth = &width
	b.winP.MaxHeight = &height
	return b
}

// TitleBarColor sets the title bar background color.
func (b *WindowBuilder) TitleBarColor(color *pb.Color) *WindowBuilder {
	b.winP.TitleBarColor = color
	return b
}

// TitleBarTextColor sets the title bar text color.
func (b *WindowBuilder) TitleBarTextColor(color *pb.Color) *WindowBuilder {
	b.winP.TitleBarTextColor = color
	return b
}

// ActiveBorderColor sets the border color when active.
func (b *WindowBuilder) ActiveBorderColor(color *pb.Color) *WindowBuilder {
	b.winP.ActiveBorderColor = color
	return b
}

// InactiveBorderColor sets the border color when inactive.
func (b *WindowBuilder) InactiveBorderColor(color *pb.Color) *WindowBuilder {
	b.winP.InactiveBorderColor = color
	return b
}

// Build creates the window widget on the server.
func (b *WindowBuilder) Build() (*Window, error) {
	b.props.TypeProperties = &pb.WidgetProperties_Window{Window: b.winP}

	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_WINDOW,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	win := &Window{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}
	b.client.registerWidget(win)
	return win, nil
}

// Move moves the window to the given position.
func (w *Window) Move(x, y int32) error {
	_, err := w.client.windowClient.WindowMove(w.client.ctx, &pb.WindowMoveRequest{
		SessionId: w.client.sessionID,
		WindowId:  w.widgetID,
		X:         x,
		Y:         y,
	})
	return err
}

// Resize resizes the window.
func (w *Window) Resize(width, height int32) error {
	_, err := w.client.windowClient.WindowResize(w.client.ctx, &pb.WindowResizeRequest{
		SessionId: w.client.sessionID,
		WindowId:  w.widgetID,
		Width:     width,
		Height:    height,
	})
	return err
}

// Maximize maximizes the window.
func (w *Window) Maximize() error {
	_, err := w.client.windowClient.WindowSetState(w.client.ctx, &pb.WindowSetStateRequest{
		SessionId: w.client.sessionID,
		WindowId:  w.widgetID,
		State:     pb.WindowState_WINDOW_MAXIMIZED,
	})
	return err
}

// Minimize minimizes the window.
func (w *Window) Minimize() error {
	_, err := w.client.windowClient.WindowSetState(w.client.ctx, &pb.WindowSetStateRequest{
		SessionId: w.client.sessionID,
		WindowId:  w.widgetID,
		State:     pb.WindowState_WINDOW_MINIMIZED,
	})
	return err
}

// Restore restores the window to normal state.
func (w *Window) Restore() error {
	_, err := w.client.windowClient.WindowSetState(w.client.ctx, &pb.WindowSetStateRequest{
		SessionId: w.client.sessionID,
		WindowId:  w.widgetID,
		State:     pb.WindowState_WINDOW_NORMAL,
	})
	return err
}

// GetState returns the window state and geometry.
func (w *Window) GetState() (*pb.WindowGetStateResponse, error) {
	return w.client.windowClient.WindowGetState(w.client.ctx, &pb.WindowGetStateRequest{
		SessionId: w.client.sessionID,
		WindowId:  w.widgetID,
	})
}

// SetConstraints sets window constraints.
func (w *Window) SetConstraints(constraints *pb.WindowConstraints) error {
	_, err := w.client.windowClient.WindowSetConstraints(w.client.ctx, &pb.WindowSetConstraintsRequest{
		SessionId:   w.client.sessionID,
		WindowId:    w.widgetID,
		Constraints: constraints,
	})
	return err
}

// WindowManager represents a window manager widget.
type WindowManager struct {
	baseWidget
}

// WindowManagerBuilder provides a fluent interface for creating window manager widgets.
type WindowManagerBuilder struct {
	client *Client
	props  *pb.WidgetProperties
	wmP    *pb.WindowManagerProperties
}

// NewWindowManager creates a new window manager builder.
func (c *Client) NewWindowManager() *WindowManagerBuilder {
	return &WindowManagerBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
		wmP:    &pb.WindowManagerProperties{},
	}
}

// Title sets the window manager title.
func (b *WindowManagerBuilder) Title(title string) *WindowManagerBuilder {
	b.props.Title = &title
	return b
}

// BackgroundColor sets the background color.
func (b *WindowManagerBuilder) BackgroundColor(color *pb.Color) *WindowManagerBuilder {
	b.wmP.BackgroundColor = color
	return b
}

// Build creates the window manager widget on the server.
func (b *WindowManagerBuilder) Build() (*WindowManager, error) {
	b.props.TypeProperties = &pb.WidgetProperties_WindowManager{WindowManager: b.wmP}

	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_WINDOW_MANAGER,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	wm := &WindowManager{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}
	b.client.registerWidget(wm)
	return wm, nil
}

// AddWindow adds a window to the manager.
func (wm *WindowManager) AddWindow(win *Window) error {
	_, err := wm.client.windowClient.WindowManagerAddWindow(wm.client.ctx, &pb.WindowManagerAddWindowRequest{
		SessionId: wm.client.sessionID,
		ManagerId: wm.widgetID,
		WindowId:  win.widgetID,
	})
	return err
}

// RemoveWindow removes a window from the manager.
func (wm *WindowManager) RemoveWindow(win *Window) error {
	_, err := wm.client.windowClient.WindowManagerRemoveWindow(wm.client.ctx, &pb.WindowManagerRemoveWindowRequest{
		SessionId: wm.client.sessionID,
		ManagerId: wm.widgetID,
		WindowId:  win.widgetID,
	})
	return err
}

// BringToFront brings a window to the front.
func (wm *WindowManager) BringToFront(win *Window) error {
	_, err := wm.client.windowClient.WindowBringToFront(wm.client.ctx, &pb.WindowBringToFrontRequest{
		SessionId: wm.client.sessionID,
		ManagerId: wm.widgetID,
		WindowId:  win.widgetID,
	})
	return err
}

// SendToBack sends a window to the back.
func (wm *WindowManager) SendToBack(win *Window) error {
	_, err := wm.client.windowClient.WindowSendToBack(wm.client.ctx, &pb.WindowSendToBackRequest{
		SessionId: wm.client.sessionID,
		ManagerId: wm.widgetID,
		WindowId:  win.widgetID,
	})
	return err
}

// GetZOrder returns the z-order of windows (front to back).
func (wm *WindowManager) GetZOrder() ([]string, error) {
	resp, err := wm.client.windowClient.WindowManagerGetZOrder(wm.client.ctx, &pb.WindowManagerGetZOrderRequest{
		SessionId: wm.client.sessionID,
		ManagerId: wm.widgetID,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(resp.WindowIds))
	for i, wid := range resp.WindowIds {
		ids[i] = wid.Id
	}
	return ids, nil
}
