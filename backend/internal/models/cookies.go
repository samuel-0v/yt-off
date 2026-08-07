package models

import "time"

type CookiesInfo struct {
	Exists    bool       `json:"exists"`
	Valid     bool       `json:"valid"`
	FileName  string     `json:"file_name"`
	Size      int64      `json:"size,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Message   string     `json:"message,omitempty"`
}
