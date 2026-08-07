package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"yt-off/backend/internal/database"
	"yt-off/backend/internal/models"
	"yt-off/backend/internal/repositories"
)

func TestSettingsServiceLoadsDefaultsAndPersistsUpdates(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	downloadsDir := t.TempDir()
	service, err := NewSettingsService(repositories.NewSettingsRepository(db), SettingsDefaults{
		DownloadDirectory: downloadsDir,
		AppName:           "yt-off",
		BackendPort:       "18080",
	})
	if err != nil {
		t.Fatalf("NewSettingsService() error = %v", err)
	}

	settings := service.CurrentSettings()
	if settings.DownloadDirectory != downloadsDir {
		t.Fatalf("DownloadDirectory = %q, want %q", settings.DownloadDirectory, downloadsDir)
	}
	if settings.MaxConcurrentDownloads != MaxConcurrentDownloads {
		t.Fatalf("MaxConcurrentDownloads = %d, want %d", settings.MaxConcurrentDownloads, MaxConcurrentDownloads)
	}

	nestedDir := filepath.Join(downloadsDir, "music")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	called := false
	service.OnChange(func(updated models.AppSettings) {
		called = updated.MaxConcurrentDownloads == 3
	})

	settings.DownloadDirectory = nestedDir
	settings.MaxConcurrentDownloads = 3
	settings.Theme = "dark"
	updated, err := service.UpdateSettings(settings)
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if updated.DownloadDirectory != nestedDir || updated.MaxConcurrentDownloads != 3 || updated.Theme != "dark" {
		t.Fatalf("updated settings = %#v", updated)
	}
	if !called {
		t.Fatal("settings change listener was not called")
	}

	reloaded, err := NewSettingsService(repositories.NewSettingsRepository(db), SettingsDefaults{
		DownloadDirectory: downloadsDir,
		AppName:           "yt-off",
		BackendPort:       "18080",
	})
	if err != nil {
		t.Fatalf("NewSettingsService() reload error = %v", err)
	}
	if reloaded.CurrentSettings().DownloadDirectory != nestedDir {
		t.Fatalf("reloaded DownloadDirectory = %q, want %q", reloaded.CurrentSettings().DownloadDirectory, nestedDir)
	}
}

func TestSettingsServiceRejectsInvalidDownloadDirectory(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	downloadsDir := t.TempDir()
	service, err := NewSettingsService(repositories.NewSettingsRepository(db), SettingsDefaults{
		DownloadDirectory: downloadsDir,
		AppName:           "yt-off",
		BackendPort:       "18080",
	})
	if err != nil {
		t.Fatalf("NewSettingsService() error = %v", err)
	}

	settings := service.CurrentSettings()
	settings.DownloadDirectory = filepath.Join(t.TempDir(), "outside")
	if err := os.Mkdir(settings.DownloadDirectory, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	_, err = service.UpdateSettings(settings)
	if !errors.Is(err, ErrInvalidDownloadDirectory) {
		t.Fatalf("UpdateSettings() error = %v, want ErrInvalidDownloadDirectory", err)
	}

	settings = service.CurrentSettings()
	settings.DownloadDirectory = "downloads"
	_, err = service.UpdateSettings(settings)
	if !errors.Is(err, ErrInvalidDownloadDirectory) {
		t.Fatalf("UpdateSettings() relative path error = %v, want ErrInvalidDownloadDirectory", err)
	}
}
