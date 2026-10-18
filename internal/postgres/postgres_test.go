package postgres_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/postgres"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestStore(t *testing.T) (context.Context, *postgres.Store, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated bookanook_test database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("setup: create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	var databaseName string
	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("setup: connect: %v", err)
	}

	if databaseName != "bookanook_test" {
		t.Fatalf("refusing to reset database %q; use bookanook_test", databaseName)
	}

	if _, err := pool.Exec(ctx, `
		TRUNCATE TABLE bookings, slots, users, resources RESTART IDENTITY
	`); err != nil {
		t.Fatalf("setup: reset tables (apply migrations first): %v", err)
	}

	return ctx, postgres.NewStore(pool), pool
}

func seedResources(t *testing.T, ctx context.Context, store *postgres.Store, ids ...string) {
	t.Helper()

	for _, id := range ids {
		if _, err := store.CreateResource(ctx, booking.CreateResourceInput{ID: id}); err != nil {
			t.Fatalf("setup: CreateResource(): %v", err)
		}
	}
}

func seedUsers(t *testing.T, ctx context.Context, store *postgres.Store, ids ...string) {
	t.Helper()

	for _, id := range ids {
		if _, err := store.CreateUser(ctx, booking.CreateUserInput{ID: id}); err != nil {
			t.Fatalf("setup: CreateUser(): %v", err)
		}
	}
}

func seedSlot(t *testing.T, ctx context.Context, store *postgres.Store) booking.Slot {
	t.Helper()
	seedResources(t, ctx, store, "nook-1")

	start := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)

	slot, err := store.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "nook-1", StartsAt: start, EndsAt: start.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot(): %v", err)
	}

	return slot
}

func seedBooking(t *testing.T, ctx context.Context, store *postgres.Store) booking.Booking {
	t.Helper()
	seedUsers(t, ctx, store, "tim")
	slot := seedSlot(t, ctx, store)

	created, err := store.CreateBooking(ctx, booking.CreateBookingInput{
		RequestKey: "request-1",
		SlotID:     slot.ID, UserID: "tim",
	})
	if err != nil {
		t.Fatalf("setup: CreateBooking(): %v", err)
	}

	return created
}

func assertBooking(t *testing.T, got, want booking.Booking) {
	t.Helper()

	if got.ID != want.ID || got.SlotID != want.SlotID ||
		got.UserID != want.UserID || got.Status != want.Status ||
		!got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("booking = %+v, want %+v", got, want)
	}
}

func TestCreateResourceValid(t *testing.T) {
	ctx, store, _ := newTestStore(t)

	in := booking.CreateResourceInput{ID: "nook-1"}

	got, err := store.CreateResource(ctx, in)
	if err != nil {
		t.Fatalf("CreateResource(): %v", err)
	}

	if got.ID != in.ID || got.CreatedAt.IsZero() {
		t.Errorf("resource = %+v, want ID %q and non-zero CreatedAt", got, in.ID)
	}
}

func TestCreateResourceDuplicate(t *testing.T) {
	ctx, store, _ := newTestStore(t)

	in := booking.CreateResourceInput{ID: "nook-1"}

	seedResources(t, ctx, store, in.ID)

	_, err := store.CreateResource(ctx, in)
	if !errors.Is(err, booking.ErrResourceExists) {
		t.Fatalf("CreateResource() error = %v, want %v", err, booking.ErrResourceExists)
	}
}

func TestCreateResourceSQLLikeInput(t *testing.T) {
	ctx, store, _ := newTestStore(t)

	in := booking.CreateResourceInput{ID: "nook'; DROP TABLE users; --"}

	got, err := store.CreateResource(ctx, in)
	if err != nil {
		t.Fatalf("CreateResource(): %v", err)
	}

	if got.ID != in.ID || got.CreatedAt.IsZero() {
		t.Errorf("resource = %+v, want ID %q and non-zero CreatedAt", got, in.ID)
	}

	if _, err := store.CreateUser(ctx, booking.CreateUserInput{ID: "tim"}); err != nil {
		t.Fatalf("CreateUser() after SQL-like input: %v", err)
	}
}

