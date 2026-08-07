package config

import (
	"net"
	"os"
)

type Config struct {
	AppEnv             string
	Host               string
	Port               string
	BackendPublicPort  string
	FrontendPort       string
	LocalNetworkIP     string
	DownloadsDir       string
	DatabasePath       string
	YTDLPContainerName string
	YTDLPJSRuntime     string
	YTDLPCookiesFile   string
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		Host:               getEnv("APP_HOST", "0.0.0.0"),
		Port:               getEnv("APP_PORT", "8080"),
		BackendPublicPort:  getEnv("BACKEND_PUBLIC_PORT", "18080"),
		FrontendPort:       getEnv("FRONTEND_PORT", "5173"),
		LocalNetworkIP:     getEnv("LOCAL_NETWORK_IP", ""),
		DownloadsDir:       getEnv("DOWNLOADS_DIR", "./downloads"),
		DatabasePath:       getEnv("DATABASE_PATH", "./data/yt-off.db"),
		YTDLPContainerName: getEnv("YTDLP_CONTAINER_NAME", "yt-off-yt-dlp"),
		YTDLPJSRuntime:     getEnv("YTDLP_JS_RUNTIME", "node"),
		YTDLPCookiesFile:   getEnv("YTDLP_COOKIES_FILE", "/cookies/youtube.txt"),
	}
}

func (c Config) ServerAddress() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
