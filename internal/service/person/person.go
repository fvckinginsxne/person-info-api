package person

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"person-info/internal/domain/model"
	"person-info/internal/lib/logger/sl"
	"person-info/internal/storage"
	"person-info/internal/transport/dto"
)

type Storage interface {
	SavePerson(ctx context.Context, person *model.Person) error
	PersonExists(ctx context.Context, person *model.Person) (bool, error)
	DeletePerson(ctx context.Context, id int64) error
	UpdatePerson(ctx context.Context, id int64, person *model.Person) (*model.Person, error)
	People(ctx context.Context,
		filters *model.PeopleFilters,
		pagination *model.Pagination,
		sort *model.SortOptions,
	) ([]*model.Person, error)
}

type AgeProvider interface {
	Age(ctx context.Context, name string) (int, error)
}

type GenderProvider interface {
	Gender(ctx context.Context, name string) (string, error)
}

type NationalityProvider interface {
	Nationality(ctx context.Context, name string) (string, error)
}

var (
	ErrPersonExists    = errors.New("person already exists")
	ErrPersonNotFound  = errors.New("person not found")
	ErrNoUpdatedFields = errors.New("no updated fields")
)

type Service struct {
	log                 *slog.Logger
	storage             Storage
	ageProvider         AgeProvider
	genderProvider      GenderProvider
	nationalityProvider NationalityProvider
}

func New(
	log *slog.Logger,
	storage Storage,
	ageProvider AgeProvider,
	genderProvider GenderProvider,
	nationalityProvider NationalityProvider,
) *Service {
	return &Service{
		log:                 log,
		storage:             storage,
		ageProvider:         ageProvider,
		genderProvider:      genderProvider,
		nationalityProvider: nationalityProvider,
	}
}

func (s *Service) Save(
	ctx context.Context,
	personReq *dto.CreatePersonRequest,
) (*dto.PersonResponse, error) {
	const op = "service.person.Save"

	log := s.log.With(slog.String("op", op))

	log.Info("saving person")

	person := dto.CreateReqToPersonModel(personReq)

	exists, err := s.storage.PersonExists(ctx, person)
	if err != nil {
		log.Error("failed check if person exists", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if exists {
		log.Error("person already exists")

		return nil, fmt.Errorf("%s: %w", op, ErrPersonExists)
	}

	age, err := s.ageProvider.Age(ctx, person.Name)
	if err != nil {
		log.Error("failed to get age", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	gender, err := s.genderProvider.Gender(ctx, person.Name)
	if err != nil {
		log.Error("failed to get gender", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	nationality, err := s.nationalityProvider.Nationality(ctx, person.Name)
	if err != nil {
		log.Error("failed to get nationality", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	person.Age = age
	person.Gender = gender
	person.Nationality = nationality

	if err := s.storage.SavePerson(ctx, person); err != nil {
		log.Error("failed to create person", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("person saved successfully")

	return dto.ToPersonResponse(person), nil
}

func (s *Service) Update(
	ctx context.Context,
	id int64,
	person *dto.UpdatePersonRequest,
) (*dto.PersonResponse, error) {
	const op = "service.person.Update"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("id", id),
	)

	log.Info("updating person")

	updatedPerson, err := s.storage.UpdatePerson(ctx, id, dto.UpdateReqToPersonModel(person))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNoUpdatedFields):
			log.Info("no updated fields")

			return nil, fmt.Errorf("%s: %w", op, ErrNoUpdatedFields)
		case errors.Is(err, storage.ErrPersonNotFound):
			log.Info("person not found")

			return nil, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		default:
			log.Error("failed update person", sl.Err(err))

			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return dto.ToPersonResponse(updatedPerson), nil
}

func (s *Service) People(ctx context.Context,
	filters *dto.PeopleFilters,
	pagination *dto.Pagination,
	sorting *dto.SortOptions,
) ([]*dto.PersonResponse, error) {
	const op = "service.person.People"

	log := s.log.With(slog.String("op", op))

	log.Info("fetching people")

	people, err := s.storage.People(ctx,
		dto.ToPeopleFiltersModel(filters),
		dto.ToPaginationModel(pagination),
		dto.ToSortOptionsModel(sorting),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("people fetched successfully")

	return dto.PeopleToPersonResponse(people), nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	const op = "service.person.Delete"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("id", id),
	)

	log.Info("deleting person")

	if err := s.storage.DeletePerson(ctx, id); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Error("person not found:", slog.Int64("id", id))

			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("person deleted successfully")

	return nil
}
