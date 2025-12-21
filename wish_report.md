# TWIN Protocol Analysis Report

## Executive Summary

TWIN (Text WINdows) is a terminal-based display server that provides a networked, widget-based windowing system for text-mode applications. It implements a custom binary protocol over TCP/IP (port 7754) or Unix domain sockets, offering capabilities similar to X11 but optimized for character-cell displays.

---

## 1. Protocol Overview

### Transport Layer

- **TCP/IP**: Port 7754 (`TW_INET_PORT`)
- **Unix Domain Sockets**: For local connections (AF_UNIX)
- **Optional Compression**: gzip via zlib for bandwidth optimization
- **Blocking & Non-blocking**: Both modes supported

### Wire Format

TWIN uses a **custom binary protocol** with type-safe field encoding:

```c
/* Core data types */
typedef unsigned char byte;      /* 1 byte  */
typedef unsigned short udat;     /* 2 bytes */
typedef unsigned int uldat;      /* 4 bytes */
typedef signed short dat;        /* 2 bytes signed */
typedef signed int ldat;         /* 4 bytes signed */
typedef uint32_t trune;          /* 32-bit Unicode character */
typedef uint64_t tcolor;         /* 64-bit: 32-bit FG + 32-bit BG RGB */
typedef uint64_t tcell;          /* 64-bit: color + character combined */
```

### Message Structure

```c
struct s_tmsg {
    uldat Len;    /* Total message length including this field */
    uldat Magic;  /* msg_magic (0xA3A61CE4) for validation */
    uldat Type;   /* Message type identifier */
    union s_tevent_any Event;  /* Actual event data */
};
```

### Function Call Encoding

```
[Request Length][Serial#][Function ID][Args...]
     4 bytes     4 bytes    4 bytes    Variable
```

### Connection Handshake

1. **ProtocolNumbers()** - Exchange protocol version (current: ~4.8+)
2. **MagicNumbers()** - Validate type sizes between client/server
3. **MagicChallenge()** - Security challenge/response

### Magic Constants

| Constant | Value | Meaning |
|----------|-------|---------|
| `TWIN_MAGIC` | 0x6E697754 | "Twin" identifier |
| `TW_GO_MAGIC` | 0x21216F47 | "Go!!" - proceed signal |
| `TW_WAIT_MAGIC` | 0x74696157 | "Wait" - hold signal |
| `TW_STOP_MAGIC` | 0x706F7453 | "Stop" - halt signal |

---

## 2. Capabilities

### Window Management

- Multi-window support with window hierarchy
- Virtual screens up to 64K x 64K character cells
- Window stacking, focus management, and z-ordering
- Configurable borders, scrollbars, and title bars
- Min/max size constraints
- Window mapping/unmapping (show/hide)

### Drawing Primitives

| Function | Purpose |
|----------|---------|
| `Tw_WriteCharsetWindow()` | Write ASCII/charset bytes |
| `Tw_WriteUtf8Window()` | Write UTF-8 encoded text |
| `Tw_WriteTRuneWindow()` | Write Unicode runes (32-bit) |
| `Tw_WriteTCellWindow()` | Write full cells (color + char) |
| `Tw_GotoXYWindow()` | Move cursor position |
| `Tw_SetColTextWindow()` | Set text foreground/background |
| `Tw_SetColorsWindow()` | Set all window color attributes |

### Color Support

- 21-bit RGB for foreground and background colors
- Color macro: `TCOL(foreground, background)`
- Separate colors for: gadgets, arrows, bars, tabs, borders, text, selections, disabled state

### Terminal Emulation

- Built-in VT100-compatible terminal emulator
- Scrollback buffer support
- Character set translation tables
- Three cursor types: `TW_NOCURSOR`, `TW_LINECURSOR`, `TW_SOLIDCURSOR`

### Display Features

- Hot-pluggable display drivers (`Tw_AttachHW()` / `Tw_DetachHW()`)
- Multi-headed display support
- Background image support for screens
- Automatic endianness conversion for heterogeneous networks

---

## 3. Widget Hierarchy

### Class Structure

```
Sobj (base object)
├── Swidget (generic widget)
│   ├── Swindow (main window with borders, scrollbars)
│   │   └── Terminal emulator support
│   ├── Sscreen (virtual screen container)
│   └── Sgadget (buttons, checkboxes, toggles)
├── Tmenu (menu bar)
│   └── Trow/Tmenuitem (menu entries)
├── Tmsgport (client message port)
├── Tmsg (message container)
├── Tmutex (inter-client synchronization)
└── Tall (topmost container)
```

