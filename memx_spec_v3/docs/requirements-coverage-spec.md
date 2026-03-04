# Requirements Coverage Spec

## 1. 目的
本仕様は、`memx_spec_v3/docs/requirements.md` と設計関連成果物の対応を比較し、受け入れレビューで使用する要件ID網羅率を算出する手順を定義する。

## 2. 入力
- `memx_spec_v3/docs/requirements.md`
- `memx_spec_v3/docs/traceability.md`
- `memx_spec_v3/docs/design.md`
- `memx_spec_v3/docs/interfaces.md`

## 3. 算出定義
- 母数: `requirements.md` に定義された `REQ-*` の総件数
- 分子: `traceability.md` で `Design Mapping` と `Interface Mapping` が有効な `REQ-*` 件数
- 網羅率: `分子 / 母数 * 100`（小数第2位四捨五入）

## 4. 判定
- pass: 網羅率 100%
- fail: 網羅率 100% 未満

## 5. 出力
- `coverage_total`: 母数
- `coverage_matched`: 分子
- `coverage_rate`: 網羅率（%）
- `missing_req_ids`: 未充足 `REQ-*` 一覧
