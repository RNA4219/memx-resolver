package db

import (
	"database/sql"
	"fmt"
	"log"

	// Pure-Go SQLite driver. If you prefer CGO-based sqlite3, you can swap this
	// import and the driver name in sql.Open.
	_ "modernc.org/sqlite"
)

// migrateFn は ATTACH 後にスキーマを適用する関数の型。
type migrateFn func(db *sql.DB, schemaName string) error

// MustOpenAll は short.db をメインとして開き、
// chronicle / memopedia / archive を ATTACH + migrate した Conn を返す。
func MustOpenAll(paths Paths) *Conn {
	c, err := OpenAll(paths)
	if err != nil {
		log.Fatalf("failed to open db(s): %v", err)
	}
	return c
}

// OpenAll は short.db をメインとして開き、
// chronicle / memopedia / archive を ATTACH + migrate した Conn を返す。
//
// v1.3 以降では、この関数を Service / API 層から呼び出し、
// CLI は直接 DB を触らない。
func OpenAll(paths Paths) (*Conn, error) {
	dsn := fmt.Sprintf("file:%s", paths.Short)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open short.db: %w", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		// WAL が使えない環境もあるため warning 扱い。
		log.Printf("warning: failed to set WAL mode: %v", err)
	}

	if err := migrateShort(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate short.db: %w", err)
	}

	if paths.Chronicle != "" {
		if err := attachAndMigrate(db, paths.Chronicle, "chronicle", migrateChronicle); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("attach/migrate chronicle.db: %w", err)
		}
	}
	if paths.Memopedia != "" {
		if err := attachAndMigrate(db, paths.Memopedia, "memopedia", migrateMemopedia); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("attach/migrate memopedia.db: %w", err)
		}
	}
	if paths.Archive != "" {
		if err := attachAndMigrate(db, paths.Archive, "archive", migrateArchive); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("attach/migrate archive.db: %w", err)
		}
	}

	return &Conn{DB: db}, nil
}

// Close は基盤となる *sql.DB をクローズする。
func (c *Conn) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// attachAndMigrate は DB を ATTACH し、指定された migrateFn を呼び出す。
func attachAndMigrate(db *sql.DB, path string, schemaName string, fn migrateFn) error {
	attachSQL := fmt.Sprintf("ATTACH DATABASE '%s' AS %s;", path, schemaName)
	if _, err := db.Exec(attachSQL); err != nil {
		return fmt.Errorf("attach %s: %w", schemaName, err)
	}

	if err := fn(db, schemaName); err != nil {
		return fmt.Errorf("migrate %s: %w", schemaName, err)
	}
	return nil
}
