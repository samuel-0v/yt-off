package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	dockerclient "yt-off/backend/internal/docker"
	"yt-off/backend/internal/models"
	"yt-off/backend/internal/repositories"
)

var (
	ErrDownloadURLRequired      = errors.New("url is required")
	ErrDownloadFormatIDRequired = errors.New("format_id is required")
	ErrDownloadNotFound         = errors.New("download not found")
)

type DownloadService struct {
	containerName string
	files         *FileService
	repository    downloadRepository
	runner        downloadRunner
	settings      downloadSettingsProvider
	users         groupUserResolver
	ytDLPOptions  YTDLPOptions
	queue         *DownloadQueue
	active        map[string]activeDownload
	activeMu      sync.Mutex
	mu            sync.Mutex
}

type activeDownload struct {
	containerID string
	cancel      context.CancelFunc
}

type downloadRepository interface {
	Create(task *models.DownloadTask) error
	FindByID(id string) (*models.DownloadTask, error)
	FindAll() ([]models.DownloadTask, error)
	FindByUserID(userID string) ([]models.DownloadTask, error)
	Update(task *models.DownloadTask) error
}

type downloadRunner interface {
	Run(ctx context.Context, downloadID string, command []string, stdoutWriter io.Writer, stderrWriter io.Writer, onContainerCreated func(string)) (dockerclient.ContainerRunResult, error)
	Stop(ctx context.Context, containerID string) error
}

type downloadSettingsProvider interface {
	CurrentSettings() models.AppSettings
}

type staticDownloadSettingsProvider struct {
	settings models.AppSettings
}

func (provider staticDownloadSettingsProvider) CurrentSettings() models.AppSettings {
	return provider.settings
}

type defaultDownloadUserResolver struct{}

func (resolver defaultDownloadUserResolver) ResolveUserID(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return models.DefaultUserID, nil
	}

	return userID, nil
}

func NewDownloadService(containerName string, downloadsDir string, repository downloadRepository) *DownloadService {
	settingsProvider := staticDownloadSettingsProvider{
		settings: models.AppSettings{
			DownloadDirectory:      downloadsDir,
			MaxConcurrentDownloads: MaxConcurrentDownloads,
			ShowHiddenFiles:        false,
		},
	}

	return newDownloadService(containerName, downloadsDir, repository, settingsProvider, defaultDownloadUserResolver{}, newDockerDownloadRunner(containerName), MaxConcurrentDownloads, YTDLPOptions{})
}

func NewDownloadServiceWithSettings(containerName string, repository downloadRepository, settingsProvider downloadSettingsProvider, options YTDLPOptions) *DownloadService {
	settings := settingsProvider.CurrentSettings()

	return newDownloadService(containerName, settings.DownloadDirectory, repository, settingsProvider, defaultDownloadUserResolver{}, newDockerDownloadRunner(containerName), settings.MaxConcurrentDownloads, options)
}

func NewDownloadServiceWithSettingsAndUsers(containerName string, repository downloadRepository, settingsProvider downloadSettingsProvider, users groupUserResolver, options YTDLPOptions) *DownloadService {
	settings := settingsProvider.CurrentSettings()

	return newDownloadService(containerName, settings.DownloadDirectory, repository, settingsProvider, users, newDockerDownloadRunner(containerName), settings.MaxConcurrentDownloads, options)
}

func newDownloadService(containerName string, downloadsDir string, repository downloadRepository, settingsProvider downloadSettingsProvider, users groupUserResolver, runner downloadRunner, maxConcurrentDownloads int, options YTDLPOptions) *DownloadService {
	if users == nil {
		users = defaultDownloadUserResolver{}
	}

	service := &DownloadService{
		containerName: containerName,
		files:         NewFileService(downloadsDir),
		repository:    repository,
		runner:        runner,
		settings:      settingsProvider,
		users:         users,
		ytDLPOptions:  options,
		active:        make(map[string]activeDownload),
	}
	service.queue = NewDownloadQueue(maxConcurrentDownloads, service.runDownload)

	return service
}