func TestCreateUserValid(t *testing.T) {
	ctx, store, _ := newTestStore(t)

	in := booking.CreateUserInput{ID: "tim"}

	got, err := store.CreateUser(ctx, in)
	if err != nil {
		t.Fatalf("CreateUser(): %v", err)
	}

	if got.ID != in.ID || got.CreatedAt.IsZero() {
		t.Errorf("user = %+v, want ID %q and non-zero CreatedAt", got, in.ID)
	}
}

func TestCreateUserDuplicate(t *testing.T) {
	ctx, store, _ := newTestStore(t)

	in := booking.CreateUserInput{ID: "tim"}

	if _, err := store.CreateUser(ctx, in); err != nil {
		t.Fatalf("setup: CreateUser(): %v", err)
	}

	_, err := store.CreateUser(ctx, in)
	if !errors.Is(err, booking.ErrUserExists) {
		t.Fatalf("CreateUser() error = %v, want %v", err, booking.ErrUserExists)
	}
}

func TestCreateSlotOverlap(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	original := seedSlot(t, ctx, store)

	_, err := store.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: original.ResourceID,
		StartsAt:   original.StartsAt.Add(30 * time.Minute),
		EndsAt:     original.EndsAt.Add(30 * time.Minute),
	})
	if !errors.Is(err, booking.ErrSlotOverlap) {
		t.Fatalf("CreateSlot() error = %v, want %v", err, booking.ErrSlotOverlap)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM slots").Scan(&count); err != nil {
		t.Fatalf("count slots: %v", err)
	}

	if count != 1 {
		t.Errorf("stored slots = %d, want 1", count)
	}
}

func TestCreateSlotNonOverlapping(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		start, end time.Duration
	}{
		{
			name:       "adjacent before",
			resourceID: "nook-1",
			start:      -time.Hour,
			end:        0,
		},
		{
			name:       "adjacent after",
			resourceID: "nook-1",
			start:      time.Hour,
			end:        2 * time.Hour,
		},
		{
			name:       "different resource",
			resourceID: "nook-2",
			start:      0,
			end:        time.Hour,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, store, pool := newTestStore(t)
			original := seedSlot(t, ctx, store)

			seedResources(t, ctx, store, "nook-2")

			in := booking.CreateSlotInput{
				ResourceID: tt.resourceID,
				StartsAt:   original.StartsAt.Add(tt.start),
				EndsAt:     original.StartsAt.Add(tt.end),
			}

			got, err := store.CreateSlot(ctx, in)
			if err != nil {
				t.Fatalf("CreateSlot(): %v", err)
			}

			if got.ID <= 0 || got.ResourceID != in.ResourceID || got.CreatedAt.IsZero() ||
				!got.StartsAt.Equal(in.StartsAt) || !got.EndsAt.Equal(in.EndsAt) {
				t.Errorf("slot = %+v, want stored input %+v with ID and CreatedAt", got, in)
			}

			var count int
			if err := pool.QueryRow(ctx, "SELECT count(*) FROM slots").Scan(&count); err != nil {
				t.Fatalf("count slots: %v", err)
			}

			if count != 2 {
				t.Errorf("stored slots = %d, want 2", count)
			}
		})
	}
}

func TestCreateSlotMissingResource(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	original := seedSlot(t, ctx, store)

	_, err := store.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: "missing",
		StartsAt:   original.EndsAt,
		EndsAt:     original.EndsAt.Add(time.Hour),
	})
	if !errors.Is(err, booking.ErrResourceNotFound) {
		t.Fatalf("CreateSlot() error = %v, want %v", err, booking.ErrResourceNotFound)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM slots").Scan(&count); err != nil {
		t.Fatalf("count slots: %v", err)
	}

	if count != 1 {
		t.Errorf("stored slots = %d, want 1", count)
	}
}

