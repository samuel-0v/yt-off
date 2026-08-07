package handlers

import (
	"errors"
	"io"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/services"
)

const maxCookiesFileSize = 2 << 20

type saveCookiesRequest struct {
	Content string `json:"content"`
}

func CookiesStatusHandler(cookies *services.CookiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		info, err := cookies.Status()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to read cookies status",
			})
		}

		return c.JSON(info)
	}
}

func SaveCookiesHandler(cookies *services.CookiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		content, err := readCookiesContent(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid cookies file",
			})
		}

		info, err := cookies.Save(content)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCookiesFile) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "cookies must be in Netscape cookies.txt format",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save cookies",
			})
		}

		return c.JSON(info)
	}
}

func DeleteCookiesHandler(cookies *services.CookiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := cookies.Delete(); err != nil {
			if errors.Is(err, services.ErrCookiesNotFound) {
				return c.SendStatus(fiber.StatusNoContent)
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to delete cookies",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func readCookiesContent(c *fiber.Ctx) (string, error) {
	if form, err := c.MultipartForm(); err == nil && form.File != nil {
		files := form.File["file"]
		if len(files) > 0 {
			file := files[0]
			if file.Size > maxCookiesFileSize {
				return "", services.ErrInvalidCookiesFile
			}

			opened, err := file.Open()
			if err != nil {
				return "", err
			}
			defer opened.Close()

			data, err := io.ReadAll(io.LimitReader(opened, maxCookiesFileSize+1))
			if err != nil {
				return "", err
			}
			if len(data) > maxCookiesFileSize {
				return "", services.ErrInvalidCookiesFile
			}

			return string(data), nil
		}
	}

	var request saveCookiesRequest
	if err := c.BodyParser(&request); err != nil {
		return "", err
	}

	return request.Content, nil
}
