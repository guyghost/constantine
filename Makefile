.PHONY: build run clean test fmt vet lint install-deps

# Build the application
build:
	@echo "Building scalping-bot..."
	@go build -o scalping-bot cmd/bot/main.go

# Run the application
run: build
	@echo "Running scalping-bot..."
	@./scalping-bot

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f scalping-bot
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build -o scalping-bot-linux-amd64 cmd/bot/main.go
	@GOOS=darwin GOARCH=amd64 go build -o scalping-bot-darwin-amd64 cmd/bot/main.go
	@GOOS=darwin GOARCH=arm64 go build -o scalping-bot-darwin-arm64 cmd/bot/main.go
	@GOOS=windows GOARCH=amd64 go build -o scalping-bot-windows-amd64.exe cmd/bot/main.go

# Development mode (with auto-reload)
dev:
	@echo "Starting in development mode..."
	@go run cmd/bot/main.go

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

# Generate documentation
docs:
	@echo "Generating documentation..."
	@godoc -http=:6060

# Help command
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Build and run the application"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run linter"
	@echo "  make install-deps   - Install dependencies"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make dev            - Run in development mode"
	@echo "  make security       - Check for security vulnerabilities"
	@echo "  make docs           - Generate documentation"
