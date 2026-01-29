---
title: "refactor: Quality Initiative -- Semantic Consistency, Acceptance Testing, Documentation"
type: refactor
date: 2026-01-28
---

# Quality Initiative: Semantic Consistency, Acceptance Testing, Documentation

## Overview

A three-phase quality initiative for the Yutani terminal display server, executed bottom-up: fix semantic consistency first (proto + Go), then build acceptance tests against the corrected API, then clean up documentation to reflect the final state.

## Problem Statement

The Yutani codebase has grown through multiple development phases and accumulated inconsistencies:

- **Naming drift:** Three different patterns for the same "get/set selection" concept across List, Table, and Tree services. Inconsistent message prefixing within and across services.
- **Missing registrations:** TestService (headless UI testing) is not registered in either test harness, making it unusable in automated tests.
- **Capabilities gap:** ServerCapabilities omits Image and ProgressBar widget types despite full implementation.
- **Dead API surface:** WidgetService.AddChild and RemoveChild are defined in proto but permanently return Unimplemented.
- **Test infrastructure debt:** Duplicated helpers across 5 files, `time.Sleep` synchronization in 3 test harnesses, no acceptance-level coverage for DebugService or TestService.
- **Documentation sprawl:** 13+ stale historical fix summaries at repo root, minimal AGENTS.md, no CLAUDE.md, no CONTRIBUTING.md.

## Proposed Solution

Three phases in strict sequence. Each phase must complete and compile before the next begins.

---

## Phase 1: Semantic Consistency

### Prerequisites

- [ ] Commit or stash all current uncommitted changes (git status shows modified and untracked files)
- [ ] Decide open naming questions (see "Decisions Required" section below)

### 1.1 Resolve Naming Conventions

**Selection operations -- standardize on `GetSelection` / `SetSelection`:**

"Selection" is a noun referring to the selected state. "Selected" is an adjective. The noun form is more precise and matches Table's existing convention. Table also returns the richest data (row + column), making it a natural model.

| Service | Current RPC | New RPC | Current Request Message | New Request Message |
|---------|------------|---------|------------------------|---------------------|
| ListService | `GetSelected` | `GetSelection` | `GetSelectedRequest` | `ListGetSelectionRequest` |
| ListService | `SetSelected` | `SetSelection` | `SetSelectedRequest` | `ListSetSelectionRequest` |
| TableService | `GetSelection` | `GetSelection` (no change) | `GetSelectionRequest` | `TableGetSelectionRequest` |
| TableService | `SetSelection` | `SetSelection` (no change) | `SetSelectionRequest` | `TableSetSelectionRequest` |
| TreeService | `GetSelected` | `GetSelection` | `GetTreeSelectedRequest` | `TreeGetSelectionRequest` |
| TreeService | `SetSelected` | `SetSelection` | `SetTreeSelectedRequest` | `TreeSetSelectionRequest` |

**Message prefix convention -- all messages prefixed with service name:**

Target rule: `{Service}{Action}Request` / `{Service}{Action}Response`. This follows standard protobuf convention for disambiguation across services that share similar operations.

Files requiring message renames (every message in every domain-specific proto):
- `api/proto/industries/loosh/yutani/v1/list.proto` -- prefix all with `List`
- `api/proto/industries/loosh/yutani/v1/table.proto` -- prefix remaining unprefixed messages with `Table`
- `api/proto/industries/loosh/yutani/v1/tree.proto` -- already prefixed, normalize the 2 missing (`SetExpandedRequest` -> `TreeSetExpandedRequest`, `GetChildrenRequest` -> `TreeGetChildrenRequest`)
- `api/proto/industries/loosh/yutani/v1/form.proto` -- prefix remaining unprefixed messages with `Form`

**Preserve field numbers** in all renamed messages. Only type names and field names change; wire format stays compatible within the same binary.

**AddChild/RemoveChild -- remove from WidgetService proto:**

These RPCs return Unimplemented and direct users to LayoutService. They are dead API surface. Remove from `widget.proto` (lines 27-30). The generated code will stop including them, and `widget.go` lines 424-435 can be deleted.

### 1.2 Proto File Changes

For each proto file, apply the prefix convention. Detailed changelist:

