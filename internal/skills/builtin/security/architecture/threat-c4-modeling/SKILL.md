---
name: threat-c4-modeling
description: Use deterministic IR and graph outputs to produce threat models, C4 models, trust boundaries, and attack paths.
---

# Threat + C4 Modeling

Use this skill when the user asks for threat modeling, architecture review, C4 diagrams, attack paths, or trust boundary analysis.

## Required execution order

1. Run deterministic extraction first through:
   - `analyze_architecture`
   - `generate_c4_model`
   - `generate_threat_model`
2. Only after deterministic outputs exist, perform higher-order reasoning.
3. Never claim findings that are not present in IR, graph, or evidence.

## Output guidance

- Report C4 L1/L2/L3.
- Include STRIDE findings per service/API/flow/boundary.
- Show attack paths as chains.
- Mark low-evidence items as hypotheses.
- Map findings to CWE and practical mitigations.

## Subagent orchestration

- Root agent should delegate:
  - extraction subagent
  - graph/paths subagent
  - reporting subagent
- Merge and deduplicate findings before final report.
