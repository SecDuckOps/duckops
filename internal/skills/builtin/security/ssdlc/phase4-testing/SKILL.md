---
name: ssdlc-phase4-testing
description: Phase 4 - Security testing, DAST, penetration testing, and vulnerability validation
---

# Phase 4: Security Testing

Validate security through automated and manual testing approaches.

## Objectives

- Execute comprehensive security tests
- Identify and validate vulnerabilities
- Measure blast radius and impact
- Automate security tests
- Remediate findings

## Phase Activities

### 4.1 Static Application Security Testing (SAST)

**SAST Verification**

```bash
# Semgrep
semgrep --config=security --json --output=semgrep.json .

# Gosec
gosec ./... -fmt json -out=gosec.json

# Bandit (Python)
bandit -r ./src -f json -o bandit.json
```

**SAST Validation Criteria**

| Tool | CRITICAL | HIGH | MEDIUM |
|------|----------|------|--------|
| Semgrep | BLOCK | BLOCK | WARN |
| Gosec | BLOCK | BLOCK | WARN |
| Bandit | BLOCK | BLOCK | WARN |

### 4.2 Dynamic Application Security Testing (DAST)

**OWASP ZAP Configuration**

```yaml
# zap-baseline.yaml
zap:
  baseline:
    target: https://staging.example.com
    spider: true
    active_scan:
      enabled: true
      policy: "API Security High"
  
  api_scan:
    target: https://staging.example.com/openapi.yaml
    auth:
      method: bearer
      header: Authorization
      value: "Bearer {{token}}"
```

**DAST in CI/CD**

```yaml
- name: OWASP ZAP Scan
  uses: zaproxy/action-baseline@v0.9.0
  with:
    target: 'https://staging.example.com'
    rules: 'zap-rules.conf'
    fail_mode: medium
    cmd_options: '-T 60'
```

### 4.3 Penetration Testing

**OWASP Testing Guide Categories**

| Category | Tests | Tools |
|----------|-------|-------|
| Information Gathering | robots.txt, source maps, tech fingerprint | nmap, ffuf, whatweb |
| Configuration | TLS config, default creds, HTTP methods | testssl, nikto |
| Identity Management | Account enum, weak registration | Burp Intruder |
| Authentication | Brute force, credential stuffing | hydra, ffuf |
| Authorization | IDOR, horizontal/vertical escalation | Burp, authz analyzer |
| Session Management | Session fixation, token exposure | Burp Sequencer |
| Input Validation | XSS, SQLi, command injection | sqlmap, xsstrike, commix |
| Business Logic | Workflow bypass, race conditions | manual + tools |
| Cryptography | Weak crypto, key management | testssl, certinfo |
| Client-Side | DOM XSS, CORS bypass | Burp, browser dev tools |

**Testing Commands**

```bash
# Authentication Testing
hydra -l admin -P rockyou.txt target.com http-post-form "/login:user=^USER^&pass=^PASS^:F=Invalid"

# SQL Injection
sqlmap -r request.txt --batch --level=5 --risk=3

# XSS
xsstrike -u https://target.com/?q=test

# Command Injection
commix -r request.txt --batch

# SSRF
ffuf -w payloads.txt -u https://target.com/?url=FUZZ
```

### 4.4 API Security Testing

**REST API Testing**

```bash
# Enumerate endpoints
ffuf -w /usr/share/wordlists/api.txt -u https://api.target.com/v1/FUZZ

# Authentication testing
curl -X POST /api/auth/login -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123"}'

# JWT Testing
# 1. none algorithm
echo -n '{"alg":"none","typ":"JWT"}' | base64
# 2. Weak secret brute force
jwttool -t https://api.target.com -m "eyJ..." -C -L wordlist.txt

# Authorization/IDOR
curl -H "Authorization: Bearer $TOKEN" /api/users/1
curl -H "Authorization: Bearer $TOKEN" /api/users/2  # Test other user

# Rate limiting
for i in {1..100}; do curl -s -o /dev/null; done

# Mass assignment
curl -X POST /api/user -H "Content-Type: application/json" \
  -d '{"name":"test","role":"admin","is_admin":true}'
```

