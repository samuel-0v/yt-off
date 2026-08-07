package handlers

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/config"
	"yt-off/backend/internal/services"
)

type formatsRequest struct {
	URL string `json:"url"`
}

func FormatsHandler(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request formatsRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		videoURL := strings.TrimSpace(request.URL)
		if videoURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "url is required",
			})
		}
		if !isValidHTTPURL(videoURL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid url",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		info, err := services.ExtractVideoFormats(ctx, cfg.YTDLPContainerName, videoURL, services.YTDLPOptions{
			JSRuntime:   cfg.YTDLPJSRuntime,
			CookiesFile: cfg.YTDLPCookiesFile,
		})
		if err != nil {
			if errors.Is(err, services.ErrYTDLPAuthenticationRequired) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "youtube requires cookies. Save exported cookies as cookies/youtube.txt and try again.",
				})
			}
			if errors.Is(err, services.ErrYTDLPRateLimited) {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "youtube is rate limiting requests. Try again later or use cookies/youtube.txt.",
				})
			}

			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "failed to extract video information",
			})
		}

		return c.JSON(info)
	}
}

func isValidHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
