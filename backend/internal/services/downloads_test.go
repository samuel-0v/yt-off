package services

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"yt-off/backend/internal/database"
	dockerclient "yt-off/backend/internal/docker"
	"yt-off/backend/internal/models"
	"yt-off/backend/internal/repositories"
)

func TestDownloadServiceCancelsQueuedDownload(t *testing.T) {
	service, runner := newTestDownloadService(t, 1)

	first, err := service.CreateDownload("https://example.com/one", "mp4", "mp4", "")
	if err != nil {
		t.Fatalf("CreateDownload() first error = %v", err)
	}
	waitForDownloadStatus(t, service, first.ID, models.DownloadStatusRunning)

	second, err := service.CreateDownload("https://example.com/two", "mp4", "mp4", "")
	if err != nil {
		t.Fatalf("CreateDownload() second error = %v", err)
	}
	waitForDownloadStatus(t, service, second.ID, models.DownloadStatusQueued)

	cancelled, err := service.CancelDownload(second.ID)
	if err != nil {
		t.Fatalf("CancelDownload() error = %v", err)
	}
	if cancelled.Status != models.DownloadStatusCancelled {
		t.Fatalf("Status = %q, want cancelled", cancelled.Status)
	}

	runner.assertStartedCount(t, 1)
	if _, err := service.CancelDownload(first.ID); err != nil {
		t.Fatalf("CancelDownload() cleanup error = %v", err)
	}
}

func TestDownloadServiceCancelsRunningDownload(t *testing.T) {
	service, runner := newTestDownloadService(t, 1)

	task, err := service.CreateDownload("https://example.com/video", "mp4", "mp4", "")
	if err != nil {
		t.Fatalf("CreateDownload() error = %v", err)
	}
	waitForDownloadStatus(t, service, task.ID, models.DownloadStatusRunning)

	cancelled, err := service.CancelDownload(task.ID)
	if err != nil {
		t.Fatalf("CancelDownload() error = %v", err)
	}
	if cancelled.Status != models.DownloadStatusCancelled {
		t.Fatalf("Status = %q, want cancelled", cancelled.Status)
	}

	runner.waitStopped(t, "container-"+task.ID)
	waitForDownloadStatus(t, service, task.ID, models.DownloadStatusCancelled)
}

func TestDownloadCommandForcesMergeOutputFormatForVideoAudio(t *testing.T) {
	command := downloadCommand(&models.DownloadTask{
		URL:       "https://example.com/video",
		FormatID:  "137+140",
		Extension: "mp4",
	}, "/downloads", YTDLPOptions{JSRuntime: "node"})

	want := []string{
		"yt-dlp",
		"--js-runtimes", "node",
		"-f", "137+140",
		"--newline",
		"--merge-output-format", "mp4",
		"-o", "/downloads/%(title)s.%(ext)s",
		"https://example.com/video",
	}

	if len(command) != len(want) {
		t.Fatalf("command = %#v, want %#v", command, want)
	}
	for index := range want {
		if command[index] != want[index] {
			t.Fatalf("command[%d] = %q, want %q; full command %#v", index, command[index], want[index], command)
		}
	}
}

func TestDownloadServiceStoresFailureOutput(t *testing.T) {
	service, runner := newTestDownloadService(t, 1)
	runner.completeWith(dockerclient.ContainerRunResult{
		ExitCode: 1,
		Stderr: strings.Join([]string{
			"[youtube] Extracting URL: https://example.com/video",
			"[download] 12.3% of 10.00MiB at 2.00MiB/s ETA 00:04",
			"ERROR: [youtube] video: Sign in to confirm you're not a bot",
			"Use --cookies-from-browser or --cookies for authenticated requests.",
		}, "\n"),
	}, nil)

	task, err := service.CreateDownload("https://example.com/video", "140", "m4a", "")
	if err != nil {
		t.Fatalf("CreateDownload() error = %v", err)
	}
	waitForDownloadStatus(t, service, task.ID, models.DownloadStatusFailed)

	failed, err := service.GetDownload(task.ID)
	if err != nil {
		t.Fatalf("GetDownload() error = %v", err)
	}

	want := "ERROR: [youtube] video: Sign in to confirm you're not a bot\nUse --cookies-from-browser or --cookies for authenticated requests."
	if failed.Error != want {
		t.Fatalf("Error = %q, want %q", failed.Error, want)
	}
}

func TestDownloadServiceListsAndCopiesDownloadsByUser(t *testing.T) {
	service, runner, users := newTestDownloadServiceWithUsers(t, 4)
	runner.completeWith(dockerclient.ContainerRunResult{
		ExitCode: 1,
		Stderr:   "ERROR: test download failed",
	}, nil)

	alice, err := users.GetOrCreateUser("Alice")
	if err != nil {
		t.Fatalf("GetOrCreateUser() Alice error = %v", err)
	}
	bob, err := users.GetOrCreateUser("Bob")
	if err != nil {
		t.Fatalf("GetOrCreateUser() Bob error = %v", err)
	}

	aliceDownload, err := service.CreateDownload("https://example.com/alice", "140", "m4a", alice.ID)
	if err != nil {
		t.Fatalf("CreateDownload() Alice error = %v", err)
	}
	bobDownload, err := service.CreateDownload("https://example.com/bob", "140", "m4a", bob.ID)
	if err != nil {
		t.Fatalf("CreateDownload() Bob error = %v", err)
	}
	waitForDownloadStatus(t, service, aliceDownload.ID, models.DownloadStatusFailed)
	waitForDownloadStatus(t, service, bobDownload.ID, models.DownloadStatusFailed)

	aliceDownloads, err := service.ListDownloads("mine", alice.ID)
	if err != nil {
		t.Fatalf("ListDownloads() Alice error = %v", err)
	}
	if len(aliceDownloads) != 1 || aliceDownloads[0].UserID != alice.ID || aliceDownloads[0].OwnerUsername != alice.Username {
		t.Fatalf("Alice downloads = %#v", aliceDownloads)
	}

	allDownloads, err := service.ListDownloads("", "")
	if err != nil {
		t.Fatalf("ListDownloads() all error = %v", err)
	}
	if len(allDownloads) != 2 {
		t.Fatalf("All downloads returned %d items, want 2", len(allDownloads))
	}

	copied, err := service.CopyDownload(bobDownload.ID, alice.ID)
	if err != nil {
		t.Fatalf("CopyDownload() error = %v", err)
	}
	waitForDownloadStatus(t, service, copied.ID, models.DownloadStatusFailed)

	copiedTask, err := service.GetDownload(copied.ID)
	if err != nil {
		t.Fatalf("GetDownload() copied error = %v", err)
	}
	if copiedTask.UserID != alice.ID || copiedTask.URL != "https://example.com/bob" {
		t.Fatalf("Copied task = %#v", copiedTask)
	}
}

func newTestDownloadService(t *testing.T, maxConcurrent int) (*DownloadService, *fakeDownloadRunner) {
	t.Helper()

	service, runner, _ := newTestDownloadServiceWithUsers(t, maxConcurrent)

	return service, runner
}

func newTestDownloadServiceWithUsers(t *testing.T, maxConcurrent int) (*DownloadService, *fakeDownloadRunner, *UserService) {
	t.Helper()

	db, err := database.OpenSQLite(filepath.Join(t.TempDir(), "yt-off.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	runner := newFakeDownloadRunner()
	repository := repositories.NewDownloadRepository(db)
	downloadsDir := t.TempDir()
	settingsProvider := staticDownloadSettingsProvider{
		settings: models.AppSettings{
			DownloadDirectory:      downloadsDir,
			MaxConcurrentDownloads: maxConcurrent,
		},
	}
	userService := NewUserService(repositories.NewUserRepository(db))
	service := newDownloadService("yt-off-yt-dlp", downloadsDir, repository, settingsProvider, userService, runner, maxConcurrent, YTDLPOptions{})

	return service, runner, userService
}

func waitForDownloadStatus(t *testing.T, service *DownloadService, id string, status string) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			task, _ := service.GetDownload(id)
			t.Fatalf("timed out waiting for status %q, got %#v", status, task)
		case <-ticker.C:
			task, err := service.GetDownload(id)
			if err != nil {
				continue
			}
			if task.Status == status {
				return
			}
		}
	}
}

type fakeDownloadRunner struct {
	started             chan string
	stopped             chan string
	mu                  sync.Mutex
	ids                 []string
	completeImmediately bool
	runResult           dockerclient.ContainerRunResult
	runError            error
}

func newFakeDownloadRunner() *fakeDownloadRunner {
	return &fakeDownloadRunner{
		started: make(chan string, 10),
		stopped: make(chan string, 10),
		ids:     make([]string, 0),
	}
}

func (runner *fakeDownloadRunner) Run(ctx context.Context, downloadID string, command []string, stdoutWriter io.Writer, stderrWriter io.Writer, onContainerCreated func(string)) (dockerclient.ContainerRunResult, error) {
	containerID := "container-" + downloadID
	if onContainerCreated != nil {
		onContainerCreated(containerID)
	}

	runner.mu.Lock()
	runner.ids = append(runner.ids, downloadID)
	completeImmediately := runner.completeImmediately
	runResult := runner.runResult
	runError := runner.runError
	runner.mu.Unlock()
	runner.started <- downloadID

	if completeImmediately {
		if runResult.ContainerID == "" {
			runResult.ContainerID = containerID
		}

		return runResult, runError
	}

	<-ctx.Done()

	return dockerclient.ContainerRunResult{
		ContainerID: containerID,
		ExitCode:    137,
	}, ctx.Err()
}

func (runner *fakeDownloadRunner) Stop(ctx context.Context, containerID string) error {
	runner.stopped <- containerID
	return nil
}

func (runner *fakeDownloadRunner) completeWith(result dockerclient.ContainerRunResult, err error) {
	runner.mu.Lock()
	defer runner.mu.Unlock()

	runner.completeImmediately = true
	runner.runResult = result
	runner.runError = err
}

func (runner *fakeDownloadRunner) assertStartedCount(t *testing.T, expected int) {
	t.Helper()

	runner.mu.Lock()
	defer runner.mu.Unlock()

	if len(runner.ids) != expected {
		t.Fatalf("started count = %d, want %d: %#v", len(runner.ids), expected, runner.ids)
	}
}

func (runner *fakeDownloadRunner) waitStopped(t *testing.T, expectedContainerID string) {
	t.Helper()

	select {
	case containerID := <-runner.stopped:
		if containerID != expectedContainerID {
			t.Fatalf("stopped container = %q, want %q", containerID, expectedContainerID)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for Stop(%q)", expectedContainerID)
	}
}
