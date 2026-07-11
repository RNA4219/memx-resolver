---
intent_id: INT-001
owner: memx-resolver
status: active
last_reviewed_at: 2026-03-10
next_review_due: 2026-04-10
---

# Runbook

## Environments

- Local: `mem` CLI直接実行、SQLiteファイルは `*.db`
- CI: テスト実行環境、一時DB使用
- Prod: API サーバーモード、`mem api serve --addr 127.0.0.1:7766`

## Execute

### 準備
```bash
cd .
go build -o mem.exe ./cmd/mem
```

Codex/sandbox などで `AppData` 配下の Go cache に書き込めない場合:

```powershell
$repoRoot = (Resolve-Path ..\..\..).Path
$env:GOCACHE = Join-Path $repoRoot ".tmp\go-build"
go test ./...
```

### 文書登録
```bash
mem docs ingest --title "Memory Import Spec" --doc-type spec --body "$(cat docs/specs/memory-import.md)"
```

### 文書解決
```bash
mem docs resolve --feature memory-import
mem docs resolve --task-id task:feature:local:123
```

### 文書検索
```bash
mem docs search "acceptance criteria" --doc-type spec --feature memory-import --tag memory
mem docs cards --query "acceptance criteria" --doc-type spec --feature memory-import --memory-type acceptance --token-budget 120
```

### Chunk取得
```bash
mem docs chunks --doc-id doc:spec:memory-import
mem docs chunks --doc-id doc:spec:memory-import --query "acceptance criteria"
mem docs chunks --chunk-id chunk:doc:spec:memory-import:001
```

`--json` では `chunks` に加えて、LLM に渡しやすい `memory_cards` も返る。

### 読了記録
```bash
mem docs ack --task-id task:feature:local:123 --doc-id doc:spec:memory-import --version 2026-03-10
```

### Stale確認
```bash
mem docs stale --task-id task:feature:local:123
```

## Observability

- ログ: 標準出力/標準エラー出力
- メトリクス: 確認方法は開発中
- インシデント: `docs/IN-YYYYMMDD-XXX.md` に記録

## Confirm

- Execute結果を主要出力と突き合わせ
- CHECKLISTS.md で整合性確認

## Rollback / Retry

- DBファイルのバックアップから復元
- インシデント記録を `docs/IN-YYYYMMDD-XXX.md` に追記
