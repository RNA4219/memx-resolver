---
intent_id: INT-001
owner: memx-resolver
status: active
last_reviewed_at: 2026-03-10
next_review_due: 2026-04-10
---

# Checklists

## Hygiene

- [ ] コードフォーマット確認（gofmt / goimports）
- [ ] Lint確認（golangci-lint）
- [ ] 型チェック確認（go vet）
- [ ] テスト実行（go test ./...）
- [ ] ドキュメント更新（必要に応じて）

## Review

- [ ] PR本文にIntent: INT-xxxを含める
- [ ] EVALUATION見出しへのリンクを本文に明示
- [ ] 変更の影響範囲を確認
- [ ] 破壊的変更の有無を確認

## Release

- [ ] 全テストが通過
- [ ] CHANGELOG.mdを更新
- [ ] バージョンタグを作成
- [ ] リリースノートを作成
- [ ] Acceptance recordを作成（`docs/acceptance/AC-YYYYMMDD-xx.md`）

## Acceptance Workflow

重大な変更の検収記録を残す：

- [ ] Acceptance recordを`docs/acceptance/AC-YYYYMMDD-xx.md`に作成
- [ ] front matter (`acceptance_id`, `status`, `reviewed_at`)を記入
- [ ] `## Scope`で対象/非対象を明示
- [ ] `## Acceptance Criteria`をチェック済み状態にする
- [ ] `## Evidence`に実行コマンドと結果を記録
- [ ] `## Verification Result`で判定 (`approved` / `rejected`)を記述

## Incident

- [ ] docs/IN-YYYYMMDD-XXX.mdを作成
- [ ] 影響範囲を記録
- [ ] 再発防止策を記録
- [ ] 関連PR/チケットへリンク