**`list.proto`** -- rename all messages:
- `AddItemRequest` -> `ListAddItemRequest`
- `AddItemResponse` -> `ListAddItemResponse`
- `RemoveItemRequest` -> `ListRemoveItemRequest`
- `RemoveItemResponse` -> `ListRemoveItemResponse`
- `ClearListRequest` -> `ListClearRequest` (reorder prefix: List first, then action)
- `ClearListResponse` -> `ListClearResponse`
- `GetItemCountRequest` -> `ListGetItemCountRequest`
- `GetItemCountResponse` -> `ListGetItemCountResponse`
- `GetSelectedRequest` -> `ListGetSelectionRequest`
- `GetSelectedResponse` -> `ListGetSelectionResponse`
- `SetSelectedRequest` -> `ListSetSelectionRequest`
- `SetSelectedResponse` -> `ListSetSelectionResponse`
- `GetItemRequest` -> `ListGetItemRequest`
- `GetItemResponse` -> `ListGetItemResponse`
- RPC names: `GetSelected` -> `GetSelection`, `SetSelected` -> `SetSelection`

**`table.proto`** -- prefix remaining unprefixed messages:
- `GetDimensionsRequest` -> `TableGetDimensionsRequest`
- `GetDimensionsResponse` -> `TableGetDimensionsResponse`
- `GetSelectionRequest` -> `TableGetSelectionRequest`
- `GetSelectionResponse` -> `TableGetSelectionResponse`
- `SetSelectionRequest` -> `TableSetSelectionRequest`
- `SetSelectionResponse` -> `TableSetSelectionResponse`
- `SetFixedRequest` -> `TableSetFixedRequest`
- `SetFixedResponse` -> `TableSetFixedResponse`

**`tree.proto`** -- fix 2 missing prefixes + rename Selected -> Selection:
- `SetExpandedRequest` -> `TreeSetExpandedRequest`
- `SetExpandedResponse` -> `TreeSetExpandedResponse`
- `GetChildrenRequest` -> `TreeGetChildrenRequest`
- `GetChildrenResponse` -> `TreeGetChildrenResponse`
- `GetTreeSelectedRequest` -> `TreeGetSelectionRequest`
- `GetTreeSelectedResponse` -> `TreeGetSelectionResponse`
- `SetTreeSelectedRequest` -> `TreeSetSelectionRequest`
- `SetTreeSelectedResponse` -> `TreeSetSelectionResponse`
- RPC names: `GetSelected` -> `GetSelection`, `SetSelected` -> `SetSelection`

**`form.proto`** -- prefix remaining unprefixed messages:
- `AddFieldRequest` -> `FormAddFieldRequest`
- `AddFieldResponse` -> `FormAddFieldResponse`
- `AddButtonRequest` -> `FormAddButtonRequest`
- `AddButtonResponse` -> `FormAddButtonResponse`
- `GetFieldValueRequest` -> `FormGetFieldValueRequest`
- `GetFieldValueResponse` -> `FormGetFieldValueResponse`
- `SetFieldValueRequest` -> `FormSetFieldValueRequest`
- `SetFieldValueResponse` -> `FormSetFieldValueResponse`
- `ClearFormRequest` -> `FormClearRequest`
- `ClearFormResponse` -> `FormClearResponse`
- `GetFormItemCountRequest` -> `FormGetItemCountRequest`
- `GetFormItemCountResponse` -> `FormGetItemCountResponse`

**`widget.proto`** -- remove AddChild and RemoveChild RPCs (lines 27-30) and their request/response messages.

### 1.3 Regenerate Proto

```bash
make proto
```

This regenerates all `pkg/proto/yutani/*.pb.go` and `*_grpc.pb.go` files.

### 1.4 Update Server Services

Update every Go service file in `pkg/services/` to use the new message type names. The RPC method signatures are defined by the generated `*_grpc.pb.go` interfaces, so Go method names must match.

Files to update:
- [ ] `pkg/services/list.go` -- update all `pb.GetSelectedRequest` -> `pb.ListGetSelectionRequest`, etc.
- [ ] `pkg/services/table.go` -- update unprefixed message references
- [ ] `pkg/services/tree.go` -- update message references + method signatures for GetSelection/SetSelection
- [ ] `pkg/services/form.go` -- update unprefixed message references
- [ ] `pkg/services/widget.go` -- delete AddChild and RemoveChild methods (lines 424-435)

### 1.5 Update Client Library

Every method in `pkg/client/` that constructs proto request messages or references proto types must be updated.

