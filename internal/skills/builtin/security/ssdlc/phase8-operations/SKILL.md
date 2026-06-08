---
name: phase8-operations
description: Phase 8 - Security operations & runtime security, log auditing, alerting thresholds, secrets rotation, and vulnerability patching SLAs
---

# Phase 8: Operations & Runtime

Enforce runtime security, configure security logging and alerting, track secrets rotation schedules, and maintain patching service level agreements (SLAs).

## Objectives

- Audit runtime applications for security logging coverage
- Establish clear security alerting thresholds
- Track secrets age and enforce 90-day rotation rules
- Maintain strict patch management SLAs based on vulnerability severity

## Phase Activities

### 8.1 Security Observability Logging Rules

Every production system must output structured logs for security-relevant actions. Ensure the following events are logged:

*   **Authentication Events**: Successful logins, logouts, MFA challenges, password reset requests.
*   **Authorization Failures**: Forbidden requests (403), access elevation attempts, role changes.
*   **Admin Actions**: API keys generation, configuration changes, user creation, deletion, or permission updates.
*   **API Errors**: Log HTTP status codes `4xx` (for anomaly tracking) and `5xx` (to detect potential exploits or database failures).

### 8.2 Security Alerting Thresholds

Configure monitoring alerts (e.g. Prometheus, SIEM, Falco) to trigger notifications immediately on:

-   **Brute Force Attacks**: Track authentication attempts. Alert on **>10 failed authentication failures per minute** per user or IP address.
-   **Impossible Travel**: Detect login attempts originating from geographically incompatible locations within a short time window.
-   **Privilege Escalation**: Track runtime authorization requests. Alert on any unauthorized role elevations.
-   **New Admin Accounts**: Alert immediately on the creation of any new administrator or root-privilege accounts.

### 8.3 Secrets Rotation Policy

Track credentials, passwords, certificates, and API tokens used by production infrastructure:

*   **90-Day Age Limit**: Audit the creation date of all active credentials.
*   **Flag for Rotation**: Any API key, database password, or secret older than 90 days must be flagged and updated.
*   **Workaround**: Use CLI commands to rotate secrets programmatically:
    ```bash
    aws secretsmanager rotate-secret --secret-id my-secret-id
    ```

### 8.4 Vulnerability Patching SLAs

When new CVEs are discovered in production environments, enforce the following patching schedule:

| Severity | CVSS Score Range | Maximum Remediation SLA |
|----------|------------------|-------------------------|
| **CRITICAL** | 9.0 - 10.0 | **<= 24 hours** |
| **HIGH** | 7.0 - 8.9 | **<= 7 days** |
| **MEDIUM** | 4.0 - 6.9 | **<= 30 days** |
| **LOW** | 0.1 - 3.9 | **<= 90 days** |

## Exit Criteria

- [ ] Security logs configured.
- [ ] Alerts thresholds configured.
- [ ] Secrets older than 90 days rotated.
- [ ] Vulnerability SLAs met.
- [ ] Runtime monitoring mapped in GraphX.
