# Five Tool Validation Gate Report

Target: `C:\Users\ryo-n\Codex_dev\memx-resolver`

Scope:

- semantic stale / impact scope
- memory card adaptive ranking
- agent-taskstate export bridge
- prompt-ready memory card bundle

## Chain Status

| Step | Status | Artifact | Notes |
| --- | --- | --- | --- |
| RanD | degraded | `docs/requirements.md`, `docs/acceptance/AC-20260702-01.md` | No fresh RanD `requirements_audit_packet.json`; used implemented requirements and acceptance as requirement hypothesis. |
| Code-to-gate | ran | `docs/validation/five-tool-20260702/ctg/` | `analyze` + `readiness` + SARIF export ran. Readiness `passed`; findings 3, critical 0, high 1. |
| HATE | ran | `docs/validation/five-tool-20260702/hate/`, `hate-qeg/` | pytest JUnit ingested through HATE P0a. Precheck `eligible`; QEG optional bundle exported with completeness 1.00. |
| manual-bb | ran | This report | Manual black-box plan derived from requirements, CTG findings, Go/Python tests, and HATE evidence. |
| QEG | degraded | `docs/validation/five-tool-20260702/hate-qeg/qeg-bundle.json` | QEG optional evidence bundle generated. Direct QEG `gate` not run because bundle is not a full `gate-input.json` fixture. |

## Evidence Map

- Requirements:
  - `docs/requirements.md`
  - `docs/interfaces.md`
  - `docs/design.md`
  - `docs/acceptance/AC-20260702-01.md`
- Static analysis:
  - `ctg/findings.json`
  - `ctg/risk-register.yaml`
  - `ctg/release-readiness.json`
  - `ctg/results.sarif`
- Auto-test evidence:
  - `hate-input/pytest-junit.xml`
  - `hate/precheck-decision.json`
  - `hate/HATE-test-results.ndjson`
  - `hate-qeg/qeg-bundle.json`
- Manual QA:
  - This report, sections below.
- QEG gate:
  - Optional evidence bundle ready.
  - Full QEG release fixture missing.

## Findings And Risks

- P0: none found in executed evidence.
- P1:
  - `finding-HARDCODED_SECRET-010`: Code-to-gate high finding in `workflow-cookbook/tools/audit/verify_log_chain.py:74`. This appears outside the changed resolver surface and is affected by existing exclude/suppression policy limitations, but it remains a review item before release claims.
- Residual:
  - RanD evidence is degraded because no fresh KanoMode / requirements audit packet was generated.
  - QEG direct gate is degraded because the generated HATE bundle is optional evidence, not a complete QEG `gate-input.json` fixture.

## Manual BB Plan

### 1. 根拠付き観点

| Viewpoint | Source refs | Oracle |
| --- | --- | --- |
| stale semantic diff | `docs/interfaces.md`, `resolver_docs_test.go`, `openapi.yaml` | Updating a read chunk returns `reason=semantic_diff`, non-empty `impact_scope`, and `changed_chunks`. |
| metadata-only stale | `docs/design.md`, `resolver_docs.go` | Version changes without read chunk changes remain `version_mismatch` with metadata impact. |
| adaptive ranking | `resolver_memory_cards.go`, `resolver_docs_test.go` | `ranking_weights` and `cards-feedback` alter ordering without breaking token budget. |
| prompt bundle | `cmd_docs.go`, `/v1/cards:bundle`, `cli-json.schema.json` | Bundle contains prompt text, cards, token estimate, and `source_refs`. |
| agent-taskstate bridge | `/v1/taskstate:export`, `docs/interfaces.md` | Export includes `task_ref`, required docs, read receipts, stale reasons, and 4-segment typed refs. |

### 2. リスク

| Risk | Priority | Rationale |
| --- | --- | --- |
| Impact scope under-reports when chunk IDs change due heading/chunking changes | P1 | Same semantic content may receive different chunk IDs after rechunking. |
| Feedback overboost can bury required acceptance cards | P1 | Large positive feedback can dominate defaults. |
| taskstate export is bridge-only, not direct persistence | P2 | Operational users may expect agent-taskstate DB mutation. |
| Prompt bundle source refs are derived from sanitized IDs | P2 | Ref round-trip needs consumer agreement. |
| Code-to-gate high finding in excluded workflow-cookbook path | P2 | Outside changed surface but should be reviewed or narrowed in policy. |

