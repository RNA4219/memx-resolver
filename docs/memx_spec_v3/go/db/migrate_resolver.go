package db

import (
	"database/sql"
	"fmt"
)

const resolverSchemaSQL = `PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS resolver_documents (
  doc_id TEXT PRIMARY KEY,
  doc_type TEXT NOT NULL,
  title TEXT NOT NULL,
  source_path TEXT NOT NULL DEFAULT '',
  version TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL,
  tags_json TEXT NOT NULL DEFAULT '[]',
  feature_keys_json TEXT NOT NULL DEFAULT '[]',
  task_ids_json TEXT NOT NULL DEFAULT '[]',
  acceptance_criteria_json TEXT NOT NULL DEFAULT '[]',
  forbidden_patterns_json TEXT NOT NULL DEFAULT '[]',
  definition_of_done_json TEXT NOT NULL DEFAULT '[]',
  dependencies_json TEXT NOT NULL DEFAULT '[]',
  importance TEXT NOT NULL DEFAULT 'reference'
);

CREATE INDEX IF NOT EXISTS idx_resolver_documents_type
  ON resolver_documents(doc_type);

CREATE INDEX IF NOT EXISTS idx_resolver_documents_version
  ON resolver_documents(version);

CREATE TABLE IF NOT EXISTS resolver_chunks (
  chunk_id TEXT PRIMARY KEY,
  doc_id TEXT NOT NULL,
  heading TEXT NOT NULL DEFAULT '',
  heading_path_json TEXT NOT NULL DEFAULT '[]',
  ordinal INTEGER NOT NULL,
  body TEXT NOT NULL,
  token_estimate INTEGER NOT NULL DEFAULT 0,
  importance TEXT NOT NULL DEFAULT 'reference',
  FOREIGN KEY (doc_id) REFERENCES resolver_documents(doc_id)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_resolver_chunks_doc
  ON resolver_chunks(doc_id, ordinal);

CREATE TABLE IF NOT EXISTS resolver_document_links (
  src_doc_id TEXT NOT NULL,
  dst_doc_id TEXT NOT NULL,
  link_type TEXT NOT NULL,
  PRIMARY KEY (src_doc_id, dst_doc_id, link_type)
);

CREATE TABLE IF NOT EXISTS resolver_read_receipts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  task_id TEXT NOT NULL,
  doc_id TEXT NOT NULL,
  version TEXT NOT NULL,
  chunk_ids_json TEXT NOT NULL DEFAULT '[]',
  chunk_snapshots_json TEXT NOT NULL DEFAULT '[]',
  reader TEXT NOT NULL,
  read_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_resolver_read_receipts_task
  ON resolver_read_receipts(task_id, read_at);

CREATE TABLE IF NOT EXISTS resolver_memory_card_feedback (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  card_id TEXT NOT NULL,
  doc_id TEXT NOT NULL DEFAULT '',
  chunk_id TEXT NOT NULL DEFAULT '',
  memory_type TEXT NOT NULL DEFAULT '',
  signal TEXT NOT NULL,
  weight INTEGER NOT NULL DEFAULT 1,
  query TEXT NOT NULL DEFAULT '',
  recorded_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_resolver_card_feedback_card
  ON resolver_memory_card_feedback(card_id, recorded_at);

CREATE INDEX IF NOT EXISTS idx_resolver_card_feedback_type
  ON resolver_memory_card_feedback(memory_type, recorded_at);

PRAGMA user_version = 1;
`

// migrateResolver は resolver store に対して resolver 系スキーマを適用する。
func migrateResolver(db *sql.DB) error {
	if _, err := db.Exec(resolverSchemaSQL); err != nil {
		return fmt.Errorf("apply resolver schema: %w", err)
	}
	if err := ensureResolverColumn(db, "resolver_read_receipts", "chunk_snapshots_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		return err
	}
	return nil
}

func ensureResolverColumn(db *sql.DB, table string, column string, definition string) error {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", table))
	if err != nil {
		return fmt.Errorf("inspect %s schema: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("scan %s schema: %w", table, err)
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate %s schema: %w", table, err)
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table, column, definition)); err != nil {
		return fmt.Errorf("add %s.%s: %w", table, column, err)
	}
	return nil
}