func TestCreateSlotInvalidTimeRange(t *testing.T) {
	tests := []struct {
		name       string
		start, end time.Duration
	}{
		{
			name:  "equal endpoints",
			start: time.Hour,
			end:   time.Hour,
		},
		{
			name:  "reversed endpoints",
			start: 2 * time.Hour,
			end:   time.Hour,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, store, pool := newTestStore(t)
			original := seedSlot(t, ctx, store)

			_, err := store.CreateSlot(ctx, booking.CreateSlotInput{
				ResourceID: original.ResourceID,
				StartsAt:   original.StartsAt.Add(tt.start),
				EndsAt:     original.StartsAt.Add(tt.end),
			})
			if !errors.Is(err, booking.ErrInvalidTimeRange) {
				t.Fatalf("CreateSlot() error = %v, want %v", err, booking.ErrInvalidTimeRange)
			}

			var count int
			if err := pool.QueryRow(ctx, "SELECT count(*) FROM slots").Scan(&count); err != nil {
				t.Fatalf("count slots: %v", err)
			}

			if count != 1 {
				t.Errorf("stored slots = %d, want 1", count)
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
			name:      "explicit limit",
			limit:     2,
			wantCount: 2,
		},
		{
			name:      "limit exceeds available slots",
			limit:     10,
			wantCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, store, _ := newTestStore(t)
			first := seedSlot(t, ctx, store)

			want := []booking.Slot{first}

			seedResources(t, ctx, store, "nook-2")

			for _, offset := range []time.Duration{2 * time.Hour, time.Hour} {
				slot, err := store.CreateSlot(ctx, booking.CreateSlotInput{
					ResourceID: first.ResourceID,
					StartsAt:   first.StartsAt.Add(offset),
					EndsAt:     first.EndsAt.Add(offset),
				})
				if err != nil {
					t.Fatalf("setup: CreateSlot(): %v", err)
				}
				want = append(want, slot)
			}

			if _, err := store.CreateSlot(ctx, booking.CreateSlotInput{
				ResourceID: "nook-2",
				StartsAt:   first.StartsAt,
				EndsAt:     first.EndsAt,
			}); err != nil {
				t.Fatalf("setup: other resource slot: %v", err)
			}

			got, err := store.ListSlots(ctx, booking.ListSlotsInput{
				ResourceID: first.ResourceID,
				Limit:      tt.limit,
			})
			if err != nil {
				t.Fatalf("ListSlots(): %v", err)
			}

			if got == nil || len(got) != tt.wantCount {
				t.Fatalf("slots = %v, want non-nil slice of length %d", got, tt.wantCount)
			}

			for i, slot := range got {
				if slot.ID != want[i].ID || slot.ResourceID != want[i].ResourceID ||
					!slot.StartsAt.Equal(want[i].StartsAt) || !slot.EndsAt.Equal(want[i].EndsAt) ||
					!slot.CreatedAt.Equal(want[i].CreatedAt) {
					t.Errorf("slot[%d] = %+v, want %+v", i, slot, want[i])
				}
			}
		})
	}
}

func TestListSlotsEmptyResource(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	seedSlot(t, ctx, store)

	seedResources(t, ctx, store, "empty")

	got, err := store.ListSlots(ctx, booking.ListSlotsInput{ResourceID: "empty", Limit: 2})
	if err != nil {
		t.Fatalf("ListSlots(): %v", err)
	}

	if got == nil || len(got) != 0 {
		t.Errorf("slots = %v, want a non-nil empty slice", got)
	}
}

func TestListSlotsMissingResource(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	seedSlot(t, ctx, store)

	got, err := store.ListSlots(ctx, booking.ListSlotsInput{ResourceID: "missing", Limit: 2})
	if !errors.Is(err, booking.ErrResourceNotFound) {
		t.Fatalf("ListSlots() error = %v, want %v", err, booking.ErrResourceNotFound)
	}

	if got != nil {
		t.Errorf("ListSlots() = %v, want nil on error", got)
	}
}

