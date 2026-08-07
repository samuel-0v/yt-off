package services

import (
	"regexp"
	"strconv"
	"strings"
)

type DownloadProgress struct {
	Progress float64
	Speed    string
	ETA      string
}

var (
	progressPercentPattern  = regexp.MustCompile(`\[download\]\s+([0-9]+(?:\.[0-9]+)?)%`)
	progressSpeedETAPattern = regexp.MustCompile(`\bat\s+(\S+)\s+ETA\s+(\S+)`)
	progressSpeedPattern    = regexp.MustCompile(`\bat\s+(\S+/s)`)
)

func ParseDownloadProgress(line string) (DownloadProgress, bool) {
	line = strings.TrimSpace(line)
	match := progressPercentPattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return DownloadProgress{}, false
	}

	progress, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return DownloadProgress{}, false
	}

	info := DownloadProgress{
		Progress: progress,
	}

	if speedETA := progressSpeedETAPattern.FindStringSubmatch(line); len(speedETA) >= 3 {
		info.Speed = speedETA[1]
		info.ETA = speedETA[2]
		return info, true
	}

	if speed := progressSpeedPattern.FindStringSubmatch(line); len(speed) >= 2 {
		info.Speed = speed[1]
	}

	return info, true
}
