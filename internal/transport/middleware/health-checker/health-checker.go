package healthchecker

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"person-info/internal/lib/logger/sl"
	"person-info/internal/transport/dto"
)

type StorageHealthChecker interface {
	Ping(ctx context.Context) error
}

func New(
	log *slog.Logger,
	pgClient StorageHealthChecker,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := pgClient.Ping(ctx); err != nil {
			log.Error("postgres health check failed", sl.Err(err))

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				dto.ErrorResponse{Error: "internal server error"},
			)
			return
		}

		c.Next()
	}
}
