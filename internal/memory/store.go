package memory

import (
	"book-a-nook/internal/booking"
	"context"
	"sync"
	"time"
)

type Store struct {
	mu            sync.Mutex
	resources     map[string]booking.Resource
	users         map[string]booking.User
	slots         map[int]booking.Slot
	bookings      map[int]booking.Booking
	nextSlotID    int
	nextBookingID int
}

func NewStore() *Store {
	return &Store{
		resources:     make(map[string]booking.Resource),
		users:         make(map[string]booking.User),
		slots:         make(map[int]booking.Slot),
		bookings:      make(map[int]booking.Booking),
		nextSlotID:    1,
		nextBookingID: 1,
	}
}

func (s *Store) CreateResource(ctx context.Context, in booking.CreateResourceInput) (booking.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.Resource{}, err
	}

	if _, exists := s.resources[in.ID]; exists {
		return booking.Resource{}, booking.ErrResourceAlreadyCreated
	}

	resource := booking.Resource{
		ID:        in.ID,
		CreatedAt: time.Now(),
	}

	s.resources[resource.ID] = resource

	return resource, nil
}

func (s *Store) CreateUser(ctx context.Context, in booking.CreateUserInput) (booking.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.User{}, err
	}

	if _, exists := s.users[in.ID]; exists {
		return booking.User{}, booking.ErrUserAlreadyCreated
	}

	user := booking.User{
		ID:        in.ID,
		CreatedAt: time.Now(),
	}

	s.users[user.ID] = user

	return user, nil
}

func (s *Store) CreateSlot(ctx context.Context, in booking.CreateSlotInput) (booking.Slot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.Slot{}, err
	}

	if _, exists := s.resources[in.ResourceID]; !exists {
		return booking.Slot{}, booking.ErrResourceNotFound
	}

	slot := booking.Slot{
		ID:         s.nextSlotID,
		ResourceID: in.ResourceID,
		StartsAt:   in.StartsAt,
		EndsAt:     in.EndsAt,
		CreatedAt:  time.Now(),
	}

	s.slots[slot.ID] = slot
	s.nextSlotID++

	return slot, nil
}

func (s *Store) CreateBooking(ctx context.Context, in booking.CreateBookingInput) (booking.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.Booking{}, err
	}

	if _, exists := s.slots[in.SlotID]; !exists {
		return booking.Booking{}, booking.ErrSlotNotFound
	}

	if _, exists := s.users[in.UserID]; !exists {
		return booking.Booking{}, booking.ErrUserNotFound
	}

	for _, b := range s.bookings {
		if b.SlotID == in.SlotID && b.Status == booking.StatusActive {
			return booking.Booking{}, booking.ErrSlotAlreadyBooked
		}
		if slotsOverlap(s.slots[b.SlotID], s.slots[in.SlotID]) && b.Status == booking.StatusActive {
			return booking.Booking{}, booking.ErrSlotsOverlap
		}
	}

	booking := booking.Booking{
		ID:        s.nextBookingID,
		SlotID:    in.SlotID,
		UserID:    in.UserID,
		Status:    booking.StatusActive,
		CreatedAt: time.Now(),
	}

	s.bookings[booking.ID] = booking
	s.nextBookingID++

	return booking, nil
}

func slotsOverlap(a, b booking.Slot) bool {
	return a.ResourceID == b.ResourceID &&
		a.StartsAt.Before(b.EndsAt) &&
		b.StartsAt.Before(a.EndsAt)
}