**GraphQL Testing**

```bash
# Introspection
query { __schema { types { name fields { name type { name } } } } }

# Depth limiting bypass
query { users { id admin { password secretData { ssn } } } }

# Injection
query { user(id: "1 OR 1=1") { name } }
query { users(where: "_repeat=print('x')") { id } }
```

### 4.5 Vulnerability Validation

**Exploitability Assessment**

| Finding | Exploitability | Impact | Priority |
|---------|----------------|--------|----------|
| SQLi in login | High - direct DB access | CRITICAL | P1 |
| Stored XSS in comments | Medium - user interaction | HIGH | P2 |
| IDOR on profile update | High - direct resource access | HIGH | P2 |
| Information disclosure | Low - config dependent | MEDIUM | P3 |
| CSRF on settings | Medium - user interaction | HIGH | P2 |
| Weak TLS config | Medium - passive only | MEDIUM | P3 |

**Proof of Concept Requirements**

For each finding, document:
1. Steps to reproduce
2. Working exploit
3. Real impact demonstration
4. Business context
5. Remediation recommendation

### 4.6 Blast Radius Analysis

**GraphX Blast Radius**

```bash
# Get blast radius of vulnerability
GraphX: GetBlastRadius("vulnerable_function_id")

# Get affected authentication flows
GraphX: Query edge_type="auth_dep" AND target="vulnerable_function_id"

# Get impacted APIs
GraphX: Query type="api" AND blast_radius > 7.0

# Get trust boundary crossings
GraphX: GetThreatModel().TrustBoundaries
```

**Impact Categories**

| Category | Description | Example |
|----------|-------------|---------|
| Auth Impact | Authentication compromise | Password hashes exposed |
| Data Impact | Data breach or corruption | PII exfiltration |
| Service Impact | Service disruption | DDoS, resource exhaustion |
| Privilege Impact | Unauthorized access | Admin privilege escalation |

### 4.7 Security Test Automation

**Test Categories**

| Type | Coverage | Automation |
|------|----------|------------|
| Unit Security | Crypto, validation, auth logic | 100% (SAST) |
| Integration | API security, auth flows | 80% (DAST) |
| End-to-End | Browser security, sessions | 60% (automated) |
| Fuzzing | Input validation, parsing | 40% (targeted) |
| Manual | Business logic, complex attacks | 0% |

**Automated Test Examples**

```go
// Security Unit Tests
func TestPasswordHash(t *testing.T) {
    hash, err := HashPassword("Weak1")
    assert.NoError(t, err)
    assert.NotEmpty(t, hash)
    assert.True(t, bcrypt.CheckPasswordHash("Weak1", hash))
}

func TestSQLInjection(t *testing.T) {
    db := setupTestDB()
    defer db.Close()
    
    userID := "1 OR 1=1"
    _, err := GetUserByID(db, userID)
    assert.Error(t, err) // Should reject, not SQLi
}

func TestXSSPrevention(t *testing.T) {
    input := "<script>alert(1)</script>"
    output := EscapeHTML(input)
    assert.NotContains(t, output, "<script>")
}

// Integration Tests
func TestAuthFlow(t *testing.T) {
    // Test complete login → session → access flow
}

func TestAuthorizationBoundary(t *testing.T) {
    // Test user cannot access other user's data
}
```

## GraphX Integration

**Track Testing Results**

```bash
# Query high-risk components needing testing
GraphX: Query risk_score > 5.0

# Get entry points for testing
GraphX: GetAPIs()

# Track vulnerabilities
GraphX: GetNodesByType("finding")
```

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| DAST Results | Automated scan findings |
| Penetration Test Report | Manual testing findings |
| Vulnerability Risk Matrix | Prioritized findings |
| PoC Documentation | Working exploits for findings |
| Remediation Plan | Prioritized fixes with SLAs |
| Test Automation Suite | Continuous security tests |

## Exit Criteria

- [ ] SAST clean (no HIGH/CRITICAL)
- [ ] DAST completed
- [ ] Penetration test done
- [ ] All HIGH/CRITICAL resolved
- [ ] Blast radius analyzed
- [ ] GraphX threat model updated