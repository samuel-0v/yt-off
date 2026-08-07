package models

type SystemInfo struct {
	OS             string `json:"os"`
	Architecture   string `json:"architecture"`
	Docker         string `json:"docker"`
	SQLite         string `json:"sqlite"`
	BackendVersion string `json:"backend_version"`
	YTDLPVersion   string `json:"yt_dlp_version"`
	FFmpegVersion  string `json:"ffmpeg_version"`
}

type YTDLPVersionInfo struct {
	Current string `json:"current"`
}

type NetworkInfo struct {
	Hostname     string `json:"hostname"`
	LocalIP      string `json:"local_ip"`
	BackendPort  string `json:"backend_port"`
	FrontendPort string `json:"frontend_port"`
	URL          string `json:"url"`
}
