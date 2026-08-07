package handlers

import (
	"context"
	"database/sql"
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/config"
	"yt-off/backend/internal/models"
	"yt-off/backend/internal/services"
)

func SystemHandler(cfg config.Config, db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()

		return c.JSON(services.BuildSystemInfo(ctx, cfg.YTDLPContainerName, sqliteStatus(ctx, db)))
	}
}

func YTDLPVersionHandler(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		version, err := services.ReadYTDLPVersion(ctx, cfg.YTDLPContainerName)
		if err != nil || version == "" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "failed to read yt-dlp version",
			})
		}

		return c.JSON(models.YTDLPVersionInfo{Current: version})
	}
}

func NetworkHandler(cfg config.Config, settings *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentSettings := settings.CurrentSettings()
		backendPort := currentSettings.BackendPort
		if backendPort == "" {
			backendPort = cfg.BackendPublicPort
		}

		return c.JSON(services.BuildNetworkInfo(networkIPFromRequest(c, cfg.LocalNetworkIP), backendPort, cfg.FrontendPort))
	}
}

func sqliteStatus(ctx context.Context, db *sql.DB) string {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return "disconnected"
	}

	return "connected"
}

func networkIPFromRequest(c *fiber.Ctx, configuredIP string) string {
	host := c.Get("X-Forwarded-Host")
	if host == "" {
		host = c.Hostname()
	}
	requestHost := cleanRequestHost(host)
	if isUsableNetworkHost(requestHost) {
		return requestHost
	}

	configuredIP = strings.TrimSpace(configuredIP)
	if isUsableNetworkHost(configuredIP) {
		return configuredIP
	}

	return ""
}

func cleanRequestHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	requestHost, _, err := net.SplitHostPort(host)
	if err != nil {
		requestHost = host
	}

	return strings.Trim(strings.TrimSpace(requestHost), "[]")
}

func isUsableNetworkHost(host string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" || host == "localhost" || host == "backend" || strings.Contains(host, "backend") {
		return false
	}
	if strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "0.") {
		return false
	}
	if host == "::1" {
		return false
	}

	return true
}
