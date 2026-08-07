package models

type StorageInfo struct {
	Total        uint64 `json:"total"`
	Used         uint64 `json:"used"`
	Free         uint64 `json:"free"`
	UsagePercent int    `json:"usage_percent"`
}
