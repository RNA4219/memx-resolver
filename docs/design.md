# cookbook-resolver design

## 1. 文書情報

- 文書名: cookbook-resolver design
- 文書種別: design
- 版: v0.1
- 作成日: 2026-03-10
- 状態: Draft

## 2. 目的

本書は `docs/requirements.md` と `docs/interfaces.md` を満たすための最小実装方針を定義する。

対象は以下とする。

- 文書登録
- chunk 化
- 文書解決
- chunk 取得
- 読了記録
- stale 判定
- 契約解決

## 3. 実装方針

### 3.1 配置

最小実装は `repository root Go module` 配下に追加し、既存の `memx-core` 系 API と同じプロセスで提供する。

- DB: resolver 用テーブルは resolver store に配置し、未設定時のみ `short.db` 同居を許可する
- service: resolver 用 usecase を追加する
- api: `/v1/docs:*` などの HTTP API を追加する
- client: in-proc / HTTP client の両方から同じ API を呼べるようにする

### 3.2 ストア境界

resolver 系テーブルは 1 つの resolver store にまとめる。

- `resolver_documents` / `resolver_chunks` / `resolver_document_links` / `resolver_read_receipts` を同一境界に置く
- 物理配置は `short.db` 同居または専用 `resolver.db` を選択可能にする
- API / CLI / Skill からは保持先の違いを見せない

### 3.3 agent-taskstate 連携

本段階では `agent-taskstate` に直接書き込まない。

代わりに以下を行う。

- `read_receipt` 相当を `memx-resolver` 内に正規化して保持する
- `task_id` をキーに stale 判定を返す API を提供する
- `taskstate-export` で `agent-taskstate:task:local:<id>` と `memx:doc/chunk/card:local:<id>` の `typed_ref`、required docs、read receipts、stale reasons をまとめて返す
- doc に紐づく `tracker:issue:*:*` と `birdseye:view:local:*` 参照も `source_refs` に含め、issue / Birdseye view 起点の再開材料として渡す
- direct write ではなく export bridge とし、`agent-taskstate` 側の `context_bundle_source` へ取り込める payload 形状を `docs/interfaces.md` に合わせる

### 3.4 chunking

chunk 生成は見出し優先とする。

- Markdown 見出しを検出して section 化する
- section が長すぎる場合のみ固定長で再分割する
- `importance` は見出し名と `doc_type` から推定する
- LLM が役割を推定しやすいように、chunk には `memory_type` と `cue` を付与する

### 3.4.1 LLM 向け memory card

chunk は文書構造を保つ単位であり、そのままでは LLM にとって「制約」「手順」「受入条件」の区別が曖昧になりやすい。

そのため `chunks:get` は、chunk に加えて `memory_cards` を返せるようにする。

- `memory_type`: `acceptance` / `constraint` / `procedure` / `dependency` / `done` / `decision` / `risk` / `concept` / `reference`
- `cue`: 見出し階層を短く連結した検索・プロンプト用手がかり
- `statement`: 箇条書き単位、または chunk 本文から作る短い記述
- `doc_id` / `chunk_id`: 出典を失わないための参照

memory card は `memory_type`、`importance`、query match、feedback log、`token_budget` に基づいて並べ替え、予算内で優先度の高いものから返す。

実利用ログは `cards-feedback` で `used` / `helpful` / `pinned` / `irrelevant` / `skipped` として蓄積し、card ID と memory type の補正値として ranking に反映する。呼び出し側は `ranking_weights` で重みを上書きできる。

`cards:bundle` / `mem docs bundle` は、選ばれた card を Markdown または JSONL の prompt-ready bundle として出力し、`source_refs` に card / chunk / doc の `typed_ref` を含める。

これにより、LLM に渡す記憶は「長い本文」ではなく、役割・手がかり・出典・ランキング根拠を持つ prompt-ready な最小単位として扱える。

### 3.5 解決ロジック

文書解決は軽量な決定的ロジックを採用する。

- `task_id` 一致を最優先
- 次に `feature_keys` 一致
- 次に `tags` / `title` / `summary` / `body` の部分一致
- `required` / `recommended` は doc importance で振り分ける

