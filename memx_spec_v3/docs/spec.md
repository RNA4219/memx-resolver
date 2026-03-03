---
owner: memx-core
status: active
last_reviewed_at: 2026-03-03
next_review_due: 2026-06-03
priority: high
---

# memx 仕様インデックス（正本/補助の定義）

本書は `memx_spec_v3/docs/` の仕様参照導線を一本化するためのインデックスである。

## 1. 文書の役割分担（正本/補助）

### 1-1. 正本（Normative）

- `requirements.md`
  - memx v1.3 の仕様正本。
  - 要件ID、スコープ定義、API/CLI/NFR/運用要件の判定根拠は本書を優先する。
- `contracts/openapi.yaml`
  - HTTP API 契約とエラー応答スキーマの正本。
- `contracts/cli-json.schema.json`
  - CLI `--json` 出力契約の正本。

### 1-2. 補助（Secondary）

- `error-contract.md`
  - エラー契約の運用要約。正本は `contracts/openapi.yaml`。
- `quickstart.md`
  - 起動/疎通確認手順。契約定義の正本ではない。

## 2. 要件別の参照導線

### 2-1. API 契約

- 正本: `contracts/openapi.yaml`
- 仕様背景と受け入れ条件: `requirements.md` の「6. API 要件（v1.3 追加）」

### 2-2. CLI 契約

- 正本: `contracts/cli-json.schema.json`
- CLI 要件: `requirements.md` の「3. CLI 要件」

### 2-3. エラー契約

- 正本: `contracts/openapi.yaml`（`components.responses` / `components.schemas.Error*`）
- 運用要約: `error-contract.md`
- 要件側の基準: `requirements.md` の「6-4. エラーモデル」

### 2-4. NFR（非機能要件）

- 正本: `requirements.md` の「5. 非機能要件」

### 2-5. 運用要件

- 正本: `requirements.md` の「11. インシデント対応要件（運用）」

## 3. 更新順序（契約変更時）

1. 正本（`requirements.md` または `contracts/*.yaml|json`）を更新する。
2. 補助文書（`error-contract.md` / `quickstart.md`）へ差分要約のみ反映する。
3. 必要に応じて `memx_spec_v3/README.md` の導線を更新する。