### Widget Types

#### Window (`Swindow`)
- Title bar with customizable colors
- Configurable borders and scrollbars
- Associated menu support
- TTY emulation capabilities
- Cursor positioning

#### Gadget (`Sgadget`)
Gadgets are interactive UI elements:

| Flag | Purpose |
|------|---------|
| `GADGETFL_BUTTON` | Standard button |
| `GADGETFL_TOGGLE` | Checkbox/radio (can be grouped) |
| `GADGETFL_DISABLED` | Grayed out state |
| `GADGETFL_PRESSED` | Currently pressed |
| `GADGETFL_TEXT_DEFCOL` | Use default colors |

#### Screen (`Sscreen`)
Virtual screen containers supporting:
- Background patterns or custom expose handlers
- Fill patterns
- Selection states

#### Menu (`Tmenu`)
- Menu bars with dropdown items
- Keyboard shortcuts
- Hierarchical submenus via `Trow`/`Tmenuitem`

### Widget Attributes

```c
WIDGET_WANT_MOUSE_MOTION  = 0x0001  /* Receive motion events */
WIDGET_WANT_KEYS          = 0x0002  /* Receive keyboard events */
WIDGET_WANT_MOUSE         = 0x0004  /* Receive mouse events */
WIDGET_WANT_CHANGES       = 0x0008  /* Receive resize/expose events */
WIDGET_AUTO_FOCUS         = 0x0010  /* Auto-focus on mouse over */
```

---

## 4. Event System

### Event Types

| Type | Code | Description |
|------|------|-------------|
| `TW_MSG_WIDGET_KEY` | 0x1000 | Keyboard input |
| `TW_MSG_WIDGET_MOUSE` | 0x1001 | Mouse event |
| `TW_MSG_WIDGET_CHANGE` | 0x1002 | Widget resize/expose |
| `TW_MSG_WIDGET_GADGET` | 0x1003 | Button/gadget activation |
| `TW_MSG_MENU_ROW` | 0x1004 | Menu selection |
| `TW_MSG_SELECTION` | 0x1005 | Paste request |
| `TW_MSG_SELECTIONNOTIFY` | 0x1006 | Clipboard data |
| `TW_MSG_SELECTIONREQUEST` | 0x1007 | Clipboard request |
| `TW_MSG_SELECTIONCLEAR` | 0x1008 | Clipboard cleared |
| `TW_MSG_USER_CONTROL` | 0x2000 | Inter-client control |
| `TW_MSG_USER_CLIENTMSG` | 0x2100 | Arbitrary inter-client |

### Keyboard Event Structure

```c
struct tevent_keyboard {
    twidget Widget;     /* Target widget */
    udat Code;          /* Key code */
    udat ShiftFlags;    /* SHIFT_FL, CTRL_FL, ALT_FL, etc. */
    byte SeqLen;        /* ASCII sequence length */
    byte AsciiSeq[];    /* Null-terminated sequence */
};
```

### Mouse Event Structure

```c
struct tevent_mouse {
    twidget Widget;     /* Target widget */
    udat Code;          /* PRESS_LEFT, RELEASE_RIGHT, DRAG_MOUSE, etc. */
    udat ShiftFlags;    /* Modifier keys */
    dat X, Y;           /* Coordinates relative to widget */
};
```

### Widget Change Event

```c
struct tevent_widget {
    twidget Widget;     /* Affected widget */
    udat Code;          /* TW_MSG_WIDGET_RESIZE or TW_MSG_WIDGET_EXPOSE */
    udat Flags;         /* e.g., TW_MSG_WIDGETFL_SHADED */
    dat XWidth, YWidth; /* New dimensions */
    dat X, Y;           /* New position */
};
```

### Selection/Clipboard

X11-style clipboard protocol with MIME type support:
- `TW_SEL_UTF8MAGIC` - UTF-8 text
- `TW_SEL_FILEMAGIC` - File paths
- `TW_SEL_URLMAGIC` - URLs

---

## 5. Client-Server Interaction

### Connection Lifecycle

