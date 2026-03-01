# Changes

## v3 (requirements v1.3)

- CLI と API を分離：CLI は API の薄いラッパ。
- API（HTTP + in-proc）を追加：`/v1/notes:ingest`, `/v1/notes:search`, `/v1/notes/{id}` など。
- Service(usecase) 層を追加：短期ストアの ingest/search/get を最小実装。
- DB 層を `go/db` に整理（OpenAll を追加、MustOpenAll は互換）。
