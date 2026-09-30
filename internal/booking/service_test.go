package booking_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/memory"
	"context"
	"errors"
	"testing"
	"time"
)

type dummyStore struct {
	booking.Store
}

// Resources
func TestCreateResourceInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input booking.CreateResourceInput
		want  error
	}{
		{
			name: "empty ID",
			input: booking.CreateResourceInput{
				ID: "",
			},
			want: booking.ErrInvalidResourceID,
		},
		{
			name: "whitespace ID",
			input: booking.CreateResourceInput{
				ID: " ",
			},
			want: booking.ErrInvalidResourceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(&dummyStore{})

			_, err := service.CreateResource(
				ctx,
				tt.input,
			)

			if !errors.Is(err, tt.want) {
				t.Errorf("CreateResource() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateResourceValidInput(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	in := booking.CreateResourceInput{ID: "nook-1"}
	got, err := service.CreateResource(ctx, in)
	if err != nil {
		t.Fatalf("setup: CreateResource() error = %v", err)
	}

	if got.ID != in.ID {
		t.Errorf("ID = %q, want %q", got.ID, in.ID)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreateResourceDuplicate(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	in := booking.CreateResourceInput{ID: "nook-1"}
	if _, err := service.CreateResource(ctx, in); err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}

	_, err := service.CreateResource(ctx, in)
	if !errors.Is(err, booking.ErrResourceExists) {
		t.Fatalf("CreateResource() error = %v, want %v",
			err,
			booking.ErrResourceExists,
		)
	}
}

// Users
func TestCreateUserInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input booking.CreateUserInput
		want  error
	}{
		{
			name: "empty ID",
			input: booking.CreateUserInput{
				ID: "",
			},
			want: booking.ErrInvalidUserID,
		},
		{
			name: "whitespace ID",
			input: booking.CreateUserInput{
				ID: " ",
			},
			want: booking.ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := booking.NewService(&dummyStore{})

			_, err := service.CreateUser(
				context.Background(),
				tt.input,
			)

			if !errors.Is(err, tt.want) {
				t.Errorf("CreateUser() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateUserValidInput(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	in := booking.CreateUserInput{ID: "tim"}
	got, err := service.CreateUser(ctx, in)
	if err != nil {
		t.Fatalf("setup: CreateUser() error = %v", err)
	}

	if got.ID != in.ID {
		t.Errorf("ID = %q, want %q", got.ID, in.ID)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreateUserDuplicate(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	in := booking.CreateUserInput{ID: "tim"}
	if _, err := service.CreateUser(ctx, in); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	_, err := service.CreateUser(ctx, in)
	if !errors.Is(err, booking.ErrUserExists) {
		t.Fatalf("CreateUser() error = %v, want %v",
			err,
			booking.ErrUserExists,
		)
	}
}

// Slots
func TestCreateSlotValidInput(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	resource, err := service.CreateResource(
		ctx,
		booking.CreateResourceInput{ID: "nook-1"},
	)
	if err != nil {
		t.Fatalf("setup: CreateResource() error = %v", err)
	}

	in := booking.CreateSlotInput{
		ResourceID: resource.ID,
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	}
	got, err := service.CreateSlot(ctx, in)
	if err != nil {
		t.Fatalf("setup: CreateSlot() error = %v", err)
	}

	if got.ID <= 0 {
		t.Errorf("ID %d is not positive", got.ID)
	}
	if got.ResourceID != resource.ID {
		t.Errorf("ID = %q, want %q", got.ResourceID, resource.ID)
	}
	if !got.StartsAt.Equal(in.StartsAt) {
		t.Errorf("StartsAt = %v, want %v", got.StartsAt, in.StartsAt)
	}
	if !got.EndsAt.Equal(in.EndsAt) {
		t.Errorf("EndsAt = %v, want %v", got.EndsAt, in.EndsAt)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreateSlotInvalidInput(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input booking.CreateSlotInput
		want  error
	}{
		{
			name: "empty ResourceID",
			input: booking.CreateSlotInput{
				ResourceID: "",
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			},
			want: booking.ErrInvalidResourceID,
		},
		{
			name: "whitespace ResourceID",
			input: booking.CreateSlotInput{
				ResourceID: " ",
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			},
			want: booking.ErrInvalidResourceID,
		},
		{
			name: "EndsAt before StartsAt",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now.Add(time.Hour),
				EndsAt:     now,
			},
			want: booking.ErrInvalidTimeRange,
		},
		{
			name: "equal StartsAt and EndsAt",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now,
				EndsAt:     now,
			},
			want: booking.ErrInvalidTimeRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(&dummyStore{})

			_, err := service.CreateSlot(ctx, tt.input)

			if !errors.Is(err, tt.want) {
				t.Errorf("CreateSlot() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateSlotMissingResource(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	resourceIn := booking.CreateResourceInput{
		ID: "nook-1",
	}
	if _, err := service.CreateResource(ctx, resourceIn); err != nil {
		t.Fatalf("setup: CreateResource() error = %v", err)
	}

	in := booking.CreateSlotInput{
		ResourceID: "nook-2",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	}

	_, err := service.CreateSlot(ctx, in)
	if !errors.Is(err, booking.ErrResourceNotFound) {
		t.Fatalf("CreateSlot() error = %v, want %v",
			err,
			booking.ErrResourceNotFound,
		)
	}
}

func TestCreateSlotOverlap(t *testing.T) {
	start := time.Now()
	tests := []struct {
		name  string
		input booking.CreateSlotInput
		want  error
	}{
		{
			name: "identical times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start,
				EndsAt:     start.Add(60 * time.Minute),
			},
			want: booking.ErrSlotOverlap,
		},
		{
			name: "overlaps start",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(-30 * time.Minute),
				EndsAt:     start.Add(30 * time.Minute),
			},
			want: booking.ErrSlotOverlap,
		},
		{
			name: "overlaps end",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(30 * time.Minute),
				EndsAt:     start.Add(90 * time.Minute),
			},
			want: booking.ErrSlotOverlap,
		},
		{
			name: "contained",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(15 * time.Minute),
				EndsAt:     start.Add(45 * time.Minute),
			},
			want: booking.ErrSlotOverlap,
		},
		{
			name: "contains",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(-60 * time.Minute),
				EndsAt:     start.Add(120 * time.Minute),
			},
			want: booking.ErrSlotOverlap,
		},
		{
			name: "adjacent before",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(-60 * time.Minute),
				EndsAt:     start,
			},
			want: nil,
		},
		{
			name: "adjacent after",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(60 * time.Minute),
				EndsAt:     start.Add(120 * time.Minute),
			},
			want: nil,
		},
		{
			name: "different times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(120 * time.Minute),
				EndsAt:     start.Add(180 * time.Minute),
			},
			want: nil,
		},
		{
			name: "different resource",
			input: booking.CreateSlotInput{
				ResourceID: "nook-2",
				StartsAt:   start,
				EndsAt:     start.Add(60 * time.Minute),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())

			for _, id := range []string{"nook-1", "nook-2"} {
				if _, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: id}); err != nil {
					t.Fatalf("setup: CreateResource(): %v", err)
				}
			}
			if _, err := service.CreateSlot(ctx, booking.CreateSlotInput{
				ResourceID: "nook-1", StartsAt: start, EndsAt: start.Add(time.Hour),
			}); err != nil {
				t.Fatalf("setup: CreateSlot(): %v", err)
			}

			_, err := service.CreateSlot(ctx, tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("CreateSlot() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateSlotOverlapDoesNotStoreSlot(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	start := time.Date(2026, time.October, 1, 10, 0, 0, 0, time.UTC)
	if _, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"}); err != nil {
		t.Fatalf("setup: CreateResource(): %v", err)
	}
	original, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1", StartsAt: start, EndsAt: start.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}
	_, err = service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1", StartsAt: start.Add(30 * time.Minute), EndsAt: start.Add(90 * time.Minute),
	})
	if !errors.Is(err, booking.ErrSlotOverlap) {
		t.Fatalf("CreateSlot() error = %v, want %v", err, booking.ErrSlotOverlap)
	}
	got, err := service.ListSlots(ctx, booking.ListSlotsInput{ResourceID: "nook-1"})
	if err != nil {
		t.Fatalf("ListSlots(): %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("length = %d, want 1", len(got))
	}
	if got[0] != original {
		t.Errorf("slot = %+v, want %+v", got[0], original)
	}
}

func TestListSlotsValidInput(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	now := time.Now()

	for _, id := range []string{"nook-1", "nook-2"} {
		if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
			ID: id,
		}); err != nil {
			t.Fatalf("setup: CreateResource(): %v", err)
		}
	}

	first, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	if _, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-2",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	}); err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	second, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now.Add(time.Hour),
		EndsAt:     now.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	got, err := service.ListSlots(ctx, booking.ListSlotsInput{
		ResourceID: " nook-1 ",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("ListSlots(): %v", err)
	}

	want := []booking.Slot{first, second}

	if len(got) != len(want) {
		t.Fatalf("length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("slot[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestListSlotsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input booking.ListSlotsInput
		want  error
	}{
		{
			name: "empty resource ID",
			input: booking.ListSlotsInput{
				ResourceID: "",
				Limit:      2,
			},
			want: booking.ErrInvalidResourceID,
		},
		{
			name: "whitespace resource ID",
			input: booking.ListSlotsInput{
				ResourceID: " ",
				Limit:      2,
			},
			want: booking.ErrInvalidResourceID,
		},
		{
			name: "negative limit",
			input: booking.ListSlotsInput{
				ResourceID: "nook-1",
				Limit:      -1,
			},
			want: booking.ErrInvalidLimit,
		},
		{
			name: "limit above maximum",
			input: booking.ListSlotsInput{
				ResourceID: "nook-1",
				Limit:      booking.MaxSlotLimit + 1,
			},
			want: booking.ErrInvalidLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := booking.NewService(&dummyStore{})

			_, err := service.ListSlots(context.Background(), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("ListSlots() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestListSlotsLimit(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{
			name:      "default limit",
			limit:     0,
			wantCount: booking.DefaultSlotLimit,
		},
		{
			name:      "explicit limit",
			limit:     2,
			wantCount: 2,
		},
		{
			name:      "limit exceeds available slots",
			limit:     booking.MaxSlotLimit,
			wantCount: booking.DefaultSlotLimit + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())
			now := time.Now()

			if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
				ID: "nook-1",
			}); err != nil {
				t.Fatalf("setup: CreateResource(): %v", err)
			}

			for i := 0; i < booking.DefaultSlotLimit+1; i++ {
				_, err := service.CreateSlot(ctx, booking.CreateSlotInput{
					ResourceID: "nook-1",
					StartsAt:   now.Add(time.Duration(i) * time.Hour),
					EndsAt:     now.Add(time.Duration(i+1) * time.Hour),
				})
				if err != nil {
					t.Fatalf("setup: CreateSlot(): %v", err)
				}
			}

			got, err := service.ListSlots(ctx, booking.ListSlotsInput{
				ResourceID: "nook-1",
				Limit:      tt.limit,
			})
			if err != nil {
				t.Fatalf("ListSlots(): %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("length = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestListSlotsEmptyResult(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{
			name:       "existing resource without slots",
			resourceID: "nook-1",
		},
		{
			name:       "unknown resource",
			resourceID: "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())

			if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
				ID: "nook-1",
			}); err != nil {
				t.Fatalf("setup: CreateResource(): %v", err)
			}

			got, err := service.ListSlots(ctx, booking.ListSlotsInput{
				ResourceID: tt.resourceID,
				Limit:      2,
			})
			if err != nil {
				t.Fatalf("ListSlots(): %v", err)
			}
			if got == nil {
				t.Fatal("ListSlots() returned nil, want an empty slice")
			}
			if len(got) != 0 {
				t.Errorf("length = %d, want 0", len(got))
			}
		})
	}
}

// Bookings
func TestCreateBookingValidInput(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	resource, err := service.CreateResource(
		ctx,
		booking.CreateResourceInput{ID: "nook-1"},
	)
	if err != nil {
		t.Fatalf("setup: CreateResource() error = %v", err)
	}

	user, err := service.CreateUser(
		ctx,
		booking.CreateUserInput{ID: "tim"},
	)
	if err != nil {
		t.Fatalf("setup: CreateUser() error = %v", err)
	}

	slot, err := service.CreateSlot(
		ctx,
		booking.CreateSlotInput{
			ResourceID: resource.ID,
			StartsAt:   now,
			EndsAt:     now.Add(time.Hour),
		},
	)
	if err != nil {
		t.Fatalf("setup: CreateSlot() error = %v", err)
	}

	got, err := service.CreateBooking(
		ctx,
		booking.CreateBookingInput{
			SlotID: slot.ID,
			UserID: user.ID,
		},
	)
	if err != nil {
		t.Fatalf("setup: CreateBooking() error = %v", err)
	}

	if got.ID <= 0 {
		t.Errorf("Id = %d, want a positive ID", got.ID)
	}
	if got.SlotID != slot.ID {
		t.Errorf("SlotID = %d, want %d", got.SlotID, slot.ID)
	}
	if got.UserID != user.ID {
		t.Errorf("UserID = %q, want %q", got.UserID, user.ID)
	}
	if got.Status != booking.StatusActive {
		t.Errorf("Status = %q, want %q", got.Status, booking.StatusActive)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreateBookingInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input booking.CreateBookingInput
		want  error
	}{
		{
			name: "zero SlotID",
			input: booking.CreateBookingInput{
				SlotID: 0,
				UserID: "tim",
			},
			want: booking.ErrInvalidSlotID,
		},
		{
			name: "negative SlotID",
			input: booking.CreateBookingInput{
				SlotID: -1,
				UserID: "tim",
			},
			want: booking.ErrInvalidSlotID,
		},
		{
			name: "empty UserID",
			input: booking.CreateBookingInput{
				SlotID: 1,
				UserID: "",
			},
			want: booking.ErrInvalidUserID,
		},
		{
			name: "whitespace UserID",
			input: booking.CreateBookingInput{
				SlotID: 1,
				UserID: " ",
			},
			want: booking.ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(&dummyStore{})

			_, err := service.CreateBooking(ctx, tt.input)

			if !errors.Is(err, tt.want) {
				t.Errorf("CreateBooking() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateBookingMissingUser(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	resource, err := service.CreateResource(
		ctx,
		booking.CreateResourceInput{ID: "nook-1"},
	)
	if err != nil {
		t.Fatalf("setup: CreateResource() error = %v", err)
	}

	_, err = service.CreateUser(
		ctx,
		booking.CreateUserInput{ID: "tim"},
	)
	if err != nil {
		t.Fatalf("setup: CreateUser() error = %v", err)
	}

	slot, err := service.CreateSlot(
		ctx,
		booking.CreateSlotInput{
			ResourceID: resource.ID,
			StartsAt:   now,
			EndsAt:     now.Add(time.Hour),
		},
	)
	if err != nil {
		t.Fatalf("setup: CreateSlot() error = %v", err)
	}

	_, err = service.CreateBooking(ctx, booking.CreateBookingInput{
		SlotID: slot.ID,
		UserID: "helen",
	})
	if !errors.Is(err, booking.ErrUserNotFound) {
		t.Fatalf("CreateBooking() error = %v, want %v",
			err,
			booking.ErrUserNotFound,
		)
	}
}

func TestCreateBookingNonExistentSlot(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	resourceIn := booking.CreateResourceInput{
		ID: "nook-1",
	}
	if _, err := service.CreateResource(ctx, resourceIn); err != nil {
		t.Fatalf("setup: CreateResource() error = %v", err)
	}

	userIn := booking.CreateUserInput{
		ID: "tim",
	}
	if _, err := service.CreateUser(ctx, userIn); err != nil {
		t.Fatalf("setup: CreateUser() error = %v", err)
	}

	slotIn := booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   time.Now(),
		EndsAt:     time.Now().Add(time.Hour),
	}
	if _, err := service.CreateSlot(ctx, slotIn); err != nil {
		t.Fatalf("setup: CreateSlot() error = %v", err)
	}

	in := booking.CreateBookingInput{
		SlotID: 2,
		UserID: "tim",
	}

	_, err := service.CreateBooking(ctx, in)
	if !errors.Is(err, booking.ErrSlotNotFound) {
		t.Fatalf("CreateBooking() error = %v, want %v",
			err,
			booking.ErrSlotNotFound,
		)
	}
}

func TestCreateBookingAlreadyBookedSlot(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	resourceIn := booking.CreateResourceInput{
		ID: "nook-1",
	}
	if _, err := service.CreateResource(ctx, resourceIn); err != nil {
		t.Fatalf("setup: CreateResource() error = %v", err)
	}

	userIn := booking.CreateUserInput{
		ID: "tim",
	}
	if _, err := service.CreateUser(ctx, userIn); err != nil {
		t.Fatalf("setup: CreateUser() error = %v", err)
	}

	slotIn := booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	}
	if _, err := service.CreateSlot(ctx, slotIn); err != nil {
		t.Fatalf("setup: CreateSlot() error = %v", err)
	}

	in := booking.CreateBookingInput{
		SlotID: 1,
		UserID: "tim",
	}
	if _, err := service.CreateBooking(ctx, in); err != nil {
		t.Fatalf("setup: CreateBooking() error = %v", err)
	}

	in = booking.CreateBookingInput{
		SlotID: 1,
		UserID: "tim",
	}

	_, err := service.CreateBooking(ctx, in)
	if !errors.Is(err, booking.ErrSlotAlreadyBooked) {
		t.Fatalf("CreateBooking() error = %v, want %v",
			err,
			booking.ErrSlotAlreadyBooked,
		)
	}
}

func TestGetBookingValidInput(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	now := time.Now()

	if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
		ID: "nook-1",
	}); err != nil {
		t.Fatalf("setup: CreateResource(): %v", err)
	}

	if _, err := service.CreateUser(ctx, booking.CreateUserInput{
		ID: "tim",
	}); err != nil {
		t.Fatalf("setup: CreateUser(): %v", err)
	}

	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	want, err := service.CreateBooking(ctx, booking.CreateBookingInput{
		SlotID: slot.ID,
		UserID: "tim",
	})
	if err != nil {
		t.Fatalf("setup: CreateBooking(): %v", err)
	}

	got, err := service.GetBooking(ctx, booking.GetBookingInput{
		ID:     want.ID,
		UserID: " tim ",
	})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}
	if got != want {
		t.Errorf("GetBooking() = %+v, want %+v", got, want)
	}
}

func TestGetBookingInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input booking.GetBookingInput
		want  error
	}{
		{
			name:  "zero ID",
			input: booking.GetBookingInput{ID: 0, UserID: "tim"},
			want:  booking.ErrInvalidBookingID,
		},
		{
			name:  "negative ID",
			input: booking.GetBookingInput{ID: -1, UserID: "tim"},
			want:  booking.ErrInvalidBookingID,
		},
		{
			name:  "empty user ID",
			input: booking.GetBookingInput{ID: 1, UserID: ""},
			want:  booking.ErrInvalidUserID,
		},
		{
			name:  "whitespace user ID",
			input: booking.GetBookingInput{ID: 1, UserID: " "},
			want:  booking.ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := booking.NewService(&dummyStore{})

			_, err := service.GetBooking(context.Background(), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("GetBooking() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestGetBookingMissing(t *testing.T) {
	tests := []struct {
		name    string
		missing bool
		userID  string
	}{
		{
			name:    "missing booking",
			missing: true,
			userID:  "tim",
		},
		{
			name:   "different user",
			userID: "helen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())
			now := time.Now()

			if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
				ID: "nook-1",
			}); err != nil {
				t.Fatalf("setup: CreateResource(): %v", err)
			}

			for _, id := range []string{"tim", "helen"} {
				if _, err := service.CreateUser(ctx, booking.CreateUserInput{
					ID: id,
				}); err != nil {
					t.Fatalf("setup: CreateUser(): %v", err)
				}
			}

			slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			})
			if err != nil {
				t.Fatalf("setup: CreateSlot(): %v", err)
			}

			created, err := service.CreateBooking(ctx, booking.CreateBookingInput{
				SlotID: slot.ID,
				UserID: "tim",
			})
			if err != nil {
				t.Fatalf("setup: CreateBooking(): %v", err)
			}

			id := created.ID
			if tt.missing {
				id++
			}

			_, err = service.GetBooking(ctx, booking.GetBookingInput{
				ID:     id,
				UserID: tt.userID,
			})
			if !errors.Is(err, booking.ErrBookingNotFound) {
				t.Fatalf("GetBooking() error = %v, want %v",
					err, booking.ErrBookingNotFound)
			}
		})
	}
}

func TestCancelBookingInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input booking.CancelBookingInput
		want  error
	}{
		{
			name:  "zero ID",
			input: booking.CancelBookingInput{ID: 0, UserID: "tim"},
			want:  booking.ErrInvalidBookingID,
		},
		{
			name:  "negative ID",
			input: booking.CancelBookingInput{ID: -1, UserID: "tim"},
			want:  booking.ErrInvalidBookingID,
		},
		{
			name:  "empty user ID",
			input: booking.CancelBookingInput{ID: 1, UserID: ""},
			want:  booking.ErrInvalidUserID,
		},
		{
			name:  "whitespace user ID",
			input: booking.CancelBookingInput{ID: 1, UserID: " "},
			want:  booking.ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := booking.NewService(&dummyStore{})

			err := service.CancelBooking(context.Background(), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("CancelBooking() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCancelBookingValidInput(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	now := time.Now()

	if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
		ID: "nook-1",
	}); err != nil {
		t.Fatalf("setup: CreateResource(): %v", err)
	}

	if _, err := service.CreateUser(ctx, booking.CreateUserInput{
		ID: "tim",
	}); err != nil {
		t.Fatalf("setup: CreateUser(): %v", err)
	}

	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	created, err := service.CreateBooking(ctx, booking.CreateBookingInput{
		SlotID: slot.ID,
		UserID: "tim",
	})
	if err != nil {
		t.Fatalf("setup: CreateBooking(): %v", err)
	}

	err = service.CancelBooking(ctx, booking.CancelBookingInput{
		ID:     created.ID,
		UserID: " tim ",
	})
	if err != nil {
		t.Fatalf("CancelBooking(): %v", err)
	}

	got, err := service.GetBooking(ctx, booking.GetBookingInput{
		ID:     created.ID,
		UserID: "tim",
	})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	want := created
	want.Status = booking.StatusCancelled

	if got != want {
		t.Errorf("booking = %+v, want %+v", got, want)
	}
}

func TestCancelBookingAlreadyCancelled(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	now := time.Now()

	if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
		ID: "nook-1",
	}); err != nil {
		t.Fatalf("setup: CreateResource(): %v", err)
	}

	if _, err := service.CreateUser(ctx, booking.CreateUserInput{
		ID: "tim",
	}); err != nil {
		t.Fatalf("setup: CreateUser(): %v", err)
	}

	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	created, err := service.CreateBooking(ctx, booking.CreateBookingInput{
		SlotID: slot.ID,
		UserID: "tim",
	})
	if err != nil {
		t.Fatalf("setup: CreateBooking(): %v", err)
	}

	in := booking.CancelBookingInput{
		ID:     created.ID,
		UserID: "tim",
	}

	if err := service.CancelBooking(ctx, in); err != nil {
		t.Fatalf("setup: CancelBooking(): %v", err)
	}

	if err := service.CancelBooking(ctx, in); err != nil {
		t.Fatalf("repeated CancelBooking(): %v", err)
	}

	got, err := service.GetBooking(ctx, booking.GetBookingInput{
		ID:     created.ID,
		UserID: "tim",
	})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	want := created
	want.Status = booking.StatusCancelled

	if got != want {
		t.Errorf("booking = %+v, want %+v", got, want)
	}
}

