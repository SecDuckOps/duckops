---
name: phase1-requirements
description: Phase 1 - Security requirements gathering, OWASP ASVS mapping, data classification, and compliance mapping
---

# Phase 1: Requirements & Planning

Define security requirements, classify data, map compliance controls, and model abuse cases before development begins.

## Objectives

- Extract security requirements using OWASP ASVS levels (L1/L2/L3)
- Define abuse cases and misuse cases for each user story
- Map requirements to compliance controls (GDPR Art. 25/32, PCI-DSS 4.0, HIPAA §164.312, ISO 27001 Annex A)
- Establish security acceptance criteria before the sprint starts
- Identify and document data classification levels

## Phase Activities

### 1.1 OWASP ASVS Mapping

Extract security requirements according to the Application Security Verification Standard (ASVS) levels:

*   **ASVS Level 1 (Opportunistic)**: Good for low-risk applications, protecting against common, easy-to-find vulnerabilities. Fully testable using automated tools.
*   **ASVS Level 2 (Standard)**: Standard for applications handling sensitive transactions (PII, commerce). Validates key security controls (AuthN, AuthZ, input validation).
*   **ASVS Level 3 (Advanced)**: Required for high-risk applications (financial, health, critical infra). Requires design modularity, strong cryptography, and active defense.

| ASVS Section | Verification Level | Key Control Focus |
|--------------|--------------------|-------------------|
| V1: Architecture, Design | L1/L2/L3 | Trust boundary validation, threat modeling |
| V2: Authentication | L2/L3 | MFA, credential strength, session security |
| V3: Session Management | L2/L3 | Cookie flags, timeout limits, token entropy |
| V4: Access Control | L2/L3 | Deny-by-default, least-privilege role validation |
| V5: Validation, Sanitization | L1/L2/L3 | Parameterized queries, context-aware escaping |
| V8: Data Protection | L2/L3 | TLS in transit, AES-256 at rest, key separation |

### 1.2 Data Classification

Classify all application data to determine security and encryption requirements:

| Classification | Sensitivity | Examples | Minimum Protection |
|----------------|-------------|----------|-------------------|
| **Public** | Low | Marketing copy, product docs | Public integrity check |
| **Internal** | Medium | Internal policies, dev manuals | Authentication required |
| **Confidential** | High | Financial reports, customer emails | Encryption at rest + audit |
| **PII (Restricted)** | Critical | Name, address, IP, SSN, email | Encryption + pseudonymization |
| **PCI (Restricted)** | Critical | Credit card numbers, CVVs | Tokenization + isolated network |
| **PHI (Restricted)** | Critical | Health records, patient IDs | Strict access trails + TLS force |

### 1.3 Abuse & Misuse Cases

Draft abuse/misuse cases to model attacker behaviors against user stories:

```
User Story: "As a registered user, I want to reset my password via email so that I can regain access."
  ├── Abuse Case 1: Attacker attempts to brute-force the password reset token
  ├── Abuse Case 2: Attacker requests password reset to trigger email bombing/DoS
  └── Misuse Case: Attacker alters username in the reset request to trigger token leaks (IDOR)
```

**Common Scenarios Checklist:**
- **Auth Flow**: Abuse of MFA bypass, session replay, token stuffing.
- **File Upload**: Uploading executable payloads, zip bomb decompression, path traversal names.
- **APIs**: Mass assignment of permissions, BOLA/IDOR on resource IDs.
- **Checkout/Payments**: Price manipulation (negative prices, coupon abuse).

### 1.4 Compliance Controls Mapping

Map requirements directly to regulatory and standard compliance frameworks:

| Compliance Control | Requirement Focus | Technical Implementation |
|--------------------|-------------------|--------------------------|
| **GDPR Art. 25** | Data Protection by Design & Default | Pseudonymization, data minimization |
| **GDPR Art. 32** | Security of Processing | Encryption, availability, regular testing |
| **PCI-DSS 4.0 §3** | Protect Cardholder Data | Strong crypto (AES-256), primary account masking |
| **PCI-DSS 4.0 §6** | Develop Secure Systems | Code reviews, SAST/SCA integration, dependency pins |
| **HIPAA §164.312(a)** | Access Controls | Unique user IDs, automatic logoff, encryption |
| **HIPAA §164.312(b)** | Audit Controls | Immutable logs tracking PHI access |
| **ISO 27001 Annex A.8** | Asset Management & Info Classification | Clear labels, access control guidelines |
| **ISO 27001 Annex A.14** | Secure System Development | Secure engineering principles, test data control |

### 1.5 Security Acceptance Criteria

Define explicit security criteria before developers start coding:

- `[ ]` Input parameters are validated against a strict allowlist.
- `[ ]` Data flow crossing trust boundaries uses TLS 1.3/HTTPS.
- `[ ]` Secrets/credentials are loaded from external configuration (never hardcoded).
- `[ ]` User identity is verified on every API request checking the specific resource ownership.
- `[ ]` Logging covers success/failure of authentication and authorization events.

## GraphX Integration

**Query Phase 1 Elements**

Initialize the knowledge graph and verify compliance scope:

```bash
# Initialize GraphX threat model
graphx init

# Query entry points to define trust boundaries
GraphX: GetAPIs()
```

## Deliverables

- **Security Requirements Document**: Mapped to ASVS L1/L2/L3.
- **Data Classification Catalog**: Listing PII, PCI, and PHI assets.
- **Abuse & Misuse Case Catalog**: Attacker stories.
- **Compliance Mapping Matrix**: Traceable controls.
- **Security Acceptance Criteria**: Integrated into sprint tasks.

## Exit Criteria

- [ ] ASVS targets established.
- [ ] Compliance scope defined.
- [ ] Security acceptance criteria defined.
- [ ] GraphX initialized.