Files to update:
- [ ] `pkg/client/list.go` -- `GetSelected()` -> `GetSelection()`, all message type references
- [ ] `pkg/client/table.go` -- message type references for unprefixed messages
- [ ] `pkg/client/tree.go` -- `GetSelected()` -> `GetSelection()`, all message type references
- [ ] `pkg/client/form.go` -- message type references
- [ ] `pkg/client/widget.go` -- remove AddChild/RemoveChild if exposed

### 1.6 Update Examples

All 10 example applications use the client library. If client method names change (e.g., `GetSelected` -> `GetSelection`), examples must be updated.

Files to update (check each for usage of renamed methods):
- [ ] `examples/simple-list/main.go`
- [ ] `examples/data-table/main.go`
- [ ] `examples/login-form/main.go`
- [ ] `examples/file-browser/main.go`
- [ ] `examples/dashboard/main.go`
- [ ] `examples/process-monitor/main.go`
- [ ] `examples/chat-app/main.go`
- [ ] `examples/text-editor/main.go`
- [ ] `examples/progress-demo/main.go`
- [ ] `examples/widgets-demo/main.go`

### 1.7 Update Test Files

- [ ] `test/e2e/e2e_test.go` -- update all proto message references
- [ ] `pkg/services/list_test.go` -- update message references
- [ ] `pkg/services/table_test.go` -- update message references
- [ ] `pkg/services/tree_test.go` -- update message references
- [ ] `pkg/services/form_test.go` -- update message references
- [ ] `pkg/services/widget_factory_test.go` -- if applicable
- [ ] `pkg/client/widgets_test.go` -- if applicable

### 1.8 Update CLI and Binaries

- [ ] `pkg/cli/debug.go` -- check for proto type references
- [ ] `pkg/cli/session.go` -- check for proto type references
- [ ] `cmd/test-client/main.go` -- update proto references
- [ ] `cmd/phase4-demo/main.go` -- update proto references

### 1.9 Fix ServerCapabilities

In `pkg/services/session.go` lines 122-127, add the missing widget types:

```go
SupportedWidgets: []string{
    "Box", "TextView", "InputField", "TextArea",
    "Button", "Checkbox", "DropDown", "List",
    "Table", "TreeView", "Form", "Flex",
    "Grid", "Pages", "Modal", "Image", "ProgressBar",
},
```

### 1.10 Register TestService in Test Harnesses

**`pkg/client/testing/integration.go`** -- add after line 80:
```go
pb.RegisterTestServiceServer(s.grpcServer, services.NewTestService(s.server))
```

**`test/e2e/e2e_test.go`** -- add after line 84:
```go
pb.RegisterTestServiceServer(grpcServer, services.NewTestService(srv))
```

### 1.11 Resolve Constructor Ambiguity

In `pkg/server/server.go`:
- Keep `New(cfg *config.Config, opts ...ServerOption)` as the primary constructor
- Keep `NewTestServer(maxSessions int)` as a test convenience
- Remove `NewServer(maxSessions int, mouseEnable, pasteEnable bool)` -- it has no callers in the committed codebase beyond `cmd/yutani-server/main.go`
- Update `cmd/yutani-server/main.go` to use `New()` with a `config.Config` directly

### 1.12 Verify

```bash
make proto && make build && go test ./...
```

All must pass before proceeding to Phase 2.

---

## Phase 2: Acceptance Testing

### 2.1 Test Helper Consolidation

Create a shared test utility package to eliminate duplication:

**New file: `pkg/testutil/helpers.go`**
```go
package testutil

func BoolPtr(b bool) *bool       { return &b }
func StrPtr(s string) *string    { return &s }
func Int32Ptr(i int32) *int32    { return &i }
```

Update these files to import from `testutil` instead of defining locally:
- [ ] `pkg/services/widget_factory_test.go` (lines 308-316)
- [ ] `test/e2e/e2e_test.go` (lines 746-756)
- [ ] `pkg/server/registry_test.go` (lines 369-373)
- [ ] `cmd/test-client/main.go` (lines 374-382)

Note: `pkg/client/widget.go` (lines 97-105) is production code, not test code. Leave it as-is since it's unexported and used internally by the client SDK.

### 2.2 Replace time.Sleep Synchronization

Replace sleep-based startup waits with a gRPC readiness poll:

**New helper in `pkg/testutil/ready.go`:**
```go
func WaitForServer(t *testing.T, addr string, timeout time.Duration) {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(),
            grpc.WithTimeout(50*time.Millisecond))
        if err == nil {
            // Verify with Ping
            client := pb.NewSessionServiceClient(conn)
            _, err = client.Ping(context.Background(), &pb.PingRequest{})
            conn.Close()
            if err == nil {
                return
            }
        }
        time.Sleep(10 * time.Millisecond)
    }
    t.Fatal("server did not become ready")
}
```

