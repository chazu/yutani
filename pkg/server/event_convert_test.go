package server

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestConvertKey(t *testing.T) {
	tests := []struct {
		name   string
		tcell  tcell.Key
		expect pb.Key
	}{
		{"Enter", tcell.KeyEnter, pb.Key_KEY_ENTER},
		{"Tab", tcell.KeyTab, pb.Key_KEY_TAB},
		{"Backspace", tcell.KeyBackspace, pb.Key_KEY_BACKSPACE},
		{"Backspace2", tcell.KeyBackspace2, pb.Key_KEY_BACKSPACE},
		{"Escape", tcell.KeyEscape, pb.Key_KEY_ESCAPE},
		{"Up", tcell.KeyUp, pb.Key_KEY_UP},
		{"Down", tcell.KeyDown, pb.Key_KEY_DOWN},
		{"Left", tcell.KeyLeft, pb.Key_KEY_LEFT},
		{"Right", tcell.KeyRight, pb.Key_KEY_RIGHT},
		{"Home", tcell.KeyHome, pb.Key_KEY_HOME},
		{"End", tcell.KeyEnd, pb.Key_KEY_END},
		{"PgUp", tcell.KeyPgUp, pb.Key_KEY_PGUP},
		{"PgDn", tcell.KeyPgDn, pb.Key_KEY_PGDN},
		{"Delete", tcell.KeyDelete, pb.Key_KEY_DELETE},
		{"Insert", tcell.KeyInsert, pb.Key_KEY_INSERT},
		{"F1", tcell.KeyF1, pb.Key_KEY_F1},
		{"F2", tcell.KeyF2, pb.Key_KEY_F2},
		{"F3", tcell.KeyF3, pb.Key_KEY_F3},
		{"F4", tcell.KeyF4, pb.Key_KEY_F4},
		{"F5", tcell.KeyF5, pb.Key_KEY_F5},
		{"F6", tcell.KeyF6, pb.Key_KEY_F6},
		{"F7", tcell.KeyF7, pb.Key_KEY_F7},
		{"F8", tcell.KeyF8, pb.Key_KEY_F8},
		{"F9", tcell.KeyF9, pb.Key_KEY_F9},
		{"F10", tcell.KeyF10, pb.Key_KEY_F10},
		{"F11", tcell.KeyF11, pb.Key_KEY_F11},
		{"F12", tcell.KeyF12, pb.Key_KEY_F12},
		{"Rune", tcell.KeyRune, pb.Key_KEY_RUNE},
		{"Unknown defaults to Rune", tcell.Key(9999), pb.Key_KEY_RUNE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertKey(tt.tcell)
			if got != tt.expect {
				t.Errorf("convertKey(%v) = %v, want %v", tt.tcell, got, tt.expect)
			}
		})
	}
}

func TestConvertModifiers_Single(t *testing.T) {
	tests := []struct {
		name   string
		mod    tcell.ModMask
		expect []pb.Modifier
	}{
		{"Shift", tcell.ModShift, []pb.Modifier{pb.Modifier_MOD_SHIFT}},
		{"Ctrl", tcell.ModCtrl, []pb.Modifier{pb.Modifier_MOD_CTRL}},
		{"Alt", tcell.ModAlt, []pb.Modifier{pb.Modifier_MOD_ALT}},
		{"Meta", tcell.ModMeta, []pb.Modifier{pb.Modifier_MOD_META}},
		{"Hyper", tcell.ModHyper, []pb.Modifier{pb.Modifier_MOD_HYPER}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertModifiers(tt.mod)
			if len(got) != len(tt.expect) {
				t.Fatalf("convertModifiers(%v) returned %d modifiers, want %d", tt.mod, len(got), len(tt.expect))
			}
			for i, m := range got {
				if m != tt.expect[i] {
					t.Errorf("convertModifiers(%v)[%d] = %v, want %v", tt.mod, i, m, tt.expect[i])
				}
			}
		})
	}
}

