# Event Handling Fixes

## Issues Identified

### 1. Only Every Other Keypress Received
**Root Cause:** The event polling loop (`pollScreenEvents`) was calling `screen.PollEvent()` which consumes events from the tcell event queue. This interfered with tview's own event handling, causing events to be split between the polling loop and tview's internal processing.

### 2. No Mouse Events
**Root Cause:** 
- Mouse support was not explicitly enabled in tview
- Mouse events were being polled via `screen.PollEvent()` but this doesn't work reliably with tview
- tview has its own mouse event handling that needs to be hooked into

## Solutions Implemented

### 1. Enabled Mouse Support in tview

**File:** `pkg/server/server.go`

Added explicit mouse enablement:
```go
if s.mouseEnable {
    s.app.EnableMouse(true)
    slog.Info("Mouse support enabled")
}
```

### 2. Added Mouse Capture Handler

Instead of polling for mouse events, we now use tview's `SetMouseCapture` callback:

```go
if s.mouseEnable {
    s.app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
        // Dispatch mouse event
        s.handleMouseEvent(event)
        // Return event to allow tview to process it too
        return event, action
    })
}
```

**Benefits:**
- Events are captured before tview processes them
- No interference with tview's event queue
- All mouse events are reliably captured

### 3. Fixed Event Polling Loop

**Changed:** `pollScreenEvents` from polling all events to only checking for resize

**Before:**
```go
// Poll for events (non-blocking)
ev := s.screen.PollEvent()
if ev == nil {
    continue
}

// Handle mouse events
if mouseEv, ok := ev.(*tcell.EventMouse); ok {
    s.handleMouseEvent(mouseEv)
}
```

**After:**
```go
// Use ticker instead of polling
ticker := time.NewTicker(100 * time.Millisecond)
defer ticker.Stop()

for {
    select {
    case <-s.stopCh:
        return
    case <-ticker.C:
        // Only check for resize
        width, height := s.screen.Size()
        if width != lastWidth || height != lastHeight {
            // Dispatch resize event
        }
    }
}
```

**Benefits:**
- No longer consumes events from the queue
- Resize detection still works
- No interference with key/mouse events

### 4. Preserved Key Event Handling

Key events continue to work via `SetInputCapture`:

```go
s.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    // Handle Ctrl+C
    if event.Key() == tcell.KeyCtrlC {
        s.app.Stop()
        return nil
    }
    // Dispatch to clients
    s.handleInputEvent(event)
    // Return event to allow tview to process it too
    return event
})
```

## Event Flow (After Fixes)

### Keyboard Events
```
User presses key
    ↓
tcell captures event
    ↓
tview's event loop
    ↓
SetInputCapture callback
    ↓
handleInputEvent() → Dispatch to clients
    ↓
Event returned to tview for widget processing
```

### Mouse Events
```
User clicks/moves mouse
    ↓
tcell captures event
    ↓
tview's event loop
    ↓
SetMouseCapture callback
    ↓
handleMouseEvent() → Dispatch to clients
    ↓
Event returned to tview for widget processing
```

### Resize Events
```
Terminal resized
    ↓
pollScreenEvents ticker fires
    ↓
Check screen.Size()
    ↓
If changed → Dispatch to clients
```

## Testing

### Test Keyboard Events
1. Start server and demo
2. Press keys - should see every keypress in logs
3. Check logs: `grep "EVENT.*KEY" phase4-demo.log`

### Test Mouse Events
1. Start server and demo
2. Move mouse over terminal
3. Click in terminal
4. Check logs: `grep "EVENT.*MOUSE" phase4-demo.log`

### Test Resize Events
1. Start server and demo
2. Resize terminal window
3. Check logs: `grep "EVENT.*RESIZE" phase4-demo.log`

## Files Modified

1. `pkg/server/server.go`
   - Added `EnableMouse(true)` call
   - Added `SetMouseCapture` handler
   - Modified `pollScreenEvents` to only check resize
   - Added imports: `log/slog`, `time`

## Expected Behavior

**Before Fixes:**
- ❌ Only ~50% of keypresses received
- ❌ No mouse events at all
- ✅ Resize events worked

**After Fixes:**
- ✅ 100% of keypresses received
- ✅ All mouse events received (move, click, scroll)
- ✅ Resize events still work

## Verification Commands

```bash
# Run with debug logging
./run-phase4-demo-debug.sh

# In another terminal, check event counts
watch -n 1 'echo "Keys: $(grep -c "EVENT.*KEY" phase4-demo.log) | Mouse: $(grep -c "EVENT.*MOUSE" phase4-demo.log) | Resize: $(grep -c "EVENT.*RESIZE" phase4-demo.log)"'

# View events in real-time
tail -f phase4-demo.log | grep EVENT
```

## Next Steps

Now that events are being captured correctly, you can:

1. **Test widget interaction** - Try navigating with Tab, arrow keys
2. **Test mouse clicks** - Click on different widgets
3. **Implement event handlers** - Add logic to respond to events
4. **Test focus changes** - Verify focus events are generated

## Known Limitations

- Mouse events are only captured when mouse support is enabled in config
- Resize events are polled every 100ms (not instant but fast enough)
- Some special key combinations might be intercepted by the terminal