Replace in:
- [ ] `pkg/client/testing/integration.go` line 116 -- replace `time.Sleep(100ms)`
- [ ] `test/e2e/e2e_test.go` line 57 -- replace `time.Sleep(100ms)`
- [ ] `pkg/services/testing_helper_test.go` line 32 -- replace `time.Sleep(50ms)`

For event subscription sleeps (`e2e_test.go` lines 369, 390), use `WaitForIdle` from TestService once it's registered.

### 2.3 Acceptance Test Structure

Add acceptance tests in `test/e2e/` alongside existing E2E tests. Use a separate file to distinguish acceptance tests from existing ones:

**New file: `test/e2e/acceptance_test.go`**

### 2.4 User Workflow Tests

Full client-to-server scenarios that exercise realistic usage patterns:

- [ ] **Session lifecycle workflow:** Create session -> verify in ListSessions -> Ping -> build a UI -> destroy session -> verify gone
- [ ] **Interactive form workflow:** Create session -> create Form -> add fields + buttons -> set field values -> get field values -> verify form state -> clear form
- [ ] **List management workflow:** Create session -> create List -> add items -> select item -> get selected -> remove item -> verify count -> clear
- [ ] **Table data workflow:** Create session -> create Table -> set cells in bulk -> get dimensions -> select cell -> get selection -> set fixed rows/columns -> verify
- [ ] **Tree navigation workflow:** Create session -> create TreeView -> set root -> add children -> expand/collapse -> select node -> get children -> remove node
- [ ] **Layout composition workflow:** Create session -> create Flex -> add items with proportions -> set direction -> create nested Grid -> add items to grid -> create Pages -> add pages -> show/hide pages
- [ ] **Multi-widget dashboard workflow:** Create session -> create multiple widget types -> arrange in Flex/Grid layout -> set root -> interact with widgets -> verify all state
- [ ] **Headless interaction workflow** (requires TestService): Create session in test mode -> create interactive widgets -> InjectKey/InjectText/InjectMouse -> verify widget state changed -> WaitForIdle

### 2.5 API Contract Tests

Verify every gRPC method returns correct responses AND correct errors:

**New file: `test/e2e/contract_test.go`**

For each service, test:
1. Happy path for every RPC method
2. Nil session_id returns `InvalidArgument`
3. Non-existent session_id returns `NotFound`
4. Widget owned by different session returns `PermissionDenied` (for widget operations)
5. Wrong widget type returns appropriate error (e.g., calling ListService.AddItem on a Table widget)

Services and their RPC coverage:

- [ ] **SessionService** (5 RPCs): CreateSession, DestroySession, ListSessions, GetServerInfo, Ping
  - Edge: max sessions exceeded -> `ResourceExhausted`
  - Edge: double-destroy -> `NotFound`
- [ ] **WidgetService** (8 RPCs after removing AddChild/RemoveChild): CreateWidget, DeleteWidget, SetProperties, GetProperties, SetRoot, SetFocus, GetFocus, ListWidgets
  - Edge: delete non-existent widget, set properties on deleted widget, SetRoot with non-existent widget
- [ ] **ScreenService** (RPCs from screen.proto): all methods with session validation
- [ ] **EventService** (RPCs from event.proto): Subscribe, InjectEvent, SetEventFilter
- [ ] **ListService** (7 RPCs): AddItem, RemoveItem, Clear, GetItemCount, GetSelection, SetSelection, GetItem
  - Edge: negative index, out-of-bounds index, operations on empty list
- [ ] **TableService** (8 RPCs): SetCell, GetCell, SetCells, Clear, GetDimensions, GetSelection, SetSelection, SetFixed
  - Edge: operations on empty table, negative row/column
- [ ] **FormService** (6 RPCs): AddField, AddButton, GetFieldValue, SetFieldValue, Clear, GetItemCount
  - Edge: get non-existent field value, operations on empty form
- [ ] **TreeService** (7 RPCs): SetRoot, AddChild, RemoveNode, SetExpanded, GetSelection, SetSelection, GetChildren
  - Edge: add child to non-existent parent, remove root node, operations on empty tree
- [ ] **LayoutService** (11 RPCs): all Flex/Grid/Pages operations
  - Edge: add item to wrong layout type, remove non-existent item
