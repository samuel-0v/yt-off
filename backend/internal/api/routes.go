package api

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/api/handlers"
	"yt-off/backend/internal/config"
	"yt-off/backend/internal/models"
	"yt-off/backend/internal/repositories"
	"yt-off/backend/internal/services"
)

func RegisterRoutes(app *fiber.App, cfg config.Config, db *sql.DB, downloadRepository *repositories.DownloadRepository, settingsRepository *repositories.SettingsRepository) error {
	settingsService, err := services.NewSettingsService(settingsRepository, services.SettingsDefaults{
		DownloadDirectory: cfg.DownloadsDir,
		AppName:           services.AppName,
		BackendPort:       cfg.BackendPublicPort,
	})
	if err != nil {
		return err
	}
	cookiesService, err := services.NewCookiesService(cfg.YTDLPCookiesFile)
	if err != nil {
		return err
	}

	currentSettings := settingsService.CurrentSettings()
	ytDLPOptions := services.YTDLPOptions{
		JSRuntime:   cfg.YTDLPJSRuntime,
		CookiesFile: cfg.YTDLPCookiesFile,
	}
	userRepository := repositories.NewUserRepository(db)
	groupRepository := repositories.NewDownloadGroupRepository(db)
	userService := services.NewUserService(userRepository)
	downloadService := services.NewDownloadServiceWithSettingsAndUsers(cfg.YTDLPContainerName, downloadRepository, settingsService, userService, ytDLPOptions)
	groupService := services.NewDownloadGroupService(groupRepository, downloadRepository, userService)
	fileService := services.NewFileService(currentSettings.DownloadDirectory)
	fileService.SetShowHiddenFiles(currentSettings.ShowHiddenFiles)

	settingsService.OnChange(func(settings models.AppSettings) {
		downloadService.ApplySettings(settings)
		fileService.SetDownloadsDir(settings.DownloadDirectory)
		fileService.SetShowHiddenFiles(settings.ShowHiddenFiles)
	})

	app.Get("/health", healthHandler)

	api := app.Group("/api")
	api.Get("/version", versionHandler(cfg))
	api.Get("/storage", storageHandler(settingsService))
	api.Get("/docker/status", dockerStatusHandler(cfg))
	api.Get("/settings", handlers.GetSettingsHandler(settingsService))
	api.Put("/settings", handlers.UpdateSettingsHandler(settingsService))
	api.Get("/system", handlers.SystemHandler(cfg, db))
	api.Get("/ytdlp/version", handlers.YTDLPVersionHandler(cfg))
	api.Get("/network", handlers.NetworkHandler(cfg, settingsService))
	api.Get("/cookies", handlers.CookiesStatusHandler(cookiesService))
	api.Put("/cookies", handlers.SaveCookiesHandler(cookiesService))
	api.Post("/cookies", handlers.SaveCookiesHandler(cookiesService))
	api.Delete("/cookies", handlers.DeleteCookiesHandler(cookiesService))
	api.Get("/users", handlers.ListUsersHandler(userService))
	api.Post("/users", handlers.CreateUserHandler(userService))
	api.Post("/formats", handlers.FormatsHandler(cfg))
	api.Get("/downloads", handlers.ListDownloadsHandler(downloadService))
	api.Post("/downloads", handlers.CreateDownloadHandler(downloadService))
	api.Post("/downloads/:id/copy", handlers.CopyDownloadHandler(downloadService))
	api.Get("/downloads/:id", handlers.GetDownloadHandler(downloadService))
	api.Delete("/downloads/:id", handlers.CancelDownloadHandler(downloadService))
	api.Get("/groups", handlers.ListGroupsHandler(groupService))
	api.Post("/groups", handlers.CreateGroupHandler(groupService))
	api.Get("/groups/:id", handlers.GetGroupHandler(groupService))
	api.Put("/groups/:id", handlers.UpdateGroupHandler(groupService))
	api.Delete("/groups/:id", handlers.DeleteGroupHandler(groupService))
	api.Post("/groups/:id/items", handlers.AddGroupItemHandler(groupService))
	api.Delete("/groups/:id/items/:item_id", handlers.RemoveGroupItemHandler(groupService))
	api.Get("/files", handlers.ListFilesHandler(fileService))
	api.Get("/files/:name", handlers.DownloadFileHandler(fileService))
	api.Delete("/files/:name", handlers.DeleteFileHandler(fileService, downloadRepository))

	return nil
}

func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func versionHandler(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(models.VersionInfo{
			Name:        services.AppName,
			Version:     services.AppVersion,
			Environment: cfg.AppEnv,
		})
	}
}

func storageHandler(settings *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		info, err := services.GetStorageInfo(settings.CurrentSettings().DownloadDirectory)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to read storage information",
			})
		}

		return c.JSON(info)
	}
}

func dockerStatusHandler(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		status, err := services.ReadDockerStatus(ctx, cfg.YTDLPContainerName)
		if err != nil && status.Docker != "connected" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(status)
		}

		return c.JSON(status)
	}
}
