package storage

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteConfig struct {
	Path string
}

func OpenSQLite(cfg SQLiteConfig) (*sql.DB, error) {
	return sql.Open("sqlite", cfg.Path)
}
