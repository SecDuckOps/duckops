---
name: strix-cli
description: STRIX command-line interface for threat modeling, attack analysis, and security visualization
---

# STRIX CLI

Command-line interface for threat modeling operations.

## Installation

```bash
# Install STRIX CLI
go install github.com/strix/strix@latest

# Or use Docker
docker run --rm -v $(pwd):/repo ghcr.io/strix/strix threat-model /repo

# Shell completion
strix completion bash > /etc/bash_completion.d/strix
```

## Global Options

```bash
--repo string       Repository path (default: .)
--format string     Output format (json|yaml|markdown|html) (default: json)
--depth string      Analysis depth (quick|full|detailed) (default: full)
--quiet            Suppress non-essential output
--verbose          Enable verbose logging
--config string    Config file path (default: .strix.yaml)
```

## Commands

### Threat Model Generation

```bash
# Generate full threat model
strix threat-model --repo ./myapp

# Quick scan
strix threat-model --depth quick

# Output formats
strix threat-model --format markdown --output threat-model.md
strix threat-model --format json --output threat-model.json

# Include specific methodologies
strix threat-model --methods stride,attack-trees,att&ck

# Filter by component
strix threat-model --components api-gateway,auth-service
```

### Attack Path Analysis

```bash
# Discover attack paths
strix attack-paths

# Find paths to specific target
strix attack-paths --target "database"

# Limit path depth
strix attack-paths --max-depth 5

# Find paths with specific techniques
strix attack-paths --technique T1552.001

# Output attack path graph
strix attack-paths --output paths.mmd --format mermaid
```

### Trust Boundary Analysis

```bash
# Visualize trust boundaries
strix trust-boundaries

# Show boundary details
strix trust-boundaries --detail

# Filter boundaries by risk level
strix trust-boundaries --risk critical,high

# Export boundary map
strix trust-boundaries --format mermaid --output boundaries.mmd
```

### Architecture Review

```bash
# Review architecture for security
strix architecture review

# Generate C4 diagrams
strix architecture c4

# Generate DFD
strix architecture dfd --level 2

# Export architecture map
strix architecture export --format json --output arch.json
```

### Abuse Case Generation

```bash
# Generate abuse cases
strix abuse-cases

# Generate for specific component
strix abuse-cases --component "ai-agent"

# AI/LLM specific abuse cases
strix abuse-cases --ai-focused

# List all abuse case templates
strix abuse-cases --list-templates
```

### Security Graph

```bash
# Interactive graph exploration
strix security-graph

# Query graph
strix security-graph query --node-type service --risk high

# Get blast radius
strix security-graph blast-radius --node api-gateway

# Find entry points
strix security-graph entry-points

# Export graph
strix security-graph export --format graphml --output graph.graphml
```

### Report Generation

```bash
# Full threat report
strix report --format markdown --output report.md

# Executive summary
strix report --type executive

# Detailed technical report
strix report --type technical --format html --output report.html

# Risk matrix
strix report --type risk-matrix --format html --output matrix.html
```

## Subcommands

### Initialize

```bash
# Initialize STRIX in repository
strix init

# With custom config
strix init --config .strix.custom.yaml

# Templates
strix init --template api-service
strix init --template microservices
strix init --template ai-platform
```

### Analyze

```bash
# Analyze specific components
strix analyze --components auth,users,payments

# Analyze changes only
strix analyze --diff HEAD~1

# Incremental update
strix analyze --incremental

# Parallel analysis
strix analyze --parallel 8
```

### Config

```bash
# View current config
strix config show

# Set options
strix config set --key analysis.depth --value detailed
strix config set --key output.format --value markdown

# Export config
strix config export --output .strix.yaml
```

### Watch

```bash
# Watch mode
strix watch

# With file patterns
strix watch --patterns "**/*.go" "**/k8s/**/*.yaml"

# Debounce delay
strix watch --delay 5s
```

## Output Examples

### Threat Model Output (JSON)

```json
{
  "repository": "./myapp",
  "generated": "2024-01-15T10:30:00Z",
  "summary": {
    "total_components": 15,
    "total_threats": 47,
    "critical": 3,
    "high": 12,
    "medium": 22,
    "low": 10
  },
  "components": [
    {
      "id": "api-gateway",
      "type": "service",
      "trust_boundary": "dmz",
      "threats": [
        {
          "id": "STRIDE-001",
          "title": "JWT Token Forgery",
          "category": "Spoofing",
          "severity": "HIGH",
          "cwE": ["CWE-347"],
          "attack_id": "T1606.002"
        }
      ]
    }
  ],
  "attack_paths": [
    {
      "id": "PATH-001",
      "source": "internet",
      "target": "database",
      "nodes": ["internet", "api-gateway", "ssrf", "metadata", "creds", "database"],
      "complexity": "HIGH"
    }
  ],
  "recommendations": [
    {
      "threat_id": "STRIDE-001",
      "mitigation": "Implement JWT signature validation with RS256",
      "priority": "HIGH"
    }
  ]
}
```

### Mermaid Diagram Output

```bash
$ strix diagram --type trust-boundaries --format mermaid

%% Generated by STRIX
graph TB
    subgraph Internet["🌐 Internet Risk: CRITICAL"]
        Client[User]
        Attacker[Attacker]
    end
    subgraph DMZ["DMZ Risk: HIGH"]
        WAF[WAF]
        LB[Load Balancer]
    end
    %% ... more elements
```

## Configuration File

```yaml
# .strix.yaml
strix:
  version: "1.0"
  
analysis:
  depth: full
  languages:
    - go
    - python
    - typescript
  frameworks:
    - kubernetes
    - terraform
    - docker
  
  ignore:
    - "**/vendor/**"
    - "**/node_modules/**"
    - "**/*_test.go"
  
threats:
  methods:
    - stride
    - attack-trees
    - att&ck
  severity_threshold: MEDIUM
  
output:
  format: json
  include_evidence: true
  export_graph: true
  
graph:
  sync_enabled: true
  update_on_change: true
  
filters:
  min_severity: LOW
  exclude_components: []
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Analysis failed |
| 2 | Invalid arguments |
| 3 | Repository not found |
| 4 | Configuration error |
| 5 | Analysis timeout |

## Examples

### Full Threat Modeling Workflow

```bash
#!/bin/bash
# Threat modeling workflow

set -e

# Initialize
strix init

# Generate threat model
strix threat-model --format json --output threat-model.json

# Find attack paths
strix attack-paths --max-depth 4 --output paths.json

# Generate visualizations
strix diagram --type architecture --format mermaid --output arch.mmd
strix diagram --type attack-paths --format html --output paths.html

# Generate report
strix report --type technical --format markdown --output report.md

# Integrate with CI/CD
if [ "$CI" = "true" ]; then
  strix threat-model --format sarif --output gl-sarif.json
fi
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `strix threat-model` | Full threat analysis |
| `strix attack-paths` | Attack path discovery |
| `strix trust-boundaries` | Boundary visualization |
| `strix architecture review` | Architecture analysis |
| `strix abuse-cases` | Abuse case generation |
| `strix security-graph` | Graph exploration |
| `strix report` | Generate reports |