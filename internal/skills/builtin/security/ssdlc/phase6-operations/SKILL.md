---
name: ssdlc-phase6-operations
description: Phase 6 - Security operations, monitoring, and incident response
---

# Phase 6: Security Operations & Monitoring

Continuously monitor production systems, detect threats, and respond to incidents.

## Objectives

- Monitor security events 24/7
- Detect and respond to threats
- Manage vulnerabilities
- Maintain security posture
- Enable continuous improvement

## Phase Activities

### 6.1 Security Monitoring

**SOC Monitoring Priorities**

| Category | Metrics | Alert Threshold |
|----------|---------|-----------------|
| Authentication | Failed logins, password resets, MFA failures | > 10/min |
| Authorization | Escalation attempts, forbidden access | > 5/min |
| Data Access | Bulk exports, unusual queries | > 1000 records/min |
| Infrastructure | New ports, unusual traffic, DDoS | > baseline 3x |
| Dependencies | New CVEs, outdated packages | Daily scan |

**SIEM Integration**

```yaml
# SIEM Configuration
siem:
  providers:
    - crowdstrike
    - elasticsearch
  
  log_sources:
    - application_logs
    - authentication_logs
    - network_flows
    - container_logs
    - cloud_trails
  
  correlation_rules:
    - name: "Credential stuffing"
      condition: |
        event_type=login 
        AND failed_attempts > 20 
        AND unique_ips > 5
        AND time_window < 5m
      severity: HIGH
      actions:
        - block_ip
        - notify_security
        - create_ticket
    
    - name: "Data exfiltration"
      condition: |
        event_type=data_access
        AND volume > 10000
        AND unusual_time
      severity: CRITICAL
      actions:
        - alert_oncall
        - snapshot_session
        - isolate_account
```

**Application Monitoring (RASP)**

```yaml
# Runtime Application Self-Protection
rasp:
  monitoring:
    - sql_injection
    - xss_attempts
    - command_injection
    - file_access
    - network_connections
  
  blocking:
    - confirmed_attacks
    - high_confidence_findings
  
  alerting:
    - all_attacks
    - anomalies
```

### 6.2 Incident Response

**Incident Classification**

| Severity | Response Time | Examples |
|----------|---------------|----------|
| P1 - Critical | 15 min | Active breach, ransomware, data exfiltration |
| P2 - High | 1 hour | RCE exploit, confirmed vulnerability |
| P3 - Medium | 4 hours | Suspicious activity, policy violation |
| P4 - Low | 24 hours | Potential issue, no immediate risk |

**Response Playbook**

```yaml
incident_response:
  phases:
    detection:
    - Verify alert validity
    - Classify severity
    - Notify team
    - Create ticket
    
    containment:
    - Isolate affected systems
    - Block malicious actors
    - Preserve evidence
    - Activate response team
    
    eradication:
    - Remove threat
    - Patch vulnerabilities
    - Rotate credentials
    
    recovery:
    - Restore from backup
    - Verify integrity
    - Monitor for recurrence
    
    lessons_learned:
    - Document timeline
    - Identify gaps
    - Update defenses
    - Notify stakeholders
```

**Evidence Collection**

```bash
# Memory acquisition
sudo avml /tmp/memory.lime

# Disk imaging
sudo dd if=/dev/sda of=/tmp/disk.img status=progress

# Network capture
sudo tcpdump -i eth0 -w /tmp/capture.pcap

# Log preservation
sudo tar -czf /tmp/logs.tar.gz /var/log/ /etc/
sudo setfattr -n security.evm -v 0 /tmp/logs.tar.gz
```

### 6.3 Vulnerability Management

**Continuous Scanning**

```yaml
vulnerability_management:
  scanning:
    sast:
      frequency: on_push
    dast:
      frequency: nightly
    container:
      frequency: on_build
    dependency:
      frequency: daily
  
  remediation_sla:
    critical: 24 hours
    high: 7 days
    medium: 30 days
    low: 90 days
  
  exceptions:
    criteria:
      - business_critical_system
      - compensating_controls
      - approved_by_ciso
    review: quarterly
```

**CVE Response**

