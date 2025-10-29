# Security Policy

## Supported Versions

We release security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of Constantine seriously. If you believe you have found a security vulnerability, please report it to us responsibly.

### Reporting Process

1. **DO NOT** create a public GitHub issue for security vulnerabilities
2. **DO NOT** disclose the vulnerability publicly until it has been addressed

Instead, please use one of these methods:

#### Preferred: GitHub Security Advisories
1. Go to https://github.com/guyghost/constantine/security/advisories/new
2. Fill out the security advisory form
3. Submit the advisory

#### Alternative: Email
Send details to the maintainers (update with actual email)

### What to Include

Please provide as much information as possible:

- **Description** - A clear description of the vulnerability
- **Impact** - What an attacker could potentially do
- **Reproduction Steps** - Detailed steps to reproduce the issue
- **Affected Versions** - Which versions are vulnerable
- **Suggested Fix** - If you have ideas on how to fix it (optional)
- **Your Contact Info** - So we can follow up with you

### Response Timeline

- **Acknowledgment**: Within 48 hours of receiving your report
- **Initial Assessment**: Within 7 days
- **Fix Development**: Depends on complexity, but we aim for 30 days
- **Public Disclosure**: Coordinated with you after fix is released

## Security Best Practices

### For Users

1. **Never commit secrets**
   - Add `.env` to `.gitignore`
   - Use environment variables or secret managers
   - Enable `LOG_SENSITIVE_DATA=false`

2. **Use dedicated wallets**
   - Don't use wallets with large amounts
   - Use separate wallets for testing

3. **Keep dependencies updated**
   - Monitor Dependabot PRs
   - Run `make vulncheck` regularly
   - Update Go version when new security patches are released

4. **Review permissions**
   - Grant minimal API permissions required
   - Use read-only keys for testing
   - Rotate API keys periodically

5. **Secure deployment**
   - Use HTTPS for telemetry endpoints
   - Run in isolated environments
   - Use firewall rules to restrict access
   - Enable authentication for monitoring endpoints

### For Developers

1. **Code Review**
   - All PRs require review before merge
   - Pay special attention to authentication/authorization
   - Review dependency changes carefully

2. **Security Scanning**
   - CI automatically runs multiple security scanners
   - Fix vulnerabilities before merging
   - Keep security tools updated

3. **Dependency Management**
   - Pin dependency versions in go.mod
   - Review dependency licenses
   - Use `make audit` before adding dependencies
   - Monitor CVE databases

4. **Secrets Management**
   - Never hardcode credentials
   - Use environment variables
   - Consider using HashiCorp Vault or similar
   - Rotate secrets regularly

5. **Input Validation**
   - Validate all external input
   - Sanitize data from exchanges
   - Use type-safe parsing
   - Implement rate limiting

## Security Tooling

Constantine uses multiple security tools:

### Automated Scanning

1. **govulncheck** - Go vulnerability database scanner
   ```bash
   make vulncheck
   ```

2. **gosec** - Go security checker
   - Runs automatically in CI
   - Checks for common security issues

3. **Trivy** - Comprehensive vulnerability scanner
   - Scans dependencies and file system
   - Runs on every PR and daily

4. **Nancy** - Sonatype dependency checker
   - Additional vulnerability database
   - Runs in CI security workflow

### Manual Checks

1. **License Compliance**
   ```bash
   make license-check
   ```

2. **SBOM Generation**
   ```bash
   make sbom
   ```

3. **Comprehensive Audit**
   ```bash
   make audit
   ```

## Known Security Considerations

### Exchange API Keys

- **Hyperliquid**: Private key gives full account access
- **Coinbase**: API keys should have minimal permissions
- **dYdX**: Mnemonic provides complete wallet control

### Trading Risks

⚠️ **This bot is for educational/research purposes**

- Automated trading involves financial risk
- Test thoroughly with small amounts first
- Use paper trading mode when available
- Monitor bot behavior continuously
- Have circuit breakers and limits in place

### Network Security

- WebSocket connections are not end-to-end encrypted by default
- Consider using VPN or TLS termination proxy
- Validate SSL certificates
- Implement connection timeout and retry logic

## Disclosure Policy

When we receive a security report:

1. **Acknowledgment** - We confirm receipt
2. **Investigation** - We investigate and validate
3. **Fix Development** - We develop and test a fix
4. **Release** - We release a patched version
5. **Advisory** - We publish a security advisory
6. **Attribution** - We credit the reporter (if desired)

### Public Disclosure

- We coordinate public disclosure with the reporter
- We aim to patch before public disclosure
- We provide migration guides if breaking changes are needed
- We notify users via GitHub Security Advisories

## Security Hall of Fame

We appreciate security researchers who help us keep Constantine secure. Reporters who follow responsible disclosure will be:

- Credited in release notes (if desired)
- Listed here (if desired)
- Thanked publicly

<!-- Security researchers will be listed here -->

## Contact

For security issues: Use GitHub Security Advisories (preferred) or email maintainers

For general questions: Use GitHub Discussions

## Updates to This Policy

This security policy may be updated from time to time. Check back regularly for changes.

Last updated: 2025-10-29
