---
name: phase6-release
description: Phase 6 - Pre-release security gate, container image validation, Docker CIS benchmarks, and dependency verification
---

# Phase 6: Release & Staging

Enforce pre-release security gates, validate container image layers, verify CIS benchmarks, and confirm dependency pinning before production rollout.

## Objectives

- Verify all CRITICAL and HIGH vulnerabilities are resolved or risk-accepted
- Confirm container image is scanned with 0 CRITICAL issues
- Ensure no secrets remain in image history layers
- Attach SBOM artifacts to the release
- Verify Dockerfile follows CIS Docker Benchmarks
- Ensure all dependencies are pinned with integrity hashes

## Phase Activities

### 6.1 Pre-Release Security Gate Checklist

Before promoting a release artifact to production, complete the following validation checklist:

-   `[ ]` **Vulnerability Resolution**: Verify that all **CRITICAL** and **HIGH** severity vulnerabilities detected by SAST, DAST, SCA, and secrets scanners are either resolved or explicitly risk-accepted with a signed justification document.
-   `[ ]` **Container Security Scan**: Run Trivy or equivalent scanner on the production image. Ensure there are **0 CRITICAL** issues.
-   `[ ]` **Image Layer Audit**: Inspect the container image structure using `docker history --no-trunc <image-name>`. Confirm no environment variables (`ENV`), build arguments (`ARG`), or COPY commands contain secrets or API keys.
-   `[ ]` **SBOM Attachment**: Ensure that the generated Software Bill of Materials (SBOM) is attached to the release artifact or metadata registry.
-   `[ ]` **Dockerfile Hardening**: Audit the production Dockerfile against the CIS Docker Benchmark guidelines.
-   `[ ]` **Dependency Integrity**: Ensure that all dependencies (npm, go mod, python requirements) are pinned to specific versions and validated with cryptographic integrity hashes (e.g. `package-lock.json`, `go.sum`).

### 6.2 CIS Docker Benchmark Rules

Validate container images against standard hardening rules:

*   **Rule 4.1 - Run as Non-Root**: Ensure the `USER` directive is defined with a non-root account:
    ```dockerfile
    USER duckops
    ```
*   **Rule 4.2 - Read-Only Filesystem**: Ensure the root filesystem is read-only. Write ephemeral data to memory or designated volume mounts:
    ```yaml
    # Kubernetes SecurityContext
    readOnlyRootFilesystem: true
    ```
*   **Rule 4.3 - No Privileged Execution**: Ensure the container cannot request root privileges at runtime:
    ```yaml
    allowPrivilegeEscalation: false
    privileged: false
    ```
*   **Rule 4.4 - Minimal Image Size**: Avoid general-purpose OS bases. Use minimal bases (`scratch`, `alpine`, or distroless) to reduce the system attack surface.

### 6.3 Release Validation Commands

Validate release artifacts using these helper commands:

```bash
# Scan production image with Trivy
trivy image --severity CRITICAL --exit-code 1 ghcr.io/org/app:v1.0.0

# Verify no secrets remain in image history layers
docker history --no-trunc ghcr.io/org/app:v1.0.0 | grep -iE "password|token|secret|key" || echo "Clean"
```

## Exit Criteria

- [ ] Security gates cleared.
- [ ] Container image verified.
- [ ] Docker CIS benchmarks enforced.
- [ ] SBOM attached to release.
- [ ] Dependencies pinned.