func TestCancelBookingMakesSlotBookable(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	now := time.Now()

	if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
		ID: "nook-1",
	}); err != nil {
		t.Fatalf("setup: CreateResource(): %v", err)
	}

	if _, err := service.CreateUser(ctx, booking.CreateUserInput{
		ID: "tim",
	}); err != nil {
		t.Fatalf("setup: CreateUser(): %v", err)
	}

	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	created, err := service.CreateBooking(ctx, booking.CreateBookingInput{
		SlotID: slot.ID,
		UserID: "tim",
	})
	if err != nil {
		t.Fatalf("setup: CreateBooking(): %v", err)
	}

	in := booking.CancelBookingInput{
		ID:     created.ID,
		UserID: "tim",
	}

	if err := service.CancelBooking(ctx, in); err != nil {
		t.Fatalf("CancelBooking(): %v", err)
	}
	_, err = service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1", StartsAt: now, EndsAt: now.Add(time.Hour),
	})
	if !errors.Is(err, booking.ErrSlotOverlap) {
		t.Fatalf("CreateSlot() after cancellation error = %v, want %v", err, booking.ErrSlotOverlap)
	}
	replacement, err := service.CreateBooking(ctx, booking.CreateBookingInput{SlotID: slot.ID, UserID: "tim"})
	if err != nil {
		t.Fatalf("CreateBooking() for original slot: %v", err)
	}
	if replacement.Status != booking.StatusActive {
		t.Errorf("status = %q, want active", replacement.Status)
	}
	previous, err := service.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: "tim"})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}
	created.Status = booking.StatusCancelled
	if previous != created {
		t.Errorf("previous booking = %+v, want %+v", previous, created)
	}
}

