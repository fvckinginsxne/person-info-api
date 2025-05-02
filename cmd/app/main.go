package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "person-info/docs"
	"person-info/internal/client/person/agify"
	"person-info/internal/client/person/genderize"
	"person-info/internal/client/person/nationalize"
	"person-info/internal/config"
	"person-info/internal/lib/logger/sl"
	personService "person-info/internal/service/person"
	"person-info/internal/storage/postgres"
	"person-info/internal/transport/handler/person/create"
	del "person-info/internal/transport/handler/person/delete"
	"person-info/internal/transport/handler/person/update"
	healthchecker "person-info/internal/transport/middleware/health-checker"
)

const (
	shutdownTimeout = 10 * time.Second
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

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer cancel()

	dbURL := postgresURL(cfg)

	log.Debug("connecting to postgres", slog.String("url", dbURL))

	storage, err := postgres.New(dbURL)
	if err != nil {
		panic(err)
	}

	ageClient := agify.New(log)
	genderClient := genderize.New(log)
	nationClient := nationalize.New(log)

	service := personService.New(log,
		storage,
		ageClient,
		genderClient,
		nationClient,
	)

	g := gin.New()

	g.Use(gin.Recovery())
	g.Use(healthchecker.New(log, storage))

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	peopleGroup := g.Group("/people")
	{
		peopleGroup.POST("/", create.New(ctx, log, service))
		peopleGroup.PATCH("/:id", update.New(ctx, log, service))
		peopleGroup.DELETE("/:id", del.New(ctx, log, service))
	}

	srvAddr := serverAddr(cfg)

	srv := &http.Server{
		Addr:         srvAddr,
		Handler:      g,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Debug("starting server", slog.String("address", srvAddr))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serverErr:
		log.Error("server error", sl.Err(err))
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown server", sl.Err(err))
	}

	if err := storage.Close(shutdownCtx); err != nil {
		log.Error("failed to shutdown storage", sl.Err(err))
	}

	log.Info("shutdown complete")
}

func serverAddr(cfg *config.Config) string {
	return fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
}

func postgresURL(cfg *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
}
