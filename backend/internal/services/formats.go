package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	dockerclient "yt-off/backend/internal/docker"
	"yt-off/backend/internal/models"
)

type ytDLPVideoInfo struct {
	Title     string        `json:"title"`
	Duration  float64       `json:"duration"`
	Thumbnail string        `json:"thumbnail"`
	Formats   []ytDLPFormat `json:"formats"`
}

type ytDLPFormat struct {
	ID             string  `json:"format_id"`
	Extension      string  `json:"ext"`
	Protocol       string  `json:"protocol"`
	FormatNote     string  `json:"format_note"`
	Resolution     string  `json:"resolution"`
	Height         int     `json:"height"`
	Width          int     `json:"width"`
	FPS            float64 `json:"fps"`
	FileSize       int64   `json:"filesize"`
	FileSizeApprox int64   `json:"filesize_approx"`
	AudioCodec     string  `json:"acodec"`
	VideoCodec     string  `json:"vcodec"`
	Bitrate        float64 `json:"tbr"`
	AudioBitrate   float64 `json:"abr"`
	VideoBitrate   float64 `json:"vbr"`
}

type formatCandidate struct {
	id          string
	extension   string
	quality     string
	resolution  string
	height      int
	width       int
	fps         int
	size        int64
	audioCodec  string
	videoCodec  string
	bitrate     int
	hasAudio    bool
	hasVideo    bool
	containerPr int
	codecPr     int
}

type downloadOptionCandidate struct {
	option    models.DownloadOption
	height    int
	extension string
	score     int
}

func ExtractVideoFormats(ctx context.Context, containerName string, videoURL string, options YTDLPOptions) (models.VideoInfo, error) {
	cli, err := dockerclient.NewClient()
	if err != nil {
		return models.VideoInfo{}, err
	}
	defer cli.Close()

	command := append(options.commandPrefix(), "--dump-json", videoURL)
	result, err := dockerclient.ExecCommand(ctx, cli, containerName, command)
	if err != nil {
		return models.VideoInfo{}, err
	}
	if result.ExitCode != 0 {
		return models.VideoInfo{}, fmt.Errorf("%w: yt-dlp exited with code %d", classifyYTDLPError(result.Stdout+"\n"+result.Stderr), result.ExitCode)
	}

	return parseYTDLPVideoInfo(result.Stdout)
}

func parseYTDLPVideoInfo(output string) (models.VideoInfo, error) {
	var raw ytDLPVideoInfo
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &raw); err != nil {
		return models.VideoInfo{}, err
	}

	return models.VideoInfo{
		Title:     raw.Title,
		Duration:  int(math.Round(raw.Duration)),
		Thumbnail: raw.Thumbnail,
		Options:   buildDownloadOptions(raw.Formats),
	}, nil
}

func buildDownloadOptions(formats []ytDLPFormat) []models.DownloadOption {
	videoOnly := make([]formatCandidate, 0)
	audioOnly := make([]formatCandidate, 0)
	complete := make([]formatCandidate, 0)

	for _, item := range formats {
		candidate := newFormatCandidate(item)
		if candidate.id == "" || (!candidate.hasAudio && !candidate.hasVideo) {
			continue
		}

		switch {
		case candidate.hasVideo && candidate.hasAudio:
			complete = append(complete, candidate)
		case candidate.hasVideo:
			videoOnly = append(videoOnly, candidate)
		case candidate.hasAudio:
			audioOnly = append(audioOnly, candidate)
		}
	}

	sort.SliceStable(audioOnly, func(i, j int) bool {
		return audioScore(audioOnly[i]) > audioScore(audioOnly[j])
	})

	videoOptions := make(map[string]downloadOptionCandidate)
	for _, video := range complete {
		option := videoDownloadOption(video, nil)
		key := optionKey(option)
		insertBestOption(videoOptions, key, downloadOptionCandidate{
			option:    option,
			height:    video.height,
			extension: video.extension,
			score:     videoScore(video, nil),
		})
	}

	for _, video := range videoOnly {
		audio := selectAudioForVideo(video, audioOnly)
		option := videoDownloadOption(video, audio)
		key := optionKey(option)
		insertBestOption(videoOptions, key, downloadOptionCandidate{
			option:    option,
			height:    video.height,
			extension: video.extension,
			score:     videoScore(video, audio),
		})
	}

	orderedVideoOptions := make([]downloadOptionCandidate, 0, len(videoOptions))
	for _, option := range videoOptions {
		orderedVideoOptions = append(orderedVideoOptions, option)
	}
	sort.SliceStable(orderedVideoOptions, func(i, j int) bool {
		left := orderedVideoOptions[i]
		right := orderedVideoOptions[j]
		if containerPriority(left.extension) != containerPriority(right.extension) {
			return containerPriority(left.extension) > containerPriority(right.extension)
		}
		if left.height != right.height {
			return left.height > right.height
		}
		return left.score > right.score
	})

	audioOptions := bestAudioOptions(audioOnly)
	options := make([]models.DownloadOption, 0, len(orderedVideoOptions)+len(audioOptions))
	for _, item := range orderedVideoOptions {
		options = append(options, item.option)
	}
	options = append(options, audioOptions...)

	return options
}

