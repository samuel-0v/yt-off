package models

import "time"

const (
	DownloadStatusQueued    = "queued"
	DownloadStatusRunning   = "running"
	DownloadStatusCompleted = "completed"
	DownloadStatusFailed    = "failed"
	DownloadStatusCancelled = "cancelled"
)

type DownloadTask struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	OwnerUsername string    `json:"owner_username,omitempty"`
	URL           string    `json:"url,omitempty"`
	FormatID      string    `json:"format_id,omitempty"`
	Status        string    `json:"status"`
	Progress      float64   `json:"progress"`
	Speed         string    `json:"speed,omitempty"`
	ETA           string    `json:"eta,omitempty"`
	FileName      string    `json:"filename,omitempty"`
	FileSize      int64     `json:"file_size,omitempty"`
	Extension     string    `json:"extension,omitempty"`
	ContainerID   string    `json:"container_id,omitempty"`
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