```bash
# CVE Assessment
./vuln-check.sh --cve CVE-2024-XXXXX

# Impact Analysis
GraphX: SearchNodes(affected_component)
GraphX: GetBlastRadius(affected_node)

# Decision Matrix
| Exploitable | In Prod | Fix Available | Action |
|------------|---------|---------------|--------|
| Yes | Yes | Yes | Emergency patch |
| Yes | Yes | No | Mitigate + escalate |
| Yes | No | - | Patch in cycle |
| No | - | - | Monitor |
```

### 6.4 Security Metrics

**Key Performance Indicators**

| Metric | Target | Current | Trend |
|--------|--------|---------|-------|
| MTTD (Mean Time to Detect) | < 1 hour | | |
| MTTR (Mean Time to Respond) | < 4 hours | | |
| Critical Vulnerabilities | 0 | | |
| High Vulnerabilities | < 5 | | |
| Security Debt | < 5% | | |
| SAST Coverage | 100% | | |
| DAST Findings Open | < 10 | | |
| Training Completion | 100% | | |

**Security Dashboard**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Security Operations Center                    │
├─────────────────────────────────────────────────────────────────┤
│ Active Alerts: 3   Open Vulns: 47   MTTD: 45min   MTTR: 2.5h   │
├────────────────┬────────────────────────────────────────────────┤
│ Severity      │ Open    │ In Progress │ Resolved This Week    │
├────────────────┼─────────┼─────────────┼────────────────────────┤
│ CRITICAL      │    0    │      0      │          5            │
│ HIGH          │    2    │      3      │         15            │
│ MEDIUM        │   10    │      5      │         45            │
│ LOW           │   12    │      2      │         67            │
├────────────────┴─────────┴─────────────┴────────────────────────┤
│ Recent Incidents                                                │
│ • P2: Auth bypass attempt (contained, investigating)            │
│ • P3: Outdated dependency found (patch scheduled)            │
│ • P4: Missing security header (ticket created)                 │
├─────────────────────────────────────────────────────────────────┤
│ Compliance Status                                               │
│ • GDPR: Compliant    • PCI-DSS: Compliant    • SOC2: Compliant  │
└─────────────────────────────────────────────────────────────────┘
```

### 6.5 Log Management

**What to Log**

| Category | Events | Retention |
|----------|--------|-----------|
| Authentication | Login, logout, MFA, password reset | 1 year |
| Authorization | Access granted, denied | 1 year |
| Data Access | Reads, writes, deletes, exports | 6 months |
| System | Service start/stop, config changes | 3 months |
| Network | Connections, blocks, rate limits | 3 months |

**What NOT to Log**

- Passwords (even failed attempts)
- Full credit card numbers
- API keys/secrets
- PII without masking
- Session tokens (use session ID)

### 6.6 Cloud Monitoring

**AWS Security Monitoring**

```yaml
cloudtrail:
  enabled: true
  multi_region: true
  log_validation: true

guardduty:
  enabled: true
  frequency: HOURLY

security_hub:
  enabled: true
  standards:
    - aws-foundational-security-best-practices
    - cis-aws-benchmark
```

## GraphX Integration

**Real-Time Security Context**

```bash
# Monitor security posture
GraphX: GetStatus() - graph health
GraphX: Query risk_score > 7.0 - critical components

# Incident response
GraphX: GetBlastRadius(affected_node)
GraphX: SearchNodes(attack_pattern)
GraphX: GetThreatModel()

# Continuous improvement
GraphX: Query new_functions_since_last_scan
GraphX: GetTrustBoundaries() - verify boundaries
GraphX: SearchNodes("auth", "admin") - changes
```

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Security Dashboard | Real-time visibility |
| Incident Runbooks | Response procedures |
| Vulnerability Report | Active vulnerabilities |
| Metrics Report | KPIs and trends |
| GraphX Status | Security graph health |

## Exit Criteria

- [ ] SIEM integrated and alerting
- [ ] Incident playbooks documented
- [ ] P1 response < 15 min
- [ ] Zero CRITICAL vulnerabilities
- [ ] Metrics meeting targets
- [ ] GraphX monitoring active