package services

import "sync"

const MaxConcurrentDownloads = 2

type DownloadQueue struct {
	maxConcurrent int
	start         func(string)
	mu            sync.Mutex
	running       int
	pending       []string
}

func NewDownloadQueue(maxConcurrent int, start func(string)) *DownloadQueue {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}

	return &DownloadQueue{
		maxConcurrent: maxConcurrent,
		start:         start,
		pending:       make([]string, 0),
	}
}

func (queue *DownloadQueue) Enqueue(id string) {
	var starts []string

	queue.mu.Lock()
	if queue.running < queue.maxConcurrent {
		queue.running++
		starts = append(starts, id)
	} else {
		queue.pending = append(queue.pending, id)
	}
	queue.mu.Unlock()

	queue.startDownloads(starts)
}

func (queue *DownloadQueue) SetMaxConcurrent(maxConcurrent int) {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}

	queue.mu.Lock()
	queue.maxConcurrent = maxConcurrent
	starts := queue.nextStartsLocked()
	queue.mu.Unlock()

	queue.startDownloads(starts)
}

func (queue *DownloadQueue) Complete(id string) {
	var starts []string

	queue.mu.Lock()
	if queue.running > 0 {
		queue.running--
	}
	starts = queue.nextStartsLocked()
	queue.mu.Unlock()

	queue.startDownloads(starts)
}

func (queue *DownloadQueue) Cancel(id string) bool {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	for index, pendingID := range queue.pending {
		if pendingID != id {
			continue
		}

		queue.pending = append(queue.pending[:index], queue.pending[index+1:]...)
		return true
	}

	return false
}

func (queue *DownloadQueue) Snapshot() (int, []string) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	pending := append([]string(nil), queue.pending...)
	return queue.running, pending
}

func (queue *DownloadQueue) nextStartsLocked() []string {
	starts := make([]string, 0)
	for len(queue.pending) > 0 && queue.running < queue.maxConcurrent {
		next := queue.pending[0]
		queue.pending = queue.pending[1:]
		queue.running++
		starts = append(starts, next)
	}

	return starts
}

func (queue *DownloadQueue) startDownloads(ids []string) {
	for _, id := range ids {
		go queue.start(id)
	}
}
