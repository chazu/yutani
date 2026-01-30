package server

import (
	"testing"
)

func TestMenuBarPrimitive_AddMenu(t *testing.T) {
	mb := NewMenuBarPrimitive()

	mb.AddMenu(&MenuEntry{ID: "menu-1", Title: "File"})
	mb.AddMenu(&MenuEntry{ID: "menu-2", Title: "Edit"})

	menus := mb.GetMenus()
	if len(menus) != 2 {
		t.Fatalf("Expected 2 menus, got %d", len(menus))
	}
	if menus[0].Title != "File" {
		t.Errorf("Expected first menu 'File', got %q", menus[0].Title)
	}
	if menus[1].Title != "Edit" {
		t.Errorf("Expected second menu 'Edit', got %q", menus[1].Title)
	}
}

func TestMenuBarPrimitive_RemoveMenu(t *testing.T) {
	mb := NewMenuBarPrimitive()
	mb.AddMenu(&MenuEntry{ID: "menu-1", Title: "File"})
	mb.AddMenu(&MenuEntry{ID: "menu-2", Title: "Edit"})

	ok := mb.RemoveMenu("menu-1")
	if !ok {
		t.Error("Expected RemoveMenu to return true")
	}

	menus := mb.GetMenus()
	if len(menus) != 1 {
		t.Fatalf("Expected 1 menu after removal, got %d", len(menus))
	}
	if menus[0].Title != "Edit" {
		t.Errorf("Expected remaining menu 'Edit', got %q", menus[0].Title)
	}
}

func TestMenuBarPrimitive_RemoveNonexistent(t *testing.T) {
	mb := NewMenuBarPrimitive()
	ok := mb.RemoveMenu("nonexistent")
	if ok {
		t.Error("Expected RemoveMenu to return false for nonexistent ID")
	}
}

func TestMenuBarPrimitive_GetMenu(t *testing.T) {
	mb := NewMenuBarPrimitive()
	mb.AddMenu(&MenuEntry{ID: "menu-1", Title: "File"})

	m := mb.GetMenu("menu-1")
	if m == nil {
		t.Fatal("Expected to find menu")
	}
	if m.Title != "File" {
		t.Errorf("Expected title 'File', got %q", m.Title)
	}

	m = mb.GetMenu("nonexistent")
	if m != nil {
		t.Error("Expected nil for nonexistent menu")
	}
}

func TestMenuBarPrimitive_CloseMenu(t *testing.T) {
	mb := NewMenuBarPrimitive()
	mb.AddMenu(&MenuEntry{
		ID: "menu-1", Title: "File",
		Items: []*MenuItemEntry{{ID: "item-1", Label: "New"}},
	})

	// Simulate opening a menu
	mb.menuOpen = true
	mb.activeMenuIndex = 0
	mb.activeItemIndex = 0

	if !mb.IsMenuOpen() {
		t.Error("Expected menu to be open")
	}

	mb.CloseMenu()
	if mb.IsMenuOpen() {
		t.Error("Expected menu to be closed")
	}
	if mb.activeMenuIndex != -1 {
		t.Error("Expected activeMenuIndex to be -1")
	}
	if mb.activeItemIndex != -1 {
		t.Error("Expected activeItemIndex to be -1")
	}
}

