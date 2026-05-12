---
name: ssdlc-phase2-design
description: Phase 2 - Secure design, threat modeling, and architecture analysis
---

# Phase 2: Secure Design & Threat Modeling

Design secure architecture, apply patterns, and define implementation approach.

## Objectives

- Create secure architecture
- Apply security patterns
- Refine threat model
- Minimize attack surface
- Define trust boundaries clearly

## Phase Activities

### 2.1 Secure Architecture Design

**Architecture Patterns by Component**

| Component | Secure Pattern | Key Controls |
|-----------|----------------|--------------|
| API Gateway | Zero-trust proxy | mTLS, JWT validation, rate limiting |
| Auth Service | Centralized identity | MFA, secure sessions, password hashing |
| Microservices | Defense in depth | Network policies, service mesh, secrets |
| Database | Encryption always on | TLS in transit, AES-256 at rest |
| Caching | Memory-safe | No secrets in cache, TTL limits |
| File Storage | Access controls + encryption | IAM, bucket policies, versioning |

**Zero-Trust Architecture**

```
User → Device → Identity → Policy → Resource → Data
  │       │        │         │          │       │
  └───────┴────────┴─────────┴──────────┴───────┘
                    │
              Every request verified
              Every connection encrypted
              Every access authorized
```

### 2.2 Security Pattern Library

**Authentication Patterns**

```go
// Pattern: Multi-Factor Authentication
type MFAConfig struct {
    Required    bool
    Methods     []MFAMethod // TOTP, WebAuthn, SMS (deprecated)
    BypassCodes int        // Emergency backup codes
    RememberDays int      // Device trust duration
}

// Pattern: Secure Password Hashing
// Use Argon2id with:
// - Memory: 64MB
// - Iterations: 3
// - Parallelism: 4
// - Salt: 16+ bytes random

// Pattern: Session Token Generation
// - 32+ bytes cryptographic random
// - HttpOnly + Secure + SameSite=Strict
// - 30-minute idle timeout, 8-hour absolute max
```

**Authorization Patterns**

```go
// Pattern: Permission-Based Access Control
type Permission string
const (
    PermRead   Permission = "read"
    PermWrite  Permission = "write"
    PermDelete Permission = "delete"
    PermAdmin  Permission = "admin"
)

// Pattern: Policy Enforcement Point
func Authorize(ctx context.Context, subject, action, resource) error {
    policy := policies.Get(resource.Type)
    if !policy.Evaluate(ctx, subject, action) {
        return ErrForbidden
    }
    return nil
}

// Pattern: Deny by Default
var defaultDeny = func() bool { return false }
```

**Input Validation Patterns**

```go
// Pattern: Allowlist Validation
func ValidateEmail(email string) error {
    if !emailRegex.MatchString(email) {
        return ErrInvalidEmail
    }
    if len(email) > 254 {
        return ErrInvalidEmail
    }
    // Additional checks for DNS MX records in production
    return nil
}

// Pattern: Type Safety
type UserID int64
func (id UserID) Validate() error {
    if id <= 0 {
        return ErrInvalidID
    }
    return nil
}

// Pattern: Parameterized Queries (always)
db.Query("SELECT * FROM users WHERE id = ?", id) // Safe
db.Query("SELECT * FROM users WHERE id = " + id) // SQLi!
```

**Cryptography Patterns**

```go
// Pattern: Approved Algorithms
var ApprovedAlgorithms = map[string]bool{
    // Encryption: AES-256-GCM, ChaCha20-Poly1305
    // Hashing: SHA-256, SHA-3
    // Signatures: ECDSA, Ed25519, RSA-PSS
    // Key Exchange: X25519, ECDHE
}

// Pattern: Secure Random
func GenerateToken() ([]byte, error) {
    b := make([]byte, 32)
    _, err := crypto_rand.Read(b)
    return b, err
}

// Pattern: Secure Delete (overwrite before delete for sensitive data)
func SecureDelete(data []byte) {
    crypto_rand.Read(data) // overwrite with random
    for i := range data {
        data[i] = 0
    }
}
```

### 2.3 Attack Surface Analysis

**Reduce Attack Surface**

