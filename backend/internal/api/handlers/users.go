package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/services"
)

type createUserRequest struct {
	Username string `json:"username"`
}

func ListUsersHandler(users *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items, err := users.ListUsers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to list users",
			})
		}

		return c.JSON(items)
	}
}

func CreateUserHandler(users *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request createUserRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		user, err := users.GetOrCreateUser(request.Username)
		if err != nil {
			if errors.Is(err, services.ErrUsernameRequired) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "username is required",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create user",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(user)
	}
}
