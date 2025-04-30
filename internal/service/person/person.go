package person

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"person-info/internal/domain/model"
	"person-info/internal/lib/logger/sl"
)

type Storage interface {
	SavePerson(ctx context.Context, person *model.Person) error
	PersonExists(ctx context.Context, person *model.Person) (bool, error)
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
	ErrPersonExists = errors.New("person already exists")
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
	person *model.Person,
) error {
	const op = "service.person.Save"

	log := s.log.With(slog.String("op", op))

	log.Info("saving person")

	exists, err := s.storage.PersonExists(ctx, person)
	if err != nil {
		log.Error("failed check if person exists", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if exists {
		log.Error("person already exists")

		return fmt.Errorf("%s: %w", op, ErrPersonExists)
	}

	age, err := s.ageProvider.Age(ctx, person.Name)
	if err != nil {
		log.Error("failed to get agify", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	gender, err := s.genderProvider.Gender(ctx, person.Name)
	if err != nil {
		log.Error("failed to get genderize", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	nationality, err := s.nationalityProvider.Nationality(ctx, person.Name)
	if err != nil {
		log.Error("failed to get nationalize", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	person.Age = age
	person.Gender = gender
	person.Nationality = nationality

	if err := s.storage.SavePerson(ctx, person); err != nil {
		log.Error("failed to save person", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("person saved successfully")

	return nil
}
