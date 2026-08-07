package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"yt-off/backend/internal/models"
)

var (
	ErrCookiesPathRequired = errors.New("cookies path is required")
	ErrInvalidCookiesFile  = errors.New("invalid cookies file")
	ErrCookiesNotFound     = errors.New("cookies file not found")
)

type CookiesService struct {
	path string
}

func NewCookiesService(path string) (*CookiesService, error) {
	path = strings.TrimSpace(path)
	if path == "" || strings.Contains(path, "\x00") {
		return nil, ErrCookiesPathRequired
	}
	if !filepath.IsAbs(path) {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		path = absolute
	}

	return &CookiesService{path: filepath.Clean(path)}, nil
}

func (service *CookiesService) Status() (models.CookiesInfo, error) {
	info := models.CookiesInfo{
		FileName: filepath.Base(service.path),
	}

	stat, err := os.Stat(service.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			info.Message = "cookies file not found"
			return info, nil
		}

		return models.CookiesInfo{}, err
	}
	if stat.IsDir() {
		info.Message = "cookies path is a directory"
		return info, nil
	}

	content, err := os.ReadFile(service.path)
	if err != nil {
		return models.CookiesInfo{}, err
	}

	updatedAt := stat.ModTime().UTC()
	info.Exists = true
	info.Size = stat.Size()
	info.UpdatedAt = &updatedAt
	info.Valid = validateCookiesContent(string(content)) == nil
	if !info.Valid {
		info.Message = "cookies file must be in Netscape format"
	}

	return info, nil
}

func (service *CookiesService) Save(content string) (models.CookiesInfo, error) {
	content = normalizeCookiesContent(content)
	if err := validateCookiesContent(content); err != nil {
		return models.CookiesInfo{}, err
	}

	if err := os.MkdirAll(filepath.Dir(service.path), 0o700); err != nil {
		return models.CookiesInfo{}, fmt.Errorf("create cookies directory: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(service.path), ".youtube-cookies-*")
	if err != nil {
		return models.CookiesInfo{}, fmt.Errorf("create temporary cookies file: %w", err)
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)

	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		return models.CookiesInfo{}, fmt.Errorf("write temporary cookies file: %w", err)
	}
	if err := tempFile.Chmod(0o600); err != nil {
		tempFile.Close()
		return models.CookiesInfo{}, fmt.Errorf("chmod temporary cookies file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return models.CookiesInfo{}, fmt.Errorf("close temporary cookies file: %w", err)
	}
	if err := os.Rename(tempName, service.path); err != nil {
		return models.CookiesInfo{}, fmt.Errorf("replace cookies file: %w", err)
	}

	return service.Status()
}

func (service *CookiesService) Delete() error {
	if err := os.Remove(service.path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrCookiesNotFound
		}

		return err
	}

	return nil
}

func normalizeCookiesContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	content = strings.TrimPrefix(content, "\ufeff")
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	return content + "\n"
}

func validateCookiesContent(content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return ErrInvalidCookiesFile
	}

	lines := strings.Split(content, "\n")
	firstLine := strings.TrimSpace(strings.TrimPrefix(lines[0], "\ufeff"))
	if firstLine != "# Netscape HTTP Cookie File" && firstLine != "# HTTP Cookie File" {
		return ErrInvalidCookiesFile
	}

	hasCookieRow := false
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		columns := strings.Split(line, "\t")
		if len(columns) >= 7 {
			hasCookieRow = true
			break
		}
	}
	if !hasCookieRow {
		return ErrInvalidCookiesFile
	}

	return nil
}
