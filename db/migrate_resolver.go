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
  version_scheme TEXT NOT NULL DEFAULT 'string',
  updated_at TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL,
  tags_json TEXT NOT NULL DEFAULT '[]',
  feature_keys_json TEXT NOT NULL DEFAULT '[]',
  task_ids_json TEXT NOT NULL DEFAULT '[]',
  tracker_refs_json TEXT NOT NULL DEFAULT '[]',
  birdseye_refs_json TEXT NOT NULL DEFAULT '[]',
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

CREATE VIRTUAL TABLE IF NOT EXISTS resolver_documents_fts USING fts5(
  doc_id UNINDEXED,
  doc_type,
  title,
  source_path,
  summary,
  body,
  tags,
  feature_keys
);

CREATE TRIGGER IF NOT EXISTS resolver_documents_ai AFTER INSERT ON resolver_documents BEGIN
  INSERT INTO resolver_documents_fts(doc_id, doc_type, title, source_path, summary, body, tags, feature_keys)
  VALUES (new.doc_id, new.doc_type, new.title, new.source_path, new.summary, new.body, new.tags_json, new.feature_keys_json);
END;

CREATE TRIGGER IF NOT EXISTS resolver_documents_au AFTER UPDATE ON resolver_documents BEGIN
  DELETE FROM resolver_documents_fts WHERE doc_id = old.doc_id;
  INSERT INTO resolver_documents_fts(doc_id, doc_type, title, source_path, summary, body, tags, feature_keys)
  VALUES (new.doc_id, new.doc_type, new.title, new.source_path, new.summary, new.body, new.tags_json, new.feature_keys_json);
END;

CREATE TRIGGER IF NOT EXISTS resolver_documents_ad AFTER DELETE ON resolver_documents BEGIN
  DELETE FROM resolver_documents_fts WHERE doc_id = old.doc_id;
END;

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

CREATE VIRTUAL TABLE IF NOT EXISTS resolver_chunks_fts USING fts5(
  chunk_id UNINDEXED,
  doc_id UNINDEXED,
  heading,
  heading_path,
  body,
  importance
);

CREATE TRIGGER IF NOT EXISTS resolver_chunks_ai AFTER INSERT ON resolver_chunks BEGIN
  INSERT INTO resolver_chunks_fts(chunk_id, doc_id, heading, heading_path, body, importance)
  VALUES (new.chunk_id, new.doc_id, new.heading, new.heading_path_json, new.body, new.importance);
END;

CREATE TRIGGER IF NOT EXISTS resolver_chunks_au AFTER UPDATE ON resolver_chunks BEGIN
  DELETE FROM resolver_chunks_fts WHERE chunk_id = old.chunk_id;
  INSERT INTO resolver_chunks_fts(chunk_id, doc_id, heading, heading_path, body, importance)
  VALUES (new.chunk_id, new.doc_id, new.heading, new.heading_path_json, new.body, new.importance);
END;

CREATE TRIGGER IF NOT EXISTS resolver_chunks_ad AFTER DELETE ON resolver_chunks BEGIN
  DELETE FROM resolver_chunks_fts WHERE chunk_id = old.chunk_id;
END;

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
  previous_receipt_hash TEXT NOT NULL DEFAULT '',
  receipt_hash TEXT NOT NULL DEFAULT '',
  reader TEXT NOT NULL,
  read_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_resolver_read_receipts_task
  ON resolver_read_receipts(task_id, read_at);

CREATE INDEX IF NOT EXISTS idx_resolver_read_receipts_hash
  ON resolver_read_receipts(receipt_hash);

CREATE TABLE IF NOT EXISTS resolver_audit_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  operation TEXT NOT NULL,
  actor TEXT NOT NULL DEFAULT '',
  target_type TEXT NOT NULL DEFAULT '',
  target_id TEXT NOT NULL DEFAULT '',
  result TEXT NOT NULL,
  receipt_hash TEXT NOT NULL DEFAULT '',
  details_json TEXT NOT NULL DEFAULT '{}',
  recorded_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_resolver_audit_log_operation
  ON resolver_audit_log(operation, recorded_at);

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
	if err := ensureResolverColumn(db, "resolver_read_receipts", "previous_receipt_hash", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureResolverColumn(db, "resolver_read_receipts", "receipt_hash", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureResolverColumn(db, "resolver_documents", "tracker_refs_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		return err
	}
	if err := ensureResolverColumn(db, "resolver_documents", "birdseye_refs_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		return err
	}
	if err := ensureResolverColumn(db, "resolver_documents", "version_scheme", "TEXT NOT NULL DEFAULT 'string'"); err != nil {
		return err
	}
	if err := rebuildResolverFTS(db); err != nil {
		return err
	}
	return nil
}

func rebuildResolverFTS(db *sql.DB) error {
	if _, err := db.Exec(`
DELETE FROM resolver_documents_fts;
INSERT INTO resolver_documents_fts(doc_id, doc_type, title, source_path, summary, body, tags, feature_keys)
SELECT doc_id, doc_type, title, source_path, summary, body, tags_json, feature_keys_json
FROM resolver_documents;

DELETE FROM resolver_chunks_fts;
INSERT INTO resolver_chunks_fts(chunk_id, doc_id, heading, heading_path, body, importance)
SELECT chunk_id, doc_id, heading, heading_path_json, body, importance
FROM resolver_chunks;
`); err != nil {
		return fmt.Errorf("rebuild resolver fts: %w", err)
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
