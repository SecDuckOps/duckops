---
name: ssdlc-phase5-deployment
description: Phase 5 - Secure deployment, infrastructure hardening, and deployment validation
---

# Phase 5: Secure Deployment

Deploy systems securely with hardened infrastructure and validated configurations.

## Objectives

- Harden infrastructure and containers
- Implement secrets management
- Secure CI/CD pipeline
- Validate deployment
- Enable monitoring

## Phase Activities

### 5.1 Container Security

**Multi-Stage Build**

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o app

# Production stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/app /app
USER 1001:1001
ENTRYPOINT ["/app"]
```

**Security Features**

| Feature | Implementation | Purpose |
|---------|---------------|---------|
| Non-root user | `USER 1001:1001` | Container breakout protection |
| Read-only root FS | `readOnly: true` | Filesystem attack prevention |
| No shell | `FROM scratch` | Reverse shell prevention |
| Minimal base | `alpine` or `scratch` | Attack surface reduction |
| Signed image | `cosign` | Image authenticity |

**Kubernetes Pod Security**

```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    fsGroup: 1001
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: app
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        drop:
          - ALL
    resources:
      limits:
        memory: "256Mi"
        cpu: "500m"
```

### 5.2 Secrets Management

**HashiCorp Vault Integration**

```yaml
# Vault Agent sidecar
vault:
  address: https://vault.internal:8200
  auth:
    method: kubernetes
    role: app-role
  
  templates:
    - secretPath: secret/data/app/database
      dest: /run/secrets/db.conf
      perm: "0400"
      owner: "999"
```

**Secret Rotation**

| Secret Type | Rotation | Automated |
|-------------|----------|-----------|
| Database passwords | 90 days | Yes |
| API keys | 180 days | Partial |
| TLS certificates | 90 days | Yes (Let's Encrypt) |
| Encryption keys | 365 days | Yes |
| Session tokens | Per session | N/A |

**Deployment Verification**

```bash
# Verify no secrets in image
trufflehog docker://registry/app:latest

# Verify no default credentials
curl -s https://app.example.com/health
test ! -f /app/config/default.json

# Verify secrets loaded at runtime
kubectl exec -it app-pod -- /app/healthcheck
vault read secret/data/app/config
```

### 5.3 Infrastructure Hardening

**Kubernetes Security**

```yaml
# Network Policy - Default Deny
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress

---
# RBAC - Least Privilege
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: app-role
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get"]
```

**Cloud Security (AWS)**

```yaml
# S3 Bucket Policy
- name: app-data-bucket
  bucket: app-data
  encryption: AES256
  versioning: true
  public_access_block: true
  lifecycle:
    transition_to_glacier: 90

# RDS Security
rds:
  storage_encrypted: true
  multi_az: true
  backup_retention: 30
  parameters:
    - name: ssl
      value: "require"
    - name: rds.force_ssl
      value: "1"

# IAM Policy - Deny Unused
- name: deny-unused-actions
  Statement:
  - Effect: Deny
    Action: "*"
    Resource: "*"
    Condition:
      Bool:
        aws:SecureTransport: false
```

### 5.4 CI/CD Security

**Pipeline Security Controls**

```yaml
# GitHub Actions Security
jobs:
  deploy:
    permissions:
      contents: none
      id-token: write
      pull-requests: read
    
    steps:
    - name: Verify signature
      run: cosign verify --certificate-identity ${{ env.SIGNING_ID }} app.tar
    
    - name: Pull secrets
      uses: hashicorp/vault-action@v2
      with:
        method: kubernetes
        role: deployer
    
    - name: Scan image
      run: trivy image --severity CRITICAL --exit-code 1 app:latest
```

**Deployment Approval Gates**

| Environment | Approver | Conditions |
|-------------|----------|------------|
| Dev | Automated | SAST pass |
| Staging | Tech Lead | Security tests pass |
| Pre-Prod | Security + DevOps | DAST clean, pen test reviewed |
| Production | Security + DevOps + CISO | Full review, compliance check |

### 5.5 Runtime Security

**Kubernetes Runtime Security**

```yaml
# Falco Rules
- rule: Terminal shell in container
  desc: A shell was spawned in a container
  condition: spawned_process and container and shell_procs
  output: |
    Shell spawned in container
    (user=%user.name container=%container.name 
     image=%container.image.repository)
  priority: WARNING

- rule: Unauthorized process
  desc: A process not in expected paths
  condition: >
    spawned_process and 
    container and 
    proc.name not in (expected_processes)
  priority: CRITICAL
```

**Deployment Validation**

```bash
#!/bin/bash
# Post-deployment security check
set -e

# 1. Verify no default credentials
curl -s https://app.example.com/api/health | jq -r '.version'
test ! -f /app/config/default.json

# 2. Verify TLS configuration
test "$(curl -sI https://app.example.com | grep -i strict-transport)" != ""
openssl s_client -connect app.example.com:443 -brief

# 3. Verify security headers
curl -sI https://app.example.com | grep -E \
  "X-Frame-Options|X-Content-Type-Options|CSP|X-XSS-Protection"

# 4. Verify no debug endpoints
test $(curl -s https://app.example.com/debug | wc -c) = 0

# 5. Verify container security
kubectl exec -it app-pod -- /app/security-check
```

### 5.6 Deployment Topology

**Secure Deployment Architecture**

```
┌─────────────────────────────────────────────────────────────────┐
│                         CI/CD Pipeline                          │
│  Build → Test → Scan → Sign → Deploy                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Container Registry                          │
│  Signed Images ← Cosign ← Notary                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Network Policies                       │   │
│  │  Default Deny → Explicit Allow                           │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│  ┌──────────┐    ┌───────────────────┐    ┌──────────────┐    │
│  │  Ingress  │    │   Application     │    │   Services    │    │
│  │  (WAF)    │───▶│   (Hardened)      │───▶│  (mTLS)       │    │
│  └──────────┘    └───────────────────┘    └──────────────┘    │
│                              │                                   │
│                              ▼                                   │
│                      ┌──────────────┐                           │
│                      │   Database   │                           │
│                      │  (Encrypted) │                           │
│                      └──────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Monitoring & Alerting                        │
│  SIEM ← Metrics ← Logs ← Traces                                │
└─────────────────────────────────────────────────────────────────┘
```

## GraphX Integration

**Deployment Tracking**

```bash
# Verify graph up-to-date
GraphX: GetStatus()

# Get new components deployed
GraphX: Query updated_at > last_deployment

# Verify security posture
GraphX: Query type="api"
GraphX: SearchNodes("auth", "admin")
GraphX: GetTrustBoundaries()
```

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Deployment Checklist | Pre/post deployment steps |
| Hardening Guide | Security configuration guide |
| Secrets Management | Vault configuration, rotation |
| Monitoring Setup | Alerts, dashboards |
| Incident Runbooks | Deployment-specific responses |

## Exit Criteria

- [ ] Containers run non-root
- [ ] All secrets in Vault
- [ ] Images signed
- [ ] Infrastructure hardened
- [ ] Monitoring active
- [ ] GraphX deployment mapped