```c
/* 1. Open connection */
tdisplay TwD = Tw_Open(NULL);  /* Uses $TWDISPLAY or localhost:7754 */

/* 2. Create message port (required for events) */
tmsgport port = Tw_CreateMsgPort(5, "myapp");

/* 3. Create UI elements */
tmenu menu = Tw_CreateMenu(...);
twindow win = Tw_CreateWindow(...);

/* 4. Display window */
Tw_MapWindow(win, screen);

/* 5. Send commands to server */
Tw_Flush();

/* 6. Event loop */
while ((msg = Tw_ReadMsg(TRUE)) != NULL) {
    /* Process events */
}

/* 7. Cleanup */
Tw_Close();
```

### Buffer Queues

| Queue | Purpose |
|-------|---------|
| `Queue[QREAD]` | Incoming data buffer |
| `Queue[QWRITE]` | Outgoing data (coalesced writes) |
| `Queue[QMSG]` | Received messages |
| `Queue[QgzREAD]` | Decompression buffer |
| `Queue[QgzWRITE]` | Compression output |

### Synchronization Functions

| Function | Behavior |
|----------|----------|
| `Tw_Flush()` | Write pending data, block until complete |
| `Tw_Sync()` | Flush and wait for all server replies |
| `Tw_TimidFlush()` | Non-blocking flush attempt |
| `TwInPanic()` | Check for fatal connection error |

### Request/Response Pattern

- Client assigns serial number to each request
- Server returns response with matching serial number
- Supports pipelined requests (multiple outstanding)
- Library matches replies automatically

### Message Passing

```c
/* Send to another client's message port */
tmsg msg = Tw_CreateMsg(TwD, type, len);
Tw_SendMsg(TwD, target_msgport, msg);      /* Blocking */
Tw_BlindSendMsg(TwD, target_msgport, msg); /* Non-blocking */

/* Receive messages */
tmsg msg = Tw_ReadMsg(TRUE);   /* Blocks */
tmsg msg = Tw_PeekMsg();       /* Non-blocking check */
```

---

## 6. API Function Reference

### Connection Management
- `Tw_Open()` / `Tw_Close()` - Connect/disconnect
- `Tw_FindFunction()` - Query server capabilities
- `Tw_ServerSizeof()` - Get server's type sizes

### Object Lifecycle
- `Tw_CreateWidget()` / `Tw_DeleteObj()` - Create/destroy
- `Tw_RecursiveDeleteWidget()` - Delete with children
- `Tw_ChangeField()` - Modify object properties

### Navigation
- `Tw_PrevObj()` / `Tw_NextObj()` - Sibling traversal
- `Tw_ParentObj()` - Get parent
- `Tw_FirstScreen()` / `Tw_FirstWidget()` / `Tw_FirstMsgPort()` - Get first child

### Window Operations
- `Tw_CreateWindow()` / `Tw_Create4MenuWindow()`
- `Tw_MapWidget()` / `Tw_UnMapWidget()`
- `Tw_SetXYWidget()` / `Tw_ResizeWidget()`
- `Tw_ConfigureWindow()` - Set constraints

### Drawing
- `Tw_WriteCharsetWindow()` / `Tw_WriteUtf8Window()` / `Tw_WriteTCellWindow()`
- `Tw_GotoXYWindow()` - Position cursor
- `Tw_SetColorsWindow()` - Set color palette
- `Tw_DrawWidget()` - Generic draw with data
- `Tw_ScrollWidget()` - Scroll content

### Gadgets
- `Tw_CreateGadget()` / `Tw_CreateButtonGadget()`
- `Tw_WriteTextsGadget()` / `Tw_WriteTRunesGadget()`

### Menus
- `Tw_CreateMenu()`
- `Tw_Create4MenuAny()` / `Tw_Create4MenuCommonMenuItem()`
- `Tw_SetInfoMenu()`

### Message Ports
- `Tw_CreateMsgPort()` / `Tw_DeleteMsgPort()`
- `Tw_FindMsgPort()` - Find by name
- `Tw_SendToMsgPort()` / `Tw_BlindSendToMsgPort()`

### Clipboard
- `Tw_GetOwnerSelection()` / `Tw_SetOwnerSelection()`
- `Tw_RequestSelection()` / `Tw_NotifySelection()`

### Display
- `Tw_GetDisplayWidth()` / `Tw_GetDisplayHeight()`
- `Tw_BgImageScreen()` - Set background
- `Tw_AttachHW()` / `Tw_DetachHW()` - Hot-plug displays

### Compression
- `Tw_EnableGzip()` / `Tw_DisableGzip()` / `Tw_CanCompress()`

---

## 7. Key Source Files

