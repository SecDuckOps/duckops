---
name: strix-engine
description: Core threat modeling engine with deterministic analysis and graph reasoning
---

# STRIX Engine Core

Deterministic threat modeling engine with incremental analysis and graph integration.

## Engine Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         STRIX Engine                             │
├─────────────┬─────────────┬─────────────┬──────────────────────┤
│  Parser     │   Threat    │   Risk      │    Mitigation         │
│  Manager    │  Enumerator │   Scorer    │    Generator          │
├─────────────┼─────────────┼─────────────┼──────────────────────┤
│                     Graph Reasoning Layer                       │
├─────────────────────────────────────────────────────────────────┤
│                      Security Knowledge Base                     │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### Architecture Extraction

**Supported Sources**
| Source | Format | Detection |
|--------|--------|-----------|
| Source Code | Go, Python, JS, TS, Rust | AST parsing |
| Kubernetes | YAML, Helm | Resource parsing |
| Docker | Dockerfile | Image analysis |
| Infrastructure | Terraform, Pulumi | IaC parsing |
| APIs | OpenAPI, gRPC | Spec parsing |
| CI/CD | GitHub Actions, GitLab CI | Pipeline analysis |
| Config | JSON, YAML, TOML | Config extraction |
| Documentation | Markdown | NLP extraction |

**Extraction Output**
```yaml
architecture:
  services:
    - name: api-gateway
      type: api
      language: go
      entry_points: [/auth, /api/v1/*]
      trust_boundary: dmz
      dependencies: [auth-service, user-service]
  
  databases:
    - name: postgres-main
      type: postgresql
      sensitivity: high
      trust_boundary: internal
      
  queues:
    - name: event-bus
      type: kafka
      trust_boundary: internal
```

### Trust Boundary Detection

**Boundary Types**
| Boundary | Risk Level | Authentication |
|----------|-----------|---------------|
| Internet | Critical | TLS + Auth |
| DMZ | High | mTLS + Auth |
| Internal Services | Medium | Service Auth |
| Kubernetes Namespace | Low | Network Policy |
| Data Layer | Critical | IAM + Encryption |

**Detection Rules**
```go
// Trust boundary detection patterns
var boundaryPatterns = []struct {
    Pattern string
    Boundary string
    RiskLevel string
}{
    {"**/public/**", "internet", "critical"},
    {"**/api/v*/**", "dmz", "high"},
    {"**/svc-cluster-internal/**", "internal", "medium"},
    {"**/kube-system/**", "kubernetes", "low"},
    {"**/prod-db/**", "data", "critical"},
}
```

## Threat Enumeration Engine

### STRIDE Implementation

**Per-Component Threat Generation**

| Component | Spoofing | Tampering | Repudiation | Information | DoS | EoP |
|-----------|----------|-----------|-------------|-------------|-----|-----|
| API Gateway | Token forgery | Request mod | Missing logs | Data leak | Rate limit | Auth bypass |
| Auth Service | Credential theft | Token tampar | No audit | Secret leak | Resource ex | Privilege esc |
| Database | SQL injection | Data corrupt | No audit | Data leak | Query flood | SQLi RCE |
| Message Queue | Token forgery | Poisoned msg | No audit | Msg leak | Queue flood | Unserialize |
| CI/CD | Secret steal | Pipeline mod | No audit | Cred leak | Resource ex | Privilege esc |
| Kubernetes | RBAC bypass | Config tampar | No audit | Secret leak | Resource ex | Namespace esc |

### Threat Model Structure

```go
type Threat struct {
    ID          string
    Category    STRIDE
    Component   string
    Title       string
    Description string
    
    // Mapping
    CWEs        []string
    ATTACKIDs   []string
    CAPECIDs    []string
    
    // Risk
    Likelihood  float64  // 0.0-1.0
    Impact      float64  // 0.0-1.0
    Severity    string   // CRITICAL/HIGH/MEDIUM/LOW/INFO
    
    // Context
    AttackPath  []string
    TrustZone   string
    DataFlow    string
}
```

## Risk Scoring Engine

### CVSS-like Calculation

