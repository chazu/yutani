package server

import (
	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// convertKeyEvent converts a tcell key event to protobuf
func (s *Server) convertKeyEvent(sessionID string, event *tcell.EventKey) *pb.Event {
	keyEvent := &pb.KeyEvent{
		Key:       convertKey(event.Key()),
		Rune:      string(event.Rune()),
		Modifiers: convertModifiers(event.Modifiers()),
	}

	return CreateEvent(sessionID, keyEvent)
}

// convertMouseEvent converts a tcell mouse event to protobuf
func (s *Server) convertMouseEvent(sessionID string, event *tcell.EventMouse) *pb.Event {
	x, y := event.Position()
	buttons := event.Buttons()

	action := pb.MouseAction_MOUSE_MOVE
	if buttons&tcell.Button1 != 0 {
		action = pb.MouseAction_MOUSE_CLICK
	} else if buttons&tcell.Button2 != 0 {
		action = pb.MouseAction_MOUSE_MIDDLE_CLICK
	} else if buttons&tcell.Button3 != 0 {
		action = pb.MouseAction_MOUSE_RIGHT_CLICK
	} else if buttons&tcell.WheelUp != 0 {
		action = pb.MouseAction_MOUSE_SCROLL_UP
	} else if buttons&tcell.WheelDown != 0 {
		action = pb.MouseAction_MOUSE_SCROLL_DOWN
	}

	mouseEvent := &pb.MouseEvent{
		Action: action,
		Position: &pb.Position{
			X: int32(x),
			Y: int32(y),
		},
		ScreenPosition: &pb.Position{
			X: int32(x),
			Y: int32(y),
		},
		Modifiers: convertModifiers(event.Modifiers()),
	}

	return CreateEvent(sessionID, mouseEvent)
}

// convertKey converts tcell.Key to protobuf Key
func convertKey(key tcell.Key) pb.Key {
	switch key {
	case tcell.KeyEnter:
		return pb.Key_KEY_ENTER
	case tcell.KeyTab:
		return pb.Key_KEY_TAB
	case tcell.KeyBacktab:
		return pb.Key_KEY_BACKTAB
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return pb.Key_KEY_BACKSPACE
	case tcell.KeyEscape:
		return pb.Key_KEY_ESCAPE
	case tcell.KeyUp:
		return pb.Key_KEY_UP
	case tcell.KeyDown:
		return pb.Key_KEY_DOWN
	case tcell.KeyLeft:
		return pb.Key_KEY_LEFT
	case tcell.KeyRight:
		return pb.Key_KEY_RIGHT
	case tcell.KeyHome:
		return pb.Key_KEY_HOME
	case tcell.KeyEnd:
		return pb.Key_KEY_END
	case tcell.KeyPgUp:
		return pb.Key_KEY_PGUP
	case tcell.KeyPgDn:
		return pb.Key_KEY_PGDN
	case tcell.KeyDelete:
		return pb.Key_KEY_DELETE
	case tcell.KeyInsert:
		return pb.Key_KEY_INSERT
	case tcell.KeyF1:
		return pb.Key_KEY_F1
	case tcell.KeyF2:
		return pb.Key_KEY_F2
	case tcell.KeyF3:
		return pb.Key_KEY_F3
	case tcell.KeyF4:
		return pb.Key_KEY_F4
	case tcell.KeyF5:
		return pb.Key_KEY_F5
	case tcell.KeyF6:
		return pb.Key_KEY_F6
	case tcell.KeyF7:
		return pb.Key_KEY_F7
	case tcell.KeyF8:
		return pb.Key_KEY_F8
	case tcell.KeyF9:
		return pb.Key_KEY_F9
	case tcell.KeyF10:
		return pb.Key_KEY_F10
	case tcell.KeyF11:
		return pb.Key_KEY_F11
	case tcell.KeyF12:
		return pb.Key_KEY_F12
	case tcell.KeyRune:
		return pb.Key_KEY_RUNE
	default:
		return pb.Key_KEY_RUNE // Default to rune for unknown keys
	}
}

// convertModifiers converts tcell modifiers to protobuf modifiers
func convertModifiers(mods tcell.ModMask) []pb.Modifier {
	var result []pb.Modifier

	if mods&tcell.ModShift != 0 {
		result = append(result, pb.Modifier_MOD_SHIFT)
	}
	if mods&tcell.ModCtrl != 0 {
		result = append(result, pb.Modifier_MOD_CTRL)
	}
	if mods&tcell.ModAlt != 0 {
		result = append(result, pb.Modifier_MOD_ALT)
	}
	if mods&tcell.ModMeta != 0 {
		result = append(result, pb.Modifier_MOD_META)
	}
	if mods&tcell.ModHyper != 0 {
		result = append(result, pb.Modifier_MOD_HYPER)
	}

	return result
}

