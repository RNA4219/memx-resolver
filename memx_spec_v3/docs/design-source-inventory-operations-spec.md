# Design Source Inventory Operations Spec

## 目的
- Phase 1 で生成する抽出表の保存・更新・承認の運用ルールを固定し、`docs/TASKS.md` への転記品質を安定化する。

## 運用ルール

### 1. 保存先
- 抽出表の保存先は `memx_spec_v3/docs/reviews/inventory/` に固定する。
- 新規作成・更新ともに上記ディレクトリ配下のみを許可する。

### 2. 命名規則
- 抽出表ファイル名は `DESIGN-SOURCE-INVENTORY-YYYYMMDD.md` とする。
- `YYYYMMDD` は作業日（JST）を使用し、同日複数回更新は同一ファイルへ追記する。

### 3. 更新粒度
- 抽出表は 1 行 1 `req_id` で管理する。
- 既存行の意味を変える差分更新は禁止し、修正時は「旧行を `deprecated` 扱いで残し、新行を追加」する。
- 許可される差分更新は次の 2 種のみ。
  - タイポ修正（`source_path#section` の綴り修正など）
  - `reviewed_at` の日付更新

### 4. 承認条件
- 承認可否は `blocked` 行数で判定し、`blocked` 行が 0 件の場合のみ承認可能とする。
- `blocked` 行が 1 件でも残る場合は Phase 1 Done 判定を不可とする。

## 準拠確認
- Phase 1 Done 判定時に、以下をチェックリストで明示する。
  - 保存先が `memx_spec_v3/docs/reviews/inventory/` である
  - ファイル名が `DESIGN-SOURCE-INVENTORY-YYYYMMDD.md` 形式である
  - 1 行 1 `req_id` を満たし、禁止された差分更新がない
  - `blocked` 行 0 件を確認済み
