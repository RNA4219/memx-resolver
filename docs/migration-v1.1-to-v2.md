# v1.1.0からv2.0.0への移行ガイド

## 変更の要点

v2.0.0ではGo実装をリポジトリルートのmoduleへ移し、typed_refとHTTP公開ポリシーをstrict化します。SQLite schemaと既存のHTTP endpointは維持します。

## typed_ref

v1.1.0は3セグメントを読み取れましたが、v2.0.0は4セグメントだけを受理します。

| v1.1.0 | v2.0.0 |
| --- | --- |
| `memx:doc:ABC` | 拒否。providerを補って `memx:doc:local:ABC` |
| `memx:doc:local:A:B` | 拒否。IDのコロンを `memx:doc:local:A%3AB` に変換 |
| `TRACKER:Issue:JIRA:Proj-ABC` | `tracker:issue:jira:Proj-ABC` として正規化 |
| IDの大文字小文字 | 保持 |
| domain/type/provider | 小文字化 |
| malformed percent encoding | 拒否 |

entity IDはRFC 3986のunreserved文字（英数字、`-`, `.`, `_`, `~`）以外をUTF-8 byte単位でpercent-encodeします。parse後のIDは論理値へ戻り、再出力時に同じcanonical表現になります。

## Go import path

v1.1.0の内部module参照を次へ変更します。

```go
import "github.com/RNA4219/memx-resolver/v2/api"
import "github.com/RNA4219/memx-resolver/v2/service"
```

CLIはリポジトリルートから実行します。

```bash
go run ./cmd/mem --help
go test ./... -count=1
```

## bind policy

loopback（`127.0.0.1`, `[::1]`, `localhost`）は従来どおり起動できます。

```bash
mem api serve --addr 127.0.0.1:7766
```

wildcard、空host、非loopback IP、`localhost`以外のhostnameは、明示的な許可がないと起動に失敗します。

```bash
mem api serve --addr 0.0.0.0:7766 --allow-non-loopback
```

許可時も認証・TLSなしの警告をstderrへ出します。v2ではJSON decoderがunknown fieldを拒否します。

## DB互換性

DB migrationは不要です。v1.1.0で作成したSQLiteファイルをv2.0.0で直接開けることを互換テストの前提とします。移行前にはDBファイルをバックアップしてください。

## rollback

1. DBファイルのコピーを保存する。
2. GitHub Releaseのv1.1.0 binaryまたはtagへ戻す。
3. 3セグメント入力を使っている呼び出し元は、v1.1形式へ戻す前に4セグメントcanonical値を保存しておく。
4. v2で拒否されたunknown fieldを除去してからv1.1 endpointへ再送する。

v2.0.0の公開はv1.1.0のGitHub Release検証とは分離したrelease判定で行います。
