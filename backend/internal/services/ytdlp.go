package services

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrYTDLPAuthenticationRequired = errors.New("yt-dlp authentication required")
	ErrYTDLPRateLimited            = errors.New("yt-dlp rate limited")
	ErrYTDLPExtractionFailed       = errors.New("yt-dlp extraction failed")
)

type YTDLPOptions struct {
	JSRuntime   string
	CookiesFile string
}

func (options YTDLPOptions) commandPrefix() []string {
	command := []string{"yt-dlp"}

	if runtime := strings.TrimSpace(options.JSRuntime); runtime != "" {
		command = append(command, "--js-runtimes", runtime)
	}
	if cookiesFile := strings.TrimSpace(options.CookiesFile); cookiesFile != "" && fileExists(cookiesFile) {
		command = append(command, "--cookies", cookiesFile)
	}

	return command
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func classifyYTDLPError(output string) error {
	normalized := strings.ToLower(output)

	if strings.Contains(normalized, "sign in to confirm") ||
		strings.Contains(normalized, "not a bot") ||
		strings.Contains(normalized, "--cookies") ||
		strings.Contains(normalized, "cookies-from-browser") {
		return ErrYTDLPAuthenticationRequired
	}

	if strings.Contains(normalized, "http error 429") ||
		strings.Contains(normalized, "too many requests") {
		return ErrYTDLPRateLimited
	}

	return ErrYTDLPExtractionFailed
}