func TestCreateBookingValid(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim")
	slot := seedSlot(t, ctx, store)

	in := booking.CreateBookingInput{RequestKey: "request-2", SlotID: slot.ID, UserID: "tim"}

	got, err := store.CreateBooking(ctx, in)
	if err != nil {
		t.Fatalf("CreateBooking(): %v", err)
	}

	if got.ID <= 0 || got.SlotID != in.SlotID || got.UserID != in.UserID ||
		got.Status != booking.StatusActive || got.CreatedAt.IsZero() {
		t.Fatalf("booking = %+v, want an active booking for input %+v", got, in)
	}

	found, err := store.GetBooking(ctx, booking.GetBookingInput{ID: got.ID, UserID: got.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	assertBooking(t, found, got)

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count); err != nil {
		t.Fatalf("count bookings: %v", err)
	}

	if count != 1 {
		t.Errorf("stored bookings = %d, want 1", count)
	}
}

func TestCreateBookingMissingSlot(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim")
	slot := seedSlot(t, ctx, store)

	_, err := store.CreateBooking(ctx, booking.CreateBookingInput{RequestKey: "request-3", SlotID: slot.ID + 1, UserID: "tim"})
	if !errors.Is(err, booking.ErrSlotNotFound) {
		t.Fatalf("CreateBooking() error = %v, want %v", err, booking.ErrSlotNotFound)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count); err != nil {
		t.Fatalf("count bookings: %v", err)
	}

	if count != 0 {
		t.Errorf("stored bookings = %d, want 0", count)
	}
}

func TestCreateBookingMissingUser(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim")
	slot := seedSlot(t, ctx, store)

	_, err := store.CreateBooking(ctx, booking.CreateBookingInput{RequestKey: "request-4", SlotID: slot.ID, UserID: "missing"})
	if !errors.Is(err, booking.ErrUserNotFound) {
		t.Fatalf("CreateBooking() error = %v, want %v", err, booking.ErrUserNotFound)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count); err != nil {
		t.Fatalf("count bookings: %v", err)
	}

	if count != 0 {
		t.Errorf("stored bookings = %d, want 0", count)
	}
}

func TestCreateBookingAlreadyBooked(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	created := seedBooking(t, ctx, store)
	seedUsers(t, ctx, store, "helen")

	_, err := store.CreateBooking(ctx, booking.CreateBookingInput{RequestKey: "request-5", SlotID: created.SlotID, UserID: "helen"})
	if !errors.Is(err, booking.ErrSlotAlreadyBooked) {
		t.Fatalf("CreateBooking() error = %v, want %v", err, booking.ErrSlotAlreadyBooked)
	}

	found, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	assertBooking(t, found, created)

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count); err != nil {
		t.Fatalf("count bookings: %v", err)
	}

	if count != 1 {
		t.Errorf("stored bookings = %d, want 1", count)
	}
}

func TestGetBookingValidInput(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)

	got, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	assertBooking(t, got, created)
}

func TestGetBookingCancelled(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)

	if err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID, UserID: created.UserID}); err != nil {
		t.Fatalf("setup: CancelBooking(): %v", err)
	}

	got, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	created.Status = booking.StatusCancelled

	assertBooking(t, got, created)
}

func TestGetBookingMissing(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)

	_, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID + 1, UserID: created.UserID})
	if !errors.Is(err, booking.ErrBookingNotFound) {
		t.Fatalf("GetBooking() error = %v, want %v", err, booking.ErrBookingNotFound)
	}
}

func TestGetBookingWrongOwner(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)
	seedUsers(t, ctx, store, "helen")

	_, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: "helen"})
	if !errors.Is(err, booking.ErrBookingNotFound) {
		t.Fatalf("GetBooking() error = %v, want %v", err, booking.ErrBookingNotFound)
	}
}

