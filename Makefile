.PHONY: proto clean build build-server build-test-client build-phase4-demo build-examples run demo install-tools dev fmt tidy run-example list-examples

# Proto compilation
proto:
	@echo "Compiling protobuf files..."
	@mkdir -p pkg/proto/yutani
	protoc \
		--proto_path=api/proto \
		--go_out=. \
		--go_opt=module=github.com/chazu/yutani \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/chazu/yutani \
		api/proto/industries/loosh/yutani/v1/*.proto
	@echo "Proto compilation complete"

# Build server only
build-server: proto
	@echo "Building yutani-server..."
	go build -o bin/yutani-server ./cmd/yutani-server
	@echo "Build complete: bin/yutani-server"

# Build all binaries (server, clients, and examples)
build: build-server build-test-client build-phase4-demo build-examples
	@echo "All builds complete!"

# Build test client
build-test-client: proto
	@echo "Building test-client..."
	go build -o bin/test-client ./cmd/test-client
	@echo "Build complete: bin/test-client"

# Build phase4 demo
build-phase4-demo: proto
	@echo "Building phase4-demo..."
	go build -o bin/phase4-demo ./cmd/phase4-demo
	@echo "Build complete: bin/phase4-demo"

# Build all examples
build-examples: proto
	@echo "Building examples..."
	@mkdir -p bin/examples
	@echo "  Building simple-list..."
	@go build -o bin/examples/simple-list ./examples/simple-list
	@echo "  Building data-table..."
	@go build -o bin/examples/data-table ./examples/data-table
	@echo "  Building login-form..."
	@go build -o bin/examples/login-form ./examples/login-form
	@echo "  Building file-browser..."
	@go build -o bin/examples/file-browser ./examples/file-browser
	@echo "  Building dashboard..."
	@go build -o bin/examples/dashboard ./examples/dashboard
	@echo "  Building process-monitor..."
	@go build -o bin/examples/process-monitor ./examples/process-monitor
	@echo "  Building chat-app..."
	@go build -o bin/examples/chat-app ./examples/chat-app
	@echo "  Building text-editor..."
	@go build -o bin/examples/text-editor ./examples/text-editor
	@echo "Examples build complete: bin/examples/"

# Run server
run: build-server
	./bin/yutani-server

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf pkg/proto/yutani/*.pb.go
	rm -rf bin/
	@echo "Clean complete"

# Install required tools
install-tools:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Tools installed. Make sure protoc is installed on your system."
	@echo "  macOS: brew install protobuf"
	@echo "  Linux: apt-get install protobuf-compiler"

# Run demo (builds everything and runs demo client - server must be running separately)
demo: build-phase4-demo
	./bin/phase4-demo

# Development helpers
dev: clean build-server run

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Run a specific example (requires server to be running)
# Usage: make run-example EXAMPLE=text-editor
run-example:
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "Error: EXAMPLE not specified"; \
		echo "Usage: make run-example EXAMPLE=<name>"; \
		echo "Available examples:"; \
		echo "  simple-list, data-table, login-form"; \
		echo "  file-browser, dashboard, process-monitor"; \
		echo "  chat-app, text-editor"; \
		exit 1; \
	fi
	@if [ ! -f "bin/examples/$(EXAMPLE)" ]; then \
		echo "Building $(EXAMPLE)..."; \
		make build-examples; \
	fi
	@echo "Running $(EXAMPLE)..."
	@echo "Make sure the server is running in another terminal!"
	./bin/examples/$(EXAMPLE)

# List all available examples
list-examples:
	@echo "Available examples:"
	@echo "  simple-list       - Basic list widget demonstration"
	@echo "  data-table        - Table widget with data"
	@echo "  login-form        - Form widget example"
	@echo "  file-browser      - TreeView file system browser"
	@echo "  dashboard         - Grid layout system monitor"
	@echo "  process-monitor   - Real-time process table"
	@echo "  chat-app          - Multi-room chat application"
	@echo "  text-editor       - Text editor with syntax highlighting"
	@echo ""
	@echo "Usage: make run-example EXAMPLE=<name>"
	@echo "   or: ./bin/examples/<name>"

