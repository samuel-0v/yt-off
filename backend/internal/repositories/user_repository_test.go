package repositories

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"yt-off/backend/internal/database"
	"yt-off/backend/internal/models"
)

func TestUserRepositoryCreateFindAndList(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	user := &models.User{
		ID:        "user-1",
		Username:  "Samuel",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repository.Create(user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repository.FindByUsername("samuel")
	if err != nil {
		t.Fatalf("FindByUsername() error = %v", err)
	}
	if found.ID != user.ID || found.Username != user.Username {
		t.Fatalf("FindByUsername() = %#v, want %#v", found, user)
	}

	users, err := repository.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("FindAll() returned %d users, want default user plus created user", len(users))
	}
}

func TestUserRepositoryFindByIDNotFound(t *testing.T) {
	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db)

	_, err = repository.FindByID("missing")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("FindByID() error = %v, want ErrUserNotFound", err)
	}
}
