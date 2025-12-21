# End-to-End Tests

This directory contains automated end-to-end tests for the Yutani Terminal Display Server.

## Overview

The E2E tests verify the complete system by:
1. Starting a real Yutani server instance
2. Creating gRPC client connections
3. Testing full request/response flows
4. Verifying state changes across services

## Test Architecture

### In-Memory gRPC Testing

The tests use **bufconn** (buffered connection) for in-memory gRPC communication:

**Benefits:**
- ✅ No network ports required
- ✅ Fast execution (no TCP overhead)
- ✅ Fully isolated (no port conflicts)
- ✅ Works in CI/CD environments
- ✅ Deterministic behavior

**How it works:**
```
Test Process
├── Yutani Server (goroutine)
│   ├── tview.Application
│   ├── tcell.Screen
│   └── Service implementations
├── gRPC Server (goroutine)
│   └── bufconn.Listener (in-memory)
└── Test Client
    └── gRPC connection via bufconn
```

### Test Coverage

#### TestE2E_SessionLifecycle
- Ping
- GetServerInfo
- CreateSession
- DestroySession
- Session count verification

#### TestE2E_ScreenOperations
- GetSize
- Clear
- SetCell / GetCell
- DrawText
- DrawBox
- Fill
- Sync

#### TestE2E_EventStreaming
- Subscribe to events
- InjectEvent
- Receive events via streaming
- Event filtering

#### TestE2E_WidgetOperations
- CreateWidget (Box, TextView)
- ListWidgets
- SetRoot
- GetProperties
- SetProperties
- DeleteWidget

## Running the Tests

```bash
# Run all E2E tests
go test ./test/e2e/... -v

# Run specific test
go test ./test/e2e/... -v -run TestE2E_SessionLifecycle

# Run with timeout
go test ./test/e2e/... -v -timeout 30s

# Run with race detector
go test ./test/e2e/... -v -race

# Run with coverage
go test ./test/e2e/... -v -cover -coverprofile=coverage.out
```

## Test Patterns

### Setup Pattern
```go
ts := newTestServer(t)
defer ts.stop()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

conn, err := ts.dial(ctx)
if err != nil {
    t.Fatalf("Failed to dial: %v", err)
}
defer conn.Close()
```

### Assertion Pattern
```go
resp, err := client.SomeRPC(ctx, req)
if err != nil {
    t.Fatalf("RPC failed: %v", err)
}
if !resp.Success {
    t.Error("Expected successful response")
}
```

## Limitations

### What E2E Tests DON'T Cover

1. **Visual Rendering** - Can't verify actual terminal output
2. **Real Terminal Events** - Uses injected events, not real keyboard/mouse
3. **Multiple Concurrent Clients** - Single client per test
4. **Network Issues** - In-memory connection doesn't test network failures
5. **Performance Under Load** - Not designed for load testing

### Complementary Testing

- **Unit Tests** (`pkg/server/*_test.go`, `pkg/services/*_test.go`) - Test business logic
- **Integration Tests** (test client) - Manual verification of visual output
- **Load Tests** (future) - Test performance with many clients

## Adding New Tests

1. Create test function with `TestE2E_` prefix
2. Use `newTestServer(t)` to create server
3. Create gRPC clients via `ts.dial(ctx)`
4. Make RPC calls and verify responses
5. Clean up resources (defer cleanup)

Example:
```go
func TestE2E_MyNewFeature(t *testing.T) {
    ts := newTestServer(t)
    defer ts.stop()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    conn, err := ts.dial(ctx)
    if err != nil {
        t.Fatalf("Failed to dial: %v", err)
    }
    defer conn.Close()

    // Your test logic here
}
```

## Debugging

### Enable Verbose Logging
```bash
go test ./test/e2e/... -v -args -log-level=debug
```

### Check for Race Conditions
```bash
go test ./test/e2e/... -race
```

### Profile Tests
```bash
go test ./test/e2e/... -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## CI/CD Integration

These tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run E2E Tests
  run: go test ./test/e2e/... -v -timeout 2m
```

No special setup required - tests are fully self-contained!

