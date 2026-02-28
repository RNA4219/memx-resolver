# memx_spec v2

指摘事項を反映した要件定義と、最低限のスキーマ／Goインターフェースをまとめた ZIP です。

## 構成

- `docs/requirements.md`  
  システム全体の要件定義（目的・アーキテクチャ・CLI・GC・LLM役割など、レビュー内容を反映）。
- `schema/short.sql`  
  `short.db` 用の CREATE TABLE スキーマ（FTS5 トリガ修正・カラム制約調整・user_version 追加済み）。
- `go/types.go`  
  DBパス・コネクション定義などの共通型（LLM/Gatekeeper の注入口を追加）。
- `go/open.go`  
  `db.MustOpenAll` と ATTACH / migrate のインターフェース。
- `go/migrate_short.go`  
  `short.db` 用マイグレーション。
- `go/migrate_other.go`  
  `chronicle` / `memopedia` / `archive` 用マイグレーション関数のシグネチャ。
- `go/recall.go`  
  Semantic Recall 用インターフェース。
- `go/gc.go`  
  GC / Observer / Reflector 用インターフェース。
- `go/llm_client.go`  
  Embedding / MiniLLM / ReflectLLM の役割分離インターフェース（package db）。
- `go/gatekeeper.go`  
  Gatekeeper（1B ガードモデル）用のインターフェース。
