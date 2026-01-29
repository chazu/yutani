# Yutani - Claude Code Context

Yutani is a Go-based terminal display server providing networked, widget-based TUI capabilities via gRPC. Clients connect over gRPC to create sessions, build widget trees, and receive input events. The server renders everything using tview/tcell.

## Key Directories

- `cmd/yutani/` - Unified CLI entry point (`yutani server`, `yutani ping`, etc.)
- `pkg/server/` - Core server (tview app, session registry, widget registry, event dispatcher)
- `pkg/services/` - gRPC service implementations (11 services) and `RegisterAllServices()` helper
- `pkg/client/` - Go client library with fluent builders
- `pkg/cli/` - CLI commands (server, debug, inspect, profile)
- `pkg/proto/yutani/` - Generated protobuf Go code (do not edit manually)
- `pkg/config/` - Configuration loading (file -> env -> flags)
- `pkg/testutil/` - Shared test utilities
- `api/proto/industries/loosh/yutani/v1/` - Protobuf source definitions
- `test/e2e/` - End-to-end, acceptance, and contract tests
- `examples/` - Example applications (8 examples)

## Build Commands

```bash
make proto          # Regenerate protobuf Go code
make build          # Build server, clients, and examples
make run            # Build and run server
go test ./...       # Run all tests
go test ./... -v    # Verbose test output
```

## Common Pitfalls

1. **tview thread safety**: Never modify tview primitives directly from gRPC handlers. Always use `app.QueueUpdateDraw()` for mutations or `app.QueueUpdate()` for reads, with a done channel.

2. **Proto regeneration**: After editing any `.proto` file, run `make proto` before `make build`. The generated `*.pb.go` and `*_grpc.pb.go` files must not be edited manually.

3. **Proto naming convention**: All messages follow `{Service}{Action}Request/Response` (e.g., `ListAddItemRequest`). All selection RPCs use `GetSelection`/`SetSelection`.

4. **Test mode**: Use `server.NewTestServer(maxSessions)` for tests. It creates a tcell SimulationScreen instead of a real terminal.

5. **Validation order**: gRPC handlers validate: nil check -> existence check -> ownership check. Use `codes.InvalidArgument`, `codes.NotFound`, `codes.PermissionDenied`.

## See Also

- `AGENTS.md` - Detailed conventions, architecture, and workflows
- `PRD.md` - Product requirements and full API specification
- `TUTORIAL.md` - Client library tutorial
- `DEBUGGING.md` - Debugging and grpcurl usage
