# memx エージェントガイド

memx は、エージェントが情報を保存・検索・再参照するためのローカルメモリ基盤です。

## 4つのストア

| Store | 用途 |
|------|------|
| `short` | 作業メモ、一時情報 |
| `journal` | 時系列ログ、進捗、意思決定 |
| `knowledge` | 定義、手順、永続知識 |
| `archive` | 退避済みノート |

## Claude Code Skills

| Skill | 用途 |
|------|------|
| `/remember` | short に保存 |
| `/recall` | `short / journal / knowledge` を横断検索 |
| `/journal` | journal に保存 |
| `/knowledge` | knowledge に保存 |
| `/show` | `short / journal / knowledge / archive` から表示 |
| `/memx-help` | 使い方表示 |

Skill 定義は `.claude/commands/` にあります。

## 最低限の使い方

```bash
cd memx_spec_v3/go
go build ./cmd/mem
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

### 要約

```bash
mem summarize <NOTE_ID>
mem summarize --ids id1,id2,id3 --json
```

## 実務上の使い分け

- 作業中の断片情報は `short`
- 進捗や出来事は `journal`
- 再利用したい事実や手順は `knowledge`
- 退避済みの確認は `archive`

迷ったら、まず `short` に入れてよいです。

## API サーバー

```bash
mem api serve --addr 127.0.0.1:7766
```

CLI を API 経由で使う場合:

```bash
mem out search --api-url http://127.0.0.1:7766 --json "query"
```

## LLM 要約

`.env` または環境変数で OpenAI / Alibaba を設定すると、要約と自動要約が有効になります。

```bash
export MEMX_LLM_PROVIDER="alibaba"
export DASHSCOPE_API_KEY="sk-..."
export MEMX_ALIBABA_BASE_URL="https://coding-intl.dashscope.aliyuncs.com/v1"
```

## 注意

- `journal` と `knowledge` は `--scope` 必須
- `secret` は保存拒否
- 既定 DB は `short.db / journal.db / knowledge.db / archive.db`

## 詳細ドキュメント

- [README.md](./README.md)
- [memx_spec_v3/docs/requirements.md](./memx_spec_v3/docs/requirements.md)
- [memx_spec_v3/docs/design.md](./memx_spec_v3/docs/design.md)