### Client Library
| File | Description |
|------|-------------|
| `libs/libtw/libtw.c` | Main implementation (~2,700 lines) |
| `libs/libtw/encode.h` | Binary encoding/decoding |
| `libs/libtw/avl.c` | AVL tree for listener management |

### Public Headers
| File | Description |
|------|-------------|
| `include/Tw/Tw.h` | Main public API |
| `include/Tw/Tw_defs.h` | Constants and defines |
| `include/Tw/datatypes.h` | Type definitions |
| `include/Tw/proto_gen.h` | Generated protocol definitions |

### Server Objects
| File | Description |
|------|-------------|
| `server/obj/widget.h` | Base widget class |
| `server/obj/window.h` | Window class |
| `server/obj/gadget.h` | Button/gadget class |
| `server/obj/menu.h` | Menu class |
| `server/obj/screen.h` | Screen class |
| `server/obj/msg.h` | Message class |

### Example Clients
| File | Description |
|------|-------------|
| `clients/cat.c` | Simple text display (~118 lines) |
| `clients/dialog.c` | Dialog boxes (~400 lines) |
| `clients/dm.c` | Display manager (~500 lines) |
| `clients/event.c` | Event tester (~300 lines) |
| `clients/term.c` | Terminal emulator (~500 lines) |

### Documentation
| File | Description |
|------|-------------|
| `docs/libtw.txt` | Comprehensive API reference (~577 lines) |
| `docs/diagram.txt` | Architecture diagram |
| `docs/Tutorial` | Usage guide (~400 lines) |

---

## 8. Unique Features

1. **Network Transparency**: Server and clients can run on different machines
2. **Hot-Pluggable Displays**: Attach/detach drivers without restart
3. **Optional Compression**: gzip for low-bandwidth connections
4. **Multi-Headed Support**: Single server, multiple displays
5. **Large Virtual Screens**: Up to 64K x 64K character cells
6. **Built-in Terminal**: VT100-compatible emulation
7. **Widget Composition**: Hierarchical nesting support
8. **X11-Style Clipboard**: MIME-typed selection protocol
9. **Thread Safety**: Optional pthread mutex support
10. **Cross-Platform**: Automatic endianness handling

---

## 9. Example: Minimal Client

```c
#include <Tw/Tw.h>

int main(void) {
    tdisplay TwD;
    tmsgport port;
    twindow win;
    tscreen screen;
    tmsg msg;

    /* Connect to server */
    if (!(TwD = Tw_Open(NULL))) {
        fprintf(stderr, "Cannot connect to twin server\n");
        return 1;
    }

    /* Create message port */
    port = Tw_CreateMsgPort(TwD, 6, "myapp");

    /* Get first screen */
    screen = Tw_FirstScreen(TwD);

    /* Create window */
    win = Tw_CreateWindow(TwD,
        5, "Hello",              /* title length, title */
        NULL,                    /* hints */
        NULL,                    /* menu */
        TW_NOCURSOR,             /* cursor type */
        TW_WINDOW_WANT_CHANGES,  /* attributes */
        TW_WINDOWFL_BORDERLESS,  /* flags */
        40, 10,                  /* width, height */
        0                        /* scroll back */
    );

    /* Display window */
    Tw_MapWindow(TwD, win, screen);

    /* Write some text */
    Tw_WriteCharsetWindow(TwD, win, 12, "Hello World!");

    /* Send to server */
    Tw_Flush(TwD);

    /* Event loop */
    while ((msg = Tw_ReadMsg(TwD, TRUE))) {
        switch (msg->Type) {
            case TW_MSG_WIDGET_KEY:
                /* Handle keyboard */
                break;
            case TW_MSG_WIDGET_MOUSE:
                /* Handle mouse */
                break;
            case TW_MSG_WIDGET_GADGET:
                /* Handle button press */
                break;
        }
    }

    Tw_Close(TwD);
    return 0;
}
```

---

## 10. Conclusion

TWIN provides a complete terminal-based windowing system with a well-designed binary protocol that prioritizes efficiency and network transparency. Its widget hierarchy supports building complex applications with windows, buttons, menus, and custom widgets. The event-driven architecture follows familiar patterns from X11 while remaining lightweight enough for text-mode operation.

The protocol's binary format with type-safe encoding, optional compression, and automatic endianness handling makes it suitable for both local and remote operation. The extensive API covers window management, drawing, event handling, and inter-client communication.

---

*Report generated from analysis of the TWIN codebase at ~/opp/c/wish (twin project)*
