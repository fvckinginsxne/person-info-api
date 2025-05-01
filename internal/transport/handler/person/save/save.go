package save

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	personClient "person-info/internal/client/person"
	"person-info/internal/domain/model"
	"person-info/internal/lib/logger/sl"
	personSevice "person-info/internal/service/person"
	"person-info/internal/transport/dto"
)

type Request struct {
	Name       string `json:"name" binding:"required" example:"John"`
	Surname    string `json:"surname" binding:"required" example:"Snow"`
	Patronymic string `json:"patronymic,omitempty" example:"Dmitrievich"`
}

type PersonSaver interface {
	Save(ctx context.Context, person *model.Person) error
}

// @Summary Save new person
// @Description Saves a person enriching with age, gender, nationality
// @Tags /people
// @Accept json
// @Produce json
// @Param input body Request true "Person request data"
// @Success 201 {object} dto.PersonResponse "Successfully saved person"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 409 {object} dto.ErrorResponse "Person already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /people [post]
func New(
	ctx context.Context,
	log *slog.Logger,
	personSaver PersonSaver,
) gin.HandlerFunc {
	const op = "handler.person.save.New"

	return func(c *gin.Context) {
		log := log.With(slog.String("op", op))

		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")

				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "request body is empty"})
				return
			}
			log.Error("failed to decode request body", sl.Err(err))

			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request"})
			return
		}

		log.Debug("request body received", slog.Any("request", req))

		person := &model.Person{
			Name:       req.Name,
			Surname:    req.Surname,
			Patronymic: req.Patronymic,
		}

		if err := personSaver.Save(ctx, person); err != nil {
			log.Error("failed to save person", sl.Err(err))

			switch {
			case errors.Is(err, personSevice.ErrPersonExists):
				c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "person already exists"})
			case errors.Is(err, personClient.ErrInvalidName):
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid name"})
			default:
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			}
			return
		}

		c.JSON(http.StatusCreated, dto.ToPersonResponse(person))
	}
}
