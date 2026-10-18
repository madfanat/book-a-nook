package booking

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSlotLimit = 20
	MaxSlotLimit     = 100
	maxID            = 1<<31 - 1
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
	After      int
}

func (in ListSlotsInput) Validate() error {
	if err := validateResourceID(in.ResourceID); err != nil {
		return err
	}
	if in.Limit < 1 || in.Limit > MaxSlotLimit {
		return ErrInvalidLimit
	}
	if in.After < 0 || in.After > maxID {
		return ErrInvalidCursor
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
	SlotID     int    `json:"slot_id"`
	UserID     string `json:"user_id"`
	RequestKey string `json:"-"`
}

func (in CreateBookingInput) Validate() error {
	if in.SlotID <= 0 || in.SlotID > maxID {
		return ErrInvalidSlotID
	}
	if err := validateUserID(in.UserID); err != nil {
		return err
	}
	if strings.TrimSpace(in.RequestKey) == "" ||
		len(in.RequestKey) > 128 ||
		strings.HasPrefix(in.RequestKey, "__legacy_booking__:") {
		return ErrInvalidRequestKey
	}
	return nil
}

// RequestHash fingerprints the validated creation payload, independently of
// JSON formatting, user identity, request key, or the booking's current status.
func (in CreateBookingInput) RequestHash() string {
	canonical := "slot_id:" + strconv.Itoa(in.SlotID)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
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
	if id <= 0 || id > maxID {
		return ErrInvalidBookingID
	}
	return nil
}
