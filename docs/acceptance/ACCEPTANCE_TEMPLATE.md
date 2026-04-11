---
acceptance_id: AC-YYYYMMDD-xx
task_id: TASK-xxx
intent_id: INT-xxx
owner: your-handle
status: draft
reviewed_at: YYYY-MM-DD
reviewed_by: reviewer-handle
---

# Acceptance Record: Title

## Scope

- 対象変更:
  - file1
  - file2
- 非対象:
  - 明示的に除外する範囲

## Acceptance Criteria

- [ ] Criteria 1
- [ ] Criteria 2
- [ ] Criteria 3

## Evidence

- 実行コマンド:
  - `go test ./cmd/mem/...`
- テスト結果:
  - All tests passed
- 参照ドキュメント:
  - `docs/requirements.md`
- 関連 Task Seed:
  - `docs/TASKS.md`（必要な場合）

## Verification Result

- 判定: draft → approved または rejected
- コメント:
  - 検収判断の理由
- フォローアップ:
  - 残課題（必要な場合）
