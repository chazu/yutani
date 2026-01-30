package server

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MenuItemEntry represents a single item in a menu.
type MenuItemEntry struct {
	ID              string
	Label           string
	ShortcutDisplay string
	Separator       bool
	Disabled        bool
}

// MenuEntry represents a menu that belongs to a menu bar.
type MenuEntry struct {
	ID    string
	Title string
	Items []*MenuItemEntry
}

// MenuBarPrimitive is a custom tview.Primitive that renders a horizontal menu bar
// with dropdown menus. It supports mouse clicks and keyboard navigation.
type MenuBarPrimitive struct {
	*tview.Box

	menus []*MenuEntry

	// State
	activeMenuIndex int  // -1 = no menu open
	activeItemIndex int  // -1 = no item highlighted
	menuOpen        bool // whether a dropdown is currently shown

	// Appearance
	bgColor       tcell.Color
	textColor     tcell.Color
	activeBg      tcell.Color
	activeText    tcell.Color
	disabledColor tcell.Color
	separatorChar rune

	// Focus management
	previousFocus tview.Primitive
	setFocusFn    func(p tview.Primitive) // set by Focus() delegate

	// Callbacks
	onItemSelected func(menuBarID, menuID, menuItemID, label string)
}

// NewMenuBarPrimitive creates a new menu bar.
func NewMenuBarPrimitive() *MenuBarPrimitive {
	mb := &MenuBarPrimitive{
		Box:             tview.NewBox(),
		activeMenuIndex: -1,
		activeItemIndex: -1,
		bgColor:         tcell.ColorDarkBlue,
		textColor:       tcell.ColorWhite,
		activeBg:        tcell.ColorWhite,
		activeText:      tcell.ColorBlack,
		disabledColor:   tcell.ColorGray,
		separatorChar:   '─',
	}
	return mb
}

// AddMenu adds a menu to the menu bar.
func (mb *MenuBarPrimitive) AddMenu(menu *MenuEntry) {
	mb.menus = append(mb.menus, menu)
}

// RemoveMenu removes a menu by ID.
func (mb *MenuBarPrimitive) RemoveMenu(menuID string) bool {
	for i, m := range mb.menus {
		if m.ID == menuID {
			mb.menus = append(mb.menus[:i], mb.menus[i+1:]...)
			return true
		}
	}
	return false
}

// GetMenu returns a menu by ID.
func (mb *MenuBarPrimitive) GetMenu(menuID string) *MenuEntry {
	for _, m := range mb.menus {
		if m.ID == menuID {
			return m
		}
	}
	return nil
}

// SetOnItemSelected sets the callback for when a menu item is selected.
func (mb *MenuBarPrimitive) SetOnItemSelected(fn func(menuBarID, menuID, menuItemID, label string)) {
	mb.onItemSelected = fn
}

// SetBackgroundColor sets the menu bar background color.
func (mb *MenuBarPrimitive) SetBarBackgroundColor(c tcell.Color) {
	mb.bgColor = c
}

// SetTextColor sets the menu bar text color.
func (mb *MenuBarPrimitive) SetBarTextColor(c tcell.Color) {
	mb.textColor = c
}

// SetActiveBackgroundColor sets the active menu title background.
func (mb *MenuBarPrimitive) SetActiveBackgroundColor(c tcell.Color) {
	mb.activeBg = c
}

// SetActiveTextColor sets the active menu title text color.
func (mb *MenuBarPrimitive) SetActiveTextColor(c tcell.Color) {
	mb.activeText = c
}

// IsMenuOpen returns whether a dropdown is currently open.
func (mb *MenuBarPrimitive) IsMenuOpen() bool {
	return mb.menuOpen
}

// CloseMenu closes any open dropdown.
func (mb *MenuBarPrimitive) CloseMenu() {
	mb.menuOpen = false
	mb.activeMenuIndex = -1
	mb.activeItemIndex = -1
}

// menuTitlePositions returns the x-offsets and widths of each menu title.
func (mb *MenuBarPrimitive) menuTitlePositions() []struct{ x, width int } {
	positions := make([]struct{ x, width int }, len(mb.menus))
	x := 0
	for i, menu := range mb.menus {
		// Each title has 1 space padding on each side
		titleWidth := len(menu.Title) + 2
		positions[i] = struct{ x, width int }{x, titleWidth}
		x += titleWidth
	}
	return positions
}