func TestGetBookingMissingUser(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)

	_, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: "missing"})
	if !errors.Is(err, booking.ErrBookingNotFound) {
		t.Fatalf("GetBooking() error = %v, want %v", err, booking.ErrBookingNotFound)
	}
}

func TestCancelBookingValid(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)

	err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("CancelBooking(): %v", err)
	}

	got, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	created.Status = booking.StatusCancelled

	assertBooking(t, got, created)
}

func TestCancelBookingAlreadyCancelled(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)

	in := booking.CancelBookingInput{ID: created.ID, UserID: created.UserID}
	if err := store.CancelBooking(ctx, in); err != nil {
		t.Fatalf("setup: CancelBooking(): %v", err)
	}

	if err := store.CancelBooking(ctx, in); err != nil {
		t.Fatalf("repeated CancelBooking(): %v", err)
	}

	got, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	created.Status = booking.StatusCancelled

	assertBooking(t, got, created)
}

func TestCancelBookingMakesSlotBookable(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)
	seedUsers(t, ctx, store, "helen")

	if err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID, UserID: created.UserID}); err != nil {
		t.Fatalf("CancelBooking(): %v", err)
	}

	replacement, err := store.CreateBooking(ctx, booking.CreateBookingInput{RequestKey: "request-6", SlotID: created.SlotID, UserID: "helen"})
	if err != nil {
		t.Fatalf("CreateBooking() after cancellation: %v", err)
	}

	if replacement.ID == created.ID || replacement.Status != booking.StatusActive {
		t.Errorf("replacement = %+v, want a new active booking", replacement)
	}

	previous, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking() after rebooking: %v", err)
	}

	created.Status = booking.StatusCancelled

	assertBooking(t, previous, created)
}

func TestCancelBookingMissing(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)

	err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID + 1, UserID: created.UserID})
	if !errors.Is(err, booking.ErrBookingNotFound) {
		t.Fatalf("CancelBooking() error = %v, want %v", err, booking.ErrBookingNotFound)
	}

	got, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	assertBooking(t, got, created)
}

func TestCancelBookingWrongOwner(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)
	seedUsers(t, ctx, store, "helen")

	err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID, UserID: "helen"})
	if !errors.Is(err, booking.ErrBookingNotFound) {
		t.Fatalf("CancelBooking() error = %v, want %v", err, booking.ErrBookingNotFound)
	}

	got, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}

	assertBooking(t, got, created)
}

func TestCancelBookingMissingUser(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	created := seedBooking(t, ctx, store)
	err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID, UserID: "missing"})
	if !errors.Is(err, booking.ErrBookingNotFound) {
		t.Fatalf("CancelBooking() error = %v, want %v", err, booking.ErrBookingNotFound)
	}
	got, err := store.GetBooking(ctx, booking.GetBookingInput{ID: created.ID, UserID: created.UserID})
	if err != nil {
		t.Fatalf("GetBooking(): %v", err)
	}
	assertBooking(t, got, created)
}

func TestCreateBookingConcurrent(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim")

	slot := seedSlot(t, ctx, store)

	const requests = 20

	start := make(chan struct{})

	results := make(chan error, requests)
	for index := range requests {
		go func() {
			<-start
			_, err := store.CreateBooking(ctx, booking.CreateBookingInput{RequestKey: fmt.Sprintf("request-%d", index), SlotID: slot.ID, UserID: "tim"})
			results <- err
		}()
	}
	close(start)

	successes := 0

	for range requests {
		err := <-results
		if err == nil {
			successes++
		} else if !errors.Is(err, booking.ErrSlotAlreadyBooked) {
			t.Errorf("CreateBooking() error = %v, want conflict", err)
		}
	}

	if successes != 1 {
		t.Errorf("successful bookings = %d, want 1", successes)
	}

	var count int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM bookings WHERE slot_id = $1 AND status = 'active'
	`, slot.ID).Scan(&count); err != nil {
		t.Fatalf("count active bookings: %v", err)
	}

	if count != 1 {
		t.Errorf("active bookings = %d, want 1", count)
	}
}
