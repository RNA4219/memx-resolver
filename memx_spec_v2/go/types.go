package db

import "database/sql"

// Paths は各ストアDBファイルのパスを保持する。
type Paths struct {
    Short     string
    Chronicle string
    Memopedia string
    Archive   string
}

// Conn は ATTACH 済みのメイン接続と、LLM/Gatekeeper クライアントをラップする。
type Conn struct {
    DB *sql.DB

    // LLM 系クライアント（必要に応じて設定される）
    Embed   EmbeddingClient
    Mini    MiniLLMClient
    Reflect ReflectLLMClient

    // Gatekeeper（保存前・出力前チェック用）
    Gate Gatekeeper
}
