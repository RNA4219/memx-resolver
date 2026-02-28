package db

import (
    "database/sql"
    "fmt"
    "log"
)

// migrateFn は ATTACH 後にスキーマを適用する関数の型。
type migrateFn func(db *sql.DB, schemaName string) error

// MustOpenAll は short.db をメインとして開き、
// chronicle / memopedia / archive を ATTACH + migrate した Conn を返す。
func MustOpenAll(paths Paths) *Conn {
    dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_journal_mode=WAL", paths.Short)

    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        log.Fatalf("failed to open short.db: %v", err)
    }

    if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
        log.Fatalf("failed to enable foreign_keys: %v", err)
    }
    if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
        log.Printf("warning: failed to set WAL mode: %v", err)
    }

    if err := migrateShort(db); err != nil {
        log.Fatalf("failed to migrate short.db: %v", err)
    }

    if paths.Chronicle != "" {
        if err := attachAndMigrate(db, paths.Chronicle, "chronicle", migrateChronicle); err != nil {
            log.Fatalf("failed to attach/migrate chronicle.db: %v", err)
        }
    }
    if paths.Memopedia != "" {
        if err := attachAndMigrate(db, paths.Memopedia, "memopedia", migrateMemopedia); err != nil {
            log.Fatalf("failed to attach/migrate memopedia.db: %v", err)
        }
    }
    if paths.Archive != "" {
        if err := attachAndMigrate(db, paths.Archive, "archive", migrateArchive); err != nil {
            log.Fatalf("failed to attach/migrate archive.db: %v", err)
        }
    }

    return &Conn{DB: db}
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
