package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"yt-off/backend/internal/api"
	"yt-off/backend/internal/config"
	"yt-off/backend/internal/database"
	"yt-off/backend/internal/repositories"
)

func main() {
	cfg := config.Load()

	db, err := database.OpenSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	downloadRepository := repositories.NewDownloadRepository(db)
	settingsRepository := repositories.NewSettingsRepository(db)

	app := fiber.New(fiber.Config{
		AppName: "yt-off",
	})

	if err := api.RegisterRoutes(app, cfg, db, downloadRepository, settingsRepository); err != nil {
		log.Fatal(err)
	}

	log.Printf("starting yt-off backend on %s", cfg.ServerAddress())
	if err := app.Listen(cfg.ServerAddress()); err != nil {
		log.Fatal(err)
	}
}
