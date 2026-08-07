package models

import "time"

type DownloadGroup struct {
	ID            string              `json:"id"`
	UserID        string              `json:"user_id"`
	OwnerUsername string              `json:"owner_username,omitempty"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	ItemCount     int                 `json:"item_count"`
	Items         []DownloadGroupItem `json:"items,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type DownloadGroupItem struct {
	ID         string        `json:"id"`
	GroupID    string        `json:"group_id"`
	DownloadID string        `json:"download_id"`
	Position   int           `json:"position"`
	Download   *DownloadTask `json:"download,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}
