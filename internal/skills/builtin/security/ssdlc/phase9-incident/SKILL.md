---
name: phase9-incident
description: Phase 9 - Incident Response playbooks, Contain, Identify, Eradicate, Recover, Lessons Learned, and reporting
---

# Phase 9: Incident Response

Coordinate response to security incidents, contain threat actions, identify affected components, eradicate vulnerabilities, and document findings.

## Objectives

- Execute the 5-step incident response playbook (Contain, Identify, Eradicate, Recover, Lessons Learned)
- Document security findings using vulnerability reports
- Generate executive summaries of incidents
- Analyze blast radius and update threat models

## Phase Activities

### 9.1 The 5-Step Incident Response Playbook

When a security incident, breach, or compromise is detected, execute the following steps in order:

```
1. CONTAIN   →  Isolate systems, revoke credentials, block attacker IPs
      ↓
2. IDENTIFY  →  Preserve logs (read-only), locate IOCs, analyze blast radius
      ↓
3. ERADICATE →  Patch/fix vulnerabilities, rotate compromised secrets
      ↓
4. RECOVER   →  Validate fix, restore from clean backups, monitor logs
      ↓
5. LESSONS   →  Perform root cause analysis, document detection gaps
```

### 9.2 Containment Activities

- **Isolate Systems**: Disconnect affected containers, network hosts, or VMs from the virtual network.
- **Revoke Credentials**: Revoke active session tokens, OAuth keys, and API tokens of compromised accounts or services.
- **Block Attacker IPs**: Inject blocking rules into edge WAFs or load balancers.

### 9.3 Identification & Preservation

*   **Preserve Evidence**: Copy log files, database logs, and server journals to a read-only secure environment.
*   **Locate Indicators of Compromise (IOCs)**: Scan log entries for unusual request payloads, suspicious HTTP user-agents, or anomalous database queries.
*   **Analyze Blast Radius**: Utilize GraphX to determine affected upstream and downstream dependencies:
    ```bash
    GraphX: GetBlastRadius(compromised_node_id)
    ```

### 9.4 Eradication & Recovery

- **Patch and Fix**: Implement secure code patches in the repository. Verify the fix using automated test suites.
- **Rotate Secrets**: Rotate all keys, tokens, and database passwords that may have been exposed.
- **Restore Safely**: Restore database or server files from a verified clean backup.
- **Post-Recovery Monitoring**: Monitor system logs for 72 hours for any signs of re-compromise or residual threat presence.

### 9.5 Reporting & Lessons Learned

- **Document Findings**: Always document the vulnerability and root cause using the `create_vulnerability_report` tool:
  ```bash
  create_vulnerability_report --title "Root cause of incident" --description "Details..."
  ```
- **Executive Summary**: Generate a high-level executive summary report mapping the incident timeline and containment steps.
- **Lessons Learned Review**:
  - Perform a root cause analysis (RCA).
  - Perform a detection gap analysis (why did alerts not trigger sooner?).
  - Propose control improvements to prevent recurrence.
  - Update the architectural threat model in GraphX.

## Exit Criteria

- [ ] Incident contained.
- [ ] Root cause identified.
- [ ] Fix deployed and validated.
- [ ] Secrets rotated and system restored.
- [ ] Vulnerability report generated.
- [ ] Lessons learned documented.
- [ ] GraphX threat model updated.
