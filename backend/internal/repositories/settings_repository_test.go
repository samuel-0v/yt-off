package repositories

import (
	"errors"
	"path/filepath"
	"testing"

	"yt-off/backend/internal/database"
)

func TestSettingsRepositoryUpsertFindAndList(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	repository := NewSettingsRepository(db)

	if err := repository.Upsert("theme", "system"); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if err := repository.Upsert("theme", "dark"); err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}
	if err := repository.UpsertMany(map[string]string{
		"language": "pt-BR",
		"app_name": "yt-off",
	}); err != nil {
		t.Fatalf("UpsertMany() error = %v", err)
	}

	setting, err := repository.FindByKey("theme")
	if err != nil {
		t.Fatalf("FindByKey() error = %v", err)
	}
	if setting.Value != "dark" {
		t.Fatalf("Value = %q, want dark", setting.Value)
	}

	values, err := repository.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if values["theme"] != "dark" || values["language"] != "pt-BR" || values["app_name"] != "yt-off" {
		t.Fatalf("FindAll() = %#v", values)
	}
}

func TestSettingsRepositoryFindByKeyNotFound(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	repository := NewSettingsRepository(db)

	_, err = repository.FindByKey("missing")
	if !errors.Is(err, ErrSettingNotFound) {
		t.Fatalf("FindByKey() error = %v, want ErrSettingNotFound", err)
	}
}
