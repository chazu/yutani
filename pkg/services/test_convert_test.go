package services

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestConvertProtoModifiersToTcell(t *testing.T) {
	tests := []struct {
		name   string
		mods   []pb.Modifier
		expect tcell.ModMask
	}{
		{"Shift", []pb.Modifier{pb.Modifier_MOD_SHIFT}, tcell.ModShift},
		{"Ctrl", []pb.Modifier{pb.Modifier_MOD_CTRL}, tcell.ModCtrl},
		{"Alt", []pb.Modifier{pb.Modifier_MOD_ALT}, tcell.ModAlt},
		{"Meta", []pb.Modifier{pb.Modifier_MOD_META}, tcell.ModMeta},
		{"Hyper", []pb.Modifier{pb.Modifier_MOD_HYPER}, tcell.ModHyper},
		{"None", []pb.Modifier{pb.Modifier_MOD_NONE}, tcell.ModNone},
		{"nil", nil, tcell.ModNone},
		{"empty", []pb.Modifier{}, tcell.ModNone},
		{
			"Ctrl+Shift",
			[]pb.Modifier{pb.Modifier_MOD_CTRL, pb.Modifier_MOD_SHIFT},
			tcell.ModCtrl | tcell.ModShift,
		},
		{
			"All modifiers",
			[]pb.Modifier{pb.Modifier_MOD_SHIFT, pb.Modifier_MOD_CTRL, pb.Modifier_MOD_ALT, pb.Modifier_MOD_META, pb.Modifier_MOD_HYPER},
			tcell.ModShift | tcell.ModCtrl | tcell.ModAlt | tcell.ModMeta | tcell.ModHyper,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertProtoModifiersToTcell(tt.mods)
			if got != tt.expect {
				t.Errorf("convertProtoModifiersToTcell(%v) = %v, want %v", tt.mods, got, tt.expect)
			}
		})
	}
}

func TestConvertProtoKeyToTcell(t *testing.T) {
	tests := []struct {
		name   string
		key    pb.Key
		expect tcell.Key
	}{
		{"Enter", pb.Key_KEY_ENTER, tcell.KeyEnter},
		{"Tab", pb.Key_KEY_TAB, tcell.KeyTab},
		{"Backtab", pb.Key_KEY_BACKTAB, tcell.KeyBacktab},
		{"Backspace", pb.Key_KEY_BACKSPACE, tcell.KeyBackspace2},
		{"Escape", pb.Key_KEY_ESCAPE, tcell.KeyEscape},
		{"Up", pb.Key_KEY_UP, tcell.KeyUp},
		{"Down", pb.Key_KEY_DOWN, tcell.KeyDown},
		{"Left", pb.Key_KEY_LEFT, tcell.KeyLeft},
		{"Right", pb.Key_KEY_RIGHT, tcell.KeyRight},
		{"Home", pb.Key_KEY_HOME, tcell.KeyHome},
		{"End", pb.Key_KEY_END, tcell.KeyEnd},
		{"PgUp", pb.Key_KEY_PGUP, tcell.KeyPgUp},
		{"PgDn", pb.Key_KEY_PGDN, tcell.KeyPgDn},
		{"Delete", pb.Key_KEY_DELETE, tcell.KeyDelete},
		{"Insert", pb.Key_KEY_INSERT, tcell.KeyInsert},
		{"F1", pb.Key_KEY_F1, tcell.KeyF1},
		{"F2", pb.Key_KEY_F2, tcell.KeyF2},
		{"F3", pb.Key_KEY_F3, tcell.KeyF3},
		{"F4", pb.Key_KEY_F4, tcell.KeyF4},
		{"F5", pb.Key_KEY_F5, tcell.KeyF5},
		{"F6", pb.Key_KEY_F6, tcell.KeyF6},
		{"F7", pb.Key_KEY_F7, tcell.KeyF7},
		{"F8", pb.Key_KEY_F8, tcell.KeyF8},
		{"F9", pb.Key_KEY_F9, tcell.KeyF9},
		{"F10", pb.Key_KEY_F10, tcell.KeyF10},
		{"F11", pb.Key_KEY_F11, tcell.KeyF11},
		{"F12", pb.Key_KEY_F12, tcell.KeyF12},
		{"Rune", pb.Key_KEY_RUNE, tcell.KeyRune},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertProtoKeyToTcell(tt.key)
			if got != tt.expect {
				t.Errorf("convertProtoKeyToTcell(%v) = %v, want %v", tt.key, got, tt.expect)
			}
		})
	}
}

func TestModifierRoundTrip(t *testing.T) {
	// Test that converting tcell → proto → tcell preserves the modifier
	modifiers := []struct {
		name string
		mod  tcell.ModMask
	}{
		{"Shift", tcell.ModShift},
		{"Ctrl", tcell.ModCtrl},
		{"Alt", tcell.ModAlt},
		{"Meta", tcell.ModMeta},
		{"Hyper", tcell.ModHyper},
		{"Ctrl+Shift", tcell.ModCtrl | tcell.ModShift},
		{"Ctrl+Alt+Shift", tcell.ModCtrl | tcell.ModAlt | tcell.ModShift},
		{"All", tcell.ModShift | tcell.ModCtrl | tcell.ModAlt | tcell.ModMeta | tcell.ModHyper},
	}

	for _, tt := range modifiers {
		t.Run(tt.name, func(t *testing.T) {
			// Use the server package's convertModifiers via the proto intermediate
			// We test the services side: proto → tcell
			// Build the proto modifiers manually to match what convertModifiers would produce
			var protoMods []pb.Modifier
			if tt.mod&tcell.ModShift != 0 {
				protoMods = append(protoMods, pb.Modifier_MOD_SHIFT)
			}
			if tt.mod&tcell.ModCtrl != 0 {
				protoMods = append(protoMods, pb.Modifier_MOD_CTRL)
			}
			if tt.mod&tcell.ModAlt != 0 {
				protoMods = append(protoMods, pb.Modifier_MOD_ALT)
			}
			if tt.mod&tcell.ModMeta != 0 {
				protoMods = append(protoMods, pb.Modifier_MOD_META)
			}
			if tt.mod&tcell.ModHyper != 0 {
				protoMods = append(protoMods, pb.Modifier_MOD_HYPER)
			}

			got := convertProtoModifiersToTcell(protoMods)
			if got != tt.mod {
				t.Errorf("round-trip for %v: got %v, want %v", tt.name, got, tt.mod)
			}
		})
	}
}
