---
name: graphx-mitigation
description: Context-aware security mitigation recommendations with framework-specific guidance
---

# Mitigation Recommendation Engine

Generate actionable, context-aware security recommendations integrated with architecture and deployment.

## Mitigation Strategy

```
Threat → Analyze Context → Map Controls → Generate Recommendations → Prioritize
```

## Control Mapping

### NIST CSF Mapping

| Category | Security Function | graphx Mitigations |
|----------|-------------------|-------------------|
| ID.AM | Asset Management | Inventory automation |
| ID.BE | Business Environment | Critical asset tagging |
| ID.GV | Governance | Security policies |
| ID.RM | Risk Strategy | Threat prioritization |
| ID.DE | Supply Chain | Dependency scanning |
| ID.RA | Risk Assessment | CVSS scoring |
| PR.AC | Identity Management | MFA, SSO |
| PR.DS | Data Security | Encryption, DLP |
| PR.PS | Platform Security | Hardening |
| PR.IR | Information Protection | Classification |
| DE.CM | Continuous Monitoring | SIEM integration |
| DE.DP | Detection Processes | Alert tuning |
| RS.RP | Response Planning | Playbooks |
| RS.CO | Communications | Incident response |
| RS.AN | Analysis | Forensics |
| RS.MI | Mitigation | Patching, configs |
| RS.IM | Improvements | Lessons learned |

## Threat-to-Mitigation Matrix

### Authentication & Session Management

| Threat | Mitigation | Framework | Priority |
|--------|-----------|-----------|----------|
| Credential Stuffing | Rate limiting + MFA | OWASP | HIGH |
| JWT Forgery | Algorithm validation + short expiry | OWASP | HIGH |
| Session Hijacking | HttpOnly + Secure + SameSite | OWASP | HIGH |
| OAuth Token Theft | PKCE + token rotation | OAuth | HIGH |

### Code Injection

| Threat | Mitigation | Implementation |
|--------|-----------|----------------|
| SQL Injection | Parameterized queries | `db.Query("SELECT * FROM t WHERE id=?", id)` |
| Command Injection | No shell=True, input validation | Allowlist validation |
| XSS | Output encoding, CSP | Context-aware escaping |
| Deserialization | Type validation, no pickle | Schema validation |

### Container & Kubernetes

| Threat | Mitigation | K8s Control |
|--------|-----------|------------|
| Container Escape | Read-only root, drop capabilities | securityContext |
| Privileged Container | deny privileged pods | PodSecurityPolicy |
| HostPath Access | Avoid hostPath | volume securityContext |
| RBAC Escalation | Principle of least privilege | RBAC audit |
| Secrets in Env | Vault sidecar injection | External secrets |

### Cloud Infrastructure

| Threat | Mitigation | AWS Control |
|--------|-----------|-------------|
| SSRF → Metadata | Use IMDSv2, block 169.254.169.254 | Instance metadata service |
| S3 Public Access | Block public, use bucket policies | s3.public_access_block |
| Overpermissive IAM | Principle of least privilege | IAM policies |
| Unencrypted Data | Enable encryption at rest | AES-256, KMS |
| Weak TLS | TLS 1.3, strong ciphers | Security policy |

## Context-Aware Recommendations

### Framework-Specific Guidance

**Go Applications**
```go
// SQL Injection Prevention
// GOOD: Parameterized query
rows, err := db.Query("SELECT * FROM users WHERE id = ?", userID)

// GOOD: Use sqlx for additional safety
query := sqlx.Rebind("SELECT * FROM users WHERE id = ?", userID)

// GOOD: Transaction with rollback
tx, err := db.BeginTxx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

// BAD: String concatenation (never do this)
// rows, err := db.Query("SELECT * FROM users WHERE id = " + userID)
```

**Kubernetes Workloads**
```yaml
# Secure Pod Specification
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 10000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: app
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
          - ALL
    resources:
      limits:
        memory: "256Mi"
        cpu: "500m"
```

**Terraform AWS Resources**
```hcl
# Secure S3 Bucket
resource "aws_s3_bucket" "app_data" {
  bucket = "app-data-${var.environment}"
  
  lifecycle {
    rule {
      expiration { days = 90 }
    }
  }
  
  versioning { enabled = true }
  
  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}

# Block public access
resource "aws_s3_bucket_public_access_block" "app_data" {
  bucket = aws_s3_bucket.app_data.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# IAM Role with least privilege
resource "aws_iam_role" "app_role" {
  name = "app-role-${var.environment}"
  
  inline_policy {
    name = "app-policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "s3:GetObject",
            "s3:PutObject",
          ]
          Resource = "${aws_s3_bucket.app_data.arn}/*"
          # NOT "Resource": "*"
        }
      ]
    })
  }
}
```