- [ ] **DebugService** (RPCs from debug.proto): GetScreenDump, GetWidgetState, GetAllWidgetBounds
- [ ] **TestService** (RPCs from test.proto): InjectKey, InjectText, InjectMouse, WaitForIdle, IsTestMode
  - Edge: calling inject methods when not in test mode

### 2.6 Verify

```bash
go test ./... -v -count=1
```

All existing and new tests must pass.

---

## Phase 3: Documentation

### 3.1 Audit Root-Level Markdown Files

Review each file individually. Recommended disposition:

**DELETE** (git history preserves content):
- [ ] `CHANGES_SUMMARY.md` -- historical session artifact
- [ ] `BUILD_FIXES_SUMMARY.md` -- historical session artifact
- [ ] `DISPLAY_FIX.md` -- historical fix, resolved
- [ ] `FINAL_FIX_SUMMARY.md` -- historical session artifact
- [ ] `COMPLETION_SUMMARY.md` -- historical session artifact
- [ ] `EVENT_HANDLING_FIXES.md` -- historical fix, resolved
- [ ] `TROUBLESHOOTING_SUMMARY.md` -- historical session artifact
- [ ] `AUDIT_SUMMARY.md` -- historical audit, resolved
- [ ] `PORT_FIX.md` -- historical fix, resolved
- [ ] `SETROOT_FIX.md` -- historical fix, resolved
- [ ] `E2E_TESTS_SUMMARY.md` -- historical session artifact
- [ ] `UNIT_TESTS_SUMMARY.md` -- historical session artifact
- [ ] `MAKEFILE_UPDATE.md` -- historical session artifact

**MERGE then delete source:**
- [ ] `RUN_SERVER.md` -- merge relevant content into `QUICKSTART.md`
- [ ] `MAKEFILE_USAGE.md` -- merge relevant content into `README.md` (Contributing section) or new `CONTRIBUTING.md`
- [ ] `DEBUG_GUIDE.md` -- merge into `DEBUGGING.md`
- [ ] `GRPCURL_GUIDE.md` -- merge into `DEBUGGING.md`

**KEEP and UPDATE:**
- [ ] `README.md` -- update Contributing section, add License
- [ ] `PRD.md` -- update Phase 6 checkboxes, directory structure, widget type list
- [ ] `QUICKSTART.md` -- verify accuracy after renames
- [ ] `TUTORIAL.md` -- verify accuracy after renames
- [ ] `DEBUGGING.md` -- merge in DEBUG_GUIDE and GRPCURL_GUIDE content
- [ ] `KEYBOARD_NAVIGATION.md` -- verify accuracy
- [ ] `bugs.md` -- keep as active tracking
- [ ] `wish_report.md` -- keep as active tracking

### 3.2 Expand AGENTS.md

Add the following sections to `AGENTS.md`:

- [ ] **Architecture Overview:** Server/services/proto layers, session-widget ownership model, tview thread safety
- [ ] **Technology Stack:** Go 1.24.5, gRPC, protobuf, tview, tcell, cobra
- [ ] **Coding Conventions:**
  - Service constructor pattern: `NewXxxService(srv *server.Server)`
  - Request validation pattern: nil check -> existence check -> ownership check
  - tview thread safety: `QueueUpdateDraw` with done channel for mutations, `QueueUpdate` for reads
  - Error codes: `InvalidArgument` for nil/missing, `NotFound` for doesn't exist, `PermissionDenied` for wrong owner
- [ ] **Proto Workflow:** edit `.proto` -> `make proto` -> update Go -> `make build` -> `go test ./...`
- [ ] **Testing Requirements:** describe test harnesses, where to add tests, test patterns
- [ ] **How to Add a Widget Type:** proto enum + properties message -> widget_factory.go createXxx + applyXxxProperties -> update ServerCapabilities

### 3.3 Create CLAUDE.md

Project-level guidance for Claude Code sessions:

- [ ] Project overview (1 paragraph)
- [ ] Key directories and their purposes
- [ ] Build commands (`make proto`, `make build`, `go test ./...`)
- [ ] Common pitfalls (tview thread safety, proto regeneration after changes)
- [ ] Reference to AGENTS.md for detailed conventions

### 3.4 Update PRD.md

- [ ] Update Phase 6 checklist to reflect current implementation status
- [ ] Update directory structure (Section 6.1) to include `debug.go`, `test.go`, `cli/`, etc.
- [ ] Update widget type list to include Image and ProgressBar
- [ ] Mark proto naming conventions as "standardized as of Phase 7"

