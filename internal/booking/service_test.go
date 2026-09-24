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
			store := &dummyStore{}
			service := booking.NewService(store)

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
			store := &dummyStore{}
			service := booking.NewService(store)

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

func TestCreateSlotNonExistentResource(t *testing.T) {
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
		StartsAt:   time.Now(),
		EndsAt:     time.Now().Add(time.Hour),
	}

	_, err := service.CreateSlot(ctx, in)
	if !errors.Is(err, booking.ErrResourceNotFound) {
		t.Fatalf("CreateSlot() error = %v, want %v",
			err,
			booking.ErrResourceNotFound,
		)
	}
}

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

func TestCreateBookingNonExistentUser(t *testing.T) {
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

func TestCreateBookingOverlap(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input booking.CreateSlotInput
		want  error
	}{
		{
			name: "same resource overlapping times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now.Add(30 * time.Minute),
				EndsAt:     now.Add(90 * time.Minute),
			},
			want: booking.ErrBookingOverlap,
		},
		{
			name: "same resource different times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now.Add(2 * time.Hour),
				EndsAt:     now.Add(3 * time.Hour),
			},
			want: nil,
		},
		{
			name: "same resource adjacent times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now.Add(time.Hour),
				EndsAt:     now.Add(2 * time.Hour),
			},
			want: nil,
		},
		{
			name: "different resource overlapping times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-2",
				StartsAt:   now.Add(30 * time.Minute),
				EndsAt:     now.Add(90 * time.Minute),
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())

			resource, err := service.CreateResource(ctx, booking.CreateResourceInput{
				ID: "nook-1",
			},
			)
			if err != nil {
				t.Fatalf("setup: CreateResource() error = %v", err)
			}

			resourceIn2 := booking.CreateResourceInput{
				ID: "nook-2",
			}
			if _, err := service.CreateResource(ctx, resourceIn2); err != nil {
				t.Fatalf("setup: CreateResource() error = %v", err)
			}

			userIn := booking.CreateUserInput{
				ID: "tim",
			}
			user, err := service.CreateUser(ctx, userIn)
			if err != nil {
				t.Fatalf("setup: CreateUser() error = %v", err)
			}

			slotIn := booking.CreateSlotInput{
				ResourceID: resource.ID,
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			}
			slot1, err := service.CreateSlot(ctx, slotIn)
			if err != nil {
				t.Fatalf("setup: CreateSlot() error = %v", err)
			}

			slot2, err := service.CreateSlot(ctx, tt.input)
			if err != nil {
				t.Fatalf("setup: CreateSlot() error = %v", err)
			}

			bookingIn := booking.CreateBookingInput{
				SlotID: slot1.ID,
				UserID: user.ID,
			}
			if _, err := service.CreateBooking(ctx, bookingIn); err != nil {
				t.Fatalf("setup: CreateBooking() error = %v", err)
			}

			bookingIn = booking.CreateBookingInput{
				SlotID: slot2.ID,
				UserID: user.ID,
			}

			_, err = service.CreateBooking(ctx, bookingIn)
			if !errors.Is(err, tt.want) {
				t.Fatalf("CreateBooking() error = %v, want %v", err, tt.want)
			}
		})
	}
}
