# Developer Quick Reference

This is a quick reference guide for developers working on Constantine. For comprehensive details, see [CI_PIPELINE.md](CI_PIPELINE.md).

## Prerequisites

- Go 1.23 or 1.24
- Python 3 (for pre-commit hooks)
- Make
- Git

## Initial Setup

```bash
# Clone the repository
git clone https://github.com/guyghost/constantine
cd constantine

# Install dependencies
make install-deps

# Install development tools (optional but recommended)
make install-tools

# Setup pre-commit hooks (optional but recommended)
pip install pre-commit
make pre-commit
```

## Development Workflow

### Before Making Changes

```bash
# Create a feature branch
git checkout -b feature/my-feature

# Ensure everything builds
make build
```

### While Developing

```bash
# Format code
make fmt

# Run tests frequently
make test

# Run tests with race detector
make test-race

# Check linting
make lint
```

### Before Committing

```bash
# Run all CI checks locally
make ci

# If using pre-commit hooks, they run automatically
# Otherwise, run manually:
make fmt-check
make vet
make test-race
make lint
```

### Commit Messages

Follow Conventional Commits:

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:** feat, fix, docs, style, refactor, perf, test, build, ci, chore

**Examples:**
```
feat: add support for new exchange
fix: resolve race condition in order processing
docs: update installation guide
```

### Before Creating PR

```bash
# Run full CI suite
make ci

# Check security
make vulncheck

# Generate coverage report
make test-coverage
```

## Common Commands

### Building

```bash
make build              # Build main binary
make build-backtest     # Build backtest binary
make build-all          # Build for all platforms
```

### Testing

```bash
make test               # Run tests
make test-race          # Run with race detector
make test-coverage      # Generate HTML coverage report
```

### Code Quality

```bash
make fmt                # Format code
make fmt-check          # Check formatting
make vet                # Run go vet
make lint               # Run golangci-lint
make quality            # Run all quality checks
```

### Security

```bash
make vulncheck          # Check vulnerabilities
make audit              # Comprehensive security audit
make sbom               # Generate SBOM
```

### Analysis

```bash
make deadcode           # Find unused code
make duplication        # Find duplicate code
make complexity         # Analyze cyclomatic complexity
```

### CI Simulation

```bash
make ci-validate        # Validation checks
make ci-test            # Test suite
make ci-lint            # Linting
make ci-build           # Build checks
make ci-security        # Security checks
make ci                 # All CI checks
```

### Cleanup

```bash
make clean              # Remove build artifacts
```

## Pull Request Checklist

- [ ] Code is formatted (`make fmt`)
- [ ] Tests pass (`make test-race`)
- [ ] Linting passes (`make lint`)
- [ ] All CI checks pass (`make ci`)
- [ ] Coverage maintained or improved
- [ ] Documentation updated
- [ ] Conventional commit messages used
- [ ] No security vulnerabilities (`make vulncheck`)
- [ ] Pre-commit hooks pass (if installed)

## CI/CD Workflows

### Main CI (runs on push/PR)
- Code validation (format, vet, tidy)
- Tests with race detector
- golangci-lint
- Multi-platform builds
- Security scanning

### Security (runs daily + push/PR)
- govulncheck
- gosec
- Trivy
- Nancy
- SBOM generation
- License compliance

### Benchmarks (runs on push/PR)
- Performance benchmarks
- Regression detection
- Benchmark history tracking

### Code Quality (runs on push/PR)
- Documentation checks
- Cyclomatic complexity
- Dead code detection
- Code duplication

### Release (runs on version tags)
- Multi-platform builds
- Checksum generation
- SBOM creation
- Changelog generation
- GitHub release creation

## Troubleshooting

### CI Failures

**Validation failed:**
```bash
make fmt        # Fix formatting
make tidy       # Fix go.mod/go.sum
make vet        # Check vet errors
```

**Tests failed:**
```bash
make test-race  # Run locally
# Check logs in GitHub Actions
```

**Linting failed:**
```bash
make lint       # Run locally
# Fix reported issues
```

**Coverage below threshold:**
```bash
make test-coverage  # Check coverage report
# Add tests for uncovered code
```

### Common Issues

**Build fails:**
- Run `go mod download`
- Check Go version (1.23 or 1.24)
- Clear module cache: `go clean -modcache`

**Tests timeout:**
- Increase timeout: `go test -timeout=5m`
- Check for deadlocks or infinite loops

**Linter errors:**
- Check `.golangci.yml` configuration
- Update golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## Resources

- [Full CI/CD Documentation](CI_PIPELINE.md)
- [Testing Guide](TESTING_GUIDE.md)
- [Security Policy](../SECURITY.md)
- [Architecture](../AGENTS.md)
- [Go Report Card](https://goreportcard.com/report/github.com/guyghost/constantine)
- [Codecov](https://codecov.io/gh/guyghost/constantine)

## Getting Help

- **Documentation**: Check `/docs` directory
- **Issues**: [GitHub Issues](https://github.com/guyghost/constantine/issues)
- **Discussions**: [GitHub Discussions](https://github.com/guyghost/constantine/discussions)

## Tips

1. **Use pre-commit hooks** - Catch issues before committing
2. **Run `make ci` before pushing** - Ensure CI will pass
3. **Keep commits small** - Easier to review and revert
4. **Write tests first** - TDD approach recommended
5. **Check coverage** - Aim for >60% for critical packages
6. **Update docs** - Keep documentation in sync with code
7. **Use conventional commits** - Enables automated changelogs
8. **Run benchmarks** - Check performance impact of changes
