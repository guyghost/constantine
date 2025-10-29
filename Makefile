.PHONY: build run clean test test-race test-coverage fmt vet lint install-deps build-all dev security vulncheck ci-validate ci-test ci-lint ci-build help quality deadcode duplication complexity install-tools pre-commit sbom audit

# Variables
BINARY_NAME=constantine
BACKTEST_BINARY=backtest
CMD_BOT=./cmd/bot
CMD_BACKTEST=./cmd/backtest
BIN_DIR=./bin
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)

# Build the main application
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	@go build -v -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_BOT)

# Build backtest binary
build-backtest:
	@echo "Building $(BACKTEST_BINARY) version $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	@go build -v -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BACKTEST_BINARY) $(CMD_BACKTEST)

# Build all binaries
build-bins: build build-backtest

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with race detector (as in CI)
test-race:
	@echo "Running tests with race detector..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Check if code is formatted
fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --config=.golangci.yml

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod verify

# Tidy go.mod and go.sum
tidy:
	@echo "Tidying go.mod and go.sum..."
	@go mod tidy

# Verify go.mod and go.sum
verify:
	@echo "Verifying go.mod and go.sum..."
	@go mod verify
	@go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod or go.sum is not tidy"; exit 1)

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build -v -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_BOT)
	@GOOS=linux GOARCH=arm64 go build -v -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_BOT)
	@GOOS=darwin GOARCH=amd64 go build -v -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_BOT)
	@GOOS=darwin GOARCH=arm64 go build -v -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_BOT)
	@GOOS=windows GOARCH=amd64 go build -v -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_BOT)

# Development mode (with auto-reload)
dev:
	@echo "Starting in development mode..."
	@go run $(CMD_BOT)

# Check for security vulnerabilities with govulncheck
vulncheck:
	@echo "Checking for vulnerabilities with govulncheck..."
	@which govulncheck > /dev/null || (echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest)
	@govulncheck ./...

# Security audit (comprehensive)
audit: vulncheck
	@echo "Running comprehensive security audit..."
	@which nancy > /dev/null || echo "Consider installing nancy for enhanced security scanning"
	@echo "Checking for sensitive data in code..."
	@grep -r "password\|secret\|token\|api.key" --include="*.go" . || echo "No obvious sensitive data found"

# Legacy security check (requires Docker)
security:
	@echo "Checking for security vulnerabilities with nancy..."
	@go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install golang.org/x/tools/cmd/godoc@latest
	@go install golang.org/x/tools/cmd/deadcode@latest
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@go install github.com/mibk/dupl@latest
	@go install github.com/google/go-licenses@latest
	@go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..."; curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin)
	@echo "Development tools installed successfully"

# Dead code detection
deadcode:
	@echo "Detecting dead code..."
	@which deadcode > /dev/null || go install golang.org/x/tools/cmd/deadcode@latest
	@deadcode -test ./... || echo "Analysis complete"

# Code duplication detection
duplication:
	@echo "Detecting code duplication..."
	@which dupl > /dev/null || go install github.com/mibk/dupl@latest
	@dupl -threshold 50 -files '**/*.go' || echo "Analysis complete"

# Code complexity analysis
complexity:
	@echo "Analyzing code complexity..."
	@which gocyclo > /dev/null || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@gocyclo -over 15 . || echo "Analysis complete"
	@gocyclo -avg .

# Run all quality checks
quality: deadcode duplication complexity
	@echo "✅ Quality checks complete"

# Generate SBOM (Software Bill of Materials)
sbom:
	@echo "Generating SBOM..."
	@which cyclonedx-gomod > /dev/null || go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
	@cyclonedx-gomod app -json -output sbom.json -licenses -std
	@echo "SBOM generated: sbom.json"

# Setup pre-commit hooks
pre-commit:
	@echo "Setting up pre-commit hooks..."
	@which pre-commit > /dev/null || (echo "Please install pre-commit: pip install pre-commit"; exit 1)
	@pre-commit install
	@pre-commit install --hook-type commit-msg
	@echo "Pre-commit hooks installed"

# Generate documentation
docs:
	@echo "Generating documentation..."
	@godoc -http=:6060

# CI validation job (mimics CI)
ci-validate: fmt-check vet verify
	@echo "✅ Validation passed"

# CI test job (mimics CI)
ci-test: test-race
	@echo "✅ Tests passed"

# CI lint job (mimics CI)
ci-lint: lint
	@echo "✅ Linting passed"

# CI build job (mimics CI)
ci-build: build build-backtest
	@echo "✅ Build passed"

# CI security job (mimics CI)
ci-security: vulncheck
	@echo "✅ Security check passed"

# Run all CI checks locally
ci: ci-validate ci-test ci-lint ci-build ci-security quality
	@echo "✅ All CI checks passed"

# Help command
help:
	@echo "Available commands:"
	@echo ""
	@echo "Build commands:"
	@echo "  make build          - Build the main application"
	@echo "  make build-backtest - Build the backtest binary"
	@echo "  make build-bins     - Build all binaries"
	@echo "  make build-all      - Build for multiple platforms"
	@echo ""
	@echo "Run commands:"
	@echo "  make run            - Build and run the application"
	@echo "  make dev            - Run in development mode"
	@echo ""
	@echo "Test commands:"
	@echo "  make test           - Run tests"
	@echo "  make test-race      - Run tests with race detector"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "Code quality:"
	@echo "  make fmt            - Format code"
	@echo "  make fmt-check      - Check code formatting"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make quality        - Run all quality checks"
	@echo "  make deadcode       - Detect dead code"
	@echo "  make duplication    - Detect code duplication"
	@echo "  make complexity     - Analyze code complexity"
	@echo ""
	@echo "Dependencies:"
	@echo "  make install-deps   - Install dependencies"
	@echo "  make install-tools  - Install development tools"
	@echo "  make tidy           - Tidy go.mod and go.sum"
	@echo "  make verify         - Verify go.mod and go.sum"
	@echo ""
	@echo "Security:"
	@echo "  make vulncheck      - Check for vulnerabilities (govulncheck)"
	@echo "  make audit          - Comprehensive security audit"
	@echo "  make security       - Check for vulnerabilities (nancy/docker)"
	@echo "  make sbom           - Generate Software Bill of Materials"
	@echo ""
	@echo "CI simulation:"
	@echo "  make ci-validate    - Run CI validation checks"
	@echo "  make ci-test        - Run CI test checks"
	@echo "  make ci-lint        - Run CI lint checks"
	@echo "  make ci-build       - Run CI build checks"
	@echo "  make ci-security    - Run CI security checks"
	@echo "  make ci             - Run all CI checks"
	@echo ""
	@echo "Development:"
	@echo "  make pre-commit     - Setup pre-commit hooks"
	@echo ""
	@echo "Other:"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docs           - Generate documentation"
	@echo "  make help           - Show this help message"
