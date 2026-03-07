# memx-core

> **ローカルLLM/エージェント向けのメモリ基盤**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## これは何？

**memx-core** は、LLMエージェントに「記憶」を提供する軽量なデータストアです。

### 解決する問題

- LLMは長い会話の文脈を忘れる
- 過去の知識を再利用できない
- ユーザー固有の情報を保持できない

**→ memx-core は「外部メモリ」としてこれらを解決します。**

---

## クイックスタート

```bash
# ビルド
cd memx-core_spec_v3/go
go build ./cmd/mem

# メモを保存
mem in short --title "会議メモ" --body "明日10時に打ち合わせ"

# 検索
mem out search "会議"

# 詳細表示
mem out show <NOTE_ID>
```

---

## 主な機能

| 機能 | コマンド | 説明 |
|------|----------|------|
| 保存（short） | `mem in short` | メモを短期ストアに保存 |
| 保存（journal） | `mem in journal --scope <scope>` | ログをjournalストアに保存 |
| 保存（knowledge） | `mem in knowledge --scope <scope>` | 知識をknowledgeストアに保存 |
| 検索（short） | `mem out search` | キーワードでメモを検索 |
| 検索（journal） | `mem out journal search` | journalを検索 |
| 検索（knowledge） | `mem out knowledge search` | knowledgeを検索 |
| 表示 | `mem out show` | メモの詳細を表示 |
| 要約 | `mem summarize` | LLMでメモを要約 |
| GC | `mem gc short --dry-run` | 古いメモの整理（確認のみ） |
| GC実行 | `mem gc short --enable-gc` | 古いメモをarchiveへ退避 |

---

## アーキテクチャ

4つのストア構成：

| ストア | DB | 用途 | typed_ref |
|--------|-----|------|-----------|
| short | `short.db` | 短期記憶（作業メモ） | evidence |
| journal | `journal.db` | 長期記憶（時系列ログ） | evidence |
| knowledge | `knowledge.db` | 知識ベース（永続情報） | knowledge |
| archive | `archive.db` | アーカイブ（検索対象外） | evidence |

**v1.3 では全ストアの CRUD を実装済み。**

---

## ドキュメント

### エージェント向け

- **[AGENT_GUIDE.md](./AGENT_GUIDE.md)** - AIエージェント向けの利用案内（まずこれを読んでください）

### 正本ドキュメント

| 種別 | ドキュメント |
|------|--------------|
| 要件 | [requirements.md](./memx_spec_v3/docs/requirements.md) |
| 設計 | [design.md](./memx_spec_v3/docs/design.md) |
| API契約 | [contracts/openapi.yaml](./memx_spec_v3/docs/contracts/openapi.yaml) |
| CLI契約 | [contracts/cli-json.schema.json](./memx_spec_v3/docs/contracts/cli-json.schema.json) |

### 参照導線

- [spec.md](./memx_spec_v3/docs/spec.md) - 正本/補助の定義と参照導線

---

## セキュリティ

- **fail-closed 方針**: `secret` 機密度のメモは保存を拒否
- **入力バリデーション**: タイトル/本文の長さ制限、enum値チェック
- **ローカル専用**: 外部公開を前提としない設計

---

## 開発状況

### v1.3 完了済み

- [x] CLI基本コマンド (in/out/search/show)
- [x] HTTP API サーバー
- [x] Gatekeeper（セキュリティチェック）
- [x] 入力バリデーション
- [x] LLM要約機能
- [x] 全ストアのスキーマ定義（short/journal/knowledge/archive）
- [x] GC（ガベージコレクション）機能
- [x] **journal ストアの CRUD実装**（Ingest, Get, Search, ListByScope）
- [x] **knowledge ストアの CRUD実装**（Ingest, Get, Search, ListByScope, Pin/Unpin）
- [x] **archive ストアの CRUD実装**（Get, List, ArchiveFromShort, Restore, Lineage）
- [x] Claude Code スキル（/remember, /recall, /journal, /knowledge, /show）

### 次期ロードマップ

KV優先ロードマップに従って以下を順次実装：

1. P1: KVキャッシュ独立化
2. P2: typed_ref 統一
3. P3: WorkX 状態履歴・バンドル監査
4. P4: WorkX/MemX コンテキスト再構築リゾルバ
5. P5: Tracker Bridge 最小統合

---

## Governance Docs

- [BLUEPRINT.md](./docs/BLUEPRINT.md) - 設計方針
- [RUNBOOK.md](./docs/RUNBOOK.md) - 運用手順
- [GUARDRAILS.md](./docs/GUARDRAILS.md) - 安全性ガイドライン

---

## License

MIT License