// dropdownWidth returns the width of the dropdown for the given menu.
func (mb *MenuBarPrimitive) dropdownWidth(menu *MenuEntry) int {
	maxWidth := 0
	for _, item := range menu.Items {
		w := len(item.Label)
		if item.ShortcutDisplay != "" {
			w += 2 + len(item.ShortcutDisplay) // 2 spaces between label and shortcut
		}
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth + 4 // 2 chars padding on each side
}

// firstSelectableItem returns the index of the first non-separator, non-disabled item.
func (mb *MenuBarPrimitive) firstSelectableItem(menu *MenuEntry) int {
	for i, item := range menu.Items {
		if !item.Separator && !item.Disabled {
			return i
		}
	}
	return -1
}

// nextSelectableItem returns the next selectable item index in the given direction.
func (mb *MenuBarPrimitive) nextSelectableItem(menu *MenuEntry, current, direction int) int {
	n := len(menu.Items)
	if n == 0 {
		return -1
	}
	for i := 1; i < n; i++ {
		idx := (current + direction*i + n) % n
		item := menu.Items[idx]
		if !item.Separator && !item.Disabled {
			return idx
		}
	}
	return current
}

// Draw renders the menu bar and any open dropdown.
func (mb *MenuBarPrimitive) Draw(screen tcell.Screen) {
	mb.Box.DrawForSubclass(screen, mb)
	x, y, width, _ := mb.GetInnerRect()

	// Draw the menu bar background
	barStyle := tcell.StyleDefault.Background(mb.bgColor).Foreground(mb.textColor)
	for col := 0; col < width; col++ {
		screen.SetContent(x+col, y, ' ', nil, barStyle)
	}

	// Draw menu titles
	positions := mb.menuTitlePositions()
	for i, menu := range mb.menus {
		if i >= len(positions) {
			break
		}
		pos := positions[i]
		style := barStyle
		if mb.menuOpen && mb.activeMenuIndex == i {
			style = tcell.StyleDefault.Background(mb.activeBg).Foreground(mb.activeText)
		}

		// Draw title with padding
		titleX := x + pos.x
		screen.SetContent(titleX, y, ' ', nil, style)
		for j, r := range menu.Title {
			screen.SetContent(titleX+1+j, y, r, nil, style)
		}
		screen.SetContent(titleX+len(menu.Title)+1, y, ' ', nil, style)
	}

	// Draw dropdown if menu is open
	if mb.menuOpen && mb.activeMenuIndex >= 0 && mb.activeMenuIndex < len(mb.menus) {
		mb.drawDropdown(screen, x, y, positions)
	}
}

// drawDropdown renders the dropdown menu for the active menu.
func (mb *MenuBarPrimitive) drawDropdown(screen tcell.Screen, barX, barY int, positions []struct{ x, width int }) {
	menu := mb.menus[mb.activeMenuIndex]
	if len(menu.Items) == 0 {
		return
	}

	pos := positions[mb.activeMenuIndex]
	dropX := barX + pos.x
	dropY := barY + 1
	dropW := mb.dropdownWidth(menu)

	// Ensure minimum width matches title
	if dropW < pos.width {
		dropW = pos.width
	}

	normalStyle := tcell.StyleDefault.Background(mb.bgColor).Foreground(mb.textColor)
	highlightStyle := tcell.StyleDefault.Background(mb.activeBg).Foreground(mb.activeText)
	disabledStyle := tcell.StyleDefault.Background(mb.bgColor).Foreground(mb.disabledColor)
	separatorStyle := tcell.StyleDefault.Background(mb.bgColor).Foreground(mb.disabledColor)

	// Draw top border
	screen.SetContent(dropX, dropY, '┌', nil, normalStyle)
	for col := 1; col < dropW-1; col++ {
		screen.SetContent(dropX+col, dropY, '─', nil, normalStyle)
	}
	screen.SetContent(dropX+dropW-1, dropY, '┐', nil, normalStyle)

	// Draw items
	for i, item := range menu.Items {
		itemY := dropY + 1 + i

		if item.Separator {
			// Draw separator line
			screen.SetContent(dropX, itemY, '├', nil, separatorStyle)
			for col := 1; col < dropW-1; col++ {
				screen.SetContent(dropX+col, itemY, '─', nil, separatorStyle)
			}
			screen.SetContent(dropX+dropW-1, itemY, '┤', nil, separatorStyle)
			continue
		}

		style := normalStyle
		if item.Disabled {
			style = disabledStyle
		} else if i == mb.activeItemIndex {
			style = highlightStyle
		}

		// Left border
		screen.SetContent(dropX, itemY, '│', nil, normalStyle)

		// Item content
		label := item.Label
		shortcut := item.ShortcutDisplay

		// Draw label
		col := 1
		screen.SetContent(dropX+col, itemY, ' ', nil, style)
		col++
		for _, r := range label {
			if dropX+col < dropX+dropW-1 {
				screen.SetContent(dropX+col, itemY, r, nil, style)
				col++
			}
		}

		// Fill space between label and shortcut
		if shortcut != "" {
			shortcutStart := dropW - 2 - len(shortcut)
			for col < shortcutStart {
				screen.SetContent(dropX+col, itemY, ' ', nil, style)
				col++
			}
			for _, r := range shortcut {
				if dropX+col < dropX+dropW-1 {
					screen.SetContent(dropX+col, itemY, r, nil, style)
					col++
				}
			}
		}

		// Fill remaining space
		for col < dropW-1 {
			screen.SetContent(dropX+col, itemY, ' ', nil, style)
			col++
		}

		// Right border
		screen.SetContent(dropX+dropW-1, itemY, '│', nil, normalStyle)
	}

	// Draw bottom border
	bottomY := dropY + 1 + len(menu.Items)
	screen.SetContent(dropX, bottomY, '└', nil, normalStyle)
	for col := 1; col < dropW-1; col++ {
		screen.SetContent(dropX+col, bottomY, '─', nil, normalStyle)
	}
	screen.SetContent(dropX+dropW-1, bottomY, '┘', nil, normalStyle)
}

// InputHandler returns the key event handler.
func (mb *MenuBarPrimitive) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return mb.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if !mb.menuOpen {
			return
		}

		menu := mb.menus[mb.activeMenuIndex]

		switch event.Key() {
		case tcell.KeyEscape:
			mb.CloseMenu()
			if mb.previousFocus != nil && setFocus != nil {
				setFocus(mb.previousFocus)
			}

		case tcell.KeyEnter:
			if mb.activeItemIndex >= 0 && mb.activeItemIndex < len(menu.Items) {
				item := menu.Items[mb.activeItemIndex]
				if !item.Separator && !item.Disabled {
					mb.selectItem(item, menu)
					if mb.previousFocus != nil && setFocus != nil {
						setFocus(mb.previousFocus)
					}
				}
			}

		case tcell.KeyUp:
			if len(menu.Items) > 0 {
				if mb.activeItemIndex < 0 {
					mb.activeItemIndex = mb.firstSelectableItem(menu)
				} else {
					mb.activeItemIndex = mb.nextSelectableItem(menu, mb.activeItemIndex, -1)
				}
			}

		case tcell.KeyDown:
			if len(menu.Items) > 0 {
				if mb.activeItemIndex < 0 {
					mb.activeItemIndex = mb.firstSelectableItem(menu)
				} else {
					mb.activeItemIndex = mb.nextSelectableItem(menu, mb.activeItemIndex, 1)
				}
			}

		case tcell.KeyLeft:
			if len(mb.menus) > 1 {
				mb.activeMenuIndex = (mb.activeMenuIndex - 1 + len(mb.menus)) % len(mb.menus)
				newMenu := mb.menus[mb.activeMenuIndex]
				mb.activeItemIndex = mb.firstSelectableItem(newMenu)
			}

		case tcell.KeyRight:
			if len(mb.menus) > 1 {
				mb.activeMenuIndex = (mb.activeMenuIndex + 1) % len(mb.menus)
				newMenu := mb.menus[mb.activeMenuIndex]
				mb.activeItemIndex = mb.firstSelectableItem(newMenu)
			}
		}
	})
}

