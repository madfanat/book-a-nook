package booking

import (
	"time"
)

const (
	DefaultSlotLimit = 20
	MaxSlotLimit     = 100
)

// Resources
type Resource struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateResourceInput struct {
	ID string `json:"id"`
}

func (in CreateResourceInput) Validate() error {
	return validateResourceID(in.ID)
}

func validateResourceID(id string) error {
	if id == "" {
		return ErrInvalidResourceID
	}
	return nil
}

// Users
type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserInput struct {
	ID string `json:"id"`
}

func (in CreateUserInput) Validate() error {
	return validateUserID(in.ID)
}

func validateUserID(id string) error {
	if id == "" {
		return ErrInvalidUserID
	}
	return nil
}

// Slots
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
	if err := validateResourceID(in.ResourceID); err != nil {
		return err
	}
	if in.StartsAt.IsZero() || in.EndsAt.IsZero() {
		return ErrInvalidTimeRange
	}
	if !in.EndsAt.After(in.StartsAt) {
		return ErrInvalidTimeRange
	}
	return nil
}

type ListSlotsInput struct {
	ResourceID string
	Limit      int
}

func (in ListSlotsInput) Validate() error {
	if err := validateResourceID(in.ResourceID); err != nil {
		return err
	}
	if in.Limit < 1 || in.Limit > MaxSlotLimit {
		return ErrInvalidLimit
	}
	return nil
}

// Bookings
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

type Booking struct {
	ID        int       `json:"id"`
	SlotID    int       `json:"slot_id"`
	UserID    string    `json:"user_id"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBookingInput struct {
	SlotID int    `json:"slot_id"`
	UserID string `json:"user_id"`
}

func (in CreateBookingInput) Validate() error {
	if in.SlotID <= 0 {
		return ErrInvalidSlotID
	}
	return validateUserID(in.UserID)
}

type GetBookingInput struct {
	ID     int
	UserID string
}

func (in GetBookingInput) Validate() error {
	if err := validateUserID(in.UserID); err != nil {
		return err
	}
	return validateBookingID(in.ID)
}

type CancelBookingInput struct {
	ID     int
	UserID string
}

func (in CancelBookingInput) Validate() error {
	if err := validateUserID(in.UserID); err != nil {
		return err
	}
	return validateBookingID(in.ID)
}

func validateBookingID(id int) error {
	if id <= 0 {
		return ErrInvalidBookingID
	}
	return nil
}