func TestMenuBarPrimitive_SelectableItemNavigation(t *testing.T) {
	mb := NewMenuBarPrimitive()
	menu := &MenuEntry{
		ID:    "menu-1",
		Title: "File",
		Items: []*MenuItemEntry{
			{ID: "item-1", Label: "New"},
			{ID: "item-2", Separator: true},
			{ID: "item-3", Label: "Open"},
			{ID: "item-4", Label: "Disabled", Disabled: true},
			{ID: "item-5", Label: "Quit"},
		},
	}
	mb.AddMenu(menu)

	// First selectable should skip separator and disabled
	first := mb.firstSelectableItem(menu)
	if first != 0 {
		t.Errorf("Expected first selectable index 0, got %d", first)
	}

	// Next from 0 should be 2 (skip separator at 1)
	next := mb.nextSelectableItem(menu, 0, 1)
	if next != 2 {
		t.Errorf("Expected next selectable from 0 to be 2, got %d", next)
	}

	// Next from 2 should be 4 (skip disabled at 3)
	next = mb.nextSelectableItem(menu, 2, 1)
	if next != 4 {
		t.Errorf("Expected next selectable from 2 to be 4, got %d", next)
	}

	// Next from 4 wraps to 0
	next = mb.nextSelectableItem(menu, 4, 1)
	if next != 0 {
		t.Errorf("Expected next selectable from 4 to wrap to 0, got %d", next)
	}

	// Previous from 0 wraps to 4
	prev := mb.nextSelectableItem(menu, 0, -1)
	if prev != 4 {
		t.Errorf("Expected prev selectable from 0 to wrap to 4, got %d", prev)
	}
}

func TestMenuBarPrimitive_MenuTitlePositions(t *testing.T) {
	mb := NewMenuBarPrimitive()
	mb.AddMenu(&MenuEntry{ID: "m1", Title: "File"})
	mb.AddMenu(&MenuEntry{ID: "m2", Title: "Edit"})
	mb.AddMenu(&MenuEntry{ID: "m3", Title: "Help"})

	positions := mb.menuTitlePositions()
	if len(positions) != 3 {
		t.Fatalf("Expected 3 positions, got %d", len(positions))
	}

	// "File" = 4 chars + 2 padding = 6 width
	if positions[0].x != 0 || positions[0].width != 6 {
		t.Errorf("Expected (x=0, w=6), got (x=%d, w=%d)", positions[0].x, positions[0].width)
	}
	// "Edit" starts at x=6, width=6
	if positions[1].x != 6 || positions[1].width != 6 {
		t.Errorf("Expected (x=6, w=6), got (x=%d, w=%d)", positions[1].x, positions[1].width)
	}
	// "Help" starts at x=12, width=6
	if positions[2].x != 12 || positions[2].width != 6 {
		t.Errorf("Expected (x=12, w=6), got (x=%d, w=%d)", positions[2].x, positions[2].width)
	}
}

func TestMenuBarPrimitive_DropdownWidth(t *testing.T) {
	mb := NewMenuBarPrimitive()
	menu := &MenuEntry{
		ID:    "menu-1",
		Title: "File",
		Items: []*MenuItemEntry{
			{ID: "i1", Label: "New"},
			{ID: "i2", Label: "Open File", ShortcutDisplay: "Ctrl+O"},
			{ID: "i3", Label: "Quit"},
		},
	}

	w := mb.dropdownWidth(menu)
	// "Open File" (9) + 2 spaces + "Ctrl+O" (6) = 17 + 4 padding = 21
	if w != 21 {
		t.Errorf("Expected dropdown width 21, got %d", w)
	}
}

func TestMenuBarPrimitive_OnItemSelected(t *testing.T) {
	mb := NewMenuBarPrimitive()

	var gotMenuID, gotItemID, gotLabel string
	mb.SetOnItemSelected(func(menuBarID, menuID, menuItemID, label string) {
		gotMenuID = menuID
		gotItemID = menuItemID
		gotLabel = label
	})

	menu := &MenuEntry{
		ID:    "menu-1",
		Title: "File",
		Items: []*MenuItemEntry{
			{ID: "item-1", Label: "New"},
		},
	}
	mb.AddMenu(menu)

	// Simulate selection
	mb.selectItem(menu.Items[0], menu)

	if gotMenuID != "menu-1" {
		t.Errorf("Expected menuID 'menu-1', got %q", gotMenuID)
	}
	if gotItemID != "item-1" {
		t.Errorf("Expected itemID 'item-1', got %q", gotItemID)
	}
	if gotLabel != "New" {
		t.Errorf("Expected label 'New', got %q", gotLabel)
	}
}
