package services

import (
	"errors"
	"path/filepath"
	"testing"
)

const validCookies = `# Netscape HTTP Cookie File
.youtube.com	TRUE	/	TRUE	2145916800	SID	value
`

func TestCookiesServiceSaveStatusAndDelete(t *testing.T) {
	service, err := NewCookiesService(filepath.Join(t.TempDir(), "youtube.txt"))
	if err != nil {
		t.Fatalf("NewCookiesService() error = %v", err)
	}

	status, err := service.Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Exists {
		t.Fatal("Status().Exists = true, want false")
	}

	status, err = service.Save(validCookies)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !status.Exists || !status.Valid || status.Size == 0 || status.UpdatedAt == nil {
		t.Fatalf("Status after Save() = %#v", status)
	}

	if err := service.Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	status, err = service.Status()
	if err != nil {
		t.Fatalf("Status() after Delete() error = %v", err)
	}
	if status.Exists {
		t.Fatal("Status().Exists after Delete() = true, want false")
	}
}

func TestCookiesServiceRejectsInvalidContent(t *testing.T) {
	service, err := NewCookiesService(filepath.Join(t.TempDir(), "youtube.txt"))
	if err != nil {
		t.Fatalf("NewCookiesService() error = %v", err)
	}

	_, err = service.Save("not cookies")
	if !errors.Is(err, ErrInvalidCookiesFile) {
		t.Fatalf("Save() error = %v, want ErrInvalidCookiesFile", err)
	}
}
