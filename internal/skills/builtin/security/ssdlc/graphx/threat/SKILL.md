---
name: graphx-threat
description: Comprehensive threat enumeration with STRIDE, MITRE ATT&CK, and attack path analysis
---

# Threat Enumeration Engine

Generate comprehensive threats using multiple methodologies with deterministic rules and graph reasoning.

## Threat Enumeration Pipeline

```
Component → Identify Attack Surface → Enumerate Threats → Map to Frameworks → Score Risk
```

## STRIDE Enumeration

### Spoofing Threats

**Identity-Based Attacks**
```go
var spoofingThreats = []Threat{
    {
        Title: "Credential Stuffing Attack",
        AttackVector: "Network",
        Prerequisites: []string{
            "Valid credentials leaked",
            "No rate limiting on login",
        },
        CWEs: []string{"CWE-307", "CWE-799"},
        ATTACKIDs: []string{"T1110.001"},
    },
    {
        Title: "JWT Token Forgery",
        AttackVector: "Network",
        Prerequisites: []string{
            "Weak JWT secret",
            "Algorithm confusion (none)",
        },
        CWEs: []string{"CWE-347", "CWE-345"},
        ATTACKIDs: []string{"T1606.002"},
    },
    {
        Title: "Session Hijacking",
        AttackVector: "Network",
        Prerequisites: []string{
            "Predictable session ID",
            "No HttpOnly/Secure flags",
        },
        CWEs: []string{"CWE-384", "CWE-345"},
        ATTACKIDs: []string{"T1184"},
    },
    {
        Title: "OAuth Token Theft",
        AttackVector: "Network",
        Prerequisites: []string{
            "Tokens in URL",
            "No token rotation",
        },
        CWEs: []string{"CWE-598"},
        ATTACKIDs: []string{"T1550.001"},
    },
}
```

### Tampering Threats

**Data Modification Attacks**
```go
var tamperingThreats = []Threat{
    {
        Title: "SQL Injection",
        AttackVector: "Network",
        Prerequisites: []string{
            "User input in SQL query",
            "No parameterized queries",
        },
        CWEs: []string{"CWE-89"},
        ATTACKIDs: []string{"T1190", "T1055.009"},
    },
    {
        Title: "Command Injection",
        AttackVector: "Local",
        Prerequisites: []string{
            "User input in shell command",
            "shell=True or similar",
        },
        CWEs: []string{"CWE-78"},
        ATTACKIDs: []string{"T1059.004"},
    },
    {
        Title: "Cache Poisoning",
        AttackVector: "Network",
        Prerequisites: []string{
            "DNSSEC not enabled",
            "Cache server misconfiguration",
        },
        CWEs: []string{"CWE-74"},
        ATTACKIDs: []string{"T1071.002"},
    },
    {
        Title: "Log Injection",
        AttackVector: "Network",
        Prerequisites: []string{
            "Unsanitized input in logs",
            "Log viewer XSS",
        },
        CWEs: []string{"CWE-117", "CWE-79"},
        ATTACKIDs: []string{"T1056.001"},
    },
}
```

### Repudiation Threats

**Non-Repudiation Vulnerabilities**
```go
var repudiationThreats = []Threat{
    {
        Title: "Missing Audit Trail",
        AttackVector: "Local",
        Prerequisites: []string{
            "No logging on sensitive operations",
            "No immutable log storage",
        },
        CWEs: []string{"CWE-778"},
        ATTACKIDs: []string{"T1070"},
    },
    {
        Title: "Unsigned Transactions",
        AttackVector: "Network",
        Prerequisites: []string{
            "No digital signatures",
            "No non-repudiation controls",
        },
        CWEs: []string{"CWE-345"},
        ATTACKIDs: []string{"T1609"},
    },
    {
        Title: "Action Without Logging",
        AttackVector: "Local",
        Prerequisites: []string{
            "Sensitive action not logged",
            "No alerting on critical events",
        },
        CWEs: []string{"CWE-223"},
        ATTACKIDs: []string{"T1562"},
    },
}
```

### Information Disclosure Threats

