# Yutani Bugs and Issues

## Fixed Issues

### ~~1. IsHealthy() Creates Sessions Instead of Pinging~~ ✓ FIXED
**File:** `pkg/client/client.go:453-468`

Fixed by using `Ping` RPC instead of `CreateSession`.

---

### ~~2. Timestamp Inconsistency~~ ✓ FIXED
**Files:** `pkg/services/widget.go`, `pkg/server/events.go`

Fixed - all timestamps now use nanoseconds.

---

### ~~3. Event Channel Close Race Condition~~ ✓ FIXED
**File:** `pkg/server/events.go`

Fixed by adding atomic `closed` flag to `EventSubscriber` and checking before send.

---

### ~~4. Unused State Management Code~~ ✓ FIXED
**File:** `pkg/client/state.go`

Fixed - file removed entirely.

---

### ~~5. Unused Connection Pool~~ ✓ FIXED
**File:** `pkg/client/pool.go`

Fixed - `pool.go` and `pool_test.go` removed.

---

### ~~6. Redundant Server Constructors~~ ✓ FIXED
**File:** `pkg/server/server.go`

Fixed - `NewServer()` and `NewTestServer()` now delegate to `New()`.

---

### ~~7. Stale TODO Comments / Dead AddChild/RemoveChild~~ ✓ FIXED
**File:** `pkg/services/widget.go`

Fixed - simplified to return `codes.Unimplemented` with helpful error message.

---

### ~~8. Service Tests Hanging Forever~~ ✓ FIXED
**Files:** `pkg/services/*_test.go`

Fixed - added `setupTestServer()` helper that runs tview event loop in goroutine.

---

### ~~9. MockClient Mutex Deadlock~~ ✓ FIXED
**File:** `pkg/client/testing/mock_client.go`

Fixed by moving `RecordCall()` outside the lock in `OnEvent()`.

---

### ~~10. E2E Screen Operations Test Failure~~ ✓ FIXED
**File:** `pkg/services/screen.go`

Fixed - `SetCell` was using `QueueUpdateDraw` which triggered a full widget redraw that overwrote the cell content. Changed to use `QueueUpdate` with explicit `screen.Show()` to commit changes without widget redraw.

---

## Medium Priority (Cleanup)

### ~~11. Data Race in Health Check Goroutine~~ ✓ FIXED
**File:** `pkg/client/client.go`

Fixed by capturing `healthCheckTicker` and `healthCheckStop` channels in local variables before starting the goroutine, preventing a data race when `StopHealthCheck` sets them to nil.

---

### ~~12. Add Jitter to Exponential Backoff~~ ✓ FIXED
**File:** `pkg/client/reconnect.go`

Fixed by adding `JitterFactor` field to `RetryOptions` (default 0.2 = 20% jitter). Delays now vary randomly to prevent thundering herd when multiple clients reconnect simultaneously.

---

### ~~13. Integration Test Helper Missing Event Loop~~ ✓ FIXED
**File:** `pkg/client/testing/integration.go`

Fixed - `StartTestServer()` now calls `srv.Start()` and `srv.Run()` in goroutine to run the tview event loop.

---

### ~~14. Data Race in EventBatcher Test~~ ✓ FIXED
**File:** `pkg/client/event_advanced_test.go`

Fixed - the race was in the test code, not the EventBatcher itself. The test's `receivedCount` variable was being accessed without synchronization. Changed to use `atomic.Int32`.

---

## Architectural Notes

- **Single session support** is sufficient for current requirements
- **State preservation** (save/restore) is not planned
- **Connection pooling** not needed currently; reconnect.go handles reconnection independently