### 3. 優先度

- P0: execute automated semantic stale / ranking / bundle / export regression before merge.
- P1: manual black-box stale and ranking scenarios below.
- P2: taskstate consumer compatibility review, source ref round-trip review.

### 4. 手動テストケース

| ID | Priority | Steps | Expected |
| --- | --- | --- | --- |
| MBB-001 | P1 | Ingest doc v1, ack one acceptance chunk, update the same chunk body to v2, run `mem docs stale --json`. | `status=stale`, reason `semantic_diff`, changed chunk listed, impact includes acceptance/heading. |
| MBB-002 | P1 | Ingest doc v1, ack acceptance chunk, update unrelated chunk only, run stale. | stale is metadata/version impact or no semantic change for the read chunk; no unrelated changed chunk is attributed. |
| MBB-003 | P1 | Search cards, record `cards-feedback --signal helpful --weight 20` for a lower-ranked card, search again. | Feedback target moves upward; token budget still respected. |
| MBB-004 | P1 | Run `mem docs cards --query ... --weight-query-exact 1 --weight-memory-type-base 40`. | Ranking follows configured weight emphasis. |
| MBB-005 | P1 | Run `mem docs bundle --json --query ... --format markdown` and `--format jsonl`. | Prompt/bundle are non-empty, cards preserve provenance, refs are present. |
| MBB-006 | P1 | Run `mem docs taskstate-export --json --task-id ... --feature ...`. | Payload contains task ref, required docs, read receipts with snapshots, stale reasons, source refs. |

### 5. 工数

- Prep: 20 minutes
- Execution: 45 minutes
- Evidence capture: 20 minutes
- Retry buffer: 30 minutes
- Total: about 2 hours

### 6. Gate 判定

`needs_review`

Reason:

- Automated tests, Code-to-gate readiness, and HATE P0a evidence pass.
- QEG optional bundle exists, but full QEG release fixture was not executed.
- One Code-to-gate high finding remains as review-required, even if outside the changed resolver surface.

### 7. Go/No-Go brief

This change set is suitable for engineering review and internal continuation. It should not be represented as final release approval until QEG receives a complete `gate-input.json` fixture and the Code-to-gate high finding is either fixed, excluded with a narrow policy, or explicitly waived.

## QEG Gate Package

- HATE QEG optional evidence:
  - `docs/validation/five-tool-20260702/hate-qeg/qeg-bundle.json`
  - `docs/validation/five-tool-20260702/hate-qeg/evidence-map.json`
  - `docs/validation/five-tool-20260702/hate-qeg/qeg-export-report.json`
- Missing for full QEG:
  - complete `gate-input.json`
  - policy hash and approval metadata for this memx-resolver validation run
  - manual-bb execution result artifact after cases are actually run by a human

## Verdict

`needs_review`

## Next Commands

```powershell
cd C:\Users\ryo-n\Codex_dev\memx-resolver
C:\Users\ryo-n\AppData\Local\Programs\Python\Python311\python.exe -m pytest tests --junitxml docs\validation\five-tool-20260702\hate-input\pytest-junit.xml
$env:GOCACHE='C:\Users\ryo-n\Codex_dev\memx-resolver\.tmp\go-build'; go test ./...
node C:\Users\ryo-n\Codex_dev\code-to-gate\dist\cli.js analyze C:\Users\ryo-n\Codex_dev\memx-resolver --emit all --out docs\validation\five-tool-20260702\ctg --cache disabled --parallel 4
uv run hate p0a --input C:\Users\ryo-n\Codex_dev\memx-resolver\docs\validation\five-tool-20260702\hate-input --out C:\Users\ryo-n\Codex_dev\memx-resolver\docs\validation\five-tool-20260702\hate --source-version memx-resolver-five-tool-20260702
uv run hate export qeg --fixture C:\Users\ryo-n\Codex_dev\memx-resolver\docs\validation\five-tool-20260702\hate-fixture --out C:\Users\ryo-n\Codex_dev\memx-resolver\docs\validation\five-tool-20260702\hate-qeg --source-version memx-resolver-five-tool-20260702
```
