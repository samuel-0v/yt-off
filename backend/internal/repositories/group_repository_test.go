package repositories

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"yt-off/backend/internal/database"
	"yt-off/backend/internal/models"
)

func TestDownloadGroupRepositoryCreateFindAndItems(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	downloadRepository := NewDownloadRepository(db)
	groupRepository := NewDownloadGroupRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	download := &models.DownloadTask{
		ID:        "download-1",
		UserID:    models.DefaultUserID,
		URL:       "https://example.com/video",
		FormatID:  "140",
		Status:    models.DownloadStatusCompleted,
		Extension: "m4a",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := downloadRepository.Create(download); err != nil {
		t.Fatalf("Create download error = %v", err)
	}

	group := &models.DownloadGroup{
		ID:          "group-1",
		UserID:      models.DefaultUserID,
		Name:        "Favoritos",
		Description: "Audio",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := groupRepository.Create(group); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	item, err := groupRepository.AddItem(&models.DownloadGroupItem{
		ID:         "item-1",
		GroupID:    group.ID,
		DownloadID: download.ID,
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if item.Position != 0 {
		t.Fatalf("Position = %d, want 0", item.Position)
	}

	duplicate, err := groupRepository.AddItem(&models.DownloadGroupItem{
		ID:         "item-duplicate",
		GroupID:    group.ID,
		DownloadID: download.ID,
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatalf("AddItem() duplicate error = %v", err)
	}
	if duplicate.ID != item.ID {
		t.Fatalf("Duplicate item ID = %q, want %q", duplicate.ID, item.ID)
	}

	found, err := groupRepository.FindByID(group.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.ItemCount != 1 || len(found.Items) != 1 {
		t.Fatalf("Found group items = %#v, count %d", found.Items, found.ItemCount)
	}
	if found.OwnerUsername != models.DefaultUserID {
		t.Fatalf("OwnerUsername = %q, want %q", found.OwnerUsername, models.DefaultUserID)
	}

	groups, err := groupRepository.FindByUserID(models.DefaultUserID)
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if len(groups) != 1 || groups[0].ItemCount != 1 {
		t.Fatalf("FindByUserID() = %#v", groups)
	}
}

func TestDownloadGroupRepositoryRemoveMissingItem(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	repository := NewDownloadGroupRepository(db)

	err = repository.RemoveItem("group", "missing")
	if !errors.Is(err, ErrDownloadGroupItemNotFound) {
		t.Fatalf("RemoveItem() error = %v, want ErrDownloadGroupItemNotFound", err)
	}
}
