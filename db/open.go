package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	// Pure-Go SQLite driver. If you prefer CGO-based sqlite3, you can swap this
	// import and the driver name in sql.Open.
	_ "modernc.org/sqlite"
)

// MustOpenAll は short.db をメインとして開き、
// journal / knowledge / archive / resolver を必要に応じて個別に初期化した Conn を返す。
func MustOpenAll(paths Paths) *Conn {
	c, err := OpenAll(paths)
	if err != nil {
		log.Fatalf("failed to open db(s): %v", err)
	}
	return c
}

// OpenAll は short.db をメインとして開き、
// journal / knowledge / archive は ATTACH、resolver は専用接続または short 同居で初期化する。
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
		log.Printf("warning: failed to set WAL mode: %v", err)
	}

	if err := migrateShort(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate short.db: %w", err)
	}

	resolverDB := db
	if paths.Resolver != "" && !sameDBPath(paths.Resolver, paths.Short) {
		resolverDB, err = openResolverDB(paths.Resolver)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("open resolver.db: %w", err)
		}
	} else {
		if err := migrateResolver(db); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("migrate resolver store in short.db: %w", err)
		}
	}

	if paths.Journal != "" {
		if err := migrateAndAttach(db, paths.Journal, "journal", migrateJournal); err != nil {
			_ = db.Close()
			if resolverDB != nil && resolverDB != db {
				_ = resolverDB.Close()
			}
			return nil, fmt.Errorf("migrate/attach journal.db: %w", err)
		}
	}
	if paths.Knowledge != "" {
		if err := migrateAndAttach(db, paths.Knowledge, "knowledge", migrateKnowledge); err != nil {
			_ = db.Close()
			if resolverDB != nil && resolverDB != db {
				_ = resolverDB.Close()
			}
			return nil, fmt.Errorf("migrate/attach knowledge.db: %w", err)
		}
	}
	if paths.Archive != "" {
		if err := migrateAndAttach(db, paths.Archive, "archive", migrateArchive); err != nil {
			_ = db.Close()
			if resolverDB != nil && resolverDB != db {
				_ = resolverDB.Close()
			}
			return nil, fmt.Errorf("migrate/attach archive.db: %w", err)
		}
	}

	return &Conn{
		DB:          db,
		ShortDB:     db,
		JournalDB:   openJournalDB(paths.Journal),
		KnowledgeDB: openKnowledgeDB(paths.Knowledge),
		ArchiveDB:   openArchiveDB(paths.Archive),
		ResolverDB:  resolverDB,
	}, nil
}

func openResolverDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s", path))
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		log.Printf("warning: failed to set resolver WAL mode: %v", err)
	}
	if err := migrateResolver(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func sameDBPath(left string, right string) bool {
	if left == "" || right == "" {
		return false
	}
	return filepath.Clean(left) == filepath.Clean(right)
}

// openJournalDB は journal.db を個別に開く。
func openJournalDB(path string) *sql.DB {
	if path == "" {
		return nil
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s", path))
	if err != nil {
		return nil
	}
	_ = migrateJournal(db)
	return db
}

// openKnowledgeDB は knowledge.db を個別に開く。
func openKnowledgeDB(path string) *sql.DB {
	if path == "" {
		return nil
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s", path))
	if err != nil {
		return nil
	}
	_ = migrateKnowledge(db)
	return db
}

// openArchiveDB は archive.db を個別に開く。
func openArchiveDB(path string) *sql.DB {
	if path == "" {
		return nil
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s", path))
	if err != nil {
		return nil
	}
	_ = migrateArchive(db)
	return db
}

// Close は開いている DB ハンドルを重複なくすべてクローズする。
func (c *Conn) Close() error {
	if c == nil {
		return nil
	}

	seen := map[*sql.DB]struct{}{}
	var errs []error
	for _, dbh := range []*sql.DB{c.ShortDB, c.JournalDB, c.KnowledgeDB, c.ArchiveDB, c.ResolverDB, c.DB} {
		if dbh == nil {
			continue
		}
		if _, ok := seen[dbh]; ok {
			continue
		}
		seen[dbh] = struct{}{}
		if err := dbh.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// migrateAndAttach は DB ファイルを個別に開いてマイグレーションし、その後 ATTACH する。
// これにより、各ストアのスキーマを独立して管理できる。
func migrateAndAttach(mainDB *sql.DB, path string, schemaName string, migrateFunc func(*sql.DB) error) error {
	dsn := fmt.Sprintf("file:%s", path)
	storeDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open %s: %w", schemaName, err)
	}
	defer storeDB.Close()

	if err := migrateFunc(storeDB); err != nil {
		return fmt.Errorf("migrate %s: %w", schemaName, err)
	}

	if err := storeDB.Close(); err != nil {
		return fmt.Errorf("close %s after migration: %w", schemaName, err)
	}

	attachSQL := fmt.Sprintf("ATTACH DATABASE '%s' AS %s;", path, schemaName)
	if _, err := mainDB.Exec(attachSQL); err != nil {
		return fmt.Errorf("attach %s: %w", schemaName, err)
	}

	return nil
}
