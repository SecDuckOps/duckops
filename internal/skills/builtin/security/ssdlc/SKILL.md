---
name: ssdlc-master
description: SSDLC Master - Autonomous Secure Software Engineering & SSDLC Intelligence System
---

# Autonomous Secure Software Engineering & SSDLC Intelligence System

An AI-native platform for secure software development lifecycle management with GraphX-powered security intelligence.

## Vision

Transform software engineering into an inherently secure process through AI-driven automation, continuous security validation, and architectural intelligence.

## Core Capabilities

| Capability | Description | AI Integration |
|------------|-------------|----------------|
| **Secure Design** | Threat modeling, architecture analysis | GraphX architecture discovery |
| **Automated Security Analysis** | SAST, DAST, dependency scanning | Continuous validation |
| **Architecture Intelligence** | Trust boundaries, data flows | GraphX knowledge graph |
| **Threat Modeling** | STRIDE, attack paths, blast radius | Real-time threat analysis |
| **Continuous Security Validation** | CI/CD gates, policy enforcement | Automated compliance |
| **Performance-Aware Engineering** | Resource optimization, scaling | Performance graph analysis |
| **Automated QA & Testing** | Security tests, fuzzing, validation | Intelligent test generation |
| **Runtime Security Monitoring** | SIEM, RASP, threat detection | GraphX context awareness |

## SSDLC Phases

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                          SSDLC Intelligence Layer                            │
├─────────┬─────────┬─────────┬─────────┬─────────┬─────────┬───────────────┤
│ Phase 1 │ Phase 2 │ Phase 3 │ Phase 4 │ Phase 5 │ Phase 6 │   Phase 7      │
│Require- │  Secure │ Secure  │ Security│  Secure │Operations│ Maintenance   │
│ments    │ Design  │ Develop-│ Testing │Deploy- │& Monitor-│& Continuous   │
│         │         │ment     │         │ment     │ing       │Improvement    │
└─────────┴─────────┴─────────┴─────────┴─────────┴─────────┴───────────────┘
                                      │
                              ┌───────────────┐
                              │    GraphX     │
                              │ Security      │
                              │ Knowledge     │
                              │    Graph      │
                              └───────────────┘
```

## Phase Orchestration

### Security Gates

Each phase gates the next:

| Gate | Transition | Criteria |
|------|------------|----------|
| Design Gate | Phase 1 → 2 | Threat model approved, risks accepted |
| Build Gate | Phase 2 → 3 | Secure design verified |
| Test Gate | Phase 3 → 4 | SAST clean, code reviewed |
| Deploy Gate | Phase 4 → 5 | All HIGH/CRITICAL resolved |
| Release Gate | Phase 5 → 6 | Deployment validated |
| Health Gate | Phase 6 → 7 | Monitoring active, metrics healthy |

### GraphX Integration Points

```
Project Start
    ↓
GraphX.Initialize() → Repository Analysis
    ↓
GraphX.BuildArchitecture() → C4 + DFD Generation
    ↓
GraphX.ThreatModel() → STRIDE + Attack Paths
    ↓
GraphX.TrackSecurity() → Continuous Code Analysis
    ↓
GraphX.AnalyzeBlastRadius() → Change Impact
    ↓
GraphX.Monitor() → Runtime Security Context
    ↓
GraphX.Improve() → Continuous Feedback Loop
```

## Phase Reference

Load specific phase skills for detailed guidance:

```
/skill:ssdlc-phase1-requirements   # Requirements gathering, risk assessment
/skill:ssdlc-phase2-design          # Secure architecture, threat modeling
/skill:ssdlc-phase3-development     # Secure coding, SAST integration
/skill:ssdlc-phase4-testing         # Security testing, DAST, pen testing
/skill:ssdlc-phase5-deployment      # Hardening, secrets management
/skill:ssdlc-phase6-operations      # Monitoring, incident response
/skill:ssdlc-phase7-maintenance     # Patch management, improvement
```

## DevSecOps Integration

### Continuous Security Pipeline

```yaml
devsecops_pipeline:
  stages:
    requirements:
      - risk_assessment
      - compliance_mapping
      - graphx_architecture
    
    design:
      - threat_modeling
      - architecture_review
      - trust_boundary_mapping
    
    development:
      - sast_integration
      - dependency_scanning
      - secrets_detection
      - code_review
    
    testing:
      - dast_automation
      - penetration_testing
      - blast_radius_analysis
      - fuzz_testing
    
    deployment:
      - artifact_signing
      - container_hardening
      - secrets_rotation
      - deployment_validation
    
    operations:
      - runtime_monitoring
      - siem_integration
      - incident_response
      - drift_detection
```

## Quick Reference

### Phase Checklist Summary

**Phase 1 - Requirements**
- [ ] Security requirements documented
- [ ] Compliance requirements mapped
- [ ] Data classification complete
- [ ] Risk assessment performed
- [ ] Abuse cases identified
- [ ] Trust boundaries defined

**Phase 2 - Design**
- [ ] Threat model created
- [ ] C4/DFD diagrams generated
- [ ] Attack surface analyzed
- [ ] Security patterns applied
- [ ] Architecture reviewed
- [ ] Trust boundaries mapped

**Phase 3 - Development**
- [ ] Secure coding standards followed
- [ ] SAST integration in CI
- [ ] Dependency scanning active
- [ ] No secrets in code
- [ ] Code review completed
- [ ] GraphX tracking enabled

**Phase 4 - Testing**
- [ ] DAST scan passed
- [ ] Penetration testing complete
- [ ] Blast radius analyzed
- [ ] HIGH/CRITICAL resolved
- [ ] Security tests automated
- [ ] GraphX threat model updated

**Phase 5 - Deployment**
- [ ] Artifacts signed
- [ ] Containers hardened
- [ ] Secrets in Vault
- [ ] Infrastructure validated
- [ ] Monitoring active
- [ ] GraphX deployment mapped

**Phase 6 - Operations**
- [ ] SIEM connected
- [ ] Alerts configured
- [ ] Incident playbooks ready
- [ ] Metrics tracking
- [ ] Drift detection active
- [ ] GraphX monitoring updated

**Phase 7 - Maintenance**
- [ ] Patches current
- [ ] Dependencies updated
- [ ] Threat model reviewed
- [ ] Architecture validated
- [ ] Technical debt addressed
- [ ] GraphX knowledge refreshed

## Core Principles

1. **Security by Default**: Every component starts secure
2. **Defense in Depth**: Layered security at each phase
3. **Shift Left**: Resolve issues early, reduce cost
4. **Continuous Validation**: Security never stops
5. **Graph-Native**: GraphX as persistent security memory
6. **AI-Augmented**: Intelligent automation throughout
7. **Zero Trust**: Verify everything, trust nothing

## Mindset

Security is not added—it's built in from the start.

When making decisions, ask:
1. Does this maintain security posture?
2. Is this tracked in GraphX?
3. Does this violate security gates?
4. What's the blast radius if this fails?
5. How do we verify this is secure?

## Platform Behavior

The system operates as:

**"AI-native DevSecOps platform powered by a persistent security knowledge graph."**

Capabilities:
- Autonomous threat modeling
- Continuous security validation
- Architecture-aware code generation
- Intelligent security testing
- Real-time security monitoring
- Predictive security intelligence
- Zero-trust enforcement
- SSDLC automation