package db

import "database/sql"

// TODO: chronicle / memopedia / archive 用スキーマを定義する。

func migrateChronicle(db *sql.DB, schemaName string) error {
    _ = schemaName
    // 将来的に chronicle 用 DDL をここで適用する。
    return nil
}

func migrateMemopedia(db *sql.DB, schemaName string) error {
    _ = schemaName
    // 将来的に memopedia 用 DDL をここで適用する。
    return nil
}

func migrateArchive(db *sql.DB, schemaName string) error {
    _ = schemaName
    // 将来的に archive 用 DDL をここで適用する。
    return nil
}
