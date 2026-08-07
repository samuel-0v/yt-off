package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileServiceListFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "video.mp4"), []byte("video"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	files, err := NewFileService(dir).ListFiles()
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("ListFiles() returned %d files, want 1", len(files))
	}
	if files[0].Name != "video.mp4" {
		t.Fatalf("Name = %q, want video.mp4", files[0].Name)
	}
	if files[0].Size != 5 {
		t.Fatalf("Size = %d, want 5", files[0].Size)
	}
	if files[0].Extension != "mp4" {
		t.Fatalf("Extension = %q, want mp4", files[0].Extension)
	}

	service := NewFileService(dir)
	service.SetShowHiddenFiles(true)
	files, err = service.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles() with hidden error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("ListFiles() with hidden returned %d files, want 2", len(files))
	}
}

func TestFileServiceBlocksInvalidPaths(t *testing.T) {
	service := NewFileService(t.TempDir())

	invalidNames := []string{
		"../video.mp4",
		"nested/video.mp4",
		"/tmp/video.mp4",
		`nested\video.mp4`,
		"",
	}

	for _, name := range invalidNames {
		if _, err := service.GetFile(name); !errors.Is(err, ErrInvalidFileName) {
			t.Fatalf("GetFile(%q) error = %v, want ErrInvalidFileName", name, err)
		}
	}
}

func TestFileServiceDeleteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audio.m4a")
	if err := os.WriteFile(path, []byte("audio"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := NewFileService(dir).DeleteFile("audio.m4a"); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat() error = %v, want os.ErrNotExist", err)
	}
}
