# CI/CD Pipeline Documentation

This document describes the comprehensive CI/CD pipeline for Constantine, implementing industry best practices inspired by Rust's high-quality tooling standards.

## Overview

Constantine's CI/CD pipeline consists of multiple workflows that ensure code quality, security, and reliability:

- **Main CI** - Core validation, testing, linting, building, and security
- **Security** - Comprehensive security scanning and SBOM generation
- **Benchmarks** - Performance regression detection
- **Code Quality** - Documentation, complexity, and duplication analysis
- **Release** - Automated release with multi-platform builds and changelogs
- **Dependabot** - Automated dependency updates

## Workflows

### 1. Main CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request to `main`.

**Jobs:**

#### Validation
- Checks code formatting with `gofmt`
- Runs `go vet` for correctness
- Verifies `go.mod` and `go.sum` tidiness
- Validates documentation completeness
- Tests on Go 1.23 and 1.24

#### Tests
- Runs all tests with race detector
- Generates code coverage reports
- Enforces minimum coverage threshold (40%)
- Uploads coverage to Codecov
- Tests on Go 1.23 and 1.24 for compatibility

#### Linting
- Runs `golangci-lint` with 19 active linters
- Timeout: 10 minutes
- Uses comprehensive caching

#### Build
- Cross-compilation for Linux, macOS, Windows
- Architectures: AMD64, ARM64
- Includes version information in binaries
- Optimized with `-ldflags="-s -w"`
- Uploads build artifacts

#### Security
- Runs `govulncheck` for vulnerability scanning
- Caches analysis results for faster subsequent runs

### 2. Security Workflow (`.github/workflows/security.yml`)

Runs on push, pull requests, and daily at 2 AM UTC.

**Jobs:**

#### Security Audit
- `govulncheck` - Go vulnerability database
- `gosec` - Security scanner with SARIF output
- Uploads results to GitHub Security tab

#### Trivy Scan
- Comprehensive filesystem vulnerability scanning
- Scans for CRITICAL, HIGH, and MEDIUM severities
- Uploads SARIF results to GitHub Security

#### Nancy Scan
- Dependency vulnerability checking
- Uses Sonatype vulnerability database

#### SBOM Generation
- Creates CycloneDX Software Bill of Materials
- Includes license information
- 90-day artifact retention

#### License Compliance
- Checks allowed licenses: MIT, Apache-2.0, BSD-2/3-Clause, ISC
- Generates license report
- Uploaded as artifact

### 3. Benchmarks Workflow (`.github/workflows/benchmarks.yml`)

Runs on push and pull requests.

**Features:**
- Runs all benchmarks with 5-second benchmark time
- Tracks performance over time
- Alerts on 150%+ performance regression
- Stores benchmark history
- Comments on commits with regression alerts

### 4. Code Quality Workflow (`.github/workflows/quality.yml`)

Runs on push and pull requests.

**Jobs:**

#### Documentation
- Checks for undocumented exports
- Generates markdown documentation
- Uses `gomarkdoc`

#### Complexity Analysis
- Runs `gocyclo` for cyclomatic complexity
- Reports functions with complexity > 15
- Generates average complexity metrics

#### Dead Code Detection
- Uses `deadcode` to find unused code
- Helps maintain clean codebase

#### Code Duplication
- Runs `dupl` with 50-line threshold
- Identifies duplicate code blocks

#### Coverage Report
- Generates per-package coverage breakdown
- Creates markdown table of results

### 5. Release Workflow (`.github/workflows/release.yml`)

Triggered on version tags (`v*.*.*`).

**Features:**
- Builds binaries for all platforms
- Generates SHA256 checksums
- Creates SBOM for release
- Auto-generates changelog from commits
- Creates GitHub release with all artifacts

### 6. Dependabot Configuration (`.github/dependabot.yml`)

**Go Modules:**
- Weekly updates on Mondays at 8:00 AM
- Groups minor and patch updates
- Max 10 open PRs

**GitHub Actions:**
- Weekly updates on Mondays
- Max 5 open PRs
- Auto-labeled

## Pre-commit Hooks

Install pre-commit hooks for local validation:

```bash
# Install pre-commit (requires Python)
pip install pre-commit

# Setup hooks
make pre-commit

# Run manually
pre-commit run --all-files
```

**Hooks included:**
- Go formatting
- Go vet
- Go imports
- Unit tests (short mode)
- Trailing whitespace
- YAML validation
- Large file detection
- Secret detection
- Markdown linting
- Conventional commit messages

## Local CI Simulation

