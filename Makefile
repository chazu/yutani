.PHONY: proto clean build build-server build-test-client build-phase4-demo run demo install-tools dev fmt tidy

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

# Build all binaries (server and clients)
build: build-server build-test-client build-phase4-demo
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

