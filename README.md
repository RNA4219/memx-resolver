# memx-resolver

memx-resolver は、`memx-core` をエージェント実行向けに拡張し、読むべき文書の解決、必要 chunk の取得、読了記録、stale 判定、契約解決まで扱える agent-first の OSS です。

## LLM-BOOTSTRAP

- このリポジトリの最初の入口はこの `README.md`
- 実装本体は [docs/memx_spec_v3/go](./docs/memx_spec_v3/go)
- 仕様の正本は [docs/requirements.md](./docs/requirements.md), [docs/interfaces.md](./docs/interfaces.md), [docs/design.md](./docs/design.md)
- 人間向けの使い方は [USER_GUIDE.md](./USER_GUIDE.md)
- ドキュメントハブは [docs/HUB.codex.md](./docs/HUB.codex.md)

## エージェント向け概要

memx は、エージェントが情報を保存、検索、再参照しながら継続作業するためのローカルメモリ基盤です。`memx-resolver` では、そこに `docs:resolve`、`chunks:get`、`reads:ack`、`docs:stale-check`、`contracts:resolve` を追加しています。

### 4つのストア

| Store | 用途 |
|------|------|
| `short` | 作業メモ、一時情報 |
| `journal` | 時系列ログ、進捗、意思決定 |
| `knowledge` | 定義、手順、永続知識 |
| `archive` | 退避済みノート |

### Claude Code Skills

| Skill | 用途 |
|------|------|
| `/remember` | short に保存 |
| `/recall` | `short / journal / knowledge` を横断検索 |
| `/journal` | journal に保存 |
| `/knowledge` | knowledge に保存 |
| `/show` | `short / journal / knowledge / archive` から表示 |
| `/memx-help` | 使い方表示 |
| `/resolve-docs` | feature / task / topic から読むべき文書を解決 |
| `/read-chunks` | 必要 chunk を取得 |
| `/ack-docs` | 読了記録を残す |
| `/stale-check` | stale を判定 |
| `/resolve-contract` | 契約情報を解決 |

Skill 定義は `.claude/commands/` にあります。

## 最低限の実行手順

```bash
cd docs/memx_spec_v3/go
go build ./cmd/mem

# resolver を分離したい場合
mem api serve --resolver resolver.db
```

### 保存

```bash
mem in short --title "メモ" --body "重要な情報"
mem in journal --title "進捗" --body "API実装完了" --scope project:memx
mem in knowledge --title "用語" --body "JWT = JSON Web Token" --scope glossary --pinned
```

### 検索 / 表示

```bash
mem out search --json "JWT"
mem out show <NOTE_ID>
mem out knowledge pinned --json
```

### resolver

```bash
mem docs resolve --feature resolver
mem docs chunks --doc-id workflow-cookbook:blueprint
mem docs ack --task-id TASK.sample --doc-id workflow-cookbook:blueprint --version v1
mem docs stale --task-id TASK.sample
mem docs contract --feature resolver
```

### API

```bash
mem api serve --addr 127.0.0.1:7766
```

## 実務上の使い分け

- 作業中の断片情報は `short`
- 進捗や出来事は `journal`
- 再利用したい事実や手順は `knowledge`
- 読むべき設計資料は resolver で解決
- 退避済みの確認は `archive`

## 守るべき前提

- `journal` と `knowledge` は `--scope` 必須
- `secret` は保存拒否
- 既定 DB は `short.db / journal.db / knowledge.db / archive.db`
- 変更前に [docs/requirements.md](./docs/requirements.md), [docs/interfaces.md](./docs/interfaces.md), [docs/design.md](./docs/design.md) を確認

## 詳細導線

- [USER_GUIDE.md](./USER_GUIDE.md)
- [docs/HUB.codex.md](./docs/HUB.codex.md)
- [docs/requirements.md](./docs/requirements.md)
- [docs/interfaces.md](./docs/interfaces.md)
- [docs/design.md](./docs/design.md)
- [docs/memx_spec_v3/docs/requirements.md](./docs/memx_spec_v3/docs/requirements.md)
- [docs/memx_spec_v3/docs/design.md](./docs/memx_spec_v3/docs/design.md)

License: MIT

