package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"person-info/internal/domain/model"
)

type Storage struct {
	db *sql.DB
}

func New(connURL string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) PersonExists(ctx context.Context, person *model.Person) (bool, error) {
	const op = "service.person.Exists"

	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM people WHERE name=$1 and surname=$2 and patronymic=$3)
	`, person.Name, person.Surname, person.Patronymic).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

func (s *Storage) SavePerson(ctx context.Context, person *model.Person) error {
	const op = "service.person.SavePerson"

	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO people (name, surname, patronymic, age, gender, nationality) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	done := make(chan struct{})

	var closeErr error
	go func() {
		closeErr = s.db.Close()
		close(done)
	}()

	select {
	case <-done:
		return closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Storage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
