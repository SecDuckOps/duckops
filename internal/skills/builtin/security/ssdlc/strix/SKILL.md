---
name: strix-threat-engine
description: AI-native Threat Modeling and Security Architecture Intelligence Engine
---

# STRIX: Security Threat Modeling Engine

Enterprise-scale threat modeling, attack simulation, and security architecture analysis.

## Overview

STRIX combines deterministic analysis with graph reasoning to provide accurate, actionable threat intelligence.

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
strix threat-model ./repo

# Attack paths analysis
strix attack-paths

# Trust boundaries
strix trust-boundaries

# Architecture review
strix architecture review

# Abuse cases
strix abuse-cases

# Security graph
strix security-graph

# Incremental update
strix update --diff HEAD~1
```

## Quick Start

```bash
# Initialize threat modeling
strix init

# Analyze repository
strix analyze --depth full

# Generate report
strix report --format markdown

# Watch mode
strix watch --patterns "**/*.go"
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
| `strix-engine` | Core threat modeling engine |
| `strix-parser` | Multi-source architecture extraction |
| `strix-threat` | Threat enumeration and scoring |
| `strix-mitigation` | Security recommendations |
| `strix-visualization` | Attack path rendering |
| `strix-cli` | Command interface |