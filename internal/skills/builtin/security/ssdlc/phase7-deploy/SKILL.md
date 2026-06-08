---
name: phase7-deploy
description: Phase 7 - Deploy & Infrastructure security, Infrastructure as Code (IaC) validation, IAM reviews, and Kubernetes hardening
---

# Phase 7: Deploy & Infrastructure

Verify infrastructure configurations, audit Infrastructure as Code (IaC) templates, enforce least-privilege IAM policies, and secure Kubernetes deployment templates.

## Objectives

- Perform automated scans on Terraform templates
- Validate IAM policies against wildcard roles and privileges
- Hardened Kubernetes pod and deployment configurations
- Detect and prevent cloud infrastructure misconfigurations
- Load cloud-specific security skills

## Phase Activities

### 7.1 Infrastructure as Code (IaC) Security Scans

Always scan configuration files before deploying resources:

```bash
# Scan IaC templates using Trivy config scanner
trivy config . --severity HIGH,CRITICAL

# Alternative using checkov
checkov -d . --framework terraform
```

### 7.2 Terraform Security Auditing

Audit Terraform files (`*.tf`) for common misconfigurations:

*   **S3 Buckets**: Flag S3 buckets without versioning, default encryption, or public access block configurations:
    ```hcl
    # ❌ Insecure bucket definition
    resource "aws_s3_bucket" "bad" {
      bucket = "my-bucket"
    }
    ```
*   **Security Groups**: Flag security groups permitting unrestricted ingress traffic from the public internet (`0.0.0.0/0`) on sensitive management ports (e.g., SSH 22, RDP 3389):
    ```hcl
    # ❌ Insecure ingress rule
    ingress {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
    ```

### 7.3 IAM Policy Hardening

Audit Identity & Access Management (IAM) templates:

*   **Wildcard Actions**: Flag wildcard statements permitting all actions:
    ```json
    "Action": "*"
    ```
*   **Wildcard Resources**: Flag permissions applied to all resources:
    ```json
    "Resource": "*"
    ```
*   **AssumeRole Permission**: Flag insecure `sts:AssumeRole` statements configured with `*` principals.

### 7.4 Kubernetes Deployment Hardening

Audit Kubernetes templates (`*.yaml`):

*   **Privileged Containers**: Flag pods configured with root capabilities:
    ```yaml
    privileged: true
    ```
*   **Missing SecurityContext**: Check that `securityContext` is defined both at the pod and container levels.
*   **ServiceAccount Tokens**: Ensure `automountServiceAccountToken: false` is defined unless the pod explicitly requires API access.
*   **ClusterRoleBindings**: Flag overly permissive bindings mapping system permissions to default service accounts.

### 7.5 Cloud & Metadata Protection

- **Metadata Service Exposure**: Verify if cloud VM configurations (e.g. AWS EC2) permit access to metadata services without IMDSv2 configuration.
- **Dynamic Skill Loading**: Always load the `cloud/aws-security` skill when auditing AWS templates to identify complex infrastructure risks.

## Exit Criteria

- [ ] IaC security scan completed.
- [ ] Terraform configs audited.
- [ ] IAM policies hardened.
- [ ] Kubernetes manifests hardened.
- [ ] Cloud configurations validated.
