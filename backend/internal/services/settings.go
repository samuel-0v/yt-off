package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"yt-off/backend/internal/models"
)

const (
	SettingDownloadDirectory      = "download_directory"
	SettingMaxConcurrentDownloads = "max_concurrent_downloads"
	SettingLanguage               = "language"
	SettingTheme                  = "theme"
	SettingAppName                = "app_name"
	SettingBackendPort            = "backend_port"
	SettingAutomaticUpdates       = "automatic_updates"
	SettingShowHiddenFiles        = "show_hidden_files"
)

var (
	ErrInvalidSettings          = errors.New("invalid settings")
	ErrInvalidDownloadDirectory = errors.New("invalid download directory")
)

type SettingsDefaults struct {
	DownloadDirectory string
	AppName           string
	BackendPort       string
}

type settingsRepository interface {
	FindAll() (map[string]string, error)
	UpsertMany(values map[string]string) error
}

type SettingsService struct {
	repository           settingsRepository
	defaults             models.AppSettings
	allowedDownloadsRoot string

	mu        sync.RWMutex
	settings  models.AppSettings
	listeners []func(models.AppSettings)
}

func NewSettingsService(repository settingsRepository, defaults SettingsDefaults) (*SettingsService, error) {
	downloadDirectory, err := normalizeDirectory(defaults.DownloadDirectory)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(downloadDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("create default download directory: %w", err)
	}

	appName := strings.TrimSpace(defaults.AppName)
	if appName == "" {
		appName = AppName
	}

	backendPort := strings.TrimSpace(defaults.BackendPort)
	if backendPort == "" {
		backendPort = "18080"
	}

	service := &SettingsService{
		repository: repository,
		defaults: models.AppSettings{
			DownloadDirectory:      downloadDirectory,
			MaxConcurrentDownloads: MaxConcurrentDownloads,
			Language:               "pt-BR",
			Theme:                  "system",
			AppName:                appName,
			BackendPort:            backendPort,
			AutomaticUpdates:       false,
			ShowHiddenFiles:        false,
		},
		allowedDownloadsRoot: downloadDirectory,
		listeners:            make([]func(models.AppSettings), 0),
	}

	if err := service.ensureDefaults(); err != nil {
		return nil, err
	}
	if err := service.reload(); err != nil {
		return nil, err
	}

	return service, nil
}

func (service *SettingsService) CurrentSettings() models.AppSettings {
	service.mu.RLock()
	defer service.mu.RUnlock()

	return service.settings
}

func (service *SettingsService) DefaultSettings() models.AppSettings {
	return service.defaults
}

func (service *SettingsService) UpdateSettings(settings models.AppSettings) (*models.AppSettings, error) {
	normalized, err := service.validateSettings(settings)
	if err != nil {
		return nil, err
	}

	if err := service.repository.UpsertMany(settingsToMap(normalized)); err != nil {
		return nil, err
	}

	service.mu.Lock()
	service.settings = normalized
	listeners := append([]func(models.AppSettings){}, service.listeners...)
	service.mu.Unlock()

	for _, listener := range listeners {
		listener(normalized)
	}

	return &normalized, nil
}

func (service *SettingsService) OnChange(listener func(models.AppSettings)) {
	service.mu.Lock()
	defer service.mu.Unlock()

	service.listeners = append(service.listeners, listener)
}

func (service *SettingsService) ensureDefaults() error {
	values, err := service.repository.FindAll()
	if err != nil {
		return err
	}

	defaultValues := settingsToMap(service.defaults)
	missing := make(map[string]string)
	for key, value := range defaultValues {
		if _, exists := values[key]; !exists {
			missing[key] = value
		}
	}
	if len(missing) == 0 {
		return nil
	}

	return service.repository.UpsertMany(missing)
}

func (service *SettingsService) reload() error {
	values, err := service.repository.FindAll()
	if err != nil {
		return err
	}

	settings := settingsFromMap(values, service.defaults)
	normalized, err := service.validateSettings(settings)
	if err != nil {
		normalized = service.defaults
		if updateErr := service.repository.UpsertMany(settingsToMap(normalized)); updateErr != nil {
			return updateErr
		}
	}

	service.mu.Lock()
	service.settings = normalized
	service.mu.Unlock()

	return nil
}

