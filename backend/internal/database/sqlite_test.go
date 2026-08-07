package database

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"yt-off/backend/internal/models"
)

func TestOpenSQLiteMigratesLegacyDownloadsUserID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "yt-off.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(
		`CREATE TABLE downloads (
			id TEXT PRIMARY KEY,
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
	); err != nil {
		t.Fatalf("create legacy downloads table error = %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO downloads (id, url, format_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"download-1",
		"https://example.com/video",
		"140",
		models.DownloadStatusCompleted,
		now,
		now,
	); err != nil {
		t.Fatalf("insert legacy download error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	migrated, err := OpenSQLite(path)
	if err != nil {
		t.Fatalf("OpenSQLite() migration error = %v", err)
	}
	defer migrated.Close()

	var userID string
	if err := migrated.QueryRow(`SELECT user_id FROM downloads WHERE id = ?`, "download-1").Scan(&userID); err != nil {
		t.Fatalf("read migrated download user_id error = %v", err)
	}
	if userID != models.DefaultUserID {
		t.Fatalf("user_id = %q, want %q", userID, models.DefaultUserID)
	}

	var username string
	if err := migrated.QueryRow(`SELECT username FROM users WHERE id = ?`, models.DefaultUserID).Scan(&username); err != nil {
		t.Fatalf("read default user error = %v", err)
	}
	if username != models.DefaultUserID {
		t.Fatalf("username = %q, want %q", username, models.DefaultUserID)
	}
}
