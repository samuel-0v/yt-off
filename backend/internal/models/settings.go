package models

import "time"

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AppSettings struct {
	DownloadDirectory      string `json:"download_directory"`
	MaxConcurrentDownloads int    `json:"max_concurrent_downloads"`
	Language               string `json:"language"`
	Theme                  string `json:"theme"`
	AppName                string `json:"app_name"`
	BackendPort            string `json:"backend_port"`
	AutomaticUpdates       bool   `json:"automatic_updates"`
	ShowHiddenFiles        bool   `json:"show_hidden_files"`
}
