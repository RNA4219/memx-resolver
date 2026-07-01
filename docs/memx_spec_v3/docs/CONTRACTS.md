---
owner: memx-core
status: active
last_reviewed_at: 2026-03-03
next_review_due: 2026-06-03
---

# CONTRACTS

本書は**正本スキーマへの索引・抜粋**を提供する。フィールド定義そのものの正本は保持しない。

- API 正本: `memx_spec_v3/docs/contracts/openapi.yaml`
- CLI `--json` 正本: `memx_spec_v3/docs/contracts/cli-json.schema.json`

## 1. 役割定義

- `interfaces.md`
  - 人間可読の I/O 説明と互換方針を示す補助仕様。
- `CONTRACTS.md`（本書）
  - 正本スキーマへの索引・抜粋のみを扱う。
  - 重複定義（フィールド型・required・制約の再定義）を持たない。

## 2. API 契約索引（openapi.yaml）

- `POST /v1/notes:ingest`
  - Request: `#/components/schemas/NotesIngestRequest`
  - Response 200: `#/components/schemas/NotesIngestResponse`
- `POST /v1/notes:search`
  - Request: `#/components/schemas/NotesSearchRequest`
  - Response 200: `#/components/schemas/NotesSearchResponse`
- `GET /v1/notes/{id}`
  - Response 200: `#/components/schemas/Note`
- `POST /v1/docs:ingest`
  - Request: `#/components/schemas/DocsIngestRequest`
  - Response 200: `#/components/schemas/DocsIngestResponse`
- `POST /v1/docs:resolve`
  - Request: `#/components/schemas/DocsResolveRequest`
  - Response 200: `#/components/schemas/DocsResolveResponse`
- `POST /v1/chunks:get`
  - Request: `#/components/schemas/ChunksGetRequest`
  - Response 200: `#/components/schemas/ChunksGetResponse`
- `POST /v1/docs:search`
  - Request: `#/components/schemas/DocsSearchRequest`
  - Response 200: `#/components/schemas/DocsSearchResponse`
- `POST /v1/cards:search`
  - Request: `#/components/schemas/CardsSearchRequest`
  - Response 200: `#/components/schemas/CardsSearchResponse`
- `POST /v1/cards:feedback`
  - Request: `#/components/schemas/CardFeedbackRequest`
  - Response 200: `#/components/schemas/CardFeedbackResponse`
- `POST /v1/cards:bundle`
  - Request: `#/components/schemas/PromptBundleRequest`
  - Response 200: `#/components/schemas/PromptBundleResponse`
- `POST /v1/taskstate:export`
  - Request: `#/components/schemas/TaskStateExportRequest`
  - Response 200: `#/components/schemas/TaskStateExportResponse`
- `POST /v1/reads:ack`
  - Request: `#/components/schemas/ReadsAckRequest`
  - Response 200: `#/components/schemas/ReadsAckResponse`
- `POST /v1/docs:stale-check`
  - Request: `#/components/schemas/DocsStaleCheckRequest`
  - Response 200: `#/components/schemas/DocsStaleCheckResponse`
- `POST /v1/contracts:resolve`
  - Request: `#/components/schemas/ContractsResolveRequest`
  - Response 200: `#/components/schemas/ContractsResolveResponse`

## 3. CLI JSON 契約索引（cli-json.schema.json）

- `mem in short --json`
  - 正本: `#/definitions/NotesIngestResponse`
- `mem out search --json`
  - 正本: `#/definitions/NotesSearchResponse`
- `mem out show --json`
  - 正本: `#/definitions/Note`
- `mem docs ingest --json`
  - 正本: `#/definitions/DocsIngestResponse`
- `mem docs resolve --json`
  - 正本: `#/definitions/DocsResolveResponse`
- `mem docs chunks --json`
  - 正本: `#/definitions/ChunksGetResponse`
- `mem docs search --json`
  - 正本: `#/definitions/DocsSearchResponse`
- `mem docs cards --json`
  - 正本: `#/definitions/CardsSearchResponse`
- `mem docs cards-feedback --json`
  - 正本: `#/definitions/CardFeedbackResponse`
- `mem docs bundle --json`
  - 正本: `#/definitions/PromptBundleResponse`
- `mem docs taskstate-export --json`
  - 正本: `#/definitions/TaskStateExportResponse`
- `mem docs ack --json`
  - 正本: `#/definitions/ReadsAckResponse`
- `mem docs stale --json`
  - 正本: `#/definitions/DocsStaleCheckResponse`
- `mem docs contract --json`
  - 正本: `#/definitions/ContractsResolveResponse`

## 4. エラー契約索引

- HTTP/API エラー正本: `contracts/openapi.yaml` の `components.responses` / `components.schemas.Error*`
- 運用向け要約: `error-contract.md`

## 5. レビュー時の必須確認

- 契約差分がある PR は、`operations-spec.md` の「契約差分チェック手順」を実施し、
  `openapi.yaml` / `cli-json.schema.json` を起点にレビューする。
