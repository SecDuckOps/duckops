---
name: ssdlc
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
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SSDLC Intelligence Layer                           │
├───────┬───────┬───────┬───────┬───────┬───────┬───────┬───────┬─────────────┤
│Phase 1│Phase 2│Phase 3│Phase 4│Phase 5│Phase 6│Phase 7│Phase 8│   Phase 9   │
│Require│Secure │Secure │Build &│Testing│Release│Deploy │Operat │  Incident   │
│-ments │Design │Develop│CI/CD  │& Val  │& Stage│& Infra│-ions  │  Response   │
└───────┴───────┴───────┴───────┴───────┴───────┴───────┴───────┴─────────────┘
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
| Requirements Gate | Phase 1 → 2 | Security requirements defined, ASVS mapped, abuse cases drafted |
| Design Gate | Phase 2 → 3 | Threat model approved, STRIDE threats mapped to controls |
| Code Gate | Phase 3 → 4 | Secure coding guidelines implemented, first review complete |
| Build Gate | Phase 4 → 5 | Secrets scan clean, SCA dependencies approved, GHA pins verified |
| Test Gate | Phase 5 → 6 | DAST completed, all HIGH/CRITICAL issues resolved |
| Release Gate | Phase 6 → 7 | Release checklist verified, Trivy image clean, SBOM attached |
| Deploy Gate | Phase 7 → 8 | IaC tfsec/k8s-bench validated, IAM roles mapped to least privilege |
| Ops Gate | Phase 8 → 9 | Runtime monitoring active, alerts threshold set, patch SLAs met |

### GraphX Integration Points

```
Project Start
    ↓
GraphX.Initialize() → Repository Analysis (Phase 1)
    ↓
GraphX.BuildArchitecture() → C4 + DFD Generation (Phase 2)
    ↓
GraphX.ThreatModel() → STRIDE + Attack Paths (Phase 2)
    ↓
GraphX.TrackSecurity() → Continuous Code Analysis (Phase 3/4)
    ↓
GraphX.AnalyzeBlastRadius() → Change Impact (Phase 5)
    ↓
GraphX.Monitor() → Runtime Security Context (Phase 8)
    ↓
GraphX.Improve() → Continuous Feedback Loop (Phase 9)
```

## Phase Reference

Load specific phase skills for detailed guidance:

```
/skill:phase1-requirements   # Requirements gathering, compliance mapping
/skill:phase2-design          # Secure architecture, STRIDE threat modeling
/skill:phase3-development     # Secure coding, framework-specific risks
/skill:phase4-build           # Secrets scan, SCA, CI/CD pipeline validation
/skill:phase5-testing         # Scan modes (quick/standard/deep), DAST, API testing
/skill:phase6-release         # Pre-release gate, container image validation, SBOM
/skill:phase7-deploy          # IaC security, Kubernetes Pod Security, Cloud IAM
/skill:phase8-operations      # Runtime observability, alerts, vuln management
/skill:phase9-incident        # Containment, identification, eradication, recovery
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
      - secure_coding_standards
      - unit_security_tests
    
    build_cicd:
      - secrets_detection
      - dependency_scanning (SCA)
      - github_actions_hardening
    
    testing_validation:
      - dast_automation
      - penetration_testing
      - api_security_testing
      - blast_radius_analysis
    
    release_staging:
      - artifact_signing
      - container_image_scanning
      - sbom_generation
    
    deploy_infrastructure:
      - iac_validation (tfsec)
      - kubernetes_security_context
      - cloud_misconfiguration_scan
    
    operations_runtime:
      - runtime_monitoring (SIEM/RASP)
      - vulnerability_sla_checks
      - secrets_rotation
    
    incident_response:
      - logs_preservation
      - containment_playbooks
      - post_mortem_analysis
```

## Quick Reference

### Phase Checklist Summary

**Phase 1 - Requirements & Planning**
- [ ] Security requirements mapped to OWASP ASVS (L1/L2/L3)
- [ ] Data classification complete (PII, PCI, PHI, Confidential)
- [ ] Compliance requirements identified (GDPR, HIPAA, PCI-DSS)
- [ ] Abuse and misuse cases defined
- [ ] Risk assessment performed

**Phase 2 - Design & Architecture**
- [ ] Threat model created (analyze_architecture)
- [ ] C4/DFD diagrams generated (generate_c4_model)
- [ ] STRIDE threats enumerated per component
- [ ] Attack surfaces and boundaries mapped (generate_attack_paths)
- [ ] Multi-tenancy and cryptographic architecture reviewed

**Phase 3 - Development & Coding**
- [ ] Secure coding standards followed (Framework skills loaded)
- [ ] Vulnerability patterns checked (injections, auth flaws)
- [ ] Unit security tests implemented
- [ ] GraphX tracking enabled

**Phase 4 - Build & CI/CD**
- [ ] Secrets scan completed (secrets__gitleaks / trufflehog)
- [ ] Software Composition Analysis (SCA) completed (sca__trivy_fs)
- [ ] GitHub Actions pinned to commit SHAs with minimal permissions
- [ ] Cache poisoning risks mitigated

**Phase 5 - Testing & Validation**
- [ ] Scan mode selected (/quick, /standard, /deep)
- [ ] DAST scans completed (OWASP ZAP / nuclei)
- [ ] API Security tested against OWASP API Top 10
- [ ] Vulnerability validation and PoC generated
- [ ] Blast radius analysis performed

**Phase 6 - Release & Staging**
- [ ] Pre-release security gate verification (0 CRITICAL/HIGH)
- [ ] Container image scanned (Trivy) and signed (Cosign)
- [ ] Secrets verified absent from image layers (docker history)
- [ ] SBOM generated (syft) and attached to release

**Phase 7 - Deploy & Infrastructure**
- [ ] IaC files validated (tfsec / checkov)
- [ ] IAM roles and permissions verified for least-privilege
- [ ] Kubernetes Pod Security Standards enforced
- [ ] Cloud infrastructure misconfiguration scan run

**Phase 8 - Operations & Runtime**
- [ ] Application and authentication event logs structured
- [ ] Alerting rules active for brute force and travel anomalies
- [ ] Secrets rotation scheduled and monitored
- [ ] Vulnerability patch SLAs enforced (CRITICAL <= 24h)

**Phase 9 - Incident Response**
- [ ] 5-step IR plan verified (Contain, Identify, Eradicate, Recover, Lessons)
- [ ] Evidence preservation scripts ready
- [ ] Vulnerability report and executive summary generated
- [ ] Post-mortem analysis and threat model updated

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