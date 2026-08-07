package models

type VideoInfo struct {
	Title     string           `json:"title"`
	Duration  int              `json:"duration"`
	Thumbnail string           `json:"thumbnail"`
	Options   []DownloadOption `json:"options"`
}

type DownloadOption struct {
	Label         string `json:"label"`
	FormatID      string `json:"format_id"`
	Quality       string `json:"quality"`
	Extension     string `json:"extension"`
	Type          string `json:"type"`
	Resolution    string `json:"resolution,omitempty"`
	HasVideo      bool   `json:"has_video"`
	HasAudio      bool   `json:"has_audio"`
	EstimatedSize int64  `json:"estimated_size,omitempty"`
	AudioCodec    string `json:"audio_codec,omitempty"`
	VideoCodec    string `json:"video_codec,omitempty"`
	Bitrate       int    `json:"bitrate,omitempty"`
}
