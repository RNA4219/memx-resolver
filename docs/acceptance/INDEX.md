# Acceptance Index

## Summary

| Status | Count | Percentage |
| --- | --- | --- |
| approved | 4 | 100.0% |
| rejected | 0 | 0.0% |
| draft | 0 | 0.0% |
| unknown | 0 | 0.0% |
| **Total** | **4** | **100%** |

## Release Mapping

| Release | Acceptances | Date |
| --- | --- | --- |
| unreleased | [AC-20260701-01](AC-20260701-01.md), [AC-20260702-01](AC-20260702-01.md), [AC-20260703-01](AC-20260703-01.md) | 2026-07-03 |
| v1.1.0 | [AC-20260711-01](AC-20260711-01.md) | 2026-07-11 |

## Records

| Acceptance | Status | Reviewed At | Scope |
| --- | --- | --- | --- |
| [AC-20260701-01](AC-20260701-01.md) | approved | 2026-07-01 | memory card resolver contract |
| [AC-20260702-01](AC-20260702-01.md) | approved | 2026-07-02 | semantic stale, adaptive ranking, taskstate export, prompt bundle |
| [AC-20260703-01](AC-20260703-01.md) | approved | 2026-07-03 | resolver FTS, receipt audit, refs export, migration, version policy |
| [AC-20260711-01](AC-20260711-01.md) | approved | 2026-07-11 | v1.1.0 safety, packaging, release gate |

## Notes

- memx-resolver uses `go test` as the primary verification method.
- Acceptance records are optional but recommended for major changes.
- Cross-repo integration tests may reference workflow-cookbook acceptance
  records.
- Use `/stale-check` command to verify docs freshness.