func TestCancelBookingMissing(t *testing.T) {
	tests := []struct {
		name    string
		missing bool
		userID  string
	}{
		{
			name:    "missing booking",
			missing: true,
			userID:  "tim",
		},
		{
			name:   "different owner",
			userID: "helen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())
			now := time.Now()

			if _, err := service.CreateResource(ctx, booking.CreateResourceInput{
				ID: "nook-1",
			}); err != nil {
				t.Fatalf("setup: CreateResource(): %v", err)
			}

			for _, id := range []string{"tim", "helen"} {
				if _, err := service.CreateUser(ctx, booking.CreateUserInput{
					ID: id,
				}); err != nil {
					t.Fatalf("setup: CreateUser(): %v", err)
				}
			}

			slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			})
			if err != nil {
				t.Fatalf("setup: CreateSlot(): %v", err)
			}

			created, err := service.CreateBooking(ctx, booking.CreateBookingInput{
				SlotID: slot.ID,
				UserID: "tim",
			})
			if err != nil {
				t.Fatalf("setup: CreateBooking(): %v", err)
			}

			id := created.ID
			if tt.missing {
				id++
			}

			err = service.CancelBooking(ctx, booking.CancelBookingInput{
				ID:     id,
				UserID: tt.userID,
			})
			if !errors.Is(err, booking.ErrBookingNotFound) {
				t.Fatalf("CancelBooking() error = %v, want %v",
					err, booking.ErrBookingNotFound)
			}

			got, err := service.GetBooking(ctx, booking.GetBookingInput{
				ID:     created.ID,
				UserID: "tim",
			})
			if err != nil {
				t.Fatalf("GetBooking(): %v", err)
			}
			if got != created {
				t.Errorf("booking changed: got %+v, want %+v", got, created)
			}
		})
	}
}
