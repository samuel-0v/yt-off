package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/services"
)

type groupRequest struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type groupItemRequest struct {
	UserID     string `json:"user_id"`
	DownloadID string `json:"download_id"`
}

func ListGroupsHandler(groups *services.DownloadGroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items, err := groups.ListGroups(c.Query("scope"), c.Query("user_id"))
		if err != nil {
			return groupError(c, err)
		}

		return c.JSON(items)
	}
}

func GetGroupHandler(groups *services.DownloadGroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		group, err := groups.GetGroup(c.Params("id"))
		if err != nil {
			return groupError(c, err)
		}

		return c.JSON(group)
	}
}

func CreateGroupHandler(groups *services.DownloadGroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request groupRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		group, err := groups.CreateGroup(request.UserID, request.Name, request.Description)
		if err != nil {
			return groupError(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(group)
	}
}

func UpdateGroupHandler(groups *services.DownloadGroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request groupRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		group, err := groups.UpdateGroup(c.Params("id"), request.UserID, request.Name, request.Description)
		if err != nil {
			return groupError(c, err)
		}

		return c.JSON(group)
	}
}

func DeleteGroupHandler(groups *services.DownloadGroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := groups.DeleteGroup(c.Params("id"), c.Query("user_id")); err != nil {
			return groupError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func AddGroupItemHandler(groups *services.DownloadGroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request groupItemRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		item, err := groups.AddItem(c.Params("id"), request.UserID, request.DownloadID)
		if err != nil {
			return groupError(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(item)
	}
}

func RemoveGroupItemHandler(groups *services.DownloadGroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := groups.RemoveItem(c.Params("id"), c.Params("item_id"), c.Query("user_id")); err != nil {
			return groupError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func groupError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	case errors.Is(err, services.ErrDownloadNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "download not found",
		})
	case errors.Is(err, services.ErrDownloadGroupNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "group not found",
		})
	case errors.Is(err, services.ErrDownloadGroupItemNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "group item not found",
		})
	case errors.Is(err, services.ErrDownloadGroupNameRequired):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "group name is required",
		})
	case errors.Is(err, services.ErrDownloadGroupForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "group belongs to another user",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to manage group",
		})
	}
}
