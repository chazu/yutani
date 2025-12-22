#!/bin/bash
# Test script to verify the server works correctly

echo "🧪 Testing Yutani Server..."
echo ""

# Build the server
echo "📦 Building server..."
make build-server
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✅ Build successful"
echo ""

# Check if binary exists
if [ ! -f "./bin/yutani-server" ]; then
    echo "❌ Binary not found at ./bin/yutani-server"
    exit 1
fi
echo "✅ Binary exists"
echo ""

echo "🚀 Server is ready to run!"
echo ""
echo "To start the server, run:"
echo "  ./bin/yutani-server"
echo ""
echo "Expected behavior:"
echo "  ✅ You should see a TUI with a welcome screen"
echo "  ✅ No log messages should appear"
echo "  ✅ Press Ctrl+C to exit gracefully"
echo ""
echo "To enable logging:"
echo "  YUTANI_LOG_FILE=server.log ./bin/yutani-server"
echo ""

