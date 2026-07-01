# code-to-gate Analysis Report

**Generated**: 2026-07-01T16:01:19.405Z
**Run ID**: ctg-202607011601-local
Repository: .
**Tool**: code-to-gate v1.5.0

---

## Summary

### Raw Findings (All Detections)

| Metric | Count |
|--------|-------|
| Total Raw Findings | 11 |
| Critical | 0 |
| High | 3 |
| Medium | 8 |
| Low | 0 |

### Effective Findings (After Suppression)

| Metric | Count |
|--------|-------|
| Effective Findings | 3 |
| Critical | 0 |
| High | 1 |
| Medium | 2 |
| Low | 0 |

### Accepted Exceptions (Suppressed)

| Metric | Count |
|--------|-------|
| Suppressed Findings | 8 |
| Critical | 0 |
| High | 2 |
| Medium | 6 |
| Low | 0 |

#### Exception Classification Breakdown

| Class | Count | Description |
|-------|-------|-------------|
| self-reference | 0 | Rule implementation files |
| fixture-intentional | 0 | Test fixtures |
| generated-artifact | 0 | Compiled output |
| accepted-design | 0 | Architecture decisions |
| temporary-debt | 8 | Needs repayment |

### Known Debt

| Debt Type | Count | Critical | High | Medium | Low |
|-----------|-------|----------|------|--------|-----|
| Suppression Debt | 2 | 0 | 0 | 2 | 0 |
| Explicit Debt Markers | 0 | 0 | 0 | 0 | 0 |

## Broad Suppression Review

**WARNING**: 1 broad suppression(s) detected. These suppress wide file scopes and may hide underlying issues.

| Rule | Path | Type | Reason | Class |
|------|------|------|--------|-------|
| LARGE_MODULE | workflow-cookbook/**/* | rule-wide | Embedded workflow-cookbook directory. Findings belong to workflow-cookbook repo, not memx-resolver. Should exclude workflow-cookbook/ in config. | temporary-debt |

## Domain Context

| Domain | Findings | High/Critical | Evidence Paths |
|--------|----------|---------------|----------------|
| Security boundary | 1 | 1 | workflow-cookbook/tools/audit/verify_log_chain.py |
| Code health and maintainability | 2 | 0 | .ctg/suppressions.yaml |

## Suppressed Findings

| ID | Rule | Severity | Title | Reason |
|----|------|----------|-------|--------|
| finding-UNSAFE_DELETE-000 | UNSAFE_DELETE | **HIGH** | Unsafe delete operation detected | (suppressed) |
| finding-LARGE_MODULE-001 | LARGE_MODULE | *MEDIUM* | Module has too many functions (22) | (suppressed) |
| finding-LARGE_MODULE-002 | LARGE_MODULE | *MEDIUM* | Module has too many functions (32) | (suppressed) |
| finding-LARGE_MODULE-003 | LARGE_MODULE | *MEDIUM* | Module exceeds line count threshold (998 lines) | (suppressed) |
| finding-LARGE_MODULE-004 | LARGE_MODULE | *MEDIUM* | Module exceeds line count threshold (861 lines) | (suppressed) |
| finding-LARGE_MODULE-005 | LARGE_MODULE | **HIGH** | Module exceeds line count threshold (1230 lines) | (suppressed) |
| finding-LARGE_MODULE-006 | LARGE_MODULE | *MEDIUM* | Module has too many functions (21) | (suppressed) |
| finding-LARGE_MODULE-007 | LARGE_MODULE | *MEDIUM* | Module has too many functions (22) | (suppressed) |

## Suppression Debt

These suppressions may hide underlying issues and should be reviewed.

| ID | Location | Severity | Title |
|----|----------|----------|-------|
| finding-SUPPRESSION_DEBT-008 | .ctg/suppressions.yaml | *MEDIUM* | Suppression may hide debt (UNSAFE_DELETE) |
| finding-SUPPRESSION_DEBT-009 | .ctg/suppressions.yaml | *MEDIUM* | Suppression may hide debt (LARGE_MODULE) |

## High-Priority Risks

| Risk ID | Title | Severity | Likelihood | Source Findings |
|---------|-------|----------|------------|-----------------|
| risk-HARDCODED_SECRET-010 | Possible secret in variable: help | **HIGH** | medium | finding-HARDCODED_SECRET-010 |

## All Findings

| ID | Rule | Category | Domain | Severity | Title | Evidence | Review Flags | LLM |
|----|------|----------|--------|----------|-------|----------|--------------|-----|
| finding-SUPPRESSION_DEBT-008 | SUPPRESSION_DEBT | maintainability | Code health and maintainability | *MEDIUM* | Suppression may hide debt (UNSAFE_DELETE) | .ctg/suppressions.yaml | evidence-linked | not-used |
| finding-SUPPRESSION_DEBT-009 | SUPPRESSION_DEBT | maintainability | Code health and maintainability | *MEDIUM* | Suppression may hide debt (LARGE_MODULE) | .ctg/suppressions.yaml | evidence-linked | not-used |
| finding-HARDCODED_SECRET-010 | HARDCODED_SECRET | security | Security boundary | **HIGH** | Possible secret in variable: help | workflow-cookbook/tools/audit/verify_log_chain.py | evidence-linked | not-used |

## False-Positive Review

| Finding | Checkpoint |
|---------|------------|
| finding-SUPPRESSION_DEBT-008 | domain=Code health and maintainability; evidence=.ctg/suppressions.yaml; confidence=0.80; flags=evidence-linked |
| finding-SUPPRESSION_DEBT-009 | domain=Code health and maintainability; evidence=.ctg/suppressions.yaml; confidence=0.80; flags=evidence-linked |
| finding-HARDCODED_SECRET-010 | domain=Security boundary; evidence=workflow-cookbook/tools/audit/verify_log_chain.py; confidence=0.70; flags=evidence-linked |

## Risk Narratives

### risk-HARDCODED_SECRET-010: Possible secret in variable: help

**Severity**: **HIGH**
**Likelihood**: medium
**Confidence**: 0.70

**Impact**:
- Variable "help" may contain hardcoded secret

**Recommended Actions**:
- Address finding finding-HARDCODED_SECRET-010: HARDCODED_SECRET

---

## Recommended Actions Summary

### Priority Order

1. **[HIGH]** Address finding finding-HARDCODED_SECRET-010: HARDCODED_SECRET

---

*This report was generated by code-to-gate. Findings are based on static analysis of the repository.*