### 3.6 契約情報

契約情報は以下の 2 系統から集約する。

- ingest 時に明示入力された配列
- 見出し名から抽出した section

抽出対象は以下とする。

- Acceptance Criteria
- Forbidden Patterns
- Definition of Done
- Dependencies

## 4. データモデル

4.1 から 4.4 の resolver 系テーブルは resolver store にまとめ、`short.db` とは独立して配置できるようにする。

### 4.1 resolver_documents

文書本体と契約系メタデータを保持する。

主な項目:

- `doc_id`
- `doc_type`
- `title`
- `source_path`
- `version`
- `version_scheme`
- `updated_at`
- `summary`
- `body`
- `tags_json`
- `feature_keys_json`
- `task_ids_json`
- `tracker_refs_json`
- `birdseye_refs_json`
- `acceptance_criteria_json`
- `forbidden_patterns_json`
- `definition_of_done_json`
- `dependencies_json`
- `importance`

### 4.2 resolver_chunks

文書の参照単位を保持する。

主な項目:

- `chunk_id`
- `doc_id`
- `heading`
- `heading_path_json`
- `ordinal`
- `body`
- `token_estimate`
- `importance`
- `memory_type`（レスポンス生成時に推定可能）
- `cue`（レスポンス生成時に推定可能）

### 4.3 resolver_document_links

文書間依存を保持する。

主な項目:

- `src_doc_id`
- `dst_doc_id`
- `link_type`

### 4.4 resolver_read_receipts

task と文書参照の対応を保持する。

主な項目:

- `task_id`
- `doc_id`
- `version`
- `chunk_ids_json`
- `chunk_snapshots_json`
- `previous_receipt_hash`
- `receipt_hash`
- `reader`
- `read_at`

`chunk_snapshots_json` は読了時点の `chunk_id`、本文 hash、heading path、memory type、importance、token estimate を保持する。stale 判定では最新版 chunk と比較し、読んだ chunk が変化した場合は `semantic_diff` として `impact_scope` / `changed_chunks` を返す。version だけが変わり読了 chunk が不変の場合は metadata impact の `version_mismatch` として扱う。

`previous_receipt_hash` / `receipt_hash` は task 単位の hash chain を形成し、`resolver_audit_log` の `reads_ack` 記録から後追い確認できる。

### 4.5 resolver_audit_log

resolver 操作の監査証跡を保持する。

主な項目:

- `operation`
- `actor`
- `target_type`
- `target_id`
- `result`
- `receipt_hash`
- `details_json`
- `recorded_at`

## 5. 既知の制約

- 既存 `short.db` 内の resolver データは `mem docs migrate-resolver-store --from short.db --to resolver.db --dry-run` で事前確認し、dry-run なしで専用 store へ移送できる
- version 比較は `version_scheme` により `semver` / `iso_datetime` / `git_revision` / `string` を区別する。scheme 未指定時は version 文字列から推定する
- stale 候補は read receipt の chunk snapshot と最新版 chunk を比較し、semantic diff と影響範囲を返す
- task dependency は外部正本に問い合わせず、ローカル保持の `task_ids_json` を優先利用する
- 全文検索は resolver 専用 FTS5 (`resolver_documents_fts` / `resolver_chunks_fts`) を使い、FTS が使えない場合に既存の軽量一致ロジックへフォールバックする

## 6. 完了条件

- `docs/interfaces.md` に記載した最小 API が呼べる
- 文書を登録すると chunk が生成される
- feature / task / topic から required / recommended docs を返せる
- read receipt 登録と stale 判定が動く
- contract resolve が acceptance / forbidden / DoD / dependencies を返せる
- resolver store を分離しても同じ API 契約で動作する
- `short.db` 同居 resolver store から専用 `resolver.db` へ CLI で移行できる
- memory card ranking が feedback と重み設定で補正できる
- prompt-ready bundle と agent-taskstate export が取得できる
- read receipt hash chain と audit log で ack 証跡を追跡できる
- tracker / Birdseye view refs と version_scheme が export / stale 判定に反映される
