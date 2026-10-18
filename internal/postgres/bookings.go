package postgres

import (
	"book-a-nook/internal/booking"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) CreateBooking(
	ctx context.Context,
	in booking.CreateBookingInput,
) (booking.Booking, error) {
	if existing, err := s.replayBooking(ctx, in); !errors.Is(err, pgx.ErrNoRows) {
		return existing, err
	}

	const query = `
		INSERT INTO bookings (slot_id, user_id, request_key, request_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, slot_id, user_id, status, created_at
	`

	var created booking.Booking

	err := s.pool.QueryRow(ctx, query, in.SlotID, in.UserID, in.RequestKey, in.RequestHash()).Scan(
		&created.ID,
		&created.SlotID,
		&created.UserID,
		&created.Status,
		&created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if existing, replayErr := s.replayBooking(ctx, in); !errors.Is(replayErr, pgx.ErrNoRows) {
				return existing, replayErr
			}
		}
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23503" &&
			pgErr.ConstraintName == "bookings_slot_fk" {
			return booking.Booking{}, booking.ErrSlotNotFound
		}
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23503" &&
			pgErr.ConstraintName == "bookings_user_fk" {
			return booking.Booking{}, booking.ErrUserNotFound
		}
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "bookings_one_active_per_slot" {
			return booking.Booking{}, booking.ErrSlotAlreadyBooked
		}
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "bookings_user_request_key_unique" {
			return booking.Booking{}, booking.ErrIdempotencyConflict
		}

		return booking.Booking{}, fmt.Errorf(
			"CreateBooking: %w", err,
		)
	}

	return created, nil
}

func (s *Store) replayBooking(ctx context.Context, in booking.CreateBookingInput) (booking.Booking, error) {
	const query = `
		SELECT id, slot_id, user_id, status, created_at, request_hash
		FROM bookings
		WHERE user_id = $1 AND request_key = $2
	`
	var existing booking.Booking
	var hash string
	err := s.pool.QueryRow(ctx, query, in.UserID, in.RequestKey).Scan(
		&existing.ID, &existing.SlotID, &existing.UserID,
		&existing.Status, &existing.CreatedAt, &hash,
	)
	if err != nil {
		return booking.Booking{}, fmt.Errorf("find booking request: %w", err)
	}
	if hash != in.RequestHash() {
		return booking.Booking{}, booking.ErrIdempotencyConflict
	}
	return existing, nil
}

func (s *Store) GetBooking(
	ctx context.Context,
	in booking.GetBookingInput,
) (booking.Booking, error) {
	const query = `
		SELECT id, slot_id, user_id, status, created_at
		FROM bookings
		WHERE id = $1 AND user_id = $2
	`

	var got booking.Booking

	err := s.pool.QueryRow(ctx, query, in.ID, in.UserID).Scan(
		&got.ID,
		&got.SlotID,
		&got.UserID,
		&got.Status,
		&got.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return booking.Booking{}, booking.ErrBookingNotFound
	}
	if err != nil {
		return booking.Booking{}, fmt.Errorf("find booking: %w", err)
	}

	return got, nil
}

func (s *Store) CancelBooking(
	ctx context.Context,
	in booking.CancelBookingInput,
) error {
	const query = `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'active'
	`

	tag, err := s.pool.Exec(ctx, query, in.ID, in.UserID)
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}

	if tag.RowsAffected() == 1 {
		return nil
	}

	existing, err := s.GetBooking(ctx, booking.GetBookingInput{
		ID:     in.ID,
		UserID: in.UserID,
	})
	if err != nil {
		return err
	}

	if existing.Status == booking.StatusCancelled {
		return nil
	}

	return fmt.Errorf(
		"cancel booking: unexpected status %q",
		existing.Status,
	)
}
