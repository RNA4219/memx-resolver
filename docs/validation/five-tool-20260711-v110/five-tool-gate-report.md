# v1.1.0 Five-Tool Validation Gate Report

- target: `memx-resolver`
- target HEAD: `57ce98361e81c68e26cc9be3b030c29546e02d08`
- clean worktree: `dirty=false`
- verdict: `go`

## Chain Status

| Step | Status | Artifact | Result |
| --- | --- | --- | --- |
| RanD | degraded | repository requirements and approved implementation plan | fresh target-specific packet absent; assumptions recorded |
| Code-to-gate | ran | `ctg-final-head/` | tree-sitter, findings 0, critical/high 0, readiness passed |
| HATE | ran | `hate-final-head-verified/`, `hate-final-head-verified-qeg/` | release profile eligible, completeness 1.0, partial=false |
| manual-bb | ran | `manual-bb/` | P1 8/8 PASS, no waiver |
| QEG | ran | `qeg-final-head/`, `qeg-final-head-verdict.json` | validate PASS, record serialization PASS, verdict GO |

## Evidence Map

- Requirements: implementation plan, `docs/requirements.md`, `docs/interfaces.md`, `docs/design.md`
- Static: `ctg-final-head/findings.json`, `release-readiness.json`, `repo-graph.json`
- Automated: `hate-input/junit.xml`, `lcov.info`, `go-test-final.json`, `go-vet-final.txt`, `hate-final-head-verified/`
- Manual: `manual-bb/report.md`, `manual-bb/manual-bb.json`, execution logs
- Final gate: `qeg-final-head-verdict.json`, `qeg-final-head/output-record.json`

## Findings And Risks

- Critical/high findings: 0
- Broad suppressions: 0
- Narrow suppression: atomic writer temp cleanup only, path-specific, expires 2026-10-01
- Residual: Windows symlink test skipped when OS privilege is unavailable; Linux CI must execute it.
- Residual: Linux race evidence is a required GitHub CI check before tag creation.

## Tool Compatibility Note

Code-to-gate 1.5.0 exported QEG schema 0.1 while installed QEG requires 0.2. The generated fixture was migrated deterministically: all qegVersion fields became 0.2, legacy approvedBy became approver with release-approver authority and approvedDecision=go, and referenced output seed hashes were recorded. QEG then validated with no DQ and returned GO.

## Manual BB

Grounded viewpoints, risks, priorities, eight executed P1 cases, effort, gate judgement, and Go/No-Go brief are recorded in `manual-bb/report.md` in the required order.

## Verdict

`go` for the local release candidate. Tagging remains blocked until GitHub Windows/Linux CI and the protected `pypi` environment approval complete.