```go
func (e *Engine) CalculateRisk(t Threat) RiskScore {
    score := RiskScore{}
    
    // Base score from exploitability
    score.Base = calculateBaseScore(t)
    
    // Temporal adjustments
    score.Temporal = calculateTemporal(t)
    
    // Environmental (from GraphX context)
    score.Environmental = calculateEnvironmental(t, e.graph)
    
    // Final score
    score.Final = combineScores(score.Base, score.Temporal, score.Environmental)
    
    return score
}

func calculateBaseScore(t Threat) float64 {
    // Attack complexity
    attackComplexity := 0.77  // AC:H
    if t.RequiresSpecialConditions {
        attackComplexity = 0.44  // AC:L
    }
    
    // Privilege required
    privReq := 0.85  // PR:H
    if t.NoPrivilegeRequired {
        privReq = 0.85  // PR:N
    }
    
    // Scope changed
    scopeChanged := 1.15  // Modified scope
    
    return 8.22 * t.AttackVector * attackComplexity * 
           t.PrivilegeRequired * t.UserInteraction * scopeChanged
}
```

### Risk Factors

| Factor | Weight | Calculation |
|--------|--------|-------------|
| Internet Exposure | 1.5x | Directly accessible |
| Trust Boundary | 2.0x | Crosses boundary |
| Data Sensitivity | 1.3x | PII/credentials |
| Auth Strength | 0.5x | Strong auth reduces |
| Blast Radius | Variable | Nodes affected |

## Attack Path Analysis

### Path Discovery

```go
type AttackPath struct {
    ID          string
    Source      string      // Entry point
    Target      string      // Objective
    Nodes       []PathNode  // Intermediate steps
    Complexity  string      // LOW/MEDIUM/HIGH
    Detections  []string   // Security controls
    Mitigations []string    // Recommendations
}

type PathNode struct {
    ID         string
    Component  string
    Action     string
    Technique  string  // ATT&CK ID
    Prerequisite string
}

func (e *Engine) DiscoverPaths(target string, maxDepth int) []AttackPath {
    paths := []AttackPath{}
    
    // BFS from entry points to target
    queue := newPathQueue()
    queue.Push(Path{Source: "internet", Steps: []PathNode{}})
    
    for queue.Len() > 0 && len(paths) < maxPaths {
        path := queue.Pop()
        
        if path.Reaches(target) {
            paths = append(paths, buildAttackPath(path))
            continue
        }
        
        if path.Depth() >= maxDepth {
            continue
        }
        
        // Expand with possible next steps
        for _, next := range e.getNextNodes(path.Current) {
            if !path.Contains(next) {
                queue.Push(path.Extend(next))
            }
        }
    }
    
    return paths
}
```

### Example Paths

**Cloud Credential Theft**
```
Internet → [SSRF on API] → [Metadata Access] → [Temp Credentials] 
    → [S3 Write Access] → [Data Exfiltration]
```

**Kubernetes Takeover**
```
CI Runner Compromise → [Pipeline Exec] → [Service Account Token] 
    → [K8s API Access] → [Cluster Admin] → [All Namespaces]
```

## Graph Integration

### Node Metadata

```bash
# Query high-risk nodes
GraphX: Query type="service" AND threat_count > 5

# Get blast radius
GraphX: GetBlastRadius("api-gateway")

# Get attack paths
GraphX: Query edge_type="attack_path"

# Get trust zones
GraphX: GetNodesByMetadata("trust_zone=internet")
```

### Incremental Updates

```go
func (e *Engine) UpdateIncremental(changes []FileChange) {
    affectedComponents := e.identifyAffectedComponents(changes)
    
    // Remove old threats for affected components
    for _, comp := range affectedComponents {
        e.removeThreatsFor(comp)
    }
    
    // Re-analyze affected components
    for _, comp := range affectedComponents {
        threats := e.enumerateThreats(comp)
        for _, threat := range threats {
            e.addThreat(threat)
        }
    }
    
    // Update attack paths
    e.recomputeAttackPaths()
    
    // Update risk scores
    e.recalculateRiskScores()
    
    // Sync to GraphX
    e.syncToGraph()
}
```

## Output Formats

| Format | Use Case |
|--------|----------|
| JSON | API, automation |
| SARIF | CI/CD integration |
| Markdown | Documentation |
| Mermaid | Diagrams |
| CSV | Analysis tools |

## Performance

- Initial scan: ~5 min for 10K files
- Incremental: ~30 sec for 100 changes
- Memory: ~500MB for large repos
- CPU: Multi-core parallel parsing

## Exit Criteria

- [ ] Architecture fully extracted
- [ ] Trust boundaries mapped
- [ ] STRIDE threats enumerated
- [ ] Attack paths discovered
- [ ] Risk scores calculated
- [ ] GraphX updated