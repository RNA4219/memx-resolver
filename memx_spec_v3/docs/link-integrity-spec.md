# Link Integrity Spec

## 1. 目的
本仕様は、設計文書群の章内リンク・章間リンク・運用リンクの到達性を検証し、受け入れレビューで使用するリンク不達件数を算出する手順を定義する。

## 2. 検証対象
- `memx_spec_v3/docs/design.md`
- `memx_spec_v3/docs/interfaces.md`
- `memx_spec_v3/docs/requirements.md`
- `memx_spec_v3/docs/traceability.md`
- `memx_spec_v3/docs/operations-spec.md`
- `RUNBOOK.md`
- `EVALUATION.md`

## 3. 計測定義
- リンク不達: 次のいずれかを満たすリンク
  - 相対パス先ファイルが存在しない
  - アンカーが存在しない
  - 参照先セクション名が一致しない

## 4. 判定
- pass: リンク不達件数 0
- fail: リンク不達件数 1 以上

## 5. 出力
- `link_total`: 検証リンク総数
- `link_unreachable_count`: リンク不達件数
- `link_unreachable_list`: 不達リンク一覧（source, target, reason）
