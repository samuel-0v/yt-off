package services

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"yt-off/backend/internal/database"
	"yt-off/backend/internal/models"
	"yt-off/backend/internal/repositories"
)

func TestDownloadGroupServiceManagesOwnGroups(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	users := NewUserService(repositories.NewUserRepository(db))
	downloads := repositories.NewDownloadRepository(db)
	groups := NewDownloadGroupService(repositories.NewDownloadGroupRepository(db), downloads, users)

	alice, err := users.GetOrCreateUser("Alice")
	if err != nil {
		t.Fatalf("GetOrCreateUser() Alice error = %v", err)
	}
	bob, err := users.GetOrCreateUser("Bob")
	if err != nil {
		t.Fatalf("GetOrCreateUser() Bob error = %v", err)
	}

	now := time.Now().UTC()
	download := &models.DownloadTask{
		ID:        "download-1",
		UserID:    alice.ID,
		URL:       "https://example.com/video",
		FormatID:  "140",
		Status:    models.DownloadStatusCompleted,
		Extension: "m4a",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := downloads.Create(download); err != nil {
		t.Fatalf("Create download error = %v", err)
	}

	group, err := groups.CreateGroup(alice.ID, " Favoritos  de audio ", "Minhas musicas")
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if group.Name != "Favoritos de audio" || group.UserID != alice.ID {
		t.Fatalf("CreateGroup() = %#v", group)
	}

	item, err := groups.AddItem(group.ID, alice.ID, download.ID)
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if item.Download == nil || item.Download.ID != download.ID {
		t.Fatalf("AddItem() download = %#v", item.Download)
	}

	found, err := groups.GetGroup(group.ID)
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	if found.ItemCount != 1 || len(found.Items) != 1 || found.Items[0].Download == nil {
		t.Fatalf("GetGroup() = %#v", found)
	}

	_, err = groups.UpdateGroup(group.ID, bob.ID, "Bob edit", "")
	if !errors.Is(err, ErrDownloadGroupForbidden) {
		t.Fatalf("UpdateGroup() as Bob error = %v, want ErrDownloadGroupForbidden", err)
	}

	aliceGroups, err := groups.ListGroups("mine", alice.ID)
	if err != nil {
		t.Fatalf("ListGroups() Alice error = %v", err)
	}
	if len(aliceGroups) != 1 || aliceGroups[0].ItemCount != 1 {
		t.Fatalf("ListGroups() Alice = %#v", aliceGroups)
	}

	allGroups, err := groups.ListGroups("", bob.ID)
	if err != nil {
		t.Fatalf("ListGroups() all error = %v", err)
	}
	if len(allGroups) != 1 {
		t.Fatalf("ListGroups() all returned %d items, want 1", len(allGroups))
	}
}
