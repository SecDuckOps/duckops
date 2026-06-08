---
name: graphx-cli
description: graphx command-line interface for threat modeling, attack analysis, and security visualization
---

# graphx CLI

Command-line interface for threat modeling operations.

## Installation

```bash
# Install graphx CLI
go install github.com/graphx/graphx@latest

# Or use Docker
docker run --rm -v $(pwd):/repo ghcr.io/graphx/graphx threat-model /repo

# Shell completion
graphx completion bash > /etc/bash_completion.d/graphx
```

## Global Options

```bash
--repo string       Repository path (default: .)
--format string     Output format (json|yaml|markdown|html) (default: json)
--depth string      Analysis depth (quick|full|detailed) (default: full)
--quiet            Suppress non-essential output
--verbose          Enable verbose logging
--config string    Config file path (default: .graphx.yaml)
```

## Commands

### Threat Model Generation

```bash
# Generate full threat model
graphx threat-model --repo ./myapp

# Quick scan
graphx threat-model --depth quick

# Output formats
graphx threat-model --format markdown --output threat-model.md
graphx threat-model --format json --output threat-model.json

# Include specific methodologies
graphx threat-model --methods stride,attack-trees,att&ck

# Filter by component
graphx threat-model --components api-gateway,auth-service
```

### Attack Path Analysis

```bash
# Discover attack paths
graphx attack-paths

# Find paths to specific target
graphx attack-paths --target "database"

# Limit path depth
graphx attack-paths --max-depth 5

# Find paths with specific techniques
graphx attack-paths --technique T1552.001

# Output attack path graph
graphx attack-paths --output paths.mmd --format mermaid
```

### Trust Boundary Analysis

```bash
# Visualize trust boundaries
graphx trust-boundaries

# Show boundary details
graphx trust-boundaries --detail

# Filter boundaries by risk level
graphx trust-boundaries --risk critical,high

# Export boundary map
graphx trust-boundaries --format mermaid --output boundaries.mmd
```

### Architecture Review

```bash
# Review architecture for security
graphx architecture review

# Generate C4 diagrams
graphx architecture c4

# Generate DFD
graphx architecture dfd --level 2

# Export architecture map
graphx architecture export --format json --output arch.json
```

### Abuse Case Generation

```bash
# Generate abuse cases
graphx abuse-cases

# Generate for specific component
graphx abuse-cases --component "ai-agent"

# AI/LLM specific abuse cases
graphx abuse-cases --ai-focused

# List all abuse case templates
graphx abuse-cases --list-templates
```

### Security Graph

```bash
# Interactive graph exploration
graphx security-graph

# Query graph
graphx security-graph query --node-type service --risk high

# Get blast radius
graphx security-graph blast-radius --node api-gateway

# Find entry points
graphx security-graph entry-points

# Export graph
graphx security-graph export --format graphml --output graph.graphml
```

### Report Generation

```bash
# Full threat report
graphx report --format markdown --output report.md

# Executive summary
graphx report --type executive

# Detailed technical report
graphx report --type technical --format html --output report.html

# Risk matrix
graphx report --type risk-matrix --format html --output matrix.html
```

## Subcommands

### Initialize

```bash
# Initialize graphx in repository
graphx init

# With custom config
graphx init --config .graphx.custom.yaml

# Templates
graphx init --template api-service
graphx init --template microservices
graphx init --template ai-platform
```

### Analyze

```bash
# Analyze specific components
graphx analyze --components auth,users,payments

# Analyze changes only
graphx analyze --diff HEAD~1

# Incremental update
graphx analyze --incremental

# Parallel analysis
graphx analyze --parallel 8
```

### Config

```bash
# View current config
graphx config show

# Set options
graphx config set --key analysis.depth --value detailed
graphx config set --key output.format --value markdown

# Export config
graphx config export --output .graphx.yaml
```

### Watch

```bash
# Watch mode
graphx watch

# With file patterns
graphx watch --patterns "**/*.go" "**/k8s/**/*.yaml"

# Debounce delay
graphx watch --delay 5s
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
$ graphx diagram --type trust-boundaries --format mermaid

%% Generated by graphx
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
# .graphx.yaml
graphx:
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
graphx init

# Generate threat model
graphx threat-model --format json --output threat-model.json

# Find attack paths
graphx attack-paths --max-depth 4 --output paths.json

# Generate visualizations
graphx diagram --type architecture --format mermaid --output arch.mmd
graphx diagram --type attack-paths --format html --output paths.html

# Generate report
graphx report --type technical --format markdown --output report.md

# Integrate with CI/CD
if [ "$CI" = "true" ]; then
  graphx threat-model --format sarif --output gl-sarif.json
fi
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `graphx threat-model` | Full threat analysis |
| `graphx attack-paths` | Attack path discovery |
| `graphx trust-boundaries` | Boundary visualization |
| `graphx architecture review` | Architecture analysis |
| `graphx abuse-cases` | Abuse case generation |
| `graphx security-graph` | Graph exploration |
| `graphx report` | Generate reports |