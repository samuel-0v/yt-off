package repositories

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"yt-off/backend/internal/database"
	"yt-off/backend/internal/models"
)

func TestDownloadRepositoryCreateFindUpdateAndList(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	repository := NewDownloadRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	task := &models.DownloadTask{
		ID:        "download-1",
		URL:       "https://example.com/video",
		FormatID:  "137+140",
		Status:    models.DownloadStatusQueued,
		Progress:  0,
		FileSize:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repository.Create(task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repository.FindByID(task.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.ID != task.ID || found.URL != task.URL || found.FormatID != task.FormatID {
		t.Fatalf("FindByID() = %#v, want task %#v", found, task)
	}

	found.Status = models.DownloadStatusRunning
	found.Progress = 55.5
	found.Speed = "5MiB/s"
	found.ETA = "00:10"
	found.FileName = "video.mp4"
	found.FileSize = 1048576
	found.Extension = "mp4"
	found.ContainerID = "container-1"
	found.UpdatedAt = now.Add(time.Minute)
	if err := repository.Update(found); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repository.FindByID(task.ID)
	if err != nil {
		t.Fatalf("FindByID() after update error = %v", err)
	}
	if updated.Status != models.DownloadStatusRunning {
		t.Fatalf("Status = %q, want %q", updated.Status, models.DownloadStatusRunning)
	}
	if updated.Progress != 55.5 {
		t.Fatalf("Progress = %v, want 55.5", updated.Progress)
	}
	if updated.Speed != "5MiB/s" || updated.ETA != "00:10" || updated.FileName != "video.mp4" {
		t.Fatalf("unexpected progress fields: %#v", updated)
	}
	if updated.FileSize != 1048576 || updated.Extension != "mp4" {
		t.Fatalf("unexpected file metadata: %#v", updated)
	}
	if updated.ContainerID != "container-1" {
		t.Fatalf("ContainerID = %q, want container-1", updated.ContainerID)
	}

	if err := repository.MarkFileRemoved("video.mp4"); err != nil {
		t.Fatalf("MarkFileRemoved() error = %v", err)
	}
	removed, err := repository.FindByID(task.ID)
	if err != nil {
		t.Fatalf("FindByID() after MarkFileRemoved error = %v", err)
	}
	if removed.FileName != "video.mp4" {
		t.Fatalf("FileName = %q, want video.mp4", removed.FileName)
	}
	if removed.FileSize != 0 {
		t.Fatalf("FileSize = %d, want 0", removed.FileSize)
	}

	downloads, err := repository.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(downloads) != 1 {
		t.Fatalf("FindAll() returned %d downloads, want 1", len(downloads))
	}
	if downloads[0].ID != task.ID {
		t.Fatalf("FindAll()[0].ID = %q, want %q", downloads[0].ID, task.ID)
	}
}

func TestDownloadRepositoryFindByIDNotFound(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	repository := NewDownloadRepository(db)

	_, err = repository.FindByID("missing")
	if !errors.Is(err, ErrDownloadNotFound) {
		t.Fatalf("FindByID() error = %v, want ErrDownloadNotFound", err)
	}
}
