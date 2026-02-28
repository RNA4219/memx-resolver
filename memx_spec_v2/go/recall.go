package db

import "context"

// Semantic Recall 用のスコープ。
type RecallScope string

const (
    ScopeSelf    RecallScope = "self"
    ScopeSession RecallScope = "session"
    // 将来: ScopeProject などを追加
)

// ストアの種類。
type StoreKind string

const (
    StoreShort     StoreKind = "short"
    StoreChronicle StoreKind = "chronicle"
    StoreMemopedia StoreKind = "memopedia"
    StoreArchive   StoreKind = "archive"
)

// RecallQuery は意味ベース検索のクエリ。
type RecallQuery struct {
    Text         string
    Scope        RecallScope
    Stores       []StoreKind
    TopK         int
    MessageRange int
}

// Note は notes テーブル相当の構造体（必要なフィールドのみ定義）。
type Note struct {
    ID       string
    Title    string
    Summary  string
    Body     string
    Store    StoreKind
    Score    float64
}

// NoteWithContext は anchor ノートと、その前後の文脈。
type NoteWithContext struct {
    Anchor Note
    Before []Note
    After  []Note
}

// Recall は指定されたクエリに基づいて Semantic Recall を実行する。
// 実装は今後追加する。Conn.Embed を利用して埋め込みを計算する想定。
func (c *Conn) Recall(ctx context.Context, q RecallQuery) ([]NoteWithContext, error) {
    // TODO: note_embeddings と FTS を組み合わせた実装を追加。
    // c.Embed が nil の場合はエラーを返す or FTS のみにフォールバックする。
    return nil, nil
}
