# Debugging Yutani Applications

This guide explains how to debug Yutani applications, particularly when troubleshooting event handling and UI interactions.

## Quick Start: Debug Mode

The easiest way to debug the phase4-demo is to use the debug script:

```bash
./run-phase4-demo-debug.sh
```

This will:
1. Start the server with logging to `yutani-server.log`
2. Run the phase4-demo with logging to `phase4-demo.log`
3. Capture all events and operations

## Server-Side Logging

### Enable Server Logging

The Yutani server logs to a file when the `YUTANI_LOG_FILE` environment variable is set:

```bash
# Start server with logging
YUTANI_LOG_FILE=yutani-server.log ./bin/yutani-server

# With debug level logging
YUTANI_LOG_FILE=yutani-server.log YUTANI_LOG_LEVEL=debug ./bin/yutani-server
```

### Log Levels

- `debug` - All events, including event dispatch details
- `info` - Important operations (session creation, widget creation, etc.)
- `warn` - Warnings (dropped events, etc.)
- `error` - Errors only

### What Gets Logged

Server logs include:
- Session creation/destruction
- Widget creation/deletion
- Event subscriptions
- Event dispatch (with debug level)
- Service operations (List, Table, Form, etc.)
- Errors and warnings

### Viewing Server Logs

```bash
# Follow logs in real-time
tail -f yutani-server.log

# View all event dispatches
grep "Event dispatch" yutani-server.log

# View warnings (like dropped events)
grep WARN yutani-server.log

# View errors
grep ERROR yutani-server.log
```

## Client-Side Logging

### phase4-demo Logging

The phase4-demo automatically logs to `phase4-demo.log` with detailed event information.

### What Gets Logged

Client logs include:
- Connection status
- Session creation
- Widget creation
- All received events with timestamps
- Event details (key, mouse, widget, etc.)

### Viewing Client Logs

```bash
# Follow logs in real-time
tail -f phase4-demo.log

# View all events
grep EVENT phase4-demo.log

# View only key events
grep "EVENT.*KEY" phase4-demo.log

# View only widget events
grep "EVENT.*WIDGET" phase4-demo.log

# Count events by type
grep EVENT phase4-demo.log | cut -d']' -f3 | cut -d':' -f1 | sort | uniq -c
```

## Common Issues and Solutions

### Issue: No Events Received

**Symptoms:**
- Client logs show "Event stream started" but no events
- Server logs show no "Event dispatch" messages

**Debugging:**
1. Check if event stream is connected:
   ```bash
   grep "Event stream started" phase4-demo.log
   ```

2. Check server-side subscription:
   ```bash
   grep "subscribing to events" yutani-server.log
   ```

3. Verify events are being captured:
   ```bash
   # Try pressing keys and check server logs
   grep "Event dispatch" yutani-server.log
   ```

**Solutions:**
- Ensure server is running before starting client
- Check that Subscribe RPC is being called (not StreamEvents)
- Verify event filter allows all event types

### Issue: Events Filtered Out

**Symptoms:**
- Some events received but not others
- Server logs show "Event dispatch: filtered out"

**Debugging:**
```bash
# Check what's being filtered
grep "filtered out" yutani-server.log
```

**Solutions:**
- Check event filter in Subscribe request
- Ensure all event types are enabled:
  ```go
  Filter: &pb.EventFilter{
      ReceiveKeyEvents:    true,
      ReceiveMouseEvents:  true,
      ReceiveResizeEvents: true,
      ReceiveFocusEvents:  true,
      ReceiveWidgetEvents: true,
  }
  ```

### Issue: Events Dropped

**Symptoms:**
- Server logs show "Event dropped: channel full"
- Some events missing from client

**Debugging:**
```bash
# Count dropped events
grep "Event dropped" yutani-server.log | wc -l
```

**Solutions:**
- Process events faster in client
- Reduce event generation rate
- Increase buffer size in EventDispatcher (default: 100)

### Issue: Focus Not Working

**Symptoms:**
- Cannot change focus between widgets
- No focus events received

**Debugging:**
1. Check if focus events are enabled:
   ```bash
   grep "FOCUS" phase4-demo.log
   ```

2. Check if SetFocus is being called:
   ```bash
   grep "SetFocus" yutani-server.log
   ```

**Solutions:**
- Ensure widgets are focusable
- Call SetFocus explicitly
- Check widget hierarchy (parent-child relationships)

## Event Flow Diagram

```
User Input (keyboard/mouse)
    ↓
tview captures event
    ↓
Server converts to protobuf Event
    ↓
EventDispatcher.Dispatch()
    ↓
Check filter
    ↓
Send to subscriber channel
    ↓
EventService.Subscribe stream
    ↓
Client receives event
    ↓
Client event handlers called
```

## Debugging Checklist

When events aren't working:

- [ ] Server is running with logging enabled
- [ ] Client successfully connects and creates session
- [ ] Event stream is started (Subscribe called)
- [ ] Event filter allows all event types
- [ ] Server logs show "Event dispatch" messages
- [ ] Client logs show "EVENT" messages
- [ ] No "Event dropped" warnings in server logs
- [ ] No "filtered out" messages for expected events

## Advanced Debugging

### Enable Debug Logging in Code

Add temporary debug logging:

```go
// In client code
log.Printf("DEBUG: Received event type: %T", event.Event)

// In server code
slog.Debug("Custom debug message", "key", value)
```

### Monitor Event Rate

```bash
# Count events per second
watch -n 1 'grep EVENT phase4-demo.log | tail -100 | wc -l'
```

### Analyze Event Timing

```bash
# Show event timestamps
grep EVENT phase4-demo.log | cut -d'@' -f2 | cut -d']' -f1
```

## Getting Help

If you're still having issues:

1. Collect logs:
   ```bash
   tar czf yutani-debug.tar.gz yutani-server.log phase4-demo.log
   ```

2. Include:
   - What you're trying to do
   - What's happening vs. what you expect
   - Relevant log excerpts
   - Steps to reproduce

3. Check existing issues or create a new one with the debug information

