package update

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"person-info/internal/lib/logger/sl"
	personService "person-info/internal/service/person"
	"person-info/internal/transport/dto"
)

type PersonUpdater interface {
	Update(ctx context.Context, id int64, person *dto.UpdatePersonRequest) (*dto.PersonResponse, error)
}

// @Summary Update a person
// @Description Updates a person by id
// @Tags /people
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Param input body dto.UpdatePersonRequest true "Update fields"
// @Success 200 {object} dto.PersonResponse "Updated person"
// @Failure 400 {object} dto.ErrorResponse "Invalid input"
// @Failure 404 {object} dto.ErrorResponse "Person not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /people/{id} [patch]
func New(
	ctx context.Context,
	log *slog.Logger,
	personUpdater PersonUpdater,
) gin.HandlerFunc {
	const op = "handler.person.update.New"

	return func(c *gin.Context) {
		log := log.With(slog.String("op", op))

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			log.Error("failed to parse id param")

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id param"})
			return
		}

		var req *dto.UpdatePersonRequest
		if err := c.ShouldBindJSON(req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")

				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "request body is empty"})
				return
			}
			log.Error("failed to decode request body", sl.Err(err))

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request"})
			return
		}

		updatedPerson, err := personUpdater.Update(ctx, id, req)
		if err != nil {
			switch {
			case errors.Is(err, personService.ErrPersonNotFound):
				log.Error("person not found")

				c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "person not found"})
			case errors.Is(err, personService.ErrNoUpdatedFields):
				log.Error("no updated fields")

				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "no updated fields"})
			default:
				log.Error("failed to update person", sl.Err(err))

				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			}
			return
		}

		c.JSON(http.StatusOK, updatedPerson)
	}
}