func TestConvertModifiers_Combined(t *testing.T) {
	tests := []struct {
		name   string
		mod    tcell.ModMask
		expect []pb.Modifier
	}{
		{
			"Ctrl+Shift",
			tcell.ModCtrl | tcell.ModShift,
			[]pb.Modifier{pb.Modifier_MOD_SHIFT, pb.Modifier_MOD_CTRL},
		},
		{
			"Ctrl+Alt+Shift",
			tcell.ModCtrl | tcell.ModAlt | tcell.ModShift,
			[]pb.Modifier{pb.Modifier_MOD_SHIFT, pb.Modifier_MOD_CTRL, pb.Modifier_MOD_ALT},
		},
		{
			"All modifiers",
			tcell.ModShift | tcell.ModCtrl | tcell.ModAlt | tcell.ModMeta | tcell.ModHyper,
			[]pb.Modifier{pb.Modifier_MOD_SHIFT, pb.Modifier_MOD_CTRL, pb.Modifier_MOD_ALT, pb.Modifier_MOD_META, pb.Modifier_MOD_HYPER},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertModifiers(tt.mod)
			if len(got) != len(tt.expect) {
				t.Fatalf("convertModifiers(%v) returned %d modifiers, want %d", tt.mod, len(got), len(tt.expect))
			}
			for i, m := range got {
				if m != tt.expect[i] {
					t.Errorf("convertModifiers(%v)[%d] = %v, want %v", tt.mod, i, m, tt.expect[i])
				}
			}
		})
	}
}

func TestConvertModifiers_None(t *testing.T) {
	got := convertModifiers(tcell.ModNone)
	if len(got) != 0 {
		t.Errorf("convertModifiers(ModNone) returned %v, want empty slice", got)
	}
}

func TestConvertKeyEvent(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name      string
		key       tcell.Key
		r         rune
		mod       tcell.ModMask
		expectKey pb.Key
		expectMod []pb.Modifier
	}{
		{
			name:      "plain rune 'a'",
			key:       tcell.KeyRune,
			r:         'a',
			mod:       tcell.ModNone,
			expectKey: pb.Key_KEY_RUNE,
			expectMod: nil,
		},
		{
			name:      "Ctrl+C",
			key:       tcell.KeyRune,
			r:         'c',
			mod:       tcell.ModCtrl,
			expectKey: pb.Key_KEY_RUNE,
			expectMod: []pb.Modifier{pb.Modifier_MOD_CTRL},
		},
		{
			name:      "Shift+Tab (Backtab)",
			key:       tcell.KeyTab,
			r:         0,
			mod:       tcell.ModShift,
			expectKey: pb.Key_KEY_BACKTAB,
			expectMod: nil, // tcell normalizes Shift+Tab → Backtab, consuming the Shift modifier
		},
		{
			name:      "Hyper+Enter",
			key:       tcell.KeyEnter,
			r:         0,
			mod:       tcell.ModHyper,
			expectKey: pb.Key_KEY_ENTER,
			expectMod: []pb.Modifier{pb.Modifier_MOD_HYPER},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := tcell.NewEventKey(tt.key, tt.r, tt.mod)
			result := s.convertKeyEvent("test-session", ev)

			if result == nil {
				t.Fatal("convertKeyEvent returned nil")
			}

			keyEvent := result.GetKey()
			if keyEvent == nil {
				t.Fatal("event does not contain KeyEvent")
			}

			if keyEvent.Key != tt.expectKey {
				t.Errorf("key = %v, want %v", keyEvent.Key, tt.expectKey)
			}

			if len(keyEvent.Modifiers) != len(tt.expectMod) {
				t.Fatalf("modifiers count = %d, want %d", len(keyEvent.Modifiers), len(tt.expectMod))
			}
			for i, m := range keyEvent.Modifiers {
				if m != tt.expectMod[i] {
					t.Errorf("modifier[%d] = %v, want %v", i, m, tt.expectMod[i])
				}
			}
		})
	}
}
