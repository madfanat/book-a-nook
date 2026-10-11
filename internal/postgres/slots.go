package postgres

import (
	"book-a-nook/internal/booking"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) CreateSlot(
	ctx context.Context,
	in booking.CreateSlotInput,
) (booking.Slot, error) {
	const query = `
		INSERT INTO slots (resource_id, starts_at, ends_at)
		VALUES ($1, $2, $3)
		RETURNING id, resource_id, starts_at, ends_at, created_at
	`

	var created booking.Slot

	err := s.pool.QueryRow(ctx, query, in.ResourceID, in.StartsAt, in.EndsAt).Scan(
		&created.ID,
		&created.ResourceID,
		&created.StartsAt,
		&created.EndsAt,
		&created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23503" &&
			pgErr.ConstraintName == "slots_resource_fk" {
			return booking.Slot{}, booking.ErrResourceNotFound
		}
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23514" &&
			pgErr.ConstraintName == "slots_valid_time_range" {
			return booking.Slot{}, booking.ErrInvalidTimeRange
		}
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23P01" &&
			pgErr.ConstraintName == "slots_no_overlap" {
			return booking.Slot{}, booking.ErrSlotOverlap
		}

		return booking.Slot{}, fmt.Errorf(
			"CreateSlot: %w", err,
		)
	}

	return created, nil
}

func (s *Store) ListSlots(
	ctx context.Context,
	in booking.ListSlotsInput,
) ([]booking.Slot, error) {
	var resourceID string

	err := s.pool.QueryRow(ctx, `
		SELECT id FROM resources WHERE id = $1`,
		in.ResourceID,
	).Scan(&resourceID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, booking.ErrResourceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find resource: %w", err)
	}

	const query = `
		SELECT id, resource_id, starts_at, ends_at, created_at
		FROM slots
		WHERE resource_id = $1
		ORDER BY id
		LIMIT $2
	`

	rows, err := s.pool.Query(ctx, query, in.ResourceID, in.Limit)
	if err != nil {
		return nil, fmt.Errorf("ListSlots: %w", err)
	}
	defer rows.Close()

	slots := make([]booking.Slot, 0)

	for rows.Next() {
		var slot booking.Slot

		if err := rows.Scan(
			&slot.ID,
			&slot.ResourceID,
			&slot.StartsAt,
			&slot.EndsAt,
			&slot.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan slot: %w", err)
		}

		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate slots: %w", err)
	}

	return slots, nil

}
