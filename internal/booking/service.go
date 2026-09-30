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
	// It returns ErrResourceNotFound if the resource does not exist,
	// or ErrSlotOverlap if the slots overlap.
	CreateSlot(ctx context.Context, in CreateSlotInput) (Slot, error)

	// ListSlots retrieves a list of slots for a given resource.
	// On success, it returns the list of slots.
	//
	// It returns ErrResourceNotFound if the resource does not exist.
	ListSlots(ctx context.Context, in ListSlotsInput) ([]Slot, error)

	// CreateBooking stores a booking from validated input.
	// On success, it returns the booking with ID, SlotID, UserID,
	// StatusActive, and CreatedAt.
	//
	// It returns ErrUserNotFound if the user does not exist,
	// ErrSlotNotFound if the slot does not exist, or
	// ErrSlotAlreadyBooked if the slot is already booked.
	//
	// Cancelled bookings do not prevent creation.
	CreateBooking(ctx context.Context, in CreateBookingInput) (Booking, error)

	// GetBooking retrieves a booking by its ID and UserID.
	// On success, it returns the booking with ID, SlotID, UserID,
	// Status, and CreatedAt.
	//
	// It returns ErrUserNotFound if the user does not exist, or
	// ErrBookingNotFound if the booking does not exist.
	GetBooking(ctx context.Context, in GetBookingInput) (Booking, error)

	// CancelBooking sets StatusCancelled for a booking.
	// On success, it returns nil.
	//
	// It returns ErrInvalidUserID if the UserID is invalid,
	// ErrInvalidBookingID is the BookingID is invalid,
	// ErrUserNotFound if the user does not exist, or
	// ErrBookingNotFound if the booking does not exist.
	CancelBooking(ctx context.Context, in CancelBookingInput) error
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
// It returns ErrInvalidResourceID if the ResourceID is invalid,
// ErrResourceNotFound if the resource does not exist,
// ErrInvalidTimeRange if the time range is invalid, or
// ErrSlotOverlap if the slots overlap.
func (s *Service) CreateSlot(ctx context.Context, in CreateSlotInput) (Slot, error) {
	in.ResourceID = strings.TrimSpace(in.ResourceID)

	if err := in.Validate(); err != nil {
		return Slot{}, err
	}

	return s.store.CreateSlot(ctx, in)
}

// ListSlots retrieves a list of slots for a given resource.
// On success, it returns the list of slots.
//
// It returns ErrInvalidLimit if the Limit is invalid,
// ErrInvalidResourceID if the ResourceID is invalid, or
// ErrResourceNotFound if the resource does not exist.
func (s *Service) ListSlots(ctx context.Context, in ListSlotsInput) ([]Slot, error) {
	in.ResourceID = strings.TrimSpace(in.ResourceID)

	if in.Limit == 0 {
		in.Limit = DefaultSlotLimit
	}

	if err := in.Validate(); err != nil {
		return nil, err
	}

	return s.store.ListSlots(ctx, in)
}

// CreateBooking trims leading and trailing whitespace from the
// UserID, validates the normalized input, and creates a resource.
//
// It returns ErrInvalidSlotID if the SlotID is invalid,
// ErrInvalidUserID if the UserID is invalid,
// ErrUserNotFound if the user does not exist,
// ErrSlotNotFound if the slot does not exist, or
// ErrSlotAlreadyBooked if the slot is already booked,
// Cancelled bookings do not prevent creation.
func (s *Service) CreateBooking(ctx context.Context, in CreateBookingInput) (Booking, error) {
	in.UserID = strings.TrimSpace(in.UserID)

	if err := in.Validate(); err != nil {
		return Booking{}, err
	}
	return s.store.CreateBooking(ctx, in)
}

// GetBooking retrieves a booking by its ID and UserID.
// On success, it returns the booking with ID, SlotID, UserID,
// Status, and CreatedAt.
//
// It returns ErrInvalidUserID if the UserID is invalid,
// ErrInvalidBookingID is the BookingID is invalid,
// ErrUserNotFound if the user does not exist, or
// ErrBookingNotFound if the booking does not exist.
func (s *Service) GetBooking(ctx context.Context, in GetBookingInput) (Booking, error) {
	in.UserID = strings.TrimSpace(in.UserID)

	if err := in.Validate(); err != nil {
		return Booking{}, err
	}

	b, err := s.store.GetBooking(ctx, in)
	if err != nil {
		return Booking{}, err
	}
	if b.UserID != in.UserID {
		return Booking{}, ErrBookingNotFound
	}
	return b, nil
}

// CancelBooking sets StatusCancelled for a booking.
// On success, it returns nil.
//
// It returns ErrInvalidUserID if the UserID is invalid,
// ErrInvalidBookingID is the BookingID is invalid,
// ErrUserNotFound if the user does not exist, or
// ErrBookingNotFound if the booking does not exist.
func (s *Service) CancelBooking(ctx context.Context, in CancelBookingInput) error {
	in.UserID = strings.TrimSpace(in.UserID)

	if err := in.Validate(); err != nil {
		return err
	}
	return s.store.CancelBooking(ctx, in)
}