| Surface | Strategy | Implementation |
|---------|----------|----------------|
| APIs | REST > GraphQL, minimal fields | Versioning, field restrictions |
| Authentication | Progressive delays, CAPTCHA | Rate limiting, account lockout |
| Features | Disable unused, modular removal | Feature flags |
| Protocols | TLS 1.3 only | Disable legacy SSL/TLS |
| Permissions | Drop privileges immediately | Principle of least privilege |

**Entry Point Hardening**

| Entry Point | Security Controls |
|-------------|------------------|
| REST API | AuthN/Z, rate limiting, input validation |
| WebSocket | Auth, subprotocol validation, message size |
| GraphQL | Depth limiting, query complexity analysis |
| gRPC | mTLS, header validation |
| Webhooks | Signature verification, retry limits |

### 2.4 Trust Boundary Implementation

**Network Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                        Edge Layer                            │
│  WAF → CDN → DDoS Protection → SSL Termination               │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      DMZ Layer                               │
│  Load Balancer → API Gateway → Rate Limiter                 │
│                       │                                      │
│                 Auth Service                                 │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  Service Mesh (mTLS) → Microservices → Message Queue        │
│         │                                    │              │
│    AuthN/Z                              Database Proxy      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                       Data Layer                             │
│  Database (encrypted) → Cache (encrypted) → Object Storage  │
│  Secrets Vault → Key Management                              │
└─────────────────────────────────────────────────────────────┘
```

**Trust Boundary Controls**

| Boundary | Controls | Verification |
|----------|----------|--------------|
| Edge → DMZ | WAF, TLS, rate limiting | TLS audit, WAF rules test |
| DMZ → App | Auth service, API gateway | Auth flow test, JWT validation |
| App → Data | Encryption, IAM, audit | Access review, query audit |
| Admin → Prod | Bastion, PAM, MFA, session record | Access log review |

### 2.5 Threat Model Refinement

**Updated Threat Analysis**

Review Phase 1 threat model against proposed design:

| Threat | Design Impact | Required Controls |
|--------|---------------|-------------------|
| SQL Injection | Database queries | Parameterized queries, WAF |
| XSS | User input display | Output encoding, CSP |
| CSRF | State-changing operations | CSRF tokens, SameSite |
| SSRF | Internal service calls | URL validation, blocklist |
| IDOR | Resource access | Authorization checks |
| Authentication Bypass | All protected resources | MFA, secure sessions |

**Attack Path Analysis**

```
Public Entry → [WAF] → API Gateway → [Auth] → Service A → Database
                                ↓
                           [IDOR?]
                                ↓
                         Access User B's Data
```

### 2.6 Secure Defaults

| Component | Secure Default | Why |
|-----------|----------------|-----|
| Auth | MFA for privileged, password policy | Credential stuffing prevention |
| Sessions | 30-min timeout, 8h max | Session hijacking mitigation |
| Passwords | Min 12, complexity, breach check | Weak password prevention |
| APIs | 100 req/min, 5 failures = block | DoS/abuse prevention |
| Errors | Generic + log details internally | Information disclosure prevention |
| Encryption | TLS 1.3, AES-256 | Forward secrecy, data protection |
| Logging | Structured, no secrets, PII masked | Security monitoring |
| Config | Secure defaults, no hardcoded secrets | Configuration security |

## GraphX Integration

**Architecture Queries**

```bash
# Get all entry points
GraphX: GetAPIs()

# Get trust boundaries
GraphX: GetThreatModel().TrustBoundaries

# Find high-risk functions
GraphX: Query risk_score > 7.0

# Get cross-references for auth flows
GraphX: GetNodesByType("function") | filter(auth_sensitive)

# Blast radius of auth compromise
GraphX: GetBlastRadius("auth_service_id")
```

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Secure Architecture Spec | Architecture with security controls |
| Security Pattern Library | Approved patterns per use case |
| Threat Model Update | Design-specific threats and controls |
| Attack Surface Report | Entry points, reducing strategies |
| Trust Boundary Map | Implementation of boundaries |
| Design Review Report | Findings and remediation |

## Exit Criteria

- [ ] Secure architecture approved
- [ ] All threats mapped to controls
- [ ] Attack surface minimized
- [ ] Trust boundaries defined
- [ ] Security patterns documented
- [ ] GraphX updated with architecture