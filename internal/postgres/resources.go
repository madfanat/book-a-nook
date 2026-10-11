package postgres

import (
	"book-a-nook/internal/booking"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) CreateResource(
	ctx context.Context,
	in booking.CreateResourceInput,
) (booking.Resource, error) {
	const query = `
		INSERT INTO resources (id)
		VALUES ($1)
		RETURNING id, created_at
	`

	var created booking.Resource

	err := s.pool.QueryRow(ctx, query, in.ID).Scan(
		&created.ID,
		&created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "resources_pkey" {
			return booking.Resource{}, booking.ErrResourceExists
		}

		return booking.Resource{}, fmt.Errorf(
			"CreateResource: %w", err,
		)
	}

	return created, nil
}