## Mitigation Categories

### Preventive Controls

| Control Type | Examples | Effectiveness |
|--------------|----------|----------------|
| Input Validation | Allowlists, type checking | High |
| Authentication | MFA, OAuth2, JWT | High |
| Authorization | RBAC, ABAC, zero-trust | High |
| Encryption | TLS, AES-256, KMS | High |
| Network Security | Firewall, WAF, mTLS | Medium-High |

### Detective Controls

| Control Type | Examples | Effectiveness |
|--------------|----------|----------------|
| Logging | Structured logs, audit trail | Medium |
| Monitoring | SIEM, RASP, EDR | Medium-High |
| Scanning | SAST, DAST, dependency | Medium |
| Alerting | Anomaly detection | Medium |

### Corrective Controls

| Control Type | Examples | Effectiveness |
|--------------|----------|----------------|
| Incident Response | Playbooks, automation | High |
| Backup/Recovery | Snapshots, point-in-time | High |
| Patching | Automated updates | High |

## Priority Calculation

```go
type MitigationPriority struct {
    ThreatSeverity float64    // 0.0-1.0
    ControlEffectiveness float64  // 0.0-1.0
    ImplementationEffort float64   // 0.0-1.0 (higher = easier)
    FalsePositiveRisk float64    // 0.0-1.0
}

func (m *Mitigation) CalculatePriority() Priority {
    score := (m.ThreatSeverity * 0.4) +
             (m.ControlEffectiveness * 0.3) +
             (m.ImplementationEffort * 0.2) +
             ((1.0 - m.FalsePositiveRisk) * 0.1)
    
    switch {
    case score >= 0.8:
        return CRITICAL
    case score >= 0.6:
        return HIGH
    case score >= 0.4:
        return MEDIUM
    default:
        return LOW
    }
}
```

## Implementation Tracking

```yaml
mitigations:
  - id: MIT-001
    threat_id: STRIDE-K8S-001
    title: "Remove hostPath volumes"
    
    status: open
    priority: HIGH
    owner: platform-team
    
    implementation:
      effort: medium
      timeline: 2 weeks
      dependencies:
        - Alternative volume type
      
    controls:
      - CIS Kubernetes 5.2.9
      - NIST PR.PS-1
      
    verification:
      test: Check no hostPath in any pod spec
      automation: "kubectl get pods -o json | jq '.items[].spec.volumes[].hostPath'"
```

## Compliance Mapping

| Mitigation | OWASP | CIS | NIST | PCI-DSS |
|------------|-------|-----|------|--------|
| MFA | A.5.1 | 1.1 | PR.AC-7 | 8.3 |
| Encryption | A.2.4 | 2.2 | PR.DS-1 | 3.4 |
| Input Validation | A.1.1 | 1.4 | DE.CM-4 | 6.5 |
| Rate Limiting | A.7.1 | - | PR.AC-5 | 6.2 |
| Logging | A.9.1 | - | DE.CM-1 | 10.1 |

## Output Format

```yaml
mitigation_report:
  generated: "2024-01-15T10:30:00Z"
  
  summary:
    total_threats: 47
    mitigated: 23
    in_progress: 12
    pending: 12
    
  recommendations:
    - id: MIT-001
      threat: "SQL Injection in /api/users"
      mitigation: "Use parameterized queries"
      priority: HIGH
      effort: low
      code_changes:
        - file: "api/handlers.go"
          line: 45
          before: "db.Query(fmt.Sprintf(\"SELECT...\", id))"
          after: "db.Query(\"SELECT * FROM users WHERE id=?\", id)"
      
    - id: MIT-002
      threat: "Container runs as root"
      mitigation: "Add security context"
      priority: HIGH
      effort: low
      k8s_changes:
        - file: "deployment.yaml"
          before: |
            securityContext: {}
          after: |
            securityContext:
              runAsNonRoot: true
              runAsUser: 10000
              readOnlyRootFilesystem: true
```

## Best Practices

1. **Defense in Depth**: Layer multiple controls
2. **Fail Secure**: Default deny, explicit allow
3. **Least Privilege**: Minimal permissions everywhere
4. **Zero Trust**: Verify every request
5. **Secure Defaults**: Opt-in security, not opt-out