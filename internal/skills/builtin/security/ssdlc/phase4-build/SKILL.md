---
name: phase4-build
description: Phase 4 - Build & CI/CD security, secrets scanning, Software Composition Analysis (SCA), GitHub Actions hardening, and SBOM checklist
---

# Phase 4: Build & CI/CD

Validate build pipelines, check dependencies for CVEs, scan repositories for credentials, and secure GitHub Actions configurations.

## Objectives

- Run secrets detection on repo history and filesystem
- Perform Software Composition Analysis (SCA) to flag CRITICAL/HIGH CVEs
- Hardened GitHub Actions configurations (pinned SHAs, minimal permissions)
- Validate container image structure and build parameters
- Generate and verify Software Bill of Materials (SBOM)

## Phase Activities

### 4.1 Secrets Scanning

Every build or commit pipeline must run automated secrets detection:

```bash
# Scan repository using gitleaks via container_scanner
container_scanner --tool secrets__gitleaks --path .

# Alternative local workaround
gitleaks detect --source . -f json
```

**Common Flags**:
- Flag API keys, private keys, database passwords, OAuth client secrets.
- Block the build if any new secrets are detected.

### 4.2 Software Composition Analysis (SCA)

Identify vulnerable dependencies in the application dependency tree:

```bash
# Run SCA scan using Trivy via container_scanner
container_scanner --tool sca__trivy_fs --path .

# local workaround
trivy fs --severity HIGH,CRITICAL .
```

**SCA Rules**:
- Fail the build pipeline if there are any unresolved **CRITICAL** or **HIGH** CVEs in the dependency list.
- Check for dependency supply chain risks (e.g. dependency confusion, outdated or hijacked packages).

### 4.3 GitHub Actions Hardening Checklist

Audit all workflow definition files (`.github/workflows/*.yml`):

*   **Pin Actions to Commit SHA**: Avoid using mutable tags like `@v4` or `@main`. Use immutable commit hashes:
    ```yaml
    - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2
    ```
*   **Minimal Permissions**: Explicitly restrict job permissions in workflows:
    ```yaml
    permissions:
      contents: read
    ```
*   **Secure Pull Requests**: Do not run untrusted PR code in workflows with access to repository secrets. Avoid `pull_request_target` with head checkout.
*   **Prevent Shell Injection**: Never interpolate `github.event` variables (like PR title or issue body) directly into run scripts. Pass them via environment variables:
    ```yaml
    env:
      PR_TITLE: ${{ github.event.pull_request.title }}
    run: echo "PR is $PR_TITLE"
    ```

### 4.4 Container Image Auditing

Verify container files (`Dockerfile`):

- **No Latest Tags**: Pin base image tags to specific digest or version (e.g. `golang:1.26-bookworm`).
- **Non-Root USER**: Ensure a `USER` directive is specified with a non-root UID (e.g. `USER duckops` or `USER 1000:1000`).
- **No Secrets in Layers**: Do not use `ENV` or `ARG` directives to pass database credentials or API tokens; they remain readable in the image history via `docker history`.

### 4.5 SBOM Generation Checklist

Before final release, generate a Software Bill of Materials (SBOM) to document package inventory:

```bash
# Generate CycloneDX SBOM using syft
syft . -o cyclonedx-json > sbom.json

# Workaround for vulnerability matching
syft . -o cyclonedx-json | grype --add-cpes-if-none
```

## Exit Criteria

- [ ] Secrets scan completed with 0 new issues.
- [ ] SCA scan clean of unmitigated CRITICAL/HIGH CVEs.
- [ ] GitHub Actions pinned and permission-scoped.
- [ ] Dockerfiles hardened.
- [ ] SBOM generated.
