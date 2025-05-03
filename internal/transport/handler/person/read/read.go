package read

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"person-info/internal/lib/logger/sl"
	"person-info/internal/transport/dto"
)

type PeopleProvider interface {
	People(ctx context.Context,
		filters *dto.PeopleFilters,
		pagination *dto.Pagination,
		sorting *dto.SortOptions,
	) ([]*dto.PersonResponse, error)
}

// @Summary Get people
// @Description Get people using filters and pagination
// @Tags /people
// @Produce json
// @Param filters query dto.PeopleFilters false "Filters"
// @Param pagination query dto.Pagination false "Pagination"
// @Param sort query dto.SortOptions false "Sorting"
// @Success 200 {object} []dto.PersonResponse "Successfully fetched people"
// @Failure 500 {object{ dto.ErrorResponse "Internal server error"
// @Router /people [get]
func New(
	ctx context.Context,
	log *slog.Logger,
	peopleProvider PeopleProvider,
) gin.HandlerFunc {
	const op = "handler.person.read.New"

	return func(c *gin.Context) {
		log := log.With("op", op)

		var (
			filters    dto.PeopleFilters
			pagination dto.Pagination
			sort       dto.SortOptions
		)

		if !parseQueryWithValidation(c, log, &filters, &pagination, &sort) {
			return
		}

		people, err := peopleProvider.People(ctx, &filters, &pagination, &sort)
		if err != nil {
			log.Error("failed get people", sl.Err(err))

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
			return
		}

		c.JSON(http.StatusOK, people)
	}
}

func parseQueryWithValidation(
	c *gin.Context,
	log *slog.Logger,
	filters *dto.PeopleFilters,
	pagination *dto.Pagination,
	sort *dto.SortOptions,
) bool {
	if err := c.ShouldBindQuery(filters); err != nil {
		log.Error("failed to bind filters:", sl.Err(err))

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid filters: " + err.Error()})
		return false
	}

	if err := c.ShouldBindQuery(pagination); err != nil {
		log.Error("failed to bind pagination:", sl.Err(err))

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid pagination: " + err.Error()})
		return false
	}

	if err := c.ShouldBindQuery(sort); err != nil {
		log.Error("failed to bind sorting:", sl.Err(err))

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid sorting: " + err.Error()})
		return false
	}

	if err := validator.New().Struct(filters); err != nil {
		log.Error("failed to validate filters:", sl.Err(err))

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid filters: " + err.Error()})
		return false
	}

	if err := validator.New().Struct(pagination); err != nil {
		log.Error("failed to validate pagination:", sl.Err(err))

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid pagination: " + err.Error()})
		return false
	}

	if err := validator.New().Struct(sort); err != nil {
		log.Error("failed to validate sorting:", sl.Err(err))

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid sorting: " + err.Error()})
		return false
	}

	return true
}