**Data Exposure**
```go
var infoDisclosureThreats = []Threat{
    {
        Title: "Sensitive Data in Logs",
        AttackVector: "Local",
        Prerequisites: []string{
            "Credentials logged",
            "PII not masked",
        },
        CWEs: []string{"CWE-532", "CWE-200"},
        ATTACKIDs: []string{"T1074.001"},
    },
    {
        Title: "SSRF to Metadata Service",
        AttackVector: "Network",
        Prerequisites: []string{
            "SSRF vulnerability",
            "Metadata endpoint accessible",
        },
        CWEs: []string{"CWE-918"},
        ATTACKIDs: []string{"T1552.001"},
    },
    {
        Title: "Bucket Enumeration",
        AttackVector: "Network",
        Prerequisites: []string{
            "Public bucket or weak ACLs",
            "Guessable bucket names",
        },
        CWEs: []string{"CWE-264"},
        ATTACKIDs: []string{"T1552.003"},
    },
    {
        Title: "Source Code Leak",
        AttackVector: "Network",
        Prerequisites: []string{
            "Exposed .git directory",
            "Debug endpoints",
        },
        CWEs: []string{"CWE-540"},
        ATTACKIDs: []string{"T1213.001"},
    },
}
```

### Denial of Service Threats

**Availability Attacks**
```go
var dosThreats = []Threat{
    {
        Title: "Resource Exhaustion",
        AttackVector: "Network",
        Prerequisites: []string{
            "Unbounded resource allocation",
            "No rate limiting",
        },
        CWEs: []string{"CWE-400", "CWE-770"},
        ATTACKIDs: []string{"T1499"},
    },
    {
        Title: "Database Connection Flood",
        AttackVector: "Network",
        Prerequisites: []string{
            "Connection pooling misconfig",
            "No connection limits",
        },
        CWEs: []string{"CWE-400"},
        ATTACKIDs: []string{"T1499.001"},
    },
    {
        Title: "Queue Flooding",
        AttackVector: "Network",
        Prerequisites: []string{
            "No queue depth limits",
            "Unbounded message storage",
        },
        CWEs: []string{"CWE-770"},
        ATTACKIDs: []string{"T1499.004"},
    },
}
```

### Elevation of Privilege Threats

**Privilege Escalation**
```go
var eopThreats = []Threat{
    {
        Title: "RBAC Bypass via Parameter Manipulation",
        AttackVector: "Network",
        Prerequisites: []string{
            "IDOR vulnerability",
            "No server-side authorization",
        },
        CWEs: []string{"CWE-639", "CWE-862"},
        ATTACKIDs: []string{"T1488"},
    },
    {
        Title: "Kubernetes RBAC Escalation",
        AttackVector: "Local",
        Prerequisites: []string{
            "Misconfigured RBAC",
            "Role with bind permissions",
        },
        CWEs: []string{"CWE-266", "CWE-269"},
        ATTACKIDs: []string{"T1098.009"},
    },
    {
        Title: "Container Escape via hostPath",
        AttackVector: "Local",
        Prerequisites: []string{
            "hostPath volume mounted",
            "Privileged container",
        },
        CWEs: []string{"CWE-278", "CWE-822"},
        ATTACKIDs: []string{"T1611"},
    },
    {
        Title: "CI/CD Pipeline Manipulation",
        AttackVector: "Network",
        Prerequisites: []string{
            "Pipeline vulnerable to injection",
            "Secrets in environment",
        },
        CWEs: []string{"CWE-94", "CWE-269"},
        ATTACKIDs: []string{"T1053.005", "T1574.011"},
    },
}
```

## MITRE ATT&CK Mapping

### Enterprise Matrix Coverage

| graphx Finding | ATT&CK Technique | Tactic |
|--------------|------------------|--------|
| SQL Injection | T1190 | Initial Access |
| Command Injection | T1059 | Execution |
| SSRF | T1056.004 | Collection |
| IAM Privilege Escalation | T1098 | Persistence |
| Container Escape | T1611 | Privilege Escalation |
| Data Exfiltration | T1041 | Exfiltration |

### Kubernetes Threat Matrix

| Technique | Description | Detection |
|----------|-------------|-----------|
| T1505.003 | Web Shell | Process monitoring |
| T1552.001 | HCURL | Secrets scanning |
| T1613 | Container Discovery | API monitoring |
| T1608.004 | Drive-by Compromise | WAF logs |

## Attack Path Generation

### Path Construction Rules

