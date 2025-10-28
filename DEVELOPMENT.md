# Development Guidelines

This document outlines the development standards and CI/CD requirements for Constantine.

## Code Formatting

### gofmt (Required)

All Go code must be properly formatted using `gofmt`.

```bash
# Format single file
gofmt -w path/to/file.go

# Format all Go files
find . -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} \;

# Check if formatting is needed
gofmt -l ./...
```

### goimports (Optional but Recommended)

`goimports` automatically formats imports correctly:

```bash
go install golang.org/x/tools/cmd/goimports@latest
goimports -w ./internal/...
```

## Linting

### golangci-lint

The project uses golangci-lint for comprehensive code quality checks.

**Configuration:** `.golangci.yml`

```bash
# Run linting locally (requires golangci-lint installed)
golangci-lint run --config=.golangci.yml

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
```

**Enabled Linters:**
- `gofmt` - Code formatting
- `goimports` - Import management
- `govet` - Suspicious constructs detection
- `staticcheck` - Advanced static analysis
- `gosimple` - Code simplification suggestions
- `errcheck` - Unchecked errors
- `gosec` - Security issues
- `revive` - Linting rules
- `gocyclo` - Cyclomatic complexity (max: 15)
- `misspell` - Spelling mistakes
- `unused` - Unused variables/functions
- And more...

**Linter Configuration Details:**

```yaml
# Cyclomatic complexity threshold
gocyclo:
  min-complexity: 15

# Minimum constant length and occurrences
goconst:
  min-len: 3
  min-occurrences: 3

# Misspelling locale
misspell:
  locale: US
```

## CI/CD Pipeline

### GitHub Actions Workflow

**File:** `.github/workflows/ci.yml`

The CI pipeline runs on every push and pull request:

1. **Build** - Compiles the Go code
2. **Test** - Runs all unit tests
3. **Lint** - Runs golangci-lint
4. **Format Check** - Verifies gofmt compliance (future)

### CI Requirements

All of these MUST pass for PR approval:

✅ **Code Compilation** - `go build ./...`
✅ **Unit Tests** - `go test ./...`
✅ **Linting** - `golangci-lint run --config=.golangci.yml`
✅ **Code Formatting** - All code formatted with `gofmt`

### Troubleshooting CI Failures

#### Problem: "golangci-lint version mismatch"
```
Error: can't load config: the Go language version (go1.23) used to build golangci-lint 
is lower than the targeted Go version (1.24)
```

**Solution:** Update `.golangci.yml` `go` field to match CI runner version:
```yaml
run:
  go: '1.23'  # Match your CI environment Go version
```

#### Problem: "Deprecated output format"
```
warning: The output format `github-actions` is deprecated, please use `colored-line-number`
```

**Solution:** Update `.golangci.yml`:
```yaml
output:
  formats:
    - format: colored-line-number  # Instead of github-actions
```

#### Problem: "Code not formatted"

**Solution:** Run gofmt before committing:
```bash
gofmt -w ./...
git add .
git commit -m "chore: format code with gofmt"
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/strategy/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Standards

- All new features MUST have tests
- Test files use `_test.go` suffix
- Follow Table-Driven Tests pattern
- Use descriptive test names: `TestFunctionName_Scenario`

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat:` A new feature
- `fix:` A bug fix
- `docs:` Documentation only changes
- `style:` Changes that don't affect code meaning (formatting, etc.)
- `refactor:` Code changes without feature/bug changes
- `perf:` Performance improvements
- `test:` Adding missing tests or fixing existing tests
- `build:` Build system or dependency changes
- `ci:` CI configuration changes
- `chore:` Other changes (version bumps, etc.)

### Examples

```
feat(strategy): implement dynamic indicator weights
fix(dydx): improve candle polling from 1m to 10s
docs: add comprehensive troubleshooting guide
chore: format code with gofmt
test: add comprehensive test suite for trading cycle
ci: fix golangci-lint configuration for Go 1.23
```

## Pre-Commit Checklist

Before committing code:

- [ ] Code compiles: `go build ./cmd/bot`
- [ ] Tests pass: `go test ./...`
- [ ] Code formatted: `gofmt -l ./...` (no output)
- [ ] Lints pass: `golangci-lint run --config=.golangci.yml`
- [ ] Commit message follows Conventional Commits
- [ ] New features have tests
- [ ] Documentation updated if needed

## IDE Setup

### VS Code

Install extensions:
- **Go** (golang.go) - Official Go support
- **golangci-lint** (nametag.golangci-lint-plus) - Linting integration

Add to `.vscode/settings.json`:
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "gofmt",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

### GoLand / IntelliJ IDEA

1. Go to Settings → Languages & Frameworks → Go → Code Style
2. Enable "Reformat code" on save
3. Set gofmt as the formatter
4. Enable golangci-lint integration

### Vim/Neovim

Use vim-go or gopls with proper linting setup.

## Code Review Guidelines

Reviewers should check:

1. **Code Quality**
   - [ ] Follows project standards
   - [ ] Proper error handling
   - [ ] No unnecessary complexity

2. **Testing**
   - [ ] Has adequate test coverage
   - [ ] Tests are meaningful and pass
   - [ ] Edge cases considered

3. **Documentation**
   - [ ] Code is well-commented
   - [ ] Public API documented
   - [ ] README updated if needed

4. **Performance**
   - [ ] No performance regressions
   - [ ] Efficient algorithms used
   - [ ] Proper resource management

5. **Security**
   - [ ] No security vulnerabilities
   - [ ] Input validation present
   - [ ] Error messages don't leak sensitive info

## Build & Release

### Building Binaries

```bash
# Build for current platform
go build -o bin/bot ./cmd/bot

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/bot_linux ./cmd/bot
GOOS=darwin GOARCH=arm64 go build -o bin/bot_darwin_arm64 ./cmd/bot

# With version info
go build -ldflags="-X main.Version=v1.0.0" -o bin/bot ./cmd/bot
```

### Generating Release Notes

1. Review commits since last release
2. Categorize by type (feat, fix, docs, etc.)
3. Update RELEASE_NOTES.md
4. Create git tag: `git tag v1.0.0`

## Useful Commands

```bash
# Format all code
find . -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} \;

# Run all checks
go build ./... && go test ./... && golangci-lint run --config=.golangci.yml

# Check specific linter
golangci-lint run --no-config --disable-all --enable=gofmt ./...

# Generate test coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Find TODOs in code
rg "TODO|FIXME|XXX" --type go

# Update dependencies
go get -u ./...
go mod tidy
```

## Resources

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [Conventional Commits](https://www.conventionalcommits.org/)
