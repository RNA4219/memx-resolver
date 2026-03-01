# memx_spec v3

要件定義（v1.3）と、CLI→API→Service→DB のレイヤリングを反映した最小実装をまとめた ZIP です。

## 構成

- `docs/requirements.md`  
  システム全体の要件定義（目的・アーキテクチャ・CLI・GC・LLM役割など、レビュー内容を反映）。
- `schema/short.sql`  
  `short.db` 用の CREATE TABLE スキーマ（FTS5 トリガ修正・カラム制約調整・user_version 追加済み）。

### Go（モジュールは `go/` 配下）

- `go/go.mod`
  - Go モジュール定義（依存は最小）。
- `go/db/*`
  - DB 接続・マイグレーション・LLM/Gatekeeper の注入口（インフラ層）。
- `go/service/*`
  - Usecase 層（短期ストアへの ingest/search/get の最小実装）。
- `go/api/*`
  - ツール/AI 向け API（HTTP と in-proc クライアント）。
- `go/cmd/mem/main.go`
  - 人間向け CLI。API の薄いラッパ。