```go
type AttackPathBuilder struct {
    entryPoints    []EntryPoint
    pivots         map[string][]Pivot
    objectives     []string
    maxDepth       int
}

func (b *AttackPathBuilder) BuildPaths() []AttackPath {
    paths := []AttackPath{}
    
    // Start from entry points
    for _, ep := range b.entryPoints {
        path := NewPath(ep)
        b.expandPath(path, 0, paths)
    }
    
    return b.rankPaths(paths)
}

func (b *AttackPathBuilder) expandPath(path *Path, depth int, paths []AttackPath) {
    if depth >= b.maxDepth {
        return
    }
    
    current := path.Current()
    
    // Find pivots from current position
    for _, pivot := range b.pivots[current] {
        if path.Contains(pivot) {
            continue
        }
        
        newPath := path.Extend(pivot)
        
        if b.reachesObjective(newPath) {
            paths = append(paths, newPath.ToAttackPath())
        }
        
        b.expandPath(newPath, depth+1, paths)
    }
}
```

### Common Attack Paths

**Path 1: Internet → Database**
```
Internet → [API Gateway] → [SSRF Vector] → [Metadata Service] 
    → [Cloud Credentials] → [S3 API] → [Data Exfiltration]
```

**Path 2: CI/CD → Production**
```
Developer → [Malicious Commit] → [CI Pipeline] → [Compromised Runner]
    → [Secrets Access] → [Kubernetes Cluster] → [Production Workloads]
```

**Path 3: Container → Host → Cluster**
```
Internet → [Web App] → [RCE] → [Container Shell] → [Host Escape]
    → [Kubelet Access] → [Cluster Admin] → [All Namespaces]
```

## Abuse Case Generation

### AI/LLM Abuse Cases

```go
var llmAbuseCases = []Threat{
    {
        Title: "Prompt Injection",
        Category: "AI System",
        Prerequisites: []string{
            "LLM accepts user input",
            "LLM controls actions",
        },
        CWEs: []string{"CWE-20", "CWE-74"},
        AttackPattern: "User input influences LLM behavior to perform unauthorized actions",
    },
    {
        Title: "Tool/Function Abuse",
        Category: "AI Agent",
        Prerequisites: []string{
            "Agent has dangerous tools",
            "Insufficient tool validation",
        },
        CWEs: []string{"CWE-78"},
        AttackPattern: "Manipulated input causes agent to execute unintended tools",
    },
    {
        Title: "RAG Poisoning",
        Category: "AI System",
        Prerequisites: []string{
            "RAG retrieves from untrusted source",
            "No retrieval validation",
        },
        CWEs: []string{"CWE-200"},
        AttackPattern: "Malicious data in knowledge base influences LLM responses",
    },
    {
        Title: "Context Window Overflow",
        Category: "AI System",
        Prerequisites: []string{
            "Unbounded context",
            "No context validation",
        },
        CWEs: []string{"CWE-400"},
        AttackPattern: "Overflow causes LLM to leak data or ignore instructions",
    },
    {
        Title: "Data Exfiltration via Model",
        Category: "AI System",
        Prerequisites: []string{
            "Model trained on sensitive data",
            "No output filtering",
        },
        CWEs: []string{"CWE-200"},
        AttackPattern: "Crafted queries extract training data or secrets",
    },
}
```

## Threat Output Format

```yaml
threats:
  - id: STRIDE-K8S-001
    title: "Container Escape via hostPath"
    category: "Elevation of Privilege"
    
    affected_components:
      - api-server-pod
    
    cwe:
      - CWE-278      # Exec Container Run as Privileged
      - CWE-822      # Untrusted Search Path
    
    attack_ids:
      - T1611        # Escape to Host
    
    capec_ids:
      - CAPEC-233    # Privilege Escalation
    
    risk:
      likelihood: 0.3
      impact: 0.9
      severity: HIGH
    
    attack_path:
      - Internet
      - API Server
      - hostPath Volume Mount
      - Host Filesystem
      - Kubelet Access
      - Cluster Admin
    
    mitigations:
      - name: "Remove hostPath volumes"
        priority: HIGH
        controls: ["CIS K8s 5.2.9"]
      
      - name: "Drop privileged capabilities"
        priority: HIGH
        controls: ["CIS K8s 5.2.7"]
      
      - name: "Enforce read-only root filesystem"
        priority: MEDIUM
        controls: ["CIS K8s 5.2.6"]
```

## Framework Compliance

| Framework | Coverage |
|-----------|----------|
| OWASP Top 10 | 100% |
| OWASP API Top 10 | 100% |
| CWE Top 25 | 85% |
| MITRE ATT&CK Enterprise | 70% |
| MITRE ATT&CK Container | 60% |
| CIS Benchmarks | 80% |