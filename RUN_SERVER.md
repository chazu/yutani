# How to Run the Yutani Server

## ✅ The Fix is Complete!

The server has been fixed to properly display the TUI and respond to Ctrl+C.

## 🚀 Running the Server

### Step 1: Build
```bash
make build-server
```

### Step 2: Run
```bash
./bin/yutani-server
```

## 📺 What You Should See

When you run the server, you should immediately see a TUI display like this:

```
┌─────────────────────────────────────────────────────────┐
│                  Yutani v0.1.0-alpha                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│                                                         │
│              Yutani Display Server                      │
│                                                         │
│          Waiting for client connections...              │
│                                                         │
│        Server is running on localhost:7755              │
│              Press Ctrl+C to exit                       │
│                                                         │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

The text should be:
- **Yellow and bold** for "Yutani Display Server"
- **Red and bold** for "Ctrl+C"
- **Centered** in the window
- **Bordered** with a title showing the version

## ⌨️ Controls

- **Ctrl+C** - Exit the server gracefully (this now works!)
- **Mouse** - Enabled (when widgets are displayed)

## 🐛 Troubleshooting

### Problem: Nothing appears, terminal is blank

**Possible causes:**
1. The terminal doesn't support the required features
2. The TERM environment variable is not set correctly
3. Running in a non-interactive environment

**Solutions:**
```bash
# Check your TERM variable
echo $TERM

# Try setting it explicitly
export TERM=xterm-256color
./bin/yutani-server

# Or try screen-256color
export TERM=screen-256color
./bin/yutani-server
```

### Problem: Ctrl+C doesn't work

**This should be fixed now!** If it still doesn't work:

1. Make sure you rebuilt: `make clean && make build-server`
2. Try pressing Ctrl+C multiple times
3. Check if you're running the right binary: `which yutani-server` or `./bin/yutani-server`

### Problem: Want to see logs

```bash
# Enable logging to a file
YUTANI_LOG_FILE=server.log ./bin/yutani-server

# In another terminal, watch the logs
tail -f server.log
```

## 🧪 Testing with the Demo Client

Once the server is running:

### Terminal 1 (Server)
```bash
./bin/yutani-server
```

### Terminal 2 (Demo Client)
```bash
make demo
# or
./bin/phase4-demo
```

You should see widgets appear in the server terminal!

## 📋 Technical Details

### What Changed

1. **TUI now runs in main goroutine** - Required for proper display
2. **tview manages screen initialization** - No manual screen.Init()
3. **Ctrl+C handled in input capture** - Calls app.Stop()
4. **Logging disabled by default** - No interference with TUI

### Architecture

```
main.go
  ├─> yutaniServer.Start()  // Initialize TUI (non-blocking)
  ├─> Start gRPC server     // In background goroutine
  └─> yutaniServer.Run()    // Run TUI (BLOCKS in main goroutine)
       └─> app.Run()        // tview event loop
            └─> Ctrl+C → app.Stop() → returns → graceful shutdown
```

## 🎯 Expected Behavior

1. **Server starts** → Welcome screen appears immediately
2. **Client connects** → Widgets replace welcome screen
3. **Client disconnects** → Welcome screen returns (future feature)
4. **Ctrl+C pressed** → Server exits gracefully

## ❓ Still Having Issues?

If the server still doesn't display properly or Ctrl+C doesn't work:

1. **Check your terminal emulator** - Try iTerm2, Terminal.app, or another emulator
2. **Check terminal size** - Make sure your terminal is at least 80x24
3. **Enable logging** - Run with `YUTANI_LOG_FILE=debug.log ./bin/yutani-server`
4. **Share the logs** - Check debug.log for error messages

The most common issue is running in a non-interactive environment (like in the background with `&`).
The server MUST run in an interactive terminal to display the TUI.