func (service *DownloadService) ApplySettings(settings models.AppSettings) {
	service.files.SetDownloadsDir(settings.DownloadDirectory)
	service.files.SetShowHiddenFiles(settings.ShowHiddenFiles)
	service.queue.SetMaxConcurrent(settings.MaxConcurrentDownloads)
}

func (service *DownloadService) CreateDownload(url string, formatID string, extension string, userID string) (*models.DownloadTask, error) {
	url = strings.TrimSpace(url)
	formatID = strings.TrimSpace(formatID)
	extension = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(extension)), ".")
	if url == "" {
		return nil, ErrDownloadURLRequired
	}
	if formatID == "" {
		return nil, ErrDownloadFormatIDRequired
	}
	resolvedUserID, err := service.users.ResolveUserID(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	task := &models.DownloadTask{
		ID:        uuid.NewString(),
		UserID:    resolvedUserID,
		URL:       url,
		FormatID:  formatID,
		Status:    models.DownloadStatusQueued,
		Progress:  0,
		Extension: extension,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := service.repository.Create(task); err != nil {
		return nil, err
	}

	createdTask := cloneDownloadTask(task)
	service.queue.Enqueue(task.ID)

	return createdTask, nil
}

func (service *DownloadService) ListDownloads(scope string, userID string) ([]models.DownloadTask, error) {
	if isMineScope(scope) {
		resolvedUserID, err := service.users.ResolveUserID(userID)
		if err != nil {
			return nil, err
		}

		return service.repository.FindByUserID(resolvedUserID)
	}

	return service.repository.FindAll()
}

func (service *DownloadService) GetDownload(id string) (*models.DownloadTask, error) {
	task, err := service.repository.FindByID(id)
	if errors.Is(err, repositories.ErrDownloadNotFound) {
		return nil, ErrDownloadNotFound
	}
	if err != nil {
		return nil, err
	}

	return cloneDownloadTask(task), nil
}

func (service *DownloadService) CopyDownload(id string, userID string) (*models.DownloadTask, error) {
	source, err := service.GetDownload(id)
	if err != nil {
		return nil, err
	}

	return service.CreateDownload(source.URL, source.FormatID, source.Extension, userID)
}

func (service *DownloadService) CancelDownload(id string) (*models.DownloadTask, error) {
	task, err := service.GetDownload(id)
	if err != nil {
		return nil, err
	}

	if isTerminalDownloadStatus(task.Status) {
		return task, nil
	}

	if task.Status == models.DownloadStatusQueued {
		service.queue.Cancel(id)
		service.updateDownload(id, func(task *models.DownloadTask) {
			task.Status = models.DownloadStatusCancelled
			task.Speed = ""
			task.ETA = ""
			task.Error = ""
		})

		return service.GetDownload(id)
	}

	if task.Status == models.DownloadStatusRunning {
		active := service.activeDownload(id)
		service.updateDownload(id, func(task *models.DownloadTask) {
			task.Status = models.DownloadStatusCancelled
			task.Speed = ""
			task.ETA = ""
			task.Error = ""
		})
		if active.cancel != nil {
			active.cancel()
		}
		if active.containerID != "" {
			go service.runner.Stop(context.Background(), active.containerID)
		}

		return service.GetDownload(id)
	}

	return task, nil
}

func (service *DownloadService) runDownload(id string) {
	defer service.queue.Complete(id)

	task, err := service.GetDownload(id)
	if err != nil {
		return
	}
	if task.Status == models.DownloadStatusCancelled {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	service.registerActiveDownload(id, activeDownload{cancel: cancel})
	defer service.unregisterActiveDownload(id)
	defer cancel()

	if service.downloadCancelled(id) {
		return
	}

	settings := service.settings.CurrentSettings()
	downloadsDir := settings.DownloadDirectory
	fileService := NewFileService(downloadsDir)
	fileService.SetShowHiddenFiles(settings.ShowHiddenFiles)

	service.updateDownload(id, func(task *models.DownloadTask) {
		task.Status = models.DownloadStatusRunning
		task.Progress = 0
		task.Error = ""
	})

	stdoutWriter := newDownloadLineWriter(func(line string) {
		service.handleDownloadOutputLine(id, line)
	})
	stderrWriter := newDownloadLineWriter(func(line string) {
		service.handleDownloadOutputLine(id, line)
	})

	result, err := service.runner.Run(ctx, task.ID, downloadCommand(task, downloadsDir, service.ytDLPOptions), stdoutWriter, stderrWriter, func(containerID string) {
		service.setActiveContainerID(id, containerID)
	})
	stdoutWriter.Flush()
	stderrWriter.Flush()
	if service.downloadCancelled(id) {
		return
	}
	if err != nil || result.ExitCode != 0 {
		service.failDownload(id, downloadFailureMessage(result, err))
		return
	}

	output := result.Stdout + "\n" + result.Stderr
	fileName := extractDownloadFileName(output)
	service.updateDownload(id, func(task *models.DownloadTask) {
		task.Status = models.DownloadStatusCompleted
		task.Progress = 100
		task.Speed = ""
		task.ETA = ""
		if fileName != "" {
			task.FileName = fileName
		}
		if task.FileName != "" {
			task.Extension = fileExtension(task.FileName)
			if file, err := fileService.GetFile(task.FileName); err == nil {
				task.FileSize = file.Size
				task.Extension = file.Extension
			}
		}
		task.Error = ""
	})
}

func (service *DownloadService) handleDownloadOutputLine(id string, line string) {
	progress, ok := ParseDownloadProgress(line)
	fileName := extractDownloadFileName(line)
	if !ok && fileName == "" {
		return
	}

	service.updateDownload(id, func(task *models.DownloadTask) {
		if task.Status != models.DownloadStatusRunning {
			return
		}
		if ok {
			task.Progress = progress.Progress
			if progress.Speed != "" {
				task.Speed = progress.Speed
			}
			if progress.ETA != "" {
				task.ETA = progress.ETA
			}
		}
		if fileName != "" {
			task.FileName = fileName
			task.Extension = fileExtension(fileName)
		}
	})
}

func (service *DownloadService) failDownload(id string, message string) {
	service.updateDownload(id, func(task *models.DownloadTask) {
		if task.Status == models.DownloadStatusCancelled {
			return
		}
		task.Status = models.DownloadStatusFailed
		task.Error = message
	})
}

const maxDownloadFailureMessageLength = 1000

func downloadFailureMessage(result dockerclient.ContainerRunResult, runErr error) string {
	if message := compactDownloadFailureOutput(result.Stderr + "\n" + result.Stdout); message != "" {
		return message
	}
	if runErr != nil {
		return limitDownloadFailureMessage(runErr.Error())
	}
	if result.ExitCode != 0 {
		return fmt.Sprintf("yt-dlp exited with code %d", result.ExitCode)
	}

	return "download failed"
}

func compactDownloadFailureOutput(output string) string {
	output = strings.ReplaceAll(output, "\r", "\n")
	lines := strings.Split(output, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isNoisyDownloadProgressLine(line) {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return ""
	}

	for index := len(kept) - 1; index >= 0; index-- {
		if isYTDLPErrorLine(kept[index]) {
			kept = kept[index:]
			break
		}
	}
	if len(kept) > 6 {
		kept = kept[len(kept)-6:]
	}

	return limitDownloadFailureMessage(strings.Join(kept, "\n"))
}

func isNoisyDownloadProgressLine(line string) bool {
	_, ok := ParseDownloadProgress(line)
	return ok
}

func isYTDLPErrorLine(line string) bool {
	normalized := strings.ToLower(strings.TrimSpace(line))
	return strings.HasPrefix(normalized, "error:") || strings.Contains(normalized, " error:")
}

func limitDownloadFailureMessage(message string) string {
	message = strings.TrimSpace(message)
	runes := []rune(message)
	if len(runes) <= maxDownloadFailureMessageLength {
		return message
	}

	return strings.TrimSpace(string(runes[:maxDownloadFailureMessageLength])) + "..."
}

func (service *DownloadService) updateDownload(id string, update func(task *models.DownloadTask)) {
	service.mu.Lock()
	defer service.mu.Unlock()

	task, err := service.repository.FindByID(id)
	if err != nil {
		return
	}

	update(task)
	task.UpdatedAt = time.Now().UTC()
	if err := service.repository.Update(task); err != nil {
		return
	}
}

func (service *DownloadService) registerActiveDownload(id string, active activeDownload) {
	service.activeMu.Lock()
	defer service.activeMu.Unlock()

	service.active[id] = active
}

func (service *DownloadService) unregisterActiveDownload(id string) {
	service.activeMu.Lock()
	defer service.activeMu.Unlock()

	delete(service.active, id)
}

func (service *DownloadService) activeDownload(id string) activeDownload {
	service.activeMu.Lock()
	defer service.activeMu.Unlock()

	return service.active[id]
}

func (service *DownloadService) setActiveContainerID(id string, containerID string) {
	service.activeMu.Lock()
	active := service.active[id]
	active.containerID = containerID
	service.active[id] = active
	service.activeMu.Unlock()

	service.updateDownload(id, func(task *models.DownloadTask) {
		if task.Status == models.DownloadStatusRunning {
			task.ContainerID = containerID
		}
	})
}

func (service *DownloadService) downloadCancelled(id string) bool {
	task, err := service.GetDownload(id)
	if err != nil {
		return false
	}

	return task.Status == models.DownloadStatusCancelled
}

func isTerminalDownloadStatus(status string) bool {
	return status == models.DownloadStatusCompleted ||
		status == models.DownloadStatusFailed ||
		status == models.DownloadStatusCancelled
}

func downloadCommand(task *models.DownloadTask, downloadsDir string, options YTDLPOptions) []string {
	command := append(options.commandPrefix(),
		"-f", task.FormatID,
		"--newline",
	)
	if strings.Contains(task.FormatID, "+") {
		command = append(command, "--merge-output-format", mergeOutputFormat(task.Extension))
	}

	command = append(command,
		"-o", filepath.ToSlash(filepath.Join(downloadsDir, "%(title)s.%(ext)s")),
		task.URL,
	)

	return command
}

func mergeOutputFormat(extension string) string {
	switch strings.TrimPrefix(strings.ToLower(strings.TrimSpace(extension)), ".") {
	case "mp4", "m4v", "":
		return "mp4"
	case "webm":
		return "webm"
	default:
		return "mkv"
	}
}

func cloneDownloadTask(task *models.DownloadTask) *models.DownloadTask {
	if task == nil {
		return nil
	}

	copied := *task
	return &copied
}

var downloadDestinationPattern = regexp.MustCompile(`(?m)(?:Destination:|Merging formats into)\s+"?(/downloads/[^"\r\n]+)"?|\[download\]\s+(/downloads/[^\r\n]+?)\s+has already been downloaded`)

func extractDownloadFileName(output string) string {
	matches := downloadDestinationPattern.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return ""
	}

	lastMatch := matches[len(matches)-1]
	for _, value := range lastMatch[1:] {
		value = strings.TrimSpace(value)
		if value != "" {
			return filepath.Base(value)
		}
	}

	return ""
}

type downloadLineWriter struct {
	mu     sync.Mutex
	buffer strings.Builder
	onLine func(string)
}

func newDownloadLineWriter(onLine func(string)) *downloadLineWriter {
	return &downloadLineWriter{onLine: onLine}
}

func (writer *downloadLineWriter) Write(data []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	for _, char := range string(data) {
		if char == '\n' || char == '\r' {
			writer.flushLocked()
			continue
		}
		writer.buffer.WriteRune(char)
	}

	return len(data), nil
}

func (writer *downloadLineWriter) Flush() {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	writer.flushLocked()
}

func (writer *downloadLineWriter) flushLocked() {
	line := strings.TrimSpace(writer.buffer.String())
	writer.buffer.Reset()
	if line != "" && writer.onLine != nil {
		writer.onLine(line)
	}
}
