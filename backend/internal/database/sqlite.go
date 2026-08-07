package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"yt-off/backend/internal/models"
)

func OpenSQLite(path string) (*sql.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}

	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	statements := []string{
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS downloads (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			url TEXT NOT NULL,
			format_id TEXT NOT NULL,
			status TEXT NOT NULL,
			progress REAL DEFAULT 0,
			speed TEXT,
			eta TEXT,
			filename TEXT,
			file_size INTEGER DEFAULT 0,
			extension TEXT,
			container_id TEXT,
			error TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS download_groups (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS download_group_items (
			id TEXT PRIMARY KEY,
			group_id TEXT NOT NULL,
			download_id TEXT NOT NULL,
			position INTEGER DEFAULT 0,
			created_at DATETIME,
			FOREIGN KEY(group_id) REFERENCES download_groups(id) ON DELETE CASCADE,
			FOREIGN KEY(download_id) REFERENCES downloads(id) ON DELETE CASCADE,
			UNIQUE(group_id, download_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_download_groups_user_id ON download_groups(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_download_group_items_group_id ON download_group_items(group_id)`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("run sqlite migration: %w", err)
		}
	}

	if err := ensureColumn(db, "downloads", "file_size", "file_size INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn(db, "downloads", "extension", "extension TEXT"); err != nil {
		return err
	}
	if err := ensureColumn(db, "downloads", "container_id", "container_id TEXT"); err != nil {
		return err
	}
	if err := ensureColumn(db, "downloads", "user_id", "user_id TEXT"); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_downloads_user_id ON downloads(user_id)`); err != nil {
		return fmt.Errorf("create downloads user index: %w", err)
	}
	if err := ensureDefaultUser(db); err != nil {
		return err
	}
	if _, err := db.Exec(
		`UPDATE downloads
		SET user_id = ?
		WHERE user_id IS NULL OR trim(user_id) = ''`,
		models.DefaultUserID,
	); err != nil {
		return fmt.Errorf("assign default user to downloads: %w", err)
	}

	return nil
}

func ensureDefaultUser(db *sql.DB) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(
		`INSERT OR IGNORE INTO users (id, username, created_at, updated_at)
		VALUES (?, ?, ?, ?)`,
		models.DefaultUserID,
		models.DefaultUserID,
		now,
		now,
	); err != nil {
		return fmt.Errorf("ensure default user: %w", err)
	}

	return nil
}

func ensureColumn(db *sql.DB, tableName string, columnName string, definition string) error {
	exists, err := columnExists(db, tableName, columnName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, definition)); err != nil {
		return fmt.Errorf("add sqlite column %s.%s: %w", tableName, columnName, err)
	}

	return nil
}

func columnExists(db *sql.DB, tableName string, columnName string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, fmt.Errorf("read sqlite table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int

		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, fmt.Errorf("scan sqlite table info: %w", err)
		}
		if name == columnName {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate sqlite table info: %w", err)
	}

	return false, nil
}
