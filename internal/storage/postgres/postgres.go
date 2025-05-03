package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"

	"person-info/internal/domain/model"
	"person-info/internal/storage"
)

type Storage struct {
	db      *sql.DB
	builder sq.StatementBuilderType
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

	return &Storage{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (s *Storage) PersonExists(ctx context.Context, person *model.Person) (bool, error) {
	const op = "storage.postgres.Exists"

	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM people WHERE name=$1 and surname=$2 and patronymic=$3)
	`, person.Name, person.Surname, person.Patronymic).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

func (s *Storage) People(
	ctx context.Context,
	filters *model.PeopleFilters,
	pagination *model.Pagination,
	sort *model.SortOptions,
) ([]*model.Person, error) {
	const op = "service.person.People"

	query := s.builder.Select(
		"name",
		"surname",
		"patronymic",
		"age",
		"gender",
		"nationality",
	).From("people")

	query = setFilters(query, filters)

	if sort.By != "" {
		order := sort.Order
		if order == "" {
			order = "ASC"
		}
		query = query.OrderBy(fmt.Sprintf("%s %s", sort.By, order))
	}

	if pagination.Size > 0 {
		query = query.Limit(uint64(pagination.Size))
	}

	if pagination.Page > 1 && pagination.Size > 0 {
		offset := (pagination.Page - 1) * pagination.Size
		query = query.Offset(uint64(offset))
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var people []*model.Person
	for rows.Next() {
		var person model.Person

		err := rows.Scan(
			&person.Name,
			&person.Surname,
			&person.Patronymic,
			&person.Age,
			&person.Gender,
			&person.Nationality,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		people = append(people, &person)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return people, nil
}

func (s *Storage) SavePerson(ctx context.Context, person *model.Person) error {
	const op = "storage.postgres.SavePerson"

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

func (s *Storage) UpdatePerson(
	ctx context.Context,
	id int64,
	person *model.Person,
) (*model.Person, error) {
	const op = "storage.postgres.UpdatePerson"

	updateBuilder := s.builder.Update("people")

	updateBuilder = setUpdatedFields(updateBuilder, person)

	if _, args, _ := updateBuilder.ToSql(); len(args) == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrNoUpdatedFields)
	}

	query, args, err := updateBuilder.
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING name, surname, patronymic, age, gender, nationality").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(
		&person.Name,
		&person.Surname,
		&person.Patronymic,
		&person.Age,
		&person.Gender,
		&person.Nationality,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return person, nil
}

func (s *Storage) DeletePerson(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeletePerson"

	result, err := s.db.ExecContext(ctx, `DELETE FROM people WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
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

func setFilters(query sq.SelectBuilder, filters *model.PeopleFilters) sq.SelectBuilder {
	if filters.Name != "" {
		query = query.Where(sq.ILike{"name": fmt.Sprintf("%%%s%%", filters.Name)})
	}

	if filters.Surname != "" {
		query = query.Where(sq.ILike{"surname": fmt.Sprintf("%%%s%%", filters.Surname)})
	}

	if filters.Age > 0 {
		query = query.Where(sq.GtOrEq{"age": filters.Age})
	}

	if filters.Gender != "" {
		query = query.Where(sq.Eq{"gender": filters.Gender})
	}

	if filters.Nationality != "" {
		query = query.Where(sq.Eq{"nationality": filters.Nationality})
	}

	return query
}

func setUpdatedFields(updateBuilder sq.UpdateBuilder, person *model.Person) sq.UpdateBuilder {
	if person.Name != "" {
		updateBuilder = updateBuilder.Set("name", person.Name)
	}

	if person.Surname != "" {
		updateBuilder = updateBuilder.Set("surname", person.Surname)
	}

	if person.Patronymic != "" {
		updateBuilder = updateBuilder.Set("patronymic", person.Patronymic)
	}

	if person.Age > 0 {
		updateBuilder = updateBuilder.Set("age", person.Age)
	}

	if person.Gender != "" {
		updateBuilder = updateBuilder.Set("gender", person.Gender)
	}

	if person.Nationality != "" {
		updateBuilder = updateBuilder.Set("nationality", person.Nationality)
	}

	return updateBuilder
}