func newFormatCandidate(item ytDLPFormat) formatCandidate {
	extension := strings.ToLower(strings.TrimSpace(item.Extension))
	if isStoryboardFormat(item, extension) {
		return formatCandidate{}
	}

	audioCodec := normalizeCodec(item.AudioCodec)
	videoCodec := normalizeCodec(item.VideoCodec)
	fileSize := item.FileSize
	if fileSize == 0 {
		fileSize = item.FileSizeApprox
	}

	hasAudio := audioCodec != ""
	hasVideo := videoCodec != "" || item.Height > 0 || item.Width > 0
	if !hasAudio && !hasVideo {
		if isAudioExtension(extension) {
			hasAudio = true
		}
		if isVideoExtension(extension) {
			hasVideo = true
			hasAudio = true
		}
	}

	return formatCandidate{
		id:          strings.TrimSpace(item.ID),
		extension:   extension,
		quality:     qualityLabel(item),
		resolution:  normalizeResolution(item),
		height:      item.Height,
		width:       item.Width,
		fps:         int(math.Round(item.FPS)),
		size:        fileSize,
		audioCodec:  audioCodec,
		videoCodec:  videoCodec,
		bitrate:     bestBitrate(item),
		hasAudio:    hasAudio,
		hasVideo:    hasVideo,
		containerPr: containerPriority(extension),
		codecPr:     codecPriority(extension, videoCodec, audioCodec),
	}
}

func videoDownloadOption(video formatCandidate, audio *formatCandidate) models.DownloadOption {
	formatID := video.id
	estimatedSize := video.size
	hasAudio := video.hasAudio
	audioCodec := video.audioCodec
	bitrate := video.bitrate
	if audio != nil {
		formatID = video.id + "+" + audio.id
		estimatedSize += audio.size
		hasAudio = true
		audioCodec = audio.audioCodec
		if bitrate == 0 {
			bitrate = audio.bitrate
		}
	}

	return models.DownloadOption{
		Label:         videoLabel(video),
		FormatID:      formatID,
		Quality:       video.quality,
		Extension:     video.extension,
		Type:          "video",
		Resolution:    video.resolution,
		HasVideo:      true,
		HasAudio:      hasAudio,
		EstimatedSize: estimatedSize,
		AudioCodec:    audioCodec,
		VideoCodec:    video.videoCodec,
		Bitrate:       bitrate,
	}
}

func bestAudioOptions(audioFormats []formatCandidate) []models.DownloadOption {
	grouped := make(map[string]formatCandidate)
	for _, audio := range audioFormats {
		if audio.extension == "" {
			continue
		}
		current, ok := grouped[audio.extension]
		if !ok || audioScore(audio) > audioScore(current) {
			grouped[audio.extension] = audio
		}
	}

	candidates := make([]formatCandidate, 0, len(grouped))
	for _, audio := range grouped {
		candidates = append(candidates, audio)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if audioContainerPriority(candidates[i].extension) != audioContainerPriority(candidates[j].extension) {
			return audioContainerPriority(candidates[i].extension) > audioContainerPriority(candidates[j].extension)
		}
		return audioScore(candidates[i]) > audioScore(candidates[j])
	})

	options := make([]models.DownloadOption, 0, len(candidates))
	for _, audio := range candidates {
		options = append(options, models.DownloadOption{
			Label:         audioLabel(audio),
			FormatID:      audio.id,
			Quality:       "audio",
			Extension:     audio.extension,
			Type:          "audio",
			HasVideo:      false,
			HasAudio:      true,
			EstimatedSize: audio.size,
			AudioCodec:    audio.audioCodec,
			Bitrate:       audio.bitrate,
		})
	}

	return options
}

func selectAudioForVideo(video formatCandidate, audioFormats []formatCandidate) *formatCandidate {
	if len(audioFormats) == 0 {
		return nil
	}

	bestIndex := -1
	bestScore := math.MinInt
	for index, audio := range audioFormats {
		score := audioScore(audio)
		if isCompatibleAudio(video, audio) {
			score += 100000000
		}
		if score > bestScore {
			bestScore = score
			bestIndex = index
		}
	}

	if bestIndex == -1 {
		return nil
	}

	return &audioFormats[bestIndex]
}

func isCompatibleAudio(video formatCandidate, audio formatCandidate) bool {
	if isMP4Family(video.extension, video.videoCodec) {
		return isMP4Audio(audio.extension, audio.audioCodec)
	}
	if isWebMFamily(video.extension, video.videoCodec) {
		return audio.extension == "webm" || strings.Contains(audio.audioCodec, "opus") || strings.Contains(audio.audioCodec, "vorbis")
	}

	return true
}

func insertBestOption(options map[string]downloadOptionCandidate, key string, candidate downloadOptionCandidate) {
	current, ok := options[key]
	if !ok || candidate.score > current.score {
		options[key] = candidate
	}
}

