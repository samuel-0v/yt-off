package services

import (
	"context"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	dockerclient "yt-off/backend/internal/docker"
	"yt-off/backend/internal/models"
)

func ReadDockerStatus(ctx context.Context, containerName string) (models.DockerStatus, error) {
	cli, err := dockerclient.NewClient()
	if err != nil {
		return models.DockerStatus{
			Docker:         "disconnected",
			YTDLPContainer: "unknown",
		}, err
	}
	defer cli.Close()

	if err := dockerclient.Ping(ctx, cli); err != nil {
		return models.DockerStatus{
			Docker:         "disconnected",
			YTDLPContainer: "unknown",
		}, err
	}

	containerInfo, found, err := dockerclient.FindContainerByName(ctx, cli, containerName)
	if err != nil {
		return models.DockerStatus{
			Docker:         "connected",
			YTDLPContainer: "unknown",
		}, err
	}

	containerStatus := "missing"
	if found {
		containerStatus = containerInfo.State
		if containerInfo.State == "running" {
			containerStatus = "available"
		}
	}

	return models.DockerStatus{
		Docker:         "connected",
		YTDLPContainer: containerStatus,
	}, nil
}

func ReadYTDLPVersion(ctx context.Context, containerName string) (string, error) {
	result, err := execContainerCommand(ctx, containerName, []string{"yt-dlp", "--version"})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", errCommandFailed
	}

	return firstLine(result.Stdout), nil
}

func ReadFFmpegVersion(ctx context.Context, containerName string) (string, error) {
	result, err := execContainerCommand(ctx, containerName, []string{"ffmpeg", "-version"})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", errCommandFailed
	}

	line := firstLine(result.Stdout)
	fields := strings.Fields(line)
	if len(fields) >= 3 && fields[0] == "ffmpeg" && fields[1] == "version" {
		return fields[2], nil
	}

	return line, nil
}

func BuildSystemInfo(ctx context.Context, containerName string, sqliteStatus string) models.SystemInfo {
	dockerStatus, _ := ReadDockerStatus(ctx, containerName)
	ytDLPVersion, _ := ReadYTDLPVersion(ctx, containerName)
	ffmpegVersion, _ := ReadFFmpegVersion(ctx, containerName)

	return models.SystemInfo{
		OS:             runtime.GOOS,
		Architecture:   runtime.GOARCH,
		Docker:         dockerStatus.Docker,
		SQLite:         sqliteStatus,
		BackendVersion: AppVersion,
		YTDLPVersion:   ytDLPVersion,
		FFmpegVersion:  ffmpegVersion,
	}
}

func BuildNetworkInfo(configuredIP string, backendPort string, frontendPort string) models.NetworkInfo {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		hostname = "yt-off"
	}

	localIP := strings.TrimSpace(configuredIP)
	if localIP == "" {
		localIP = detectLocalIP()
	}
	if localIP == "" {
		localIP = "127.0.0.1"
	}

	if strings.TrimSpace(frontendPort) == "" {
		frontendPort = "5173"
	}
	if strings.TrimSpace(backendPort) == "" {
		backendPort = "18080"
	}

	return models.NetworkInfo{
		Hostname:     hostname,
		LocalIP:      localIP,
		BackendPort:  backendPort,
		FrontendPort: frontendPort,
		URL:          "http://" + net.JoinHostPort(localIP, frontendPort),
	}
}

var errCommandFailed = &commandFailedError{}

type commandFailedError struct{}

func (err *commandFailedError) Error() string {
	return "container command failed"
}

func execContainerCommand(ctx context.Context, containerName string, command []string) (dockerclient.ExecResult, error) {
	cli, err := dockerclient.NewClient()
	if err != nil {
		return dockerclient.ExecResult{}, err
	}
	defer cli.Close()

	commandCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return dockerclient.ExecCommand(commandCtx, cli, containerName, command)
}

func firstLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}

	return ""
}

func detectLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, item := range interfaces {
		if item.Flags&net.FlagUp == 0 || item.Flags&net.FlagLoopback != 0 {
			continue
		}

		addresses, err := item.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			ipNet, ok := address.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}

			return ip.String()
		}
	}

	return ""
}
