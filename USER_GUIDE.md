# memx-resolver ユーザーガイド

memx-resolver は、ローカル LLM / エージェント向けの軽量メモリ基盤を、coding agent 向け resolver 機能で拡張した実験実装です。4つの SQLite ストアで、保存、検索、参照、要約を行い、docs/chunks/read receipts/stale check も扱えます。

エージェント向けの入口は [README.md](./README.md) です。人間が使い方を追う場合はこのガイドを参照してください。

## ストア

| Store | 用途 |
|------|------|
| `short` | 短期メモ、作業中の情報 |
| `journal` | 時系列ログ、進捗、判断 |
| `knowledge` | 知識、定義、手順 |
| `archive` | 検索対象外の退避先 |

## クイックスタート

```bash
cd docs/memx_spec_v3/go
go build ./cmd/mem

# resolver を分離したい場合
mem api serve --resolver resolver.db

# short に保存
mem in short --title "会議メモ" --body "明日10時に打ち合わせ"

# 横断検索
mem out search --json "会議"

# ID を指定して表示
mem out show <NOTE_ID>
```

`mem out search` は `short / journal / knowledge` を横断検索します。  
`mem out show` は `short / journal / knowledge / archive` を解決します。

## Claude Code Skills

`.claude/commands/` に以下の Skill を含みます。

| Skill | 用途 |
|------|------|
| `/remember` | short に保存 |
| `/recall` | 横断検索 |
| `/journal` | journal に保存 |
| `/knowledge` | knowledge に保存 |
| `/show` | ノート詳細表示 |
| `/memx-help` | 使い方表示 |
| `/resolve-docs` | 読むべき文書を解決 |
| `/read-chunks` | 必要 chunk を取得 |
| `/ack-docs` | 読了記録を残す |
| `/stale-check` | stale を判定 |
| `/resolve-contract` | 契約情報を解決 |

## CLI

```bash
# short
mem in short --title "Title" --body "Body"

# journal / knowledge
mem in journal --title "進捗" --body "API実装完了" --scope project:memx
mem in knowledge --title "用語" --body "JWT = JSON Web Token" --scope glossary --pinned

# 検索 / 表示
mem out search --json "JWT"
mem out show <NOTE_ID>

# store ごとの操作
mem out journal list --scope project:memx
mem out knowledge pinned --json

# 要約
mem summarize <NOTE_ID>
mem summarize --ids id1,id2,id3 --json

# resolver
mem docs resolve --feature resolver
mem docs chunks --doc-id workflow-cookbook:blueprint
mem docs ack --task-id TASK.sample --doc-id workflow-cookbook:blueprint --version v1
mem docs stale --task-id TASK.sample
mem docs contract --feature resolver

# resolver 専用 DB を使う場合
mem docs ingest --resolver resolver.db --title "Spec" --body "# Spec" --doc-type spec --version 2026-03-10

# GC
mem gc short --dry-run
mem gc short --enable-gc
```

既定 DB は `short.db / journal.db / knowledge.db / archive.db` です。

## API

```bash
mem api serve --addr 127.0.0.1:7766
```

CLI から API サーバー経由で使う場合:

```bash
mem in short --api-url http://127.0.0.1:7766 --title "test" --body "body"
mem out search --api-url http://127.0.0.1:7766 --json "test"
```

主なエンドポイント:

- `POST /v1/notes:ingest`
- `POST /v1/notes:search`
- `GET /v1/notes/{id}`
- `POST /v1/notes:summarize`
- `POST /v1/journal:ingest`
- `POST /v1/journal:search`
- `GET /v1/journal/{id}`
- `POST /v1/knowledge:ingest`
- `POST /v1/knowledge:search`
- `GET /v1/knowledge/{id}`
- `GET /v1/archive/{id}`
- `POST /v1/docs:ingest`
- `POST /v1/docs:resolve`
- `POST /v1/chunks:get`
- `POST /v1/docs:search`
- `POST /v1/reads:ack`
- `POST /v1/docs:stale-check`
- `POST /v1/contracts:resolve`

## LLM 要約

`mem summarize` と保存時の自動要約で使います。  
`mem` CLI と API サーバーは、`memx-core` 配下で起動した場合に `.env` を自動読込します。

### OpenAI

```bash
export MEMX_LLM_PROVIDER="openai"
export OPENAI_API_KEY="sk-..."
export MEMX_OPENAI_MODEL="gpt-5-mini"
```

### Alibaba Cloud Model Studio

```bash
export MEMX_LLM_PROVIDER="alibaba"
export DASHSCOPE_API_KEY="sk-..."
export MEMX_ALIBABA_MODEL="glm-5"
export MEMX_ALIBABA_BASE_URL="https://coding-intl.dashscope.aliyuncs.com/v1"
```

Alibaba 互換モードでは `chat/completions` を使います。

## セキュリティ

- `secret` は保存拒否
- 既定の `sensitivity` は `internal`
- タイトル、本文、列挙値にバリデーションあり

## 参照先

- [README.md](./README.md)
- [docs/HUB.codex.md](./docs/HUB.codex.md)
- [docs/requirements.md](./docs/requirements.md)
- [docs/interfaces.md](./docs/interfaces.md)
- [docs/design.md](./docs/design.md)
- [docs/memx_spec_v3/docs/requirements.md](./docs/memx_spec_v3/docs/requirements.md)
- [docs/memx_spec_v3/docs/design.md](./docs/memx_spec_v3/docs/design.md)
- [docs/memx_spec_v3/docs/contracts/openapi.yaml](./docs/memx_spec_v3/docs/contracts/openapi.yaml)
- [docs/memx_spec_v3/docs/contracts/cli-json.schema.json](./docs/memx_spec_v3/docs/contracts/cli-json.schema.json)