func optionKey(option models.DownloadOption) string {
	return strings.Join([]string{option.Type, option.Quality, option.Extension}, ":")
}

func videoScore(video formatCandidate, audio *formatCandidate) int {
	score := video.containerPr*10000000 + video.height*10000 + video.codecPr*100 + video.bitrate
	if audio != nil {
		score += audioScore(*audio)
	}

	return score
}

func audioScore(audio formatCandidate) int {
	return audioContainerPriority(audio.extension)*100000 + audio.codecPr*1000 + audio.bitrate
}

func bestBitrate(item ytDLPFormat) int {
	bitrate := item.Bitrate
	if bitrate == 0 {
		bitrate = item.VideoBitrate
	}
	if bitrate == 0 {
		bitrate = item.AudioBitrate
	}

	return int(math.Round(bitrate))
}

func videoLabel(video formatCandidate) string {
	extension := strings.ToUpper(video.extension)
	if video.quality != "" && video.quality != "video" {
		return fmt.Sprintf("%s %s", video.quality, extension)
	}

	return fmt.Sprintf("Video %s", extension)
}

func audioLabel(audio formatCandidate) string {
	return fmt.Sprintf("Áudio %s", strings.ToUpper(audio.extension))
}

func qualityLabel(item ytDLPFormat) string {
	if item.Height > 0 {
		return fmt.Sprintf("%dp", item.Height)
	}

	resolution := normalizeResolution(item)
	if resolution != "" {
		return resolution
	}

	return "video"
}

func normalizeCodec(codec string) string {
	codec = strings.TrimSpace(codec)
	if codec == "" || codec == "none" {
		return ""
	}

	return codec
}

func normalizeResolution(item ytDLPFormat) string {
	resolution := strings.TrimSpace(item.Resolution)
	if resolution != "" && resolution != "audio only" && resolution != "unknown" {
		return resolution
	}
	if item.Height > 0 {
		return fmt.Sprintf("%dp", item.Height)
	}

	return ""
}

func codecPriority(extension string, videoCodec string, audioCodec string) int {
	switch {
	case isMP4Family(extension, videoCodec) && isH264(videoCodec) && isMP4Audio(extension, audioCodec):
		return 700
	case isMP4Family(extension, videoCodec) && isH264(videoCodec):
		return 650
	case isMP4Family(extension, videoCodec) && isMP4Audio(extension, audioCodec):
		return 550
	case isMP4Family(extension, videoCodec):
		return 450
	case isWebMFamily(extension, videoCodec) && (strings.Contains(videoCodec, "vp9") || strings.Contains(videoCodec, "av01") || strings.Contains(videoCodec, "av1")):
		return 350
	case isWebMFamily(extension, videoCodec):
		return 300
	case isMP4Audio(extension, audioCodec):
		return 250
	case strings.Contains(audioCodec, "opus"):
		return 220
	default:
		return 100
	}
}

func containerPriority(extension string) int {
	switch extension {
	case "mp4", "m4v":
		return 300
	case "webm":
		return 200
	case "mkv":
		return 100
	default:
		return 50
	}
}

func audioContainerPriority(extension string) int {
	switch extension {
	case "m4a", "mp4", "aac":
		return 300
	case "webm", "opus":
		return 200
	case "mp3":
		return 150
	default:
		return 50
	}
}

func isMP4Family(extension string, videoCodec string) bool {
	return extension == "mp4" || extension == "m4v" || strings.HasPrefix(videoCodec, "avc1") || strings.Contains(videoCodec, "h264")
}

func isWebMFamily(extension string, videoCodec string) bool {
	return extension == "webm" || strings.Contains(videoCodec, "vp9") || strings.Contains(videoCodec, "vp8") || strings.Contains(videoCodec, "av01") || strings.Contains(videoCodec, "av1")
}

func isH264(videoCodec string) bool {
	videoCodec = strings.ToLower(videoCodec)
	return strings.HasPrefix(videoCodec, "avc1") || strings.Contains(videoCodec, "h264") || strings.Contains(videoCodec, "avc")
}

func isMP4Audio(extension string, audioCodec string) bool {
	return extension == "m4a" || extension == "aac" || (extension == "mp4" && audioCodec != "") || strings.HasPrefix(audioCodec, "mp4a") || strings.Contains(audioCodec, "aac")
}

func isAudioExtension(extension string) bool {
	switch extension {
	case "m4a", "mp3", "aac", "opus", "ogg", "webm":
		return true
	default:
		return false
	}
}

func isVideoExtension(extension string) bool {
	switch extension {
	case "mp4", "webm", "mkv", "m4v", "mov":
		return true
	default:
		return false
	}
}

func isStoryboardFormat(item ytDLPFormat, extension string) bool {
	formatNote := strings.ToLower(item.FormatNote)
	protocol := strings.ToLower(strings.TrimSpace(item.Protocol))

	return extension == "mhtml" || protocol == "mhtml" || strings.Contains(formatNote, "storyboard")
}
