package memory

import (
	"book-a-nook/internal/booking"
	"context"
	"sort"
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

// Resources
func (s *Store) CreateResource(ctx context.Context, in booking.CreateResourceInput) (booking.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.Resource{}, err
	}

	if _, exists := s.resources[in.ID]; exists {
		return booking.Resource{}, booking.ErrResourceExists
	}

	resource := booking.Resource{
		ID:        in.ID,
		CreatedAt: time.Now(),
	}

	s.resources[resource.ID] = resource

	return resource, nil
}

// Users
func (s *Store) CreateUser(ctx context.Context, in booking.CreateUserInput) (booking.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.User{}, err
	}

	if _, exists := s.users[in.ID]; exists {
		return booking.User{}, booking.ErrUserExists
	}

	user := booking.User{
		ID:        in.ID,
		CreatedAt: time.Now(),
	}

	s.users[user.ID] = user

	return user, nil
}

// Slots
func (s *Store) CreateSlot(ctx context.Context, in booking.CreateSlotInput) (booking.Slot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.Slot{}, err
	}

	if _, exists := s.resources[in.ResourceID]; !exists {
		return booking.Slot{}, booking.ErrResourceNotFound
	}

	for _, sl := range s.slots {
		if sl.ResourceID == in.ResourceID && sl.StartsAt.Before(in.EndsAt) && in.StartsAt.Before(sl.EndsAt) {
			return booking.Slot{}, booking.ErrSlotOverlap
		}
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

func (s *Store) ListSlots(ctx context.Context, in booking.ListSlotsInput) ([]booking.Slot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, exists := s.resources[in.ResourceID]; !exists {
		return nil, booking.ErrResourceNotFound
	}

	result := make([]booking.Slot, 0)
	for _, slot := range s.slots {
		if slot.ResourceID == in.ResourceID {
			result = append(result, slot)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	if len(result) > in.Limit {
		result = result[:in.Limit]
	}

	return result, nil
}

// Bookings
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

func (s *Store) GetBooking(ctx context.Context, in booking.GetBookingInput) (booking.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return booking.Booking{}, err
	}

	b, exists := s.bookings[in.ID]
	if !exists {
		return booking.Booking{}, booking.ErrBookingNotFound
	}

	return b, nil
}

func (s *Store) CancelBooking(ctx context.Context, in booking.CancelBookingInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	b, exists := s.bookings[in.ID]
	if !exists {
		return booking.ErrBookingNotFound
	}

	if b.UserID != in.UserID {
		return booking.ErrBookingNotFound
	}

	b.Status = booking.StatusCancelled
	s.bookings[in.ID] = b

	return nil
}
