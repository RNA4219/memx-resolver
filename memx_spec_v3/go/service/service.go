package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"memx/db"
)

// Service は memx の Usecase 層。
// DB/LLM/Gatekeeper などの実装詳細を隠蔽し、API から呼び出される。
type Service struct {
	Conn *db.Conn
}

func New(paths db.Paths) (*Service, error) {
	c, err := db.OpenAll(paths)
	if err != nil {
		return nil, err
	}
	return &Service{Conn: c}, nil
}

func (s *Service) Close() error {
	if s == nil || s.Conn == nil {
		return nil
	}
	return s.Conn.Close()
}

// IngestNoteRequest は short への投入（最小）。
// v1.3 では CLI はこれを API に渡すだけ。
type IngestNoteRequest struct {
	Title       string
	Body        string
	Summary     string
	SourceType  string
	Origin      string
	SourceTrust string
	Sensitivity string
	Tags        []string
}

func (r *IngestNoteRequest) normalize() {
	r.Title = strings.TrimSpace(r.Title)
	r.Body = strings.TrimSpace(r.Body)
	r.SourceType = strings.TrimSpace(r.SourceType)
	r.Origin = strings.TrimSpace(r.Origin)
	r.SourceTrust = strings.TrimSpace(r.SourceTrust)
	r.Sensitivity = strings.TrimSpace(r.Sensitivity)

	if r.SourceType == "" {
		r.SourceType = "manual"
	}
	if r.SourceTrust == "" {
		r.SourceTrust = "user_input"
	}
	if r.Sensitivity == "" {
		r.Sensitivity = "internal"
	}
}

// IngestShort は short.db にノートを保存する。
// LLM / Gatekeeper は今後差し込む（v1.3 ではフックのみ）。
func (s *Service) IngestShort(ctx context.Context, req IngestNoteRequest) (Note, error) {
	req.normalize()
	if req.Title == "" || req.Body == "" {
		return Note{}, fmt.Errorf("%w: title/body is required", ErrInvalidArgument)
	}

	id, err := newUUIDLike()
	if err != nil {
		return Note{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := s.Conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return Note{}, err
	}
	defer func() { _ = tx.Rollback() }()

	// NOTE: summary は v1.3 時点では空でもよい。
	_, err = tx.ExecContext(ctx, `
INSERT INTO notes(
  id, title, summary, body,
  created_at, updated_at, last_accessed_at,
  access_count,
  source_type, origin, source_trust, sensitivity
) VALUES(?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
`, id, req.Title, req.Summary, req.Body, now, now, now, req.SourceType, req.Origin, req.SourceTrust, req.Sensitivity)
	if err != nil {
		return Note{}, err
	}

	for _, t := range req.Tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if err := upsertTagAndBind(ctx, tx, id, t, now); err != nil {
			return Note{}, err
		}
	}

	// meta（近似でOK）
	_, _ = tx.ExecContext(ctx, `UPDATE short_meta SET value = CAST(CAST(value AS INTEGER) + 1 AS TEXT) WHERE key = 'note_count';`)

	if err := tx.Commit(); err != nil {
		return Note{}, err
	}

	return Note{
		ID:             id,
		Title:          req.Title,
		Summary:        req.Summary,
		Body:           req.Body,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastAccessedAt: now,
		AccessCount:    0,
		SourceType:     req.SourceType,
		Origin:         req.Origin,
		SourceTrust:    req.SourceTrust,
		Sensitivity:    req.Sensitivity,
	}, nil
}

func upsertTagAndBind(ctx context.Context, tx *sql.Tx, noteID, tag, now string) error {
	// route は最小実装として short_only。
	_, err := tx.ExecContext(ctx, `
INSERT INTO tags(name, route, parent_id, created_at, updated_at, usage_count)
VALUES(?, 'short_only', NULL, ?, ?, 0)
ON CONFLICT(name) DO UPDATE SET updated_at = excluded.updated_at;
`, tag, now, now)
	if err != nil {
		return err
	}

	var tagID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?;`, tag).Scan(&tagID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO note_tags(note_id, tag_id) VALUES(?, ?);`, noteID, tagID); err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE tags SET usage_count = usage_count + 1 WHERE id = ?;`, tagID)
	return nil
}

// SearchShort は FTS5 を使った検索。
func (s *Service) SearchShort(ctx context.Context, query string, topK int) ([]Note, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrInvalidArgument)
	}
	if topK <= 0 {
		topK = 20
	}

	rows, err := s.Conn.DB.QueryContext(ctx, `
SELECT n.id, n.title, n.summary, n.body,
       n.created_at, n.updated_at, n.last_accessed_at, n.access_count,
       n.source_type, n.origin, n.source_trust, n.sensitivity
FROM notes_fts
JOIN notes n ON notes_fts.rowid = n.rowid
WHERE notes_fts MATCH ?
ORDER BY bm25(notes_fts)
LIMIT ?;
`, query, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Note, 0, topK)
	for rows.Next() {
		var n Note
		if err := rows.Scan(
			&n.ID, &n.Title, &n.Summary, &n.Body,
			&n.CreatedAt, &n.UpdatedAt, &n.LastAccessedAt, &n.AccessCount,
			&n.SourceType, &n.Origin, &n.SourceTrust, &n.Sensitivity,
		); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetShort は id 指定でノートを取得し、アクセス情報を更新する。
func (s *Service) GetShort(ctx context.Context, id string) (Note, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Note{}, fmt.Errorf("%w: id is required", ErrInvalidArgument)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := s.Conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return Note{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var n Note
	err = tx.QueryRowContext(ctx, `
SELECT id, title, summary, body,
       created_at, updated_at, last_accessed_at, access_count,
       source_type, origin, source_trust, sensitivity
FROM notes WHERE id = ?;
`, id).Scan(
		&n.ID, &n.Title, &n.Summary, &n.Body,
		&n.CreatedAt, &n.UpdatedAt, &n.LastAccessedAt, &n.AccessCount,
		&n.SourceType, &n.Origin, &n.SourceTrust, &n.Sensitivity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Note{}, ErrNotFound
		}
		return Note{}, err
	}

	_, _ = tx.ExecContext(ctx, `
UPDATE notes
SET last_accessed_at = ?, access_count = access_count + 1
WHERE id = ?;
`, now, id)

	if err := tx.Commit(); err != nil {
		return Note{}, err
	}

	n.LastAccessedAt = now
	n.AccessCount++
	return n, nil
}

func newUUIDLike() (string, error) {
	// 32 hex = 128-bit。UUIDv4 に厳密に合わせる必要は v1 ではない。
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
