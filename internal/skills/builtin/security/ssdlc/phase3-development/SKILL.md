---
name: ssdlc-phase3-development
description: Phase 3 - Secure development, SAST integration, and secure coding practices
---

# Phase 3: Secure Development

Implement secure code, integrate security tools, and establish development security practices.

## Objectives

- Write secure code following standards
- Integrate SAST in CI/CD
- Detect secrets and vulnerabilities early
- Establish security code review process
- Track security in GraphX

## Phase Activities

### 3.1 Secure Coding Standards

**Language-Specific Standards**

**Go (Golang)**

| Vulnerability | Insecure | Secure |
|--------------|----------|--------|
| SQL Injection | `fmt.Sprintf("SELECT * FROM t WHERE id=%s", id)` | `db.Query("SELECT * FROM t WHERE id=?", id)` |
| Command Injection | `exec.Command("sh", "-c", input)` | `exec.Command("ls", input)` (no shell) |
| Path Traversal | `ioutil.ReadFile(userPath)` | `path.Clean(base + "/" + userPath)` |
| Hardcoded Secret | `key := "abc123"` | `key := os.Getenv("API_KEY")` |
| Weak Crypto | `md5.Sum(data)` | `sha256.Sum256(data)` |
| Constant-time | `a == b` | `subtle.ConstantTimeCompare(a, b)` |

**Python**

| Vulnerability | Insecure | Secure |
|--------------|----------|--------|
| Code Injection | `eval(user_input)` | Never eval user input |
| YAML Deserialization | `yaml.load(data)` | `yaml.safe_load(data)` |
| Path Traversal | `open(base + user_path)` | `pathlib.Path(base).joinpath(user_path)` |
| SQL Injection | `cursor.execute(f"SELECT * FROM t WHERE id={id}")` | `cursor.execute("SELECT * FROM t WHERE id=?", (id,))` |
| Password Hashing | `hashlib.md5(pwd)` | `bcrypt.hashpw(pwd, bcrypt.gensalt())` |

**JavaScript/TypeScript**

| Vulnerability | Insecure | Secure |
|--------------|----------|--------|
| Prototype Pollution | `Object.assign({}, req.body)` | `structuredClone(req.body)` or deep clone |
| SQL Injection | `` `SELECT * FROM t WHERE id=${id}` `` | `db.query('SELECT * FROM t WHERE id=$1', [id])` |
| XSS | `element.innerHTML = userInput` | `element.textContent = userInput` |
| Path Traversal | `fs.readFile(userPath)` | `path.isInside(base, userPath)` |
| Weak Random | `Math.random()` | `crypto.randomBytes()` |

### 3.2 SAST Integration

**SAST Configuration**

```yaml
# .github/workflows/sast.yml
name: Static Analysis
on: [push, pull_request]

jobs:
  semgrep:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: returntocorp/semgrep-action@v1
        with:
          config: |
            p/security-audit
            p/secrets
            p/ci
          autocompile_pr_limits: MEDIUM

  gosec:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: securego/gosec-action@v2
        with:
          args: '-severity HIGH -confidence HIGH -no-fail'
```

**Critical Rules**

| Rule | Language | Severity | Blocks Build |
|------|----------|----------|--------------|
| sql-injection | Go, Python, JS | CRITICAL | Yes |
| command-injection | All | CRITICAL | Yes |
| hardcoded-credential | All | CRITICAL | Yes |
| weak-crypto | All | HIGH | Yes |
| path-traversal | All | HIGH | Yes |
| xss-through-html | JS, Python | HIGH | Yes |
| xx-insecure-deserialization | All | CRITICAL | Yes |
| cors-misconfiguration | JS | MEDIUM | Yes (if critical) |

### 3.3 Dependency Security

**Dependency Scanning**

```yaml
# GitHub Actions - Dependency Review
- name: Dependency Review
  uses: actions/dependency-review-action@v4
  with:
    fail-on-severity: critical
    
# npm audit in CI
- run: npm audit --audit-level=high

# pip-audit for Python
- run: pip-audit --severity=high --format=json > results.json
```

**Approved Libraries**

| Category | Approved | Blocked |
|----------|----------|---------|
| Crypto | `crypto`, `golang.org/x/crypto` | Custom crypto, MD5 for passwords |
| Auth | `golang-jwt/jwt`, `pquerna/otp` | DIY auth |
| DB | `database/sql`, SQL drivers | String concat queries |
| HTTP | `net/http`, `fasthttp` | No SSL verify false |
| Serialization | `encoding/json`, `protobuf` | `pickle`, `yaml.load` |

### 3.4 Secrets Detection

**Pre-commit Hook**

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/trufflesecurity/trufflehog
    rev: v3.80.0
    hooks:
      - id: trufflehog
        args: ['--no-update']

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: no-commit-to-branch
        args: ['--branch', 'main', '--pattern', '^release/']
```

**CI Secrets Scan**

```yaml
- name: TruffleHog
  uses: trufflesecurity/trufflehog@main
  with:
    path: ./
    base: ${{ github.event.repository.default_branch }}
    head: HEAD
    extra_config: '--no-update'
```

### 3.5 Security Code Review

**Review Checklist**

| Category | Check | Priority |
|----------|-------|----------|
| Input | All input validated against allowlist? | Required |
| Input | Type coercion prevented? | Required |
| Input | Length limits enforced? | Required |
| Auth | MFA required for sensitive ops? | Required |
| Auth | Passwords hashed with Argon2id/bcrypt? | Required |
| Auth | Sessions timeout properly? | Required |
| Auth | Authorization checked per-resource? | Required |
| Crypto | Approved algorithms only? | Required |
| Crypto | Secrets from Vault/env, not hardcoded? | Required |
| Output | Proper encoding/escaping? | Required |
| Output | No sensitive data in errors? | Required |
| Logging | No secrets/PII logged? | Required |
| Dependencies | All dependencies approved? | Required |

**Review Focus Areas**

1. Authentication flows
2. Authorization checks
3. Input validation
4. Output encoding
5. Cryptography usage
6. Error handling
7. Logging practices

### 3.6 Infrastructure as Code Validation

**Terraform Security**

```yaml
# tfsec configuration
---
checks:
  - code: AWS012
    description: S3 bucket should have versioning enabled
  - code: EXP001
    description: Terraform external providers should include version constraint
  - code: GEN002
    description: AWS provider should enforce SSL

exclude:
  - aws-s3-enable-bucket-encryption
```

**Dockerfile Security**

```dockerfile
# Multi-stage build
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s"

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/app /app
USER 1001
ENTRYPOINT ["/app"]

# Security: No root, minimal image, no shell
```

## GraphX Integration

**Track Security Code**

```bash
# Get functions added since last review
GraphX: Query created_at > last_review_date

# Search for security-relevant code
GraphX: SearchNodes("auth", "password", "token", "secret")

# Get high-risk functions
GraphX: Query risk_score > 5.0 type=function

# Get blast radius of changes
GraphX: GetBlastRadius("new_function_id")
```

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Secure Coding Standards | Language-specific guidelines |
| SAST Configuration | CI/CD integration config |
| Dependency Approval | List of approved/blocked packages |
| Secrets Policy | Detection and prevention config |
| Code Review Checklist | Pre-merge security checks |
| GraphX Tracking | Security code tracking enabled |

## Exit Criteria

- [ ] SAST integrated, blocks on HIGH/CRITICAL
- [ ] No hardcoded secrets in codebase
- [ ] All dependencies from approved list
- [ ] Code review checklist followed
- [ ] Security tests added
- [ ] GraphX tracks security code