// MouseHandler returns the mouse event handler.
func (mb *MenuBarPrimitive) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return mb.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		mx, my := event.Position()
		bx, by, _, _ := mb.GetInnerRect()

		positions := mb.menuTitlePositions()

		if action == tview.MouseLeftClick {
			// Check if click is on the menu bar
			if my == by {
				for i, pos := range positions {
					if mx >= bx+pos.x && mx < bx+pos.x+pos.width {
						if mb.menuOpen && mb.activeMenuIndex == i {
							// Toggle off
							mb.CloseMenu()
							if mb.previousFocus != nil && setFocus != nil {
								setFocus(mb.previousFocus)
							}
						} else {
							// Open menu
							mb.activeMenuIndex = i
							mb.menuOpen = true
							menu := mb.menus[i]
							mb.activeItemIndex = mb.firstSelectableItem(menu)
							if setFocus != nil {
								setFocus(mb)
							}
						}
						return true, nil
					}
				}
			}

			// Check if click is on a dropdown item
			if mb.menuOpen && mb.activeMenuIndex >= 0 && mb.activeMenuIndex < len(mb.menus) {
				menu := mb.menus[mb.activeMenuIndex]
				pos := positions[mb.activeMenuIndex]
				dropX := bx + pos.x
				dropY := by + 1
				dropW := mb.dropdownWidth(menu)
				if dropW < pos.width {
					dropW = pos.width
				}

				// Check if within dropdown bounds (including border)
				if mx >= dropX && mx < dropX+dropW && my > dropY && my <= dropY+len(menu.Items) {
					itemIdx := my - dropY - 1
					if itemIdx >= 0 && itemIdx < len(menu.Items) {
						item := menu.Items[itemIdx]
						if !item.Separator && !item.Disabled {
							mb.selectItem(item, menu)
							if mb.previousFocus != nil && setFocus != nil {
								setFocus(mb.previousFocus)
							}
							return true, nil
						}
					}
				}

				// Click outside dropdown - close it
				mb.CloseMenu()
				if mb.previousFocus != nil && setFocus != nil {
					setFocus(mb.previousFocus)
				}
				return false, nil
			}
		}

		// Mouse move for hover highlighting
		if action == tview.MouseMove && mb.menuOpen && mb.activeMenuIndex >= 0 && mb.activeMenuIndex < len(mb.menus) {
			// Check hover on menu titles
			if my == by {
				for i, pos := range positions {
					if mx >= bx+pos.x && mx < bx+pos.x+pos.width {
						if i != mb.activeMenuIndex {
							mb.activeMenuIndex = i
							menu := mb.menus[i]
							mb.activeItemIndex = mb.firstSelectableItem(menu)
						}
						return true, nil
					}
				}
			}

			// Check hover on dropdown items
			menu := mb.menus[mb.activeMenuIndex]
			pos := positions[mb.activeMenuIndex]
			dropX := bx + pos.x
			dropY := by + 1
			dropW := mb.dropdownWidth(menu)
			if dropW < pos.width {
				dropW = pos.width
			}

			if mx >= dropX && mx < dropX+dropW && my > dropY && my <= dropY+len(menu.Items) {
				itemIdx := my - dropY - 1
				if itemIdx >= 0 && itemIdx < len(menu.Items) {
					item := menu.Items[itemIdx]
					if !item.Separator && !item.Disabled {
						mb.activeItemIndex = itemIdx
					}
				}
				return true, nil
			}
		}

		return false, nil
	})
}

// selectItem fires the selection callback and closes the menu.
func (mb *MenuBarPrimitive) selectItem(item *MenuItemEntry, menu *MenuEntry) {
	if mb.onItemSelected != nil {
		mb.onItemSelected("", menu.ID, item.ID, item.Label)
	}
	mb.CloseMenu()
}

// Focus is called when the menu bar receives focus.
func (mb *MenuBarPrimitive) Focus(delegate func(p tview.Primitive)) {
	mb.setFocusFn = delegate
	mb.Box.Focus(delegate)
}

// SetPreviousFocus stores the widget that had focus before the menu opened.
func (mb *MenuBarPrimitive) SetPreviousFocus(p tview.Primitive) {
	mb.previousFocus = p
}

// HasFocus returns true if the menu bar has focus.
func (mb *MenuBarPrimitive) HasFocus() bool {
	return mb.Box.HasFocus()
}

// GetMenus returns all menus (for testing/introspection).
func (mb *MenuBarPrimitive) GetMenus() []*MenuEntry {
	return mb.menus
}
