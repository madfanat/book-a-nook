package booking

import (
	"time"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCancelled Status = "cancelled"
)

func (s Status) Validate() error {
	switch s {
	case StatusActive, StatusCancelled:
		return nil
	default:
		return ErrInvalidStatus
	}
}

type Slot struct {
	ID         int       `json:"id"`
	ResourceID int       `json:"resource_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
}

type CreateSlotInput struct {
	ResourceID int
	StartsAt   time.Time
	EndsAt     time.Time
}

func (in CreateSlotInput) Validate() error {
	if in.ResourceID <= 0 {
		return ErrInvalidResourceID
	}
	if in.StartsAt.IsZero() || in.EndsAt.IsZero() {
		return ErrInvalidTimeRange
	}
	if !in.EndsAt.After(in.StartsAt) {
		return ErrInvalidTimeRange
	}
	return nil
}

type CreateBookingInput struct {
	SlotID int
	UserID int
}

type Booking struct {
	ID          int       `json:"id"`
	SlotID      int       `json:"slot_id"`
	UserID      int       `json:"user_id"`
	Status      Status    `json:"status"`
	RequestKey  string    `json:"request_key"`
	RequestHash string    `json:"request_hash"`
	CreatedAt   time.Time `json:"created_at"`
}

func (in CreateBookingInput) Validate() error {
	if in.SlotID <= 0 {
		return ErrInvalidSlotID
	}
	if in.UserID <= 0 {
		return ErrInvalidUserID
	}
	return nil
}
