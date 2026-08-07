package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/models"
	"yt-off/backend/internal/services"
)

func GetSettingsHandler(settings *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(settings.CurrentSettings())
	}
}

func UpdateSettingsHandler(settings *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request models.AppSettings
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		updated, err := settings.UpdateSettings(request)
		if err != nil {
			if errors.Is(err, services.ErrInvalidSettings) || errors.Is(err, services.ErrInvalidDownloadDirectory) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid settings",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update settings",
			})
		}

		return c.JSON(updated)
	}
}
