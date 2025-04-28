package main

import (
	"log/slog"
	"os"

	"person-info/internal/config"
)

// @title Person Info API
// @version 1.0
// @description API for getting most probable age, gender, nationality for a person
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	cfg := config.MustLoad()

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	log.Info("starting service")

	log.Debug("config loaded", slog.Any("config", cfg))
}
