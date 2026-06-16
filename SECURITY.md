# Security Policy for DuckOps

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest| :x:                |

## Reporting a Vulnerability

DuckOps is a security testing tool. We take vulnerabilities in our own code seriously.

**Do not** open public GitHub issues for security vulnerabilities.

### How to report

1. **Email**: Send details to `security@duckops.dev`
2. **PGP**: Encrypt using our public key (available at `https://duckops.dev/security/pgp-key.asc`)
3. **Private advisory**: Use the [GitHub Security Advisory](https://github.com/SecDuckOps/duckops/security/advisories/new)

### What to include

- Affected version(s)
- Type of vulnerability (RCE, XSS, privilege escalation, etc.)
- Steps to reproduce or PoC
- Impact assessment
- Suggested fix (if available)

### Response timeline

- **24h**: Acknowledgement of receipt
- **7 days**: Initial triage and severity assessment
- **30 days**: Fix released for critical/high severity
- **90 days**: Fix released for medium/low severity

### Disclosure policy

We follow coordinated disclosure:

1. Reporter submits vulnerability privately
2. We validate and triage
3. Fix developed and tested
4. Fix released with advisory
5. Public disclosure after 30 days (or sooner if a fix is released)

## Security Features

DuckOps implements the following security measures:

- **Non-root containers**: Default runtime user is `appuser` (UID 1000)
- **Minimal base images**: `scratch`/`alpine` for production builds
- **SBOM generation**: SPDX-format SBOM included with every release
- **Provenance attestation**: Build provenance via GitHub Actions
- **SAST scanning**: gosec + CodeQL + semgrep in CI
- **SCA scanning**: Trivy filesystem + container scanning
- **Secrets scanning**: Gitleaks on every commit and PR
- **Dependency review**: Automated review on every PR
- **Signed releases**: Checksums signed with cosign
