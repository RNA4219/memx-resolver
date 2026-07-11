# v1.1.0 Manual Black-Box Gate

- target_head: `c36170f7957494db1a5e6c0b3eb5a73fe848c4d4`
- environment: Windows local clean detached worktree
- profile: standard
- intake_status: ok

## 1. 根拠付き観点

| ID | 観点 | View | Technique | Source |
| --- | --- | --- | --- | --- |
| OBS-REF-01 | legacy 3-segmentを警告付きで読みcanonical 4-segmentを出力する | black | equivalence/state | typed_ref v1.1 contract |
| OBS-PATH-01 | absolute/`..`/symlink escapeを拒否し一括ackを保存しない | black | invalid-single-fault | workflow path contract |
| OBS-STORE-01 | concurrent ack、corrupt JSON、lock timeoutがfail-closedになる | black | state/boundary | atomic store contract |
| OBS-HTTP-01 | body上限超過は413、複数JSON値は400になる | black | boundary/equivalence | HTTP v1.1 contract |
| OBS-BIND-01 | loopbackと外部bindを正しく分類する | black | decision table | CLI bind contract |
| OBS-CONTRACT-01 | CLIの実responseがOpenAPI schemaへ適合する | black | contract | OpenAPI |
| OBS-PACKAGE-01 | wheelをrepo外からimportしfactoryを解決できる | black | install smoke | Python package contract |

## 2. リスク

| ID | Scenario | Score | Priority | Rationale |
| --- | --- | ---: | --- | --- |
| RISK-01 | path escapeでrepo外を読み書きする | 68 | P1 | データ境界への影響が大きい |
| RISK-02 | concurrent更新でreceiptを欠落・破損する | 61 | P1 | 再現困難で監査情報を損なう |
| RISK-03 | oversized/multi-value JSONでHTTP資源を消費する | 58 | P1 | 公開bind時の可用性に影響する |
| RISK-04 | 配布wheelがsource treeなしでimport不能 | 55 | P1 | PyPI利用者全体を阻害する |

## 3. 優先度

- P0: なし
- P1: path/store、HTTP境界、OpenAPI契約、wheel installを100%実行
- P2: legacy warning文言、platform固有symlink作成可否

## 4. 手動テストケース

| TC | Priority | Result | Oracle | Evidence |
| --- | --- | --- | --- | --- |
| TC-REF-01 | P1 | PASS | typed_ref v1.1 contract | `python-blackbox.txt` |
| TC-PATH-01 | P1 | PASS | repository boundary contract | `python-blackbox.txt` |
| TC-STORE-01 | P1 | PASS | atomic JSON contract | `python-blackbox.txt` |
| TC-HTTP-01 | P1 | PASS | HTTP 413/400 contract | `http-blackbox.txt` |
| TC-BIND-01 | P1 | PASS | loopback decision table | `http-blackbox.txt` |
| TC-CONTRACT-01 | P1 | PASS | OpenAPI schemas | `openapi-blackbox.txt` |
| TC-PACKAGE-01 | P1 | PASS | factory import path | `wheel-install-blackbox.txt` |
| TC-PROC-01 | P1 | PASS | residual count must be zero | `residual-processes.txt` |

Windowsでsymlink権限がない1ケースはSKIP。absolute/`..`とWindows拡張pathは別ケースでPASSしており、Linux CIでsymlinkケースを必須実行する。

## 5. 工数

- prep: 20分
- execution: 35分
- evidence: 20分
- retry buffer: 15分
- total: 90分

## 6. Gate 判定

- decision: `go`（local manual-BB）
- P0/P1: 8/8 PASS
- blocking_risks: なし
- waivers: なし
- release条件: Linux raceおよびGitHub CI成功、QEG final GO

## 7. Go/No-Go Brief

- feature: memx-resolver v1.1.0
- decision: local manual gate GO
- top risks: path escape、concurrent receipt、HTTP resource limit、package install
- evidence: CTG zero unresolved findings、HATE eligible/completeness 1.0、P1 8/8 PASS
- residual risk: Windowsでsymlinkケース1件SKIP、Linux CIで解消必須
- required follow-up: full QEG fixtureを実行し、GO以外ならtagを禁止する
