package booking

import (
	"context"
	"strings"
)

type Store interface {
	// CreateResource stores a resource from validated input.
	// On success, it returns the resource with ID and CreatedAt.
	//
	// It returns ErrResourceExists if the resource is
	// already created.
	CreateResource(ctx context.Context, in CreateResourceInput) (Resource, error)

	// CreateUser stores a user from validated input.
	// On success, it returns the user with ID and CreatedAt.
	//
	// It returns ErrUserExists if the user is
	// already created.
	CreateUser(ctx context.Context, in CreateUserInput) (User, error)

	// CreateSlot stores a slot from validated input.
	// On success, it returns the slot with ID, ResourceID,
	// StartsAt, EndsAt, and CreatedAt.
	//
	// It returns ErrResourceNotFound if the resource does not exist.
	CreateSlot(ctx context.Context, in CreateSlotInput) (Slot, error)

	// CreateBooking stores a booking from validated input.
	// On success, it returns the booking with ID, SlotID, UserID,
	// StatusActive, and CreatedAt.
	//
	// It returns ErrUserNotFound if the user does not exist,
	// ErrSlotNotFound if the slot does not exist,
	// ErrSlotAlreadyBooked if the slot is already booked, or
	// ErrBookingOverlap if the bookings overlap.
	//
	// Cancelled bookings do not prevent creation.
	CreateBooking(ctx context.Context, in CreateBookingInput) (Booking, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// CreateResource trims leading and trailing whitespace from the ID,
// validates the normalized input, and creates a resource.
//
// It returns ErrInvalidResourceID if the ID is invalid,
// or ErrResourceExists if the resource is
// already created.
func (s *Service) CreateResource(ctx context.Context, in CreateResourceInput) (Resource, error) {
	in.ID = strings.TrimSpace(in.ID)

	if err := in.Validate(); err != nil {
		return Resource{}, err
	}

	return s.store.CreateResource(ctx, in)
}

// CreateUser trims leading and trailing whitespace from the ID,
// validates the normalized input, and creates a user.
//
// It returns ErrInvalidUserID if the ID is invalid,
// or ErrUserExists if the user is
// already created.
func (s *Service) CreateUser(ctx context.Context, in CreateUserInput) (User, error) {
	in.ID = strings.TrimSpace(in.ID)

	if err := in.Validate(); err != nil {
		return User{}, err
	}

	return s.store.CreateUser(ctx, in)
}

// CreateSlot trims leading and trailing whitespace from the
// ResourceID, validates the normalized input, and creates a slot.
//
// It returns ErrInvalidResourceID if the resource ID is invalid,
// ErrResourceNotFound if the resource does not exist,
// or ErrInvalidTimeRange if the time range is invalid.
func (s *Service) CreateSlot(ctx context.Context, in CreateSlotInput) (Slot, error) {
	in.ResourceID = strings.TrimSpace(in.ResourceID)

	if err := in.Validate(); err != nil {
		return Slot{}, err
	}

	return s.store.CreateSlot(ctx, in)
}

// CreateBooking trims leading and trailing whitespace from the
// UserID, validates the normalized input, and creates a resource.
//
// It returns ErrInvalidSlotID if the slot ID is invalid,
// ErrInvalidUserID if the user ID is invalid,
// ErrUserNotFound if the user does not exist,
// ErrSlotNotFound if the slot does not exist,
// ErrSlotAlreadyBooked if the slot is already booked,
// or ErrBookingOverlap if the bookings overlap.
// Cancelled bookings do not prevent creation.
func (s *Service) CreateBooking(ctx context.Context, in CreateBookingInput) (Booking, error) {
	in.UserID = strings.TrimSpace(in.UserID)

	if err := in.Validate(); err != nil {
		return Booking{}, err
	}
	return s.store.CreateBooking(ctx, in)
}
