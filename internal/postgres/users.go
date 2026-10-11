package postgres

import (
	"book-a-nook/internal/booking"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) CreateUser(
	ctx context.Context,
	in booking.CreateUserInput,
) (booking.User, error) {
	const query = `
		INSERT INTO users (id)
		VALUES ($1)
		RETURNING id, created_at
	`

	var created booking.User

	err := s.pool.QueryRow(ctx, query, in.ID).Scan(
		&created.ID,
		&created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_pkey" {
			return booking.User{}, booking.ErrUserExists
		}

		return booking.User{}, fmt.Errorf(
			"CreateUser: %w", err,
		)
	}

	return created, nil
}