### 3.5 Create CONTRIBUTING.md

- [ ] Prerequisites (Go 1.24.5, protoc, grpc tools)
- [ ] Setup instructions (clone, `make install-tools`, `make proto`, `make build`)
- [ ] Development workflow (make proto after proto changes, run tests before committing)
- [ ] Code style (Go standard, refer to AGENTS.md for project-specific patterns)
- [ ] PR process

### 3.6 Verify

```bash
# Verify no broken links in docs
# Verify build still works
make build && go test ./...
```

---

## Decisions Required

These open questions from the brainstorm are resolved in this plan:

| Question | Decision | Rationale |
|----------|----------|-----------|
| Selection naming pattern | `GetSelection`/`SetSelection` | Noun form; matches Table's existing pattern |
| Message prefix convention | `{Service}{Action}Request` | Standard protobuf disambiguation |
| AddChild/RemoveChild | Remove from proto | Dead API surface; Unimplemented RPCs are confusing |
| Constructor ambiguity | Remove `NewServer`, keep `New` + `NewTestServer` | `NewServer` has only one caller easily migrated |
| Test location | `test/e2e/` (separate files) | Avoids directory proliferation |
| Readiness mechanism | gRPC dial + Ping poll loop | Deterministic, no sleep |
| Stale docs | Individual audit (see Section 3.1) | Per brainstorm decision |
| CLAUDE.md vs AGENTS.md | CLAUDE.md = project context, AGENTS.md = conventions + workflow | Complementary purposes |
| License | Defer (out of scope for this initiative) | Requires project owner decision |

## Acceptance Criteria

### Phase 1
- [ ] All proto message types follow `{Service}{Action}Request/Response` convention
- [ ] All selection RPCs use `GetSelection`/`SetSelection` naming
- [ ] AddChild/RemoveChild removed from WidgetService proto
- [ ] ServerCapabilities includes Image and ProgressBar
- [ ] TestService registered in both test harnesses
- [ ] `NewServer` constructor removed; `cmd/yutani-server/main.go` uses `New`
- [ ] `make proto && make build && go test ./...` all pass

### Phase 2
- [ ] Test helpers consolidated into `pkg/testutil/`
- [ ] `time.Sleep` removed from test startup synchronization
- [ ] User workflow acceptance tests cover all 8 scenarios listed
- [ ] API contract tests cover all ~60 RPC methods across 11 services
- [ ] Error case tests cover nil session, non-existent session, wrong owner, wrong widget type, boundary conditions
- [ ] `go test ./... -count=1` passes

### Phase 3
- [ ] 13 stale markdown files deleted
- [ ] 4 docs merged into their targets
- [ ] AGENTS.md expanded with architecture, conventions, proto workflow, testing sections
- [ ] CLAUDE.md created with project context
- [ ] PRD.md updated to reflect current state
- [ ] CONTRIBUTING.md created
- [ ] `make build && go test ./...` still passes

## Dependencies & Risks

**Risk 1: Proto rename cascade is large.** Each rename touches proto, generated code, server, client, examples, tests, CLI. Mitigate by doing all renames in a single commit per proto file, with `make proto && make build` after each to catch errors immediately.

**Risk 2: Uncommitted changes in git.** The working tree has significant modifications and untracked files. Mitigate by committing or stashing before starting Phase 1.

**Risk 3: No CI pipeline.** Breaking changes are caught only by local `make build` and `go test ./...`. Mitigate by running the full build after every sub-step within Phase 1.

**Risk 4: Test helper consolidation during rename.** Changing imports simultaneously with proto renames makes debugging harder. Mitigate by doing test helper consolidation as the first step of Phase 2, after Phase 1 is complete.

## References

- Brainstorm: `docs/brainstorms/2026-01-28-quality-initiative-brainstorm.md`
- Proto definitions: `api/proto/industries/loosh/yutani/v1/*.proto`
- Server services: `pkg/services/*.go`
- Client library: `pkg/client/*.go`
- E2E tests: `test/e2e/e2e_test.go`
- Integration harness: `pkg/client/testing/integration.go`
- Server constructors: `pkg/server/server.go:28` (New), `pkg/server/server.go:76` (NewServer), `pkg/server/server.go:87` (NewTestServer)
- ServerCapabilities: `pkg/services/session.go:122-127`
