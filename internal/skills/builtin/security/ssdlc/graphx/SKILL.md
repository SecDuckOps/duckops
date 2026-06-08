---
name: graphx-threat-engine
description: AI-native Threat Modeling and Security Architecture Intelligence Engine
---

# graphx: Security Threat Modeling Engine

Enterprise-scale threat modeling, attack simulation, and security architecture analysis.

## Overview

graphx combines deterministic analysis with graph reasoning to provide accurate, actionable threat intelligence.

```
Repository → Parsing Layer → Architecture Extraction
    ↓
Trust Boundary Detection → Threat Enumeration
    ↓
Attack Path Analysis → Risk Scoring
    ↓
Mitigation Generation → Security Graph Integration
```

## Core Capabilities

| Capability | Deterministic | Graph-Based | AI-Enhanced |
|------------|--------------|-------------|-------------|
| Architecture Extraction | AST parsing, IaC analysis | Dependency graphs | Documentation understanding |
| Trust Boundary Detection | Config parsing | Cross-service flows | Natural language analysis |
| Threat Enumeration | Pattern matching | Attack paths | MITRE ATT&CK mapping |
| Risk Scoring | CVSS calculation | Blast radius analysis | Context-aware severity |
| Mitigation Generation | Best practices | Architecture-aware | Framework-specific |

## Supported Methodologies

| Methodology | Use Case |
|-------------|----------|
| STRIDE | General threat enumeration |
| DFD Threat Modeling | Data flow analysis |
| Attack Trees | Complex attack scenarios |
| MITRE ATT&CK | Adversary emulation |
| Kill Chain | Attack phase analysis |
| PASTA | Risk-centric analysis |
| LINDDUN | Privacy threats |
| AI/LLM Threats | AI system security |

## CLI Commands

```bash
# Full threat model
graphx threat-model ./repo

# Attack paths analysis
graphx attack-paths

# Trust boundaries
graphx trust-boundaries

# Architecture review
graphx architecture review

# Abuse cases
graphx abuse-cases

# Security graph
graphx security-graph

# Incremental update
graphx update --diff HEAD~1
```

## Quick Start

```bash
# Initialize threat modeling
graphx init

# Analyze repository
graphx analyze --depth full

# Generate report
graphx report --format markdown

# Watch mode
graphx watch --patterns "**/*.go"
```

## Integration Points

**GraphX**
```bash
GraphX: Query type="api" OR type="function"
GraphX: GetTrustBoundaries()
GraphX: GetBlastRadius(node_id)
GraphX: GetThreatModel()
```

**SSDLC Phases**
- Phase 1: Requirements → Threat model generation
- Phase 2: Design → Trust boundary analysis
- Phase 3: Development → Real-time threat tracking
- Phase 6: Operations → Attack path monitoring

## Sub-Engine Reference

| Engine | Purpose |
|--------|---------|
| `graphx-engine` | Core threat modeling engine |
| `graphx-parser` | Multi-source architecture extraction |
| `graphx-threat` | Threat enumeration and scoring |
| `graphx-mitigation` | Security recommendations |
| `graphx-visualization` | Attack path rendering |
| `graphx-cli` | Command interface |