func (service *SettingsService) validateSettings(settings models.AppSettings) (models.AppSettings, error) {
	downloadDirectory, err := service.validateDownloadDirectory(settings.DownloadDirectory)
	if err != nil {
		return models.AppSettings{}, err
	}
	settings.DownloadDirectory = downloadDirectory

	if settings.MaxConcurrentDownloads < 1 || settings.MaxConcurrentDownloads > 10 {
		return models.AppSettings{}, fmt.Errorf("%w: max_concurrent_downloads must be between 1 and 10", ErrInvalidSettings)
	}
	if !validTheme(settings.Theme) {
		return models.AppSettings{}, fmt.Errorf("%w: invalid theme", ErrInvalidSettings)
	}
	if !validLanguage(settings.Language) {
		return models.AppSettings{}, fmt.Errorf("%w: invalid language", ErrInvalidSettings)
	}

	settings.AppName = strings.TrimSpace(settings.AppName)
	if settings.AppName == "" {
		return models.AppSettings{}, fmt.Errorf("%w: app_name is required", ErrInvalidSettings)
	}

	settings.BackendPort = strings.TrimSpace(settings.BackendPort)
	port, err := strconv.Atoi(settings.BackendPort)
	if err != nil || port < 1 || port > 65535 {
		return models.AppSettings{}, fmt.Errorf("%w: invalid backend_port", ErrInvalidSettings)
	}

	return settings, nil
}

func (service *SettingsService) validateDownloadDirectory(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "\x00") || !filepath.IsAbs(value) {
		return "", ErrInvalidDownloadDirectory
	}
	path := filepath.Clean(value)

	root, err := filepath.EvalSymlinks(service.allowedDownloadsRoot)
	if err != nil {
		return "", fmt.Errorf("%w: shared download root is unavailable", ErrInvalidDownloadDirectory)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("%w: directory does not exist", ErrInvalidDownloadDirectory)
	}
	if !isPathInside(root, resolvedPath) {
		return "", fmt.Errorf("%w: directory must be inside %s", ErrInvalidDownloadDirectory, service.allowedDownloadsRoot)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("%w: directory does not exist", ErrInvalidDownloadDirectory)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%w: path is not a directory", ErrInvalidDownloadDirectory)
	}
	if err := ensureDirectoryWritable(resolvedPath); err != nil {
		return "", fmt.Errorf("%w: directory is not writable", ErrInvalidDownloadDirectory)
	}

	return resolvedPath, nil
}

func normalizeDirectory(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "\x00") {
		return "", ErrInvalidDownloadDirectory
	}

	cleaned := filepath.Clean(value)
	if !filepath.IsAbs(cleaned) {
		absolute, err := filepath.Abs(cleaned)
		if err != nil {
			return "", err
		}
		cleaned = absolute
	}

	return cleaned, nil
}

func ensureDirectoryWritable(path string) error {
	file, err := os.CreateTemp(path, ".yt-off-write-test-*")
	if err != nil {
		return err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		os.Remove(name)
		return err
	}

	return os.Remove(name)
}

func isPathInside(root string, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)))
}

func validTheme(value string) bool {
	switch value {
	case "system", "light", "dark":
		return true
	default:
		return false
	}
}

func validLanguage(value string) bool {
	switch value {
	case "pt-BR", "en-US", "es":
		return true
	default:
		return false
	}
}

func settingsToMap(settings models.AppSettings) map[string]string {
	return map[string]string{
		SettingDownloadDirectory:      settings.DownloadDirectory,
		SettingMaxConcurrentDownloads: strconv.Itoa(settings.MaxConcurrentDownloads),
		SettingLanguage:               settings.Language,
		SettingTheme:                  settings.Theme,
		SettingAppName:                settings.AppName,
		SettingBackendPort:            settings.BackendPort,
		SettingAutomaticUpdates:       strconv.FormatBool(settings.AutomaticUpdates),
		SettingShowHiddenFiles:        strconv.FormatBool(settings.ShowHiddenFiles),
	}
}

func settingsFromMap(values map[string]string, defaults models.AppSettings) models.AppSettings {
	settings := defaults

	if value := strings.TrimSpace(values[SettingDownloadDirectory]); value != "" {
		settings.DownloadDirectory = value
	}
	if value, err := strconv.Atoi(values[SettingMaxConcurrentDownloads]); err == nil {
		settings.MaxConcurrentDownloads = value
	}
	if value := strings.TrimSpace(values[SettingLanguage]); value != "" {
		settings.Language = value
	}
	if value := strings.TrimSpace(values[SettingTheme]); value != "" {
		settings.Theme = value
	}
	if value := strings.TrimSpace(values[SettingAppName]); value != "" {
		settings.AppName = value
	}
	if value := strings.TrimSpace(values[SettingBackendPort]); value != "" {
		settings.BackendPort = value
	}
	if value, err := strconv.ParseBool(values[SettingAutomaticUpdates]); err == nil {
		settings.AutomaticUpdates = value
	}
	if value, err := strconv.ParseBool(values[SettingShowHiddenFiles]); err == nil {
		settings.ShowHiddenFiles = value
	}

	return settings
}
