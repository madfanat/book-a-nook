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
		return ErrInvalidBookingStatus
	}
}

type Resource struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateResourceInput struct {
	ID string `json:"id"`
}

func (in CreateResourceInput) Validate() error {
	if in.ID == "" {
		return ErrInvalidResourceID
	}
	return nil
}

type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserInput struct {
	ID string `json:"id"`
}

func validateUserID(id string) error {
	if id == "" {
		return ErrInvalidUserID
	}
	return nil
}

func (in CreateUserInput) Validate() error {
	return validateUserID(in.ID)
}

type Slot struct {
	ID         int       `json:"id"`
	ResourceID string    `json:"resource_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateSlotInput struct {
	ResourceID string    `json:"resource_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
}

func (in CreateSlotInput) Validate() error {
	if in.ResourceID == "" {
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
	SlotID int    `json:"slot_id"`
	UserID string `json:"user_id"`
}

type Booking struct {
	ID        int       `json:"id"`
	SlotID    int       `json:"slot_id"`
	UserID    string    `json:"user_id"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (in CreateBookingInput) Validate() error {
	if in.SlotID <= 0 {
		return ErrInvalidSlotID
	}
	return validateUserID(in.UserID)
}
