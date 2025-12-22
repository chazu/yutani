# Troubleshooting Summary: Phase 4 Demo Event Handling

## Issue

The phase4-demo was not responding to user input - unable to change focus or interact with the UI.

## Root Causes Identified

### 1. Incorrect RPC Method Name

**Problem:** The phase4-demo was calling `StreamEvents` but the proto defines `Subscribe`.

**Location:** `cmd/phase4-demo/main.go` and `pkg/client/client.go`

**Fix:** Changed from:
```go
eventClient.StreamEvents(&pb.StreamEventsRequest{...})
```

To:
```go
eventClient.Subscribe(&pb.SubscribeRequest{...})
```

### 2. Missing Event Filter

**Problem:** No event filter was specified, which might cause events to be filtered out.

**Fix:** Added explicit event filter:
```go
Filter: &pb.EventFilter{
    ReceiveKeyEvents:    true,
    ReceiveMouseEvents:  true,
    ReceiveResizeEvents: true,
    ReceiveFocusEvents:  true,
    ReceiveWidgetEvents: true,
}
```

### 3. No Event Logging

**Problem:** No visibility into whether events were being received or dispatched.

**Fix:** Added comprehensive logging on both client and server sides.

## Changes Made

### 1. Enhanced Client Logging (`cmd/phase4-demo/main.go`)

**Added:**
- Logging to file (`phase4-demo.log`)
- Dual output (stdout + file)
- Detailed event logging with timestamps
- Event counter
- Event type identification

**New Functions:**
- `logEvent()` - Logs detailed event information
- `formatWidgetEvent()` - Formats widget events

### 2. Enhanced Server Logging (`pkg/server/events.go`)

**Added:**
- Debug logging in `Dispatch()` method
- Logging for:
  - Events with no session ID
  - Events with no subscriber
  - Filtered events
  - Successfully dispatched events
  - Dropped events (channel full)
- `getEventTypeName()` helper function

### 3. Fixed Client Library (`pkg/client/client.go`)

**Changed:**
- Updated `StartEventStream()` to use `Subscribe` RPC
- Changed stream type from `EventService_StreamEventsClient` to `EventService_SubscribeClient`
- Added explicit event filter

### 4. Created Debug Tools

**New Files:**
- `run-phase4-demo-debug.sh` - Script to run demo with full logging
- `DEBUGGING.md` - Comprehensive debugging guide
- `TROUBLESHOOTING_SUMMARY.md` - This file

## How to Use

### Run with Debug Logging

```bash
./run-phase4-demo-debug.sh
```

This will:
1. Build the binaries
2. Start server with logging to `yutani-server.log`
3. Run phase4-demo with logging to `phase4-demo.log`
4. Clean up on exit

### View Logs

```bash
# Follow client events
tail -f phase4-demo.log

# Follow server events
tail -f yutani-server.log

# View all events
grep EVENT phase4-demo.log

# View event dispatches
grep "Event dispatch" yutani-server.log
```

### Manual Server Start with Logging

```bash
# Start server with debug logging
YUTANI_LOG_FILE=yutani-server.log YUTANI_LOG_LEVEL=debug ./bin/yutani-server

# In another terminal, run the demo
./bin/phase4-demo
```

## Expected Behavior After Fixes

### On Startup

**Client Log:**
```
=== Phase 4 Demo Starting ===
Logging to: phase4-demo.log
✓ Session created: <session-id>

=== Starting Event Stream ===
✓ Event stream started
```

**Server Log:**
```
Client subscribing to events session_id=<session-id>
```

### On User Input

**Client Log (example key press):**
```
[EVENT #1 @ 15:04:05.123] KEY: key=Rune rune=a (97) mod=0
```

**Server Log (debug level):**
```
Event dispatched session_id=<session-id> event_type=KEY
```

### On Widget Interaction

**Client Log:**
```
[EVENT #5 @ 15:04:10.456] WIDGET: id=<widget-id> type=SELECTED data=map[]
```

## Verification Steps

1. **Start the debug session:**
   ```bash
   ./run-phase4-demo-debug.sh
   ```

2. **Press some keys** and verify events appear in logs:
   ```bash
   # In another terminal
   tail -f phase4-demo.log | grep KEY
   ```

3. **Check server is dispatching:**
   ```bash
   tail -f yutani-server.log | grep "Event dispatch"
   ```

4. **Try interacting with widgets** (Tab, Enter, arrow keys)

5. **Review event counts:**
   ```bash
   grep EVENT phase4-demo.log | wc -l
   ```

## Next Steps

If events are still not working after these fixes:

1. **Check event capture** - Verify tview is capturing events
2. **Check focus** - Ensure widgets are focusable
3. **Check widget hierarchy** - Verify parent-child relationships
4. **Review logs** - Look for "filtered out" or "dropped" messages

See `DEBUGGING.md` for detailed troubleshooting steps.

## Files Modified

1. `cmd/phase4-demo/main.go` - Added logging and fixed RPC call
2. `pkg/client/client.go` - Fixed RPC call and stream type
3. `pkg/server/events.go` - Added debug logging

## Files Created

1. `run-phase4-demo-debug.sh` - Debug script
2. `DEBUGGING.md` - Debugging guide
3. `TROUBLESHOOTING_SUMMARY.md` - This summary

