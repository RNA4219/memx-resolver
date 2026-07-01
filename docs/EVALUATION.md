---
intent_id: INT-001
owner: memx-resolver
status: active
last_reviewed_at: 2026-03-10
next_review_due: 2026-04-10
---

# Evaluation

## Acceptance Criteria

- [x] workflow-cookbookの主要文書をdocとして登録できる
- [x] feature名からrequired / recommended docsを返せる
- [x] docから必要chunkを取得できる
- [x] `memory_type` / importance / query match / token budget に基づいて memory_cards をランキングできる
- [x] memory_cards ranking を実利用 feedback と ranking weights で補正できる
- [x] `mem docs cards --query ...` で LLM 向け memory_cards を取得できる
- [x] `mem docs bundle --query ...` で prompt-ready bundle を取得できる
- [x] 読了したdoc versionとchunk_idsをtaskに紐づけて記録できる
- [x] doc version更新時に semantic diff / impact scope 付きで stale 判定できる
- [x] `mem docs taskstate-export --task-id ...` で agent-taskstate 連携 payload を取得できる
- [x] docs/interfaces.md に記載した最小APIが呼べる
- [x] 文書を登録するとchunkが生成される
- [x] CLI JSON / HTTP 実レスポンスが `cli-json.schema.json` / OpenAPI schema に適合する
- [x] contract resolveがacceptance / forbidden / DoD / dependenciesを返せる

## KPIs

| 指標 | 目的 | 目標値 |
| --- | --- | --- |
| doc解決精度 | feature/taskから適切なdocを返す割合 | 90%以上 |
| chunk取得時間 | chunk取得のレスポンス時間 | 100ms以下 |
| stale検知率 | 更新されたdocのstale検知率 | 100% |

## Test Outline

- 単体: 各APIの入出力テスト
- 結合: doc登録→解決→chunk取得→読了記録→stale判定のフロー
- 回帰: 既存機能への影響確認

## Verification Checklist

- [x] 主要フローが動作する（自動テスト確認）
- [x] エラー時挙動が明示されている
- [x] 依存関係が再現できる環境である

## Acceptance Records

- [AC-20260701-01](acceptance/AC-20260701-01.md)
- [AC-20260702-01](acceptance/AC-20260702-01.md)
