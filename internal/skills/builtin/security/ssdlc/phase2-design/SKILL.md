---
name: phase2-design
description: Phase 2 - Secure design, architecture threat modeling, STRIDE analysis, and attack paths
---

# Phase 2: Design & Architecture

Model security threats, analyze architecture, trace attack paths, and verify design isolation before writing code.

## Objectives

- Generate deterministic architecture and C4 diagrams
- Model STRIDE threats against components
- Map attacker kill-chains and attack paths
- Review trust boundaries and multi-tenancy isolation
- Identify and flag insecure design patterns

## Phase Activities

### 2.1 Threat Modeling Pipeline

When reviewing or modeling system architecture, always execute the following pipeline in sequence:

```
1. analyze_architecture  →  Extract components, boundaries, and dependencies
       ↓
2. generate_c4_model    →  Build C4 L1/L2/L3 diagrams
       ↓
3. generate_threat_model →  Enumerate STRIDE threats per component
       ↓
4. generate_attack_paths →  Map attack kill-chains and entry points
```

### 2.2 STRIDE Threat Enumeration

Analyze each system element against STRIDE threats:

| Element Type | STRIDE Focus | Threat Scenario | Key Mitigations |
|--------------|--------------|-----------------|-----------------|
| **Services** | Spoofing / Tampering | Attacker impersonates service; manipulates binaries | mTLS, secure boot, signed packages, service mesh |
| **APIs** | Repudiation / Info Disclosure | Attacker makes anonymous actions; leaks data | Structured JWT validation, audit trails, Rate limits |
| **Data Flows** | Denial of Service / Elevation | Attacker exhausts network; elevates roles | Network isolation, input checks, RBAC validations |

### 2.3 Attack Kill-Chains

Build and analyze multi-hop attack chains:

```
[Initial Access] ──▶ [Lateral Movement] ──▶ [Privilege Escalation] ──▶ [Data Exfiltration]
   Public API           Internal Service          Admin Role            S3 Bucket/DB
   Vulnerability        mTLS Bypass/SSRF          BFLA/IDOR             Sensitive PII
```

**Common Precondition Rules:**
- **Initial Access**: Vulnerable edge service, missing authentication, exposed credential.
- **Lateral Movement**: Cleartext database credentials in memory/config, unsecured internal REST API.
- **Privilege Escalation**: Missing authorization checks, admin endpoints accessible by standard users.

### 2.4 Trust Boundaries & Multi-Tenancy

**Isolation Design Patterns:**
*   **Trust Boundaries**: Never cross trust zones without explicit authentication and input validation. External networks (untrusted) -> Edge/DMZ (partially trusted) -> Application/Data (trusted).
*   **Multi-Tenancy Isolation**: Verify database schemas or storage layers enforce tenant key filtering (`tenant_id`). Ensure cross-tenant data leaks are prevented.
*   **Cryptographic Architecture**: Ensure strong cryptographic algorithms are used (e.g. Argon2id for password storage, AES-256-GCM for data protection). Avoid custom crypto.

### 2.5 Insecure Design Patterns Checklist

Audit architecture against common anti-patterns:

- **Secrets in Config**: Hardcoded database URLs, passwords, or API keys in configuration templates/files.
- **Database exposure**: Exposing database ports directly to the internet or edge zone.
- **Missing Rate Limits**: Public APIs without rate limit or request size validation (susceptible to DoS).
- **Direct Database URLs in ENV**: Standard practice is fine, but storing credentials in plaintext without vault injection should be avoided.

## GraphX Integration

**Queries to Verify Design & Architecture**

```bash
# Review C4 model components
graphx architecture review

# Extract attack paths starting from public entry points
graphx attack-paths

# Query trust boundary violations
GraphX: GetTrustBoundaries()
```

## Deliverables

- **Threat Model Report**: Full STRIDE analysis.
- **C4 Models**: L1/L2/L3 architecture diagrams.
- **Attack Path Maps**: Detailed kill-chains.
- **Mitigation Matrix**: Mapping threats to security controls.

## Exit Criteria

- [ ] C4 models generated and validated.
- [ ] STRIDE threats mapped.
- [ ] Attack kill-chains analyzed.
- [ ] No insecure design patterns detected.
- [ ] Threat model approved.