package services

import (
	"testing"
	"time"
)

func TestDownloadQueueLimitsConcurrentDownloads(t *testing.T) {
	started := make(chan string, 3)
	queue := NewDownloadQueue(2, func(id string) {
		started <- id
	})

	queue.Enqueue("a")
	queue.Enqueue("b")
	queue.Enqueue("c")

	first := receiveStarted(t, started)
	second := receiveStarted(t, started)
	if first == second {
		t.Fatalf("started duplicate id %q", first)
	}
	if first != "a" && first != "b" {
		t.Fatalf("unexpected first start %q", first)
	}
	if second != "a" && second != "b" {
		t.Fatalf("unexpected second start %q", second)
	}

	assertNoStart(t, started)

	running, pending := queue.Snapshot()
	if running != 2 {
		t.Fatalf("running = %d, want 2", running)
	}
	if len(pending) != 1 || pending[0] != "c" {
		t.Fatalf("pending = %#v, want [c]", pending)
	}
}

func TestDownloadQueueStartsNextAfterCompletion(t *testing.T) {
	started := make(chan string, 3)
	queue := NewDownloadQueue(2, func(id string) {
		started <- id
	})

	queue.Enqueue("a")
	queue.Enqueue("b")
	queue.Enqueue("c")
	receiveStarted(t, started)
	receiveStarted(t, started)

	queue.Complete("a")

	if next := receiveStarted(t, started); next != "c" {
		t.Fatalf("next = %q, want c", next)
	}
}

func TestDownloadQueueCancelsPendingDownload(t *testing.T) {
	started := make(chan string, 3)
	queue := NewDownloadQueue(1, func(id string) {
		started <- id
	})

	queue.Enqueue("a")
	queue.Enqueue("b")
	receiveStarted(t, started)

	if cancelled := queue.Cancel("b"); !cancelled {
		t.Fatal("Cancel() = false, want true")
	}

	queue.Complete("a")
	assertNoStart(t, started)
}

func TestDownloadQueueUpdatesMaxConcurrentDownloads(t *testing.T) {
	started := make(chan string, 4)
	queue := NewDownloadQueue(1, func(id string) {
		started <- id
	})

	queue.Enqueue("a")
	queue.Enqueue("b")
	queue.Enqueue("c")
	receiveStarted(t, started)
	assertNoStart(t, started)

	queue.SetMaxConcurrent(2)

	if next := receiveStarted(t, started); next != "b" {
		t.Fatalf("next = %q, want b", next)
	}

	queue.SetMaxConcurrent(1)
	queue.Complete("a")
	assertNoStart(t, started)

	queue.Complete("b")
	if next := receiveStarted(t, started); next != "c" {
		t.Fatalf("next after lowering limit = %q, want c", next)
	}
}

func receiveStarted(t *testing.T, started <-chan string) string {
	t.Helper()

	select {
	case id := <-started:
		return id
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for download to start")
		return ""
	}
}

func assertNoStart(t *testing.T, started <-chan string) {
	t.Helper()

	select {
	case id := <-started:
		t.Fatalf("unexpected start for %q", id)
	case <-time.After(50 * time.Millisecond):
	}
}
