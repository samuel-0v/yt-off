package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/models"
	"yt-off/backend/internal/services"
)

type createDownloadRequest struct {
	URL       string `json:"url"`
	FormatID  string `json:"format_id"`
	Extension string `json:"extension"`
	UserID    string `json:"user_id"`
}

type copyDownloadRequest struct {
	UserID string `json:"user_id"`
}

type createDownloadResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type downloadResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	OwnerUsername string    `json:"owner_username,omitempty"`
	Status        string    `json:"status"`
	Progress      float64   `json:"progress"`
	Speed         string    `json:"speed,omitempty"`
	ETA           string    `json:"eta,omitempty"`
	FileName      string    `json:"filename,omitempty"`
	FileSize      int64     `json:"file_size,omitempty"`
	Extension     string    `json:"extension,omitempty"`
	ContainerID   string    `json:"container_id,omitempty"`
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func CreateDownloadHandler(downloads *services.DownloadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request createDownloadRequest
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

		task, err := downloads.CreateDownload(videoURL, request.FormatID, request.Extension, request.UserID)
		if err != nil {
			return createDownloadError(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(createDownloadResponse{
			ID:     task.ID,
			Status: task.Status,
		})
	}
}

func ListDownloadsHandler(downloads *services.DownloadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tasks, err := downloads.ListDownloads(c.Query("scope"), c.Query("user_id"))
		if err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "user not found",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to list downloads",
			})
		}

		response := make([]downloadResponse, 0, len(tasks))
		for index := range tasks {
			response = append(response, downloadTaskResponse(&tasks[index]))
		}

		return c.JSON(response)
	}
}

func CopyDownloadHandler(downloads *services.DownloadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request copyDownloadRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		task, err := downloads.CopyDownload(c.Params("id"), request.UserID)
		if err != nil {
			return createDownloadError(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(downloadTaskResponse(task))
	}
}

func CancelDownloadHandler(downloads *services.DownloadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		task, err := downloads.CancelDownload(c.Params("id"))
		if err != nil {
			if errors.Is(err, services.ErrDownloadNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "download not found",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to cancel download",
			})
		}

		return c.JSON(downloadTaskResponse(task))
	}
}

func GetDownloadHandler(downloads *services.DownloadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		task, err := downloads.GetDownload(c.Params("id"))
		if err != nil {
			if errors.Is(err, services.ErrDownloadNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "download not found",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to read download",
			})
		}

		return c.JSON(downloadTaskResponse(task))
	}
}

func createDownloadError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrDownloadURLRequired):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "url is required",
		})
	case errors.Is(err, services.ErrDownloadFormatIDRequired):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "format_id is required",
		})
	case errors.Is(err, services.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	case errors.Is(err, services.ErrDownloadNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "download not found",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create download",
		})
	}
}

func downloadTaskResponse(task *models.DownloadTask) downloadResponse {
	return downloadResponse{
		ID:            task.ID,
		UserID:        task.UserID,
		OwnerUsername: task.OwnerUsername,
		Status:        task.Status,
		Progress:      task.Progress,
		Speed:         task.Speed,
		ETA:           task.ETA,
		FileName:      task.FileName,
		FileSize:      task.FileSize,
		Extension:     task.Extension,
		ContainerID:   task.ContainerID,
		Error:         task.Error,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
	}
}
