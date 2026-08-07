package services

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"yt-off/backend/internal/models"
)

var (
	ErrInvalidFileName = errors.New("invalid file name")
	ErrFileNotFound    = errors.New("file not found")
)

type FileService struct {
	downloadsDir    string
	showHiddenFiles bool
	mu              sync.RWMutex
}

func NewFileService(downloadsDir string) *FileService {
	return &FileService{downloadsDir: downloadsDir}
}

func (service *FileService) SetDownloadsDir(downloadsDir string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	service.downloadsDir = downloadsDir
}

func (service *FileService) SetShowHiddenFiles(showHiddenFiles bool) {
	service.mu.Lock()
	defer service.mu.Unlock()

	service.showHiddenFiles = showHiddenFiles
}

func (service *FileService) DownloadsDir() string {
	service.mu.RLock()
	defer service.mu.RUnlock()

	return service.downloadsDir
}

func (service *FileService) ListFiles() ([]models.File, error) {
	downloadsDir, showHiddenFiles := service.currentSettings()
	entries, err := os.ReadDir(downloadsDir)
	if err != nil {
		return nil, err
	}

	files := make([]models.File, 0)
	for _, entry := range entries {
		if entry.IsDir() || (!showHiddenFiles && strings.HasPrefix(entry.Name(), ".")) {
			continue
		}

		file, err := service.getFileFromDir(downloadsDir, entry.Name())
		if err != nil {
			continue
		}

		files = append(files, *file)
	}

	sort.Slice(files, func(i int, j int) bool {
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})

	return files, nil
}

func (service *FileService) GetFile(name string) (*models.File, error) {
	return service.getFileFromDir(service.DownloadsDir(), name)
}

func (service *FileService) getFileFromDir(downloadsDir string, name string) (*models.File, error) {
	path, err := service.safePath(downloadsDir, name)
	if err != nil {
		return nil, err
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotFound
		}

		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, ErrInvalidFileName
	}
	if info.IsDir() {
		return nil, ErrFileNotFound
	}

	return &models.File{
		Name:       name,
		Path:       path,
		Size:       info.Size(),
		Extension:  fileExtension(name),
		ModifiedAt: info.ModTime().UTC(),
	}, nil
}

func (service *FileService) DeleteFile(name string) error {
	path, err := service.safePath(service.DownloadsDir(), name)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrFileNotFound
		}

		return err
	}

	return nil
}

func (service *FileService) currentSettings() (string, bool) {
	service.mu.RLock()
	defer service.mu.RUnlock()

	return service.downloadsDir, service.showHiddenFiles
}

func (service *FileService) safePath(downloadsDir string, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.Contains(name, "\x00") || filepath.IsAbs(name) {
		return "", ErrInvalidFileName
	}
	if strings.ContainsAny(name, `/\`) || filepath.Clean(name) != name || filepath.Base(name) != name {
		return "", ErrInvalidFileName
	}

	baseDir, err := filepath.Abs(downloadsDir)
	if err != nil {
		return "", err
	}

	path := filepath.Join(baseDir, name)
	resolvedPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolvedPath != filepath.Join(baseDir, filepath.Base(resolvedPath)) {
		return "", ErrInvalidFileName
	}

	return resolvedPath, nil
}

func fileExtension(name string) string {
	extension := filepath.Ext(name)
	return strings.TrimPrefix(strings.ToLower(extension), ".")
}
