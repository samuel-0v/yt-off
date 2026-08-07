package services

import (
	"math"
	"syscall"

	"yt-off/backend/internal/models"
)

func GetStorageInfo(path string) (models.StorageInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return models.StorageInfo{}, err
	}

	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	free := stat.Bavail * blockSize
	used := total - free

	usagePercent := 0
	if total > 0 {
		usagePercent = int(math.Round((float64(used) / float64(total)) * 100))
	}

	return models.StorageInfo{
		Total:        total,
		Used:         used,
		Free:         free,
		UsagePercent: usagePercent,
	}, nil
}
