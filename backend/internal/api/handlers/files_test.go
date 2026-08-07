package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/services"
)

func TestDownloadFileHandlerDecodesFileName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test space.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	app := fiber.New()
	app.Get("/api/files/:name", DownloadFileHandler(services.NewFileService(dir)))

	request := httptest.NewRequest(http.MethodGet, "/api/files/test%20space.txt", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := response.Header.Get("Content-Disposition"); got != `attachment; filename="test space.txt"` {
		t.Fatalf("Content-Disposition = %q, want attachment filename", got)
	}
}

func TestDownloadFileHandlerRejectsEncodedPathTraversal(t *testing.T) {
	app := fiber.New()
	app.Get("/api/files/:name", DownloadFileHandler(services.NewFileService(t.TempDir())))

	request := httptest.NewRequest(http.MethodGet, "/api/files/..%2Fsecret.txt", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
}

func TestDownloadFileHandlerSupportsInlinePlayback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "video.mp4"), []byte("video"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	app := fiber.New()
	app.Get("/api/files/:name", DownloadFileHandler(services.NewFileService(dir)))

	request := httptest.NewRequest(http.MethodGet, "/api/files/video.mp4?inline=1", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := response.Header.Get("Content-Disposition"); got != `inline; filename=video.mp4` {
		t.Fatalf("Content-Disposition = %q, want inline filename", got)
	}
	if got := response.Header.Get("Content-Type"); got != "video/mp4" {
		t.Fatalf("Content-Type = %q, want video/mp4", got)
	}
}