Run the same checks locally that run in CI:

```bash
# Run all CI checks
make ci

# Run individual checks
make ci-validate    # Formatting, vet, mod verification
make ci-test        # Tests with race detector
make ci-lint        # golangci-lint
make ci-build       # Build for current platform
make ci-security    # Vulnerability scanning

# Additional quality checks
make quality        # Dead code, duplication, complexity
make audit          # Security audit
make sbom           # Generate SBOM
```

## Code Coverage

**Enforcement:**
- Minimum threshold: 40% total coverage
- CI fails if coverage drops below threshold
- Coverage reports uploaded to Codecov
- Per-package coverage breakdown available

**View coverage:**
```bash
make test-coverage
# Opens coverage.html in browser
```

## Security Scanning

**Multiple layers:**

1. **govulncheck** - Go-specific vulnerability database
2. **gosec** - Static security analysis
3. **Trivy** - Comprehensive filesystem scanner
4. **Nancy** - Dependency vulnerability checking
5. **License compliance** - Ensures approved licenses only

**SBOM (Software Bill of Materials):**
- Generated on every security scan
- Includes all dependencies
- CycloneDX format (industry standard)
- Available as build artifact

## Build Artifacts

**Retention periods:**
- Build binaries: 7 days
- Coverage reports: 30 days
- Quality reports: 30 days
- Benchmark results: 30 days
- SBOM: 90 days
- License reports: 30 days

## Performance Optimization

**Caching strategy:**
- Go module cache
- golangci-lint cache
- Build cache
- govulncheck analysis cache

**Parallelization:**
- Matrix builds for multiple Go versions
- Matrix builds for multiple platforms
- Independent job execution

## Best Practices

### For Contributors

1. **Before committing:**
   ```bash
   make fmt           # Format code
   make test-race     # Run tests
   make lint          # Check linting
   ```

2. **Use pre-commit hooks:**
   ```bash
   make pre-commit    # Setup once
   ```

3. **Check security:**
   ```bash
   make vulncheck     # Before adding dependencies
   ```

4. **Verify locally:**
   ```bash
   make ci            # Run full CI suite
   ```

### Commit Messages

Follow Conventional Commits format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `perf`: Performance improvement
- `test`: Tests
- `build`: Build system
- `ci`: CI configuration
- `chore`: Maintenance

**Examples:**
```
feat: add support for new exchange API
fix: resolve race condition in order processing
docs: update installation instructions
ci: add benchmark tracking workflow
```

## Release Process

1. **Create version tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **CI automatically:**
   - Runs full test suite
   - Builds multi-platform binaries
   - Generates checksums
   - Creates SBOM
   - Generates changelog
   - Creates GitHub release

3. **Release includes:**
   - Binaries for Linux, macOS, Windows (AMD64/ARM64)
   - SHA256 checksums
   - SBOM (CycloneDX)
   - Auto-generated changelog

## Troubleshooting

### Failed CI Checks

**Validation failed:**
- Run `make fmt` to fix formatting
- Run `make tidy` to fix go.mod/go.sum
- Run `go vet ./...` to see issues

**Tests failed:**
- Run `make test` locally
- Check for race conditions with `make test-race`
- Review test logs in GitHub Actions

**Linting failed:**
- Run `make lint` locally
- Fix issues reported by golangci-lint
- Check `.golangci.yml` for configuration

**Security failed:**
- Run `make vulncheck`
- Update vulnerable dependencies
- Check if vulnerability is a false positive

### Coverage Below Threshold

If coverage drops below 40%:
- Add tests for new code
- Improve existing test coverage
- Check coverage report: `make test-coverage`
- Review per-package coverage in CI artifacts

## Monitoring

**GitHub Actions:**
- View workflow runs: Actions tab
- Check logs for failures
- Download artifacts

**Codecov:**
- View coverage trends: https://codecov.io/gh/guyghost/constantine
- Track coverage over time
- Per-file coverage reports

**Security:**
- GitHub Security tab for vulnerability alerts
- Dependabot PRs for dependency updates
- SBOM artifacts for supply chain visibility

## Maintenance

**Weekly:**
- Review Dependabot PRs
- Check security alerts
- Monitor benchmark trends

**Monthly:**
- Review code quality reports
- Update development tools
- Check for outdated workflows

**Quarterly:**
- Update Go versions in CI
- Review and update linter configuration
- Audit dependencies manually

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [Go Vulnerability Database](https://pkg.go.dev/vuln/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [CycloneDX SBOM](https://cyclonedx.org/)
