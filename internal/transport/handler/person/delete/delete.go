package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"person-info/internal/lib/logger/sl"
	personSevice "person-info/internal/service/person"
	"person-info/internal/transport/dto"
)

type PersonDeleter interface {
	Delete(ctx context.Context, id int64) error
}

// @Summary Delete a person
// @Description Deletes a person by person id
// @Tags /people
// @Param id path int true "Person ID"
// @Success 204 "Person deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Missing or invalid id"
// @Failure 404 {object} dto.ErrorResponse "Person not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /people/{id} [delete]
func New(
	ctx context.Context,
	log *slog.Logger,
	personDeleter PersonDeleter,
) gin.HandlerFunc {
	const op = "handler.person.delete.New"

	return func(c *gin.Context) {

		log = log.With(slog.String("op", op))

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Error("failed parse id", sl.Err(err))

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
			return
		}

		log.Debug("delete person with id:", slog.Int("id", id))

		if err := personDeleter.Delete(ctx, int64(id)); err != nil {
			if errors.Is(err, personSevice.ErrPersonNotFound) {
				c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "person not found"})
				return
			}

			log.Error("failed delete person", sl.Err(err))

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
