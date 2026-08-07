package models

import "time"

type File struct {
	Name       string    `json:"name"`
	Path       string    `json:"-"`
	Size       int64     `json:"size"`
	Extension  string    `json:"extension"`
	ModifiedAt time.Time `json:"modified_at"`
}
