---
name: phase5-testing
description: Phase 5 - Security testing and validation, scan modes (quick/standard/deep), DAST crawling, API top 10 auditing, and code tracing
---

# Phase 5: Testing & Validation

Perform automated scans and manual validation across target systems using quick, standard, or deep assessment modes.

## Objectives

- Choose the appropriate scan mode (/quick, /standard, /deep)
- Execute web application vulnerability discovery pipelines
- Test API endpoints against the OWASP API Top 10
- Trace static data flows to validation sinks
- Validate issues with working Proofs of Concept (PoCs)

## Phase Activities

### 5.1 Scan Mode Selection

Assess the target using one of three scan modes based on the time budget:

| Scan Mode | Target Timing | Actions & Tools | Focus Area |
|-----------|---------------|-----------------|------------|
| `/quick` | < 5 minutes | `httpx` fingerprinting + `nuclei` CVE templates + exposed secrets scan | High-impact, low-hanging fruits |
| `/standard` | 15-30 minutes | `/quick` + `ffuf` content discovery + `sqlmap` on forms + auth bypass checks | Standard vulnerability assessment |
| `/deep` | Hours (Exhaustive) | Full recon → content discovery → crawl → business logic → exploitation → chaining | Exhaustive penetration test |

### 5.2 Web Application Penetration Testing Flow

Follow this systematic testing pipeline for web targets:

1.  **Reconnaissance**: Fingerprint technologies, headers, and infrastructure details using `httpx`.
2.  **Content Discovery**: Bruteforce directories and files using `ffuf` to locate hidden paths.
3.  **Crawl & Spider**: Map endpoints, inputs, forms, and client scripts using `katana` or `gospider`.
4.  **Port Scanning**: Identify exposed services and endpoints using `naabu` or `nmap`.
5.  **Vulnerability Scanning**: Run targeted YAML templates using `nuclei`.
6.  **Manual Testing**: Manually validate injection parameters, CSRF tokens, and security header policies.

### 5.3 API Security (OWASP API Top 10)

Validate API endpoints against critical flaws:

-   **BOLA / IDOR (API1)**: Test accessing resources of other users by altering IDs (e.g. `/api/users/1` -> `/api/users/2`).
-   **Broken User Authentication (API2)**: Inspect session token strength, check for token none-algorithm support or weak signing secrets.
-   **Excessive Data Exposure (API3)**: Verify if endpoints return complete object structures containing sensitive fields not shown in the UI.
-   **Lack of Resources & Rate Limiting (API4)**: Stress test endpoints to determine if missing rate limits permit brute-force or DoS.
-   **Mass Assignment (API6)**: Attempt to update sensitive fields by injecting properties into request bodies (e.g., `{"is_admin": true}`).

### 5.5 Source Code Audit (White-box Tracking)

When reviewing source code:

-   **Load Deep Scan Skill**: Load `scan_modes/deep` skill.
-   **Map Entry Points**: Identify all HTTP handlers, middleware, event listeners, and API endpoints.
-   **Trace Data Flows**: Track data from untrusted sources (request body, headers, params) to security sinks (database queries, shell executions, HTML templates).
-   **Verify Controls**: Confirm that security filters, validations, and sanitizer functions are applied correctly before the sink.

## Exit Criteria

- [ ] Scanning mode completed.
- [ ] DAST / API scans executed.
- [ ] Vulnerability findings validated with PoCs.
- [ ] Code data flows traced.
- [ ] Target security posture mapped in GraphX.
