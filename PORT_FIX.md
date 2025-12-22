# Port Configuration Fix

**Date**: December 22, 2024  
**Issue**: Examples were connecting to wrong port  
**Status**: ✅ **FIXED**

## Problem

All example applications were trying to connect to `localhost:50051`, but the Yutani server actually listens on `localhost:7755` by default.

This caused connection failures when running examples:
```
Failed to connect: connection refused
```

## Root Cause

The server configuration defaults to port **7755**:
- File: `pkg/config/config.go`
- Default: `Address: ":7755"`

But all examples were hardcoded to connect to port **50051** (a common gRPC default).

## Solution

Updated all example applications to connect to the correct port: **7755**

### Files Changed

**Examples Updated** (8 files):
1. `examples/simple-list/main.go` - Changed to `localhost:7755`
2. `examples/data-table/main.go` - Changed to `localhost:7755`
3. `examples/login-form/main.go` - Changed to `localhost:7755`
4. `examples/file-browser/main.go` - Changed to `localhost:7755`
5. `examples/dashboard/main.go` - Changed to `localhost:7755`
6. `examples/process-monitor/main.go` - Changed to `localhost:7755`
7. `examples/chat-app/main.go` - Changed to `localhost:7755`
8. `examples/text-editor/main.go` - Changed to `localhost:7755`

**Documentation Updated** (2 files):
1. `examples/README.md` - Updated connection examples and troubleshooting
2. `pkg/client/README.md` - Updated connection example

## Verification

All examples now compile with the correct port:

```bash
$ make build-examples
# ✅ Success

$ strings bin/examples/text-editor | grep localhost
localhost:7755
# ✅ Correct port
```

## Usage

Now examples work correctly:

```bash
# Terminal 1: Start server (listens on :7755)
make run

# Terminal 2: Run example (connects to :7755)
make run-example EXAMPLE=text-editor
# ✅ Connects successfully
```

## Server Port Configuration

The server port can be configured via environment variable:

```bash
# Default (7755)
go run cmd/yutani-server/main.go

# Custom port
YUTANI_ADDRESS=:8080 go run cmd/yutani-server/main.go
```

If you change the server port, update the client connection accordingly:

```go
// Connect to custom port
c, err := client.Connect("localhost:8080")
```

## Testing

Tested all examples:

```bash
# Terminal 1
make run

# Terminal 2
make run-example EXAMPLE=simple-list      # ✅ Works
make run-example EXAMPLE=data-table       # ✅ Works
make run-example EXAMPLE=login-form       # ✅ Works
make run-example EXAMPLE=file-browser     # ✅ Works
make run-example EXAMPLE=dashboard        # ✅ Works
make run-example EXAMPLE=process-monitor  # ✅ Works
make run-example EXAMPLE=chat-app         # ✅ Works
make run-example EXAMPLE=text-editor      # ✅ Works
```

## Summary

✅ All examples now use the correct port (7755)  
✅ Documentation updated  
✅ All examples compile successfully  
✅ Connection issues resolved  

The examples will now connect to the server without any port mismatch errors.

