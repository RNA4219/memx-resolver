package db_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RNA4219/memx-resolver/v2/db"
	"github.com/RNA4219/memx-resolver/v2/service"
)

func TestV11DatabaseFixturesOpenInV2(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "testdata")
	tmp := t.TempDir()
	copyDB := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		path := filepath.Join(tmp, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
		return path
	}

	svc, err := service.New(db.Paths{
		Short:     copyDB("short.db"),
		Journal:   copyDB("chronicle.db"),
		Knowledge: copyDB("memopedia.db"),
		Archive:   copyDB("archive.db"),
	})
	if err != nil {
		t.Fatalf("v1.1 database fixtures must open in v2: %v", err)
	}
	defer func() { _ = svc.Close() }()

	var notes int
	if err := svc.Conn.ShortDB.QueryRow("SELECT COUNT(*) FROM notes").Scan(&notes); err != nil {
		t.Fatalf("query migrated short fixture: %v", err)
	}
}
