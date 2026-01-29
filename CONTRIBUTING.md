# Contributing to Yutani

## Prerequisites

- Go 1.24.5 or later
- Protocol Buffers compiler (`protoc`)
- gRPC Go plugins

### macOS

```bash
brew install protobuf
```

### Linux

```bash
apt-get install protobuf-compiler
```

## Setup

```bash
git clone <repo-url>
cd yutani
make install-tools   # Install protoc-gen-go and protoc-gen-go-grpc
make proto           # Generate protobuf Go code
make build           # Build everything
```

## Development Workflow

1. Edit code or proto files
2. If proto files changed: `make proto`
3. Build: `make build`
4. Run tests: `go test ./...`
5. Format code: `make fmt`
6. Tidy deps: `make tidy`

## Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- See `AGENTS.md` for project-specific patterns:
  - Service constructor: `NewXxxService(srv *server.Server)`
  - Request validation: nil -> exists -> ownership
  - tview thread safety: `QueueUpdateDraw` / `QueueUpdate` with done channels
  - Proto messages: `{Service}{Action}Request/Response`

## Testing

Run the full test suite before submitting changes:

```bash
go test ./... -count=1
```

Tests are organized as:
- Unit tests in `pkg/services/*_test.go` and `pkg/server/*_test.go`
- E2E tests in `test/e2e/e2e_test.go`
- Acceptance tests in `test/e2e/acceptance_test.go`
- Contract tests in `test/e2e/contract_test.go`

Use shared helpers from `pkg/testutil/` instead of defining local pointer helpers.

## Adding a New Widget Type

See the "How to Add a Widget Type" section in `AGENTS.md`.

## Adding a New gRPC Service

1. Define the service in a new `.proto` file under `api/proto/industries/loosh/yutani/v1/`
2. Run `make proto`
3. Implement the service in `pkg/services/`
4. Register it in `pkg/services/register.go` (`RegisterAllServices` function)
5. Registration will automatically apply to the CLI server, test harnesses, and E2E tests
6. Add unit tests and E2E test coverage
7. Update `AGENTS.md` with the new service count

## Pull Requests

- Keep changes focused on a single concern
- Include tests for new functionality
- Ensure `go test ./...` passes
- Run `make fmt` before submitting
