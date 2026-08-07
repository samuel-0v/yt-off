package handlers

import (
	"errors"
	"mime"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/repositories"
	"yt-off/backend/internal/services"
)

func ListFilesHandler(files *services.FileService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items, err := files.ListFiles()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to list files",
			})
		}

		return c.JSON(items)
	}
}

func DownloadFileHandler(files *services.FileService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileName, err := fileNameParam(c)
		if err != nil {
			return fileError(c, err)
		}

		file, err := files.GetFile(fileName)
		if err != nil {
			return fileError(c, err)
		}

		dispositionType := "attachment"
		if wantsInlineFile(c) {
			dispositionType = "inline"
		}

		disposition := mime.FormatMediaType(dispositionType, map[string]string{
			"filename": file.Name,
		})
		c.Set(fiber.HeaderContentDisposition, disposition)
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(file.Size, 10))
		c.Set(fiber.HeaderContentType, fileContentType(file.Extension))

		return c.SendFile(file.Path, false)
	}
}

func DeleteFileHandler(files *services.FileService, downloads *repositories.DownloadRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileName, err := fileNameParam(c)
		if err != nil {
			return fileError(c, err)
		}

		if err := files.DeleteFile(fileName); err != nil {
			return fileError(c, err)
		}

		if err := downloads.MarkFileRemoved(fileName); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update download history",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func fileNameParam(c *fiber.Ctx) (string, error) {
	fileName, err := url.PathUnescape(c.Params("name"))
	if err != nil {
		return "", services.ErrInvalidFileName
	}

	return fileName, nil
}

func wantsInlineFile(c *fiber.Ctx) bool {
	value := strings.ToLower(strings.TrimSpace(c.Query("inline")))
	if value == "1" || value == "true" || value == "yes" {
		return true
	}

	return strings.EqualFold(strings.TrimSpace(c.Query("disposition")), "inline")
}

func fileContentType(extension string) string {
	extension = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(extension)), ".")
	if extension == "" {
		return fiber.MIMEOctetStream
	}

	switch extension {
	case "mp4", "m4v":
		return "video/mp4"
	case "mov":
		return "video/quicktime"
	case "webm":
		return "video/webm"
	case "m4a", "aac":
		return "audio/mp4"
	case "mp3":
		return "audio/mpeg"
	case "opus":
		return "audio/ogg"
	case "ogg", "oga":
		return "audio/ogg"
	case "wav":
		return "audio/wav"
	case "flac":
		return "audio/flac"
	}

	if contentType := mime.TypeByExtension("." + extension); contentType != "" {
		return contentType
	}

	return fiber.MIMEOctetStream
}

func fileError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrInvalidFileName):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file name",
		})
	case errors.Is(err, services.ErrFileNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to access file",
		})
	}
}
