---
name: ssdlc-phase1-requirements
description: Phase 1 - Security requirements gathering, risk assessment, and threat modeling before development
---

# Phase 1: Requirements & Threat Modeling

Define security requirements, identify assets, and model threats before development begins.

## Objectives

- Identify all security requirements
- Classify data and assets
- Assess and prioritize risks
- Define trust boundaries
- Create foundation for secure development

## Phase Activities

### 1.1 Security Requirements Gathering

**Functional Security Requirements**

| Category | Requirements | Examples |
|----------|--------------|----------|
| Authentication | MFA, password policy, session management | TOTP, WebAuthn, 12+ char passwords |
| Authorization | RBAC, ABAC, resource permissions | Least privilege, per-resource checks |
| Data Protection | Encryption, PII handling, retention | AES-256, GDPR compliance, 90-day retention |
| Audit Logging | What's logged, retention, tamper protection | Immutable logs, 1-year retention |
| Input Validation | Allowlist vs blocklist, sanitization | Type coercion prevention, length limits |
| Error Handling | Generic messages, no information disclosure | 500 on error, details server-side |

**Compliance Requirements**

| Framework | Key Requirements | Data Handling |
|-----------|-----------------|---------------|
| GDPR | Consent, right to erasure, data minimization | PII encryption, access controls |
| HIPAA | PHI protection, audit trails | Medical data encryption |
| PCI-DSS | Cardholder data, network segmentation | Tokenization, isolated networks |
| SOC 2 | Availability, confidentiality, integrity | Access logging, incident response |
| CCPA | Data disclosure, opt-out rights | Data inventory, deletion capability |

### 1.2 Asset Identification & Classification

**Data Assets**

| Classification | Examples | Protection Level |
|----------------|----------|------------------|
| Public | Marketing content, public docs | None |
| Internal | Internal docs, processes | Integrity only |
| Confidential | Customer data, financials | Encryption + access control |
| Restricted | Credentials, PII, health records | Encryption + MFA + audit |

**System Assets**

| Asset Type | Examples | Security Requirements |
|------------|----------|----------------------|
| Entry Points | APIs, web interfaces, mobile apps | Auth + validation + rate limiting |
| Trust Boundaries | Internet-facing, internal, privileged | Network segmentation, zero-trust |
| Dependencies | Third-party services, SDKs, cloud | Vetting, updates, isolation |
| Data Stores | Databases, caches, file storage | Encryption, access control, backup |

### 1.3 Threat Modeling (STRIDE)

**Data Flow Diagram Elements**

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   External  │────▶│   Process   │────▶│  Data Store │
│   Entity    │     │             │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │   Process   │
                    │             │
                    └─────────────┘
```

**STRIDE Threat Categories**

| Threat | Attack | Impact | Mitigations |
|--------|--------|--------|-------------|
| **Spoofing** | Credential theft, session hijacking | Impersonation | MFA, secure sessions, certificate pinning |
| **Tampering** | SQL injection, data poisoning | Data corruption | Parameterized queries, integrity checks |
| **Repudiation** | Deny actions, unsigned transactions | No accountability | Audit logs, digital signatures |
| **Information Disclosure** | Data leaks, side-channel | Confidentiality loss | Encryption, access controls, minimal error messages |
| **Denial of Service** | Resource exhaustion, service disruption | Availability loss | Rate limiting, redundancy, auto-scaling |
| **Elevation of Privilege** | RBAC bypass, parameter injection | Unauthorized access | Least privilege, input validation, sandboxing |

### 1.4 Risk Assessment

**Risk Matrix**

| Impact ↓ / Likelihood → | Low | Medium | High |
|-------------------------|-----|--------|------|
| **Critical** | High | Critical | Critical |
| **High** | Medium | High | Critical |
| **Medium** | Low | Medium | High |
| **Low** | Low | Low | Medium |

**Risk Treatment Options**

| Option | When to Use | Example |
|--------|-------------|---------|
| Mitigate | Can add controls | Add MFA to reduce auth risk |
| Transfer | Insurance or contracts | Cyber insurance for breach costs |
| Accept | Low risk, high cost to fix | Known legacy system limitation |
| Avoid | Can redesign process | Use managed service instead of self-hosted |

### 1.5 Trust Boundary Definition

**Common Trust Boundaries**

```
Internet ──────────────────────────── DMZ ──────────────────────────── Internal
   │                                    │                                    │
   ▼                                    ▼                                    ▼
Public APIs                         WAF/Load Balancer              Private Services
                                   (untrusted)                      (trusted)
                                         │
                                    Auth Service
                                         │
                                    Database
                                    (highly trusted)
```

**Boundary Mapping**

| Boundary | Between | Security Controls |
|----------|---------|-------------------|
| External → DMZ | Internet, third parties | WAF, rate limiting, input validation |
| DMZ → Internal | Web to application | Auth service, API gateway |
| Internal → Data | App to database | Encryption, access controls |
| Admin → Production | Privileged access | MFA, jump hosts, audit logging |

### 1.6 Abuse Case Identification

**Common Abuse Cases**

| Abuse Case | Attack Vector | Impact |
|------------|---------------|--------|
| Account takeover | Credential stuffing, phishing | Data breach, fraud |
| Privilege escalation | IDOR, parameter manipulation | Unauthorized access |
| Data exfiltration | SQL injection, misconfigured S3 | PII exposure |
| Service disruption | DDoS, resource exhaustion | Availability loss |
| Supply chain attack | Compromised dependency | Backdoor, persistence |

## GraphX Integration

**Query Architecture**

```bash
# Entry points discovery
GraphX: GetAPIs()

# Trust boundaries
GraphX: GetTrustBoundaries()

# Asset classification
GraphX: Query type="function" security_sensitivity>7

# Auth-related functions
GraphX: SearchNodes("auth", "login", "password")

# Data handling functions
GraphX: SearchNodes("pii", "personal", "credential")
```

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Security Requirements | Functional and compliance requirements |
| Data Classification | Asset inventory with classification |
| Threat Model | DFD with STRIDE threats |
| Risk Register | Prioritized risks with treatment |
| Trust Boundary Map | Security zones and controls |
| Abuse Case Catalog | Identified attack scenarios |

## Exit Criteria

- [ ] All assets identified and classified
- [ ] Threat model covers all entry points
- [ ] Risk register reviewed and approved
- [ ] Trust boundaries defined
- [ ] Security requirements traceable
- [ ] GraphX initialized with architecture