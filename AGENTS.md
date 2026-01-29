# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

## Architecture Overview

Yutani is a Go-based terminal display server. The architecture has three layers:

1. **Server layer** (`pkg/server/`): Core server with tview Application, tcell Screen, SessionRegistry, WidgetRegistry, and EventDispatcher. Entry point is `New(cfg *config.Config, opts ...ServerOption)`. Test mode uses `NewTestServer(maxSessions int)` with tcell SimulationScreen.

2. **Services layer** (`pkg/services/`): gRPC service implementations. Each service follows the constructor pattern `NewXxxService(srv *server.Server)`. There are 11 services: Session, Screen, Widget, Event, List, Table, Form, Tree, Layout, Debug, Test.

3. **Proto layer** (`api/proto/industries/loosh/yutani/v1/`): Protocol buffer definitions. Generated Go code lives in `pkg/proto/yutani/`.

Additional packages:
- `pkg/client/`: Go client library with fluent builder pattern
- `pkg/cli/`: Cobra-based CLI tool (debug, inspect, profile, session management)
- `pkg/config/`: Three-tier configuration (file -> env -> flags)
- `pkg/testutil/`: Shared test utilities

**Session-widget ownership model**: Each session owns the widgets it creates. Widget operations validate: nil check -> existence check -> ownership check. Sessions are identified by UUID strings. When a session is destroyed, all its widgets are cleaned up.

**tview thread safety**: tview is NOT thread-safe. All gRPC handlers that modify tview primitives MUST use `app.QueueUpdateDraw()` (for mutations that need a redraw) or `app.QueueUpdate()` (for reads) with a done channel to synchronize.

## Technology Stack

- Go 1.24.5
- gRPC + Protocol Buffers (protoc)
- tview (high-level TUI widgets)
- tcell v2 (low-level terminal control)
- cobra (CLI framework)
- bufconn (in-memory gRPC for E2E tests)

## Coding Conventions

### Service Constructor Pattern
```go
func NewXxxService(srv *server.Server) *XxxService {
    return &XxxService{server: srv}
}
```

### Request Validation Pattern
Every gRPC handler validates in this order:
1. `req.SessionId == nil` -> `codes.InvalidArgument`
2. Session doesn't exist -> `codes.NotFound`
3. Widget doesn't exist -> `codes.NotFound`
4. Widget owned by different session -> `codes.PermissionDenied`

### tview Thread Safety
```go
// Write operation (mutation + redraw):
done := make(chan error, 1)
srv.App().QueueUpdateDraw(func() {
    // modify tview primitive here
    done <- nil
})
err := <-done

// Read operation (no redraw needed):
done := make(chan *result, 1)
srv.App().QueueUpdate(func() {
    // read from tview primitive here
    done <- &result{...}
})
r := <-done
```

### Proto Message Naming
All messages follow `{Service}{Action}Request/Response`:
- `ListAddItemRequest`, `ListAddItemResponse`
- `TableGetSelectionRequest`, `TableGetSelectionResponse`
- `TreeSetExpandedRequest`, `TreeSetExpandedResponse`

### Error Codes
- `InvalidArgument`: nil or missing required fields
- `NotFound`: session or widget doesn't exist
- `PermissionDenied`: widget owned by different session
- `ResourceExhausted`: max sessions exceeded
- `Unimplemented`: feature not yet implemented

## Proto Workflow

When modifying `.proto` files:
```bash
# 1. Edit proto files
vim api/proto/industries/loosh/yutani/v1/xxx.proto

# 2. Regenerate Go code
make proto

# 3. Update Go service implementations in pkg/services/
# 4. Update client library in pkg/client/
# 5. Build and test
make build && go test ./...
```

## Testing

### Test Infrastructure
- **Unit tests** (`pkg/services/*_test.go`, `pkg/server/*_test.go`): Test individual service methods using test helpers
- **E2E tests** (`test/e2e/`): Full client-server tests using bufconn (in-memory gRPC transport)
- **Acceptance tests** (`test/e2e/acceptance_test.go`): Workflow-level tests covering realistic user scenarios
- **Contract tests** (`test/e2e/contract_test.go`): Verify every gRPC method's happy path and error cases
- **Integration tests** (`pkg/client/testing/`): Real TCP gRPC for integration scenarios

### Running Tests
```bash
go test ./...              # All tests
go test ./test/e2e/... -v  # E2E tests with verbose output
go test ./pkg/services/... -v -run TestListService  # Specific test
```

### Shared Test Utilities
Import from `pkg/testutil`:
```go
import "github.com/chazu/yutani/pkg/testutil"

testutil.BoolPtr(true)
testutil.StrPtr("hello")
testutil.Int32Ptr(42)
```

### Test Server Setup
For tests that need a full server with simulation screen:
```go
srv, err := server.NewTestServer(10)
```

## How to Add a Widget Type

1. Add enum value to `WidgetType` in `api/proto/industries/loosh/yutani/v1/widget.proto`
2. Add properties message if needed (e.g., `ImageProperties`)
3. Add to `WidgetProperties.type_properties` oneof
4. Run `make proto`
5. Add `createXxx()` function in `pkg/services/widget_factory.go`
6. Add `applyXxxProperties()` function in `pkg/services/widget_factory.go`
7. Add case to `createWidget` switch in widget_factory.go
8. Add case to `applyProperties` switch in widget_factory.go
9. Update `SupportedWidgets` in `pkg/services/session.go` ServerCapabilities
10. Add tests in `pkg/services/widget_factory_test.go`
11. Run `make build && go test ./...`

