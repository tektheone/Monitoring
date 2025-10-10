.PHONY: build run test clean deps

# Build the server binary
build:
	@echo "Building server..."
	@mkdir -p bin
	go build -o bin/server ./cmd/server

# Run the server with default settings
run: build
	@echo "Running server..."
	./bin/server --devices ./devices.csv --bind 127.0.0.1 --port 6733

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Run the simulator (requires device-simulator-mac-amd64 in current directory)
simulate: build
	@echo "Starting server in background..."
	./bin/server --devices ./devices.csv --bind 127.0.0.1 --port 6733 &
	@echo "Waiting for server to start..."
	sleep 2
	@echo "Running simulator..."
	./device-simulator-mac-amd64 --server http://127.0.0.1:6733/api/v1
	@echo "Stopping server..."
	pkill -f "bin/server" || true

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the server binary"
	@echo "  run           - Build and run the server"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code (requires golangci-lint)"
	@echo "  simulate      - Run server and simulator"
	@echo "  help          - Show this help"
