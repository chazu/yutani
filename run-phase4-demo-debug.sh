#!/bin/bash

# Script to run phase4-demo with full debug logging

echo "=== Yutani Phase 4 Demo - Debug Mode ==="
echo ""
echo "This script will:"
echo "  1. Start the Yutani server with logging to yutani-server.log"
echo "  2. Run the phase4-demo client with logging to phase4-demo.log"
echo ""
echo "Logs will be written to:"
echo "  - yutani-server.log (server events and operations)"
echo "  - phase4-demo.log (client events and operations)"
echo ""

# Clean up old logs
rm -f yutani-server.log phase4-demo.log

# Build the binaries
echo "Building binaries..."
go build -o bin/yutani ./cmd/yutani
go build -o bin/phase4-demo ./cmd/phase4-demo

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "✓ Build successful"
echo ""

# Start the server in the background with logging
echo "Starting Yutani server with logging..."
./bin/yutani server --log-file yutani-server.log --log-level debug &
SERVER_PID=$!

echo "✓ Server started (PID: $SERVER_PID)"
echo "  Logging to: yutani-server.log"
echo ""

# Give server time to start
sleep 2

# Run the demo
echo "Starting phase4-demo..."
echo "  Logging to: phase4-demo.log"
echo ""
echo "Press Ctrl+C to stop"
echo ""

./bin/phase4-demo

# Cleanup
echo ""
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null

echo ""
echo "=== Debug Session Complete ==="
echo ""
echo "Review the logs:"
echo "  tail -f yutani-server.log"
echo "  tail -f phase4-demo.log"
echo ""
echo "Or view all events:"
echo "  grep EVENT phase4-demo.log"
echo "  grep 'Event dispatch' yutani-server.log"
