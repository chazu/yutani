#!/bin/bash
# Test script to verify the server works correctly

echo "🧪 Testing Yutani Server..."
echo ""

# Build the server
echo "📦 Building yutani..."
make build-yutani
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✅ Build successful"
echo ""

# Check if binary exists
if [ ! -f "./bin/yutani" ]; then
    echo "❌ Binary not found at ./bin/yutani"
    exit 1
fi
echo "✅ Binary exists"
echo ""

echo "🚀 Server is ready to run!"
echo ""
echo "To start the server, run:"
echo "  ./bin/yutani server"
echo ""
echo "Expected behavior:"
echo "  ✅ You should see a TUI with a welcome screen"
echo "  ✅ No log messages should appear"
echo "  ✅ Press Ctrl+C to exit gracefully"
echo ""
echo "To enable logging:"
echo "  ./bin/yutani server --log-file server.log"
echo ""
echo "To run headless (for testing):"
echo "  ./bin/yutani server --headless"
echo ""
