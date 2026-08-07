package services

import (
	"testing"

	"yt-off/backend/internal/models"
)

func TestParseYTDLPVideoInfoBuildsFriendlyOptions(t *testing.T) {
	payload := `{
		"title": "Example video",
		"duration": 300,
		"thumbnail": "https://example.com/thumb.jpg",
		"formats": [
			{"format_id":"137","ext":"mp4","height":1080,"width":1920,"fps":30,"vcodec":"avc1.640028","acodec":"none","filesize":1000,"tbr":3000},
			{"format_id":"136","ext":"mp4","height":720,"width":1280,"fps":30,"vcodec":"avc1.4d401f","acodec":"none","filesize":700,"tbr":1800},
			{"format_id":"248","ext":"webm","height":1080,"width":1920,"fps":30,"vcodec":"vp9","acodec":"none","filesize":900,"tbr":2500},
			{"format_id":"140","ext":"m4a","vcodec":"none","acodec":"mp4a.40.2","filesize":100,"abr":129},
			{"format_id":"251","ext":"webm","vcodec":"none","acodec":"opus","filesize":80,"abr":160}
		]
	}`

	info, err := parseYTDLPVideoInfo(payload)
	if err != nil {
		t.Fatalf("parseYTDLPVideoInfo() error = %v", err)
	}

	if info.Title != "Example video" {
		t.Fatalf("Title = %q, want %q", info.Title, "Example video")
	}

	mp4Option := findOptionByFormatID(info.Options, "137+140")
	if mp4Option == nil {
		t.Fatalf("expected 1080p MP4 option using 137+140, got %#v", info.Options)
	}
	if mp4Option.Label != "1080p MP4" || !mp4Option.HasVideo || !mp4Option.HasAudio || mp4Option.Quality != "1080p" {
		t.Fatalf("unexpected MP4 option: %#v", mp4Option)
	}
	if mp4Option.VideoCodec != "avc1.640028" || mp4Option.AudioCodec != "mp4a.40.2" {
		t.Fatalf("unexpected MP4 codecs: %#v", mp4Option)
	}

	webmOption := findOptionByFormatID(info.Options, "248+251")
	if webmOption == nil {
		t.Fatalf("expected 1080p WebM option using 248+251, got %#v", info.Options)
	}

	audioOption := findOptionByFormatID(info.Options, "140")
	if audioOption == nil {
		t.Fatalf("expected audio-only M4A option, got %#v", info.Options)
	}
	if audioOption.HasVideo || !audioOption.HasAudio || audioOption.Quality != "audio" {
		t.Fatalf("unexpected audio option: %#v", audioOption)
	}
	if audioOption.AudioCodec != "mp4a.40.2" {
		t.Fatalf("AudioCodec = %q, want mp4a.40.2", audioOption.AudioCodec)
	}
}

func TestParseYTDLPVideoInfoPrefersH264MP4OverAV1MP4(t *testing.T) {
	payload := `{
		"title": "Codec choice",
		"duration": 120,
		"formats": [
			{"format_id":"399","ext":"mp4","height":1080,"width":1920,"fps":30,"vcodec":"av01.0.08M.08","acodec":"none","filesize":800,"tbr":1200},
			{"format_id":"137","ext":"mp4","height":1080,"width":1920,"fps":30,"vcodec":"avc1.640028","acodec":"none","filesize":1000,"tbr":2000},
			{"format_id":"140","ext":"m4a","vcodec":"none","acodec":"mp4a.40.2","filesize":100,"abr":129}
		]
	}`

	info, err := parseYTDLPVideoInfo(payload)
	if err != nil {
		t.Fatalf("parseYTDLPVideoInfo() error = %v", err)
	}

	option := findOptionByQualityAndExtension(info.Options, "1080p", "mp4")
	if option == nil {
		t.Fatalf("expected 1080p MP4 option, got %#v", info.Options)
	}
	if option.FormatID != "137+140" {
		t.Fatalf("FormatID = %q, want 137+140", option.FormatID)
	}
	if !option.HasAudio {
		t.Fatalf("HasAudio = false, want true: %#v", option)
	}
}

func TestParseYTDLPVideoInfoIgnoresStoryboardFormats(t *testing.T) {
	payload := `{
		"title": "Storyboard filter",
		"duration": 120,
		"formats": [
			{"format_id":"sb0","ext":"mhtml","protocol":"mhtml","format_note":"storyboard","height":180,"width":320,"vcodec":"none","acodec":"none"},
			{"format_id":"137","ext":"mp4","height":1080,"width":1920,"fps":30,"vcodec":"avc1.640028","acodec":"none","filesize":1000,"tbr":2000},
			{"format_id":"140","ext":"m4a","vcodec":"none","acodec":"mp4a.40.2","filesize":100,"abr":129}
		]
	}`

	info, err := parseYTDLPVideoInfo(payload)
	if err != nil {
		t.Fatalf("parseYTDLPVideoInfo() error = %v", err)
	}

	for _, option := range info.Options {
		if option.Extension == "mhtml" || option.FormatID == "sb0+140" {
			t.Fatalf("storyboard option should be ignored, got %#v", info.Options)
		}
	}
}

func findOptionByFormatID(options []models.DownloadOption, formatID string) *models.DownloadOption {
	for index := range options {
		if options[index].FormatID == formatID {
			return &options[index]
		}
	}

	return nil
}

func findOptionByQualityAndExtension(options []models.DownloadOption, quality string, extension string) *models.DownloadOption {
	for index := range options {
		if options[index].Quality == quality && options[index].Extension == extension {
			return &options[index]
		}
	}

	return nil
}
