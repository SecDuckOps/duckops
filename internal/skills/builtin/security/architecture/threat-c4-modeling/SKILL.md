---
name: threat-c4-modeling
description: Deterministic-first threat modeling and C4 analysis skill for subagents.
---

# Threat + C4 Modeling

Run deterministic extraction first, then reason:

1. `analyze_architecture`
2. `generate_c4_model`
3. `generate_threat_model`
4. `generate_attack_paths`
5. `export_security_report`

Use IR and graph evidence as source of truth. Do not infer components not present in the generated model.

## Core analysis techniques

- Asset-centric decomposition:
  - Identify crown-jewel assets (credentials, tokens, PII, payment data, signing keys).
  - Map where each asset is created, transformed, stored, and exposed.
- Entry-point decomposition:
  - Enumerate all inbound entry points (public APIs, webhooks, CI triggers, admin paths, agent tools).
  - Track which trust boundary each entry point crosses first.
- Data-flow-first modeling:
  - Prioritize review of flows crossing trust zones and flows with privileged identity changes.
  - Mark protocol + auth + sensitivity on each critical flow.
- Attack precondition mapping:
  - For each high-risk finding, list minimum attacker preconditions (auth state, network position, role).
  - Use preconditions to rank realistic exploit paths.

## Threat modeling techniques

- STRIDE per element type:
  - Services: identity, lateral movement, privilege scope.
  - APIs: authn/authz, injection vectors, abuse of business operations.
  - Data flows: tampering, replay, downgrade, eavesdropping.
  - Trust boundaries: boundary bypass and policy drift.
- Kill-chain chaining:
  - Build multi-hop paths (initial access -> privilege escalation -> data access -> impact).
  - Prefer graph-supported chains over isolated findings.
- Abuse-case generation:
  - Produce attacker stories for upload, webhook, deserialization, CI artifact poisoning, and tenant escape.
  - Tie each abuse case to at least one concrete endpoint/component.

Required outputs:

- C4 L1/L2/L3 diagrams
- STRIDE threat set
- Trust boundaries
- Attack surfaces
- Abuse cases
- Attack paths
- Mitigation guidance mapped to CWE when available

## Architecture and C4 techniques

- C4 Level 1 (Context):
  - Show external actors/systems and internet-facing trust edges.
- C4 Level 2 (Container):
  - Include runtime responsibilities, auth boundaries, and data stores.
- C4 Level 3 (Component):
  - Highlight controllers/services/repositories and policy enforcement points.
- Trust boundary overlays:
  - Distinguish internet, edge, application, data, and control-plane zones.
  - Explicitly mark boundary crossings and missing controls.

## Attack-surface heuristics

- Flag high-risk patterns:
  - URL fetch/proxy features (SSRF potential)
  - File upload and parsing paths
  - Webhook endpoints with weak signature validation
  - Admin/debug interfaces
  - Dangerous deserialization sinks
  - Metadata service and cloud credential reachability
- Prioritize exploitable public-to-privileged paths first.

## AI/LLM-specific security techniques

- Detect AI components and only then run AI-specific checks:
  - Prompt injection handling in tool-calling agents
  - RAG poisoning paths (indexing, retrieval, trust of documents)
  - Unsafe tool execution boundaries
  - Secret leakage into prompts, memory, logs, and traces
  - Missing policy checks in agent action approval flows

## Validation and confidence rules

- Confirm controls deterministically before final severity:
  - auth middleware coverage
  - rate limiting
  - RBAC least privilege
  - container privilege settings
  - hardcoded secrets
  - public cloud resource exposure
- Mark unsupported claims as `hypothesis`.
- Every high/critical issue must include:
  - evidence path(s)
  - exploit scenario
  - impact
  - mitigation
  - CWE mapping

## Subagent execution pattern

1. Extraction subagent: build IR + evidence.
2. Graph subagent: derive trust boundaries + attack paths.
3. Threat subagent: STRIDE + abuse cases + prioritization draft.
4. Validation subagent: deterministic control checks and confidence scoring.
5. Reporting subagent: produce final report bundle with C4 and findings.
