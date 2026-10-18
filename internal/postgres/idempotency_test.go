package postgres_test

import (
	"book-a-nook/internal/booking"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestCreateBookingPersistsRequestMetadata(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim", "helen")
	first := seedSlot(t, ctx, store)

	second, err := store.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: first.ResourceID, StartsAt: first.EndsAt, EndsAt: first.EndsAt.Add(time.Hour),
	})

	if err != nil {
		t.Fatal(err)
	}

	in := booking.CreateBookingInput{SlotID: first.ID, UserID: "tim", RequestKey: " Request-A "}
	created, err := store.CreateBooking(ctx, in)
	if err != nil {
		t.Fatal(err)
	}

	var key, hash string
	if err := pool.QueryRow(ctx, "SELECT request_key, request_hash FROM bookings WHERE id = $1", created.ID).Scan(&key, &hash); err != nil {
		t.Fatal(err)
	}

	if key != in.RequestKey || hash != in.RequestHash() {
		t.Fatalf("stored metadata = (%q, %q), want (%q, %q)", key, hash, in.RequestKey, in.RequestHash())
	}

	if err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID, UserID: in.UserID}); err != nil {
		t.Fatal(err)
	}

	in.SlotID = second.ID
	if _, err := store.CreateBooking(ctx, in); !errors.Is(err, booking.ErrIdempotencyConflict) {
		t.Fatalf("reusing cancelled booking's key = %v, want key conflict", err)
	}

	in.UserID = "helen"
	if _, err := store.CreateBooking(ctx, in); err != nil {
		t.Fatalf("different user reusing key: %v", err)
	}
}

func TestBookingIdempotencyMigration(t *testing.T) {
	ctx, _, pool := newTestStore(t)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanupCtx)
	})

	exec := func(sql string) {
		t.Helper()
		if _, err := tx.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}

	exec("CREATE SCHEMA booking_idempotency_migration_test")
	exec("SET LOCAL search_path TO booking_idempotency_migration_test, public")
	initialUp, _ := readMigration(t, "20261001073813_initial_schema.sql")
	up, down := readMigration(t, "20261018000000_booking_idempotency.sql")
	exec(initialUp)
	exec(`
		INSERT INTO resources (id) VALUES ('nook-1');
		INSERT INTO users (id) VALUES ('tim');
		INSERT INTO slots (resource_id, starts_at, ends_at)
		VALUES ('nook-1', '2026-11-07 12:00:00+00', '2026-11-07 13:00:00+00');
		INSERT INTO bookings (slot_id, user_id, status)
		VALUES (1, 'tim', 'cancelled'), (1, 'tim', 'active');
	`)
	exec(up)

	for id, status := range []string{"cancelled", "active"} {
		var gotStatus, key, hash string
		if err := tx.QueryRow(ctx, "SELECT status, request_key, request_hash FROM bookings WHERE id = $1", id+1).Scan(&gotStatus, &key, &hash); err != nil {
			t.Fatal(err)
		}
		wantHash := (booking.CreateBookingInput{SlotID: 1}).RequestHash()
		if gotStatus != status || key != fmt.Sprintf("__legacy_booking__:%d", id+1) || hash != wantHash {
			t.Fatalf("legacy booking %d = (%q, %q, %q)", id+1, gotStatus, key, hash)
		}
	}

	exec("SAVEPOINT missing_key")

	_, err = tx.Exec(ctx, "INSERT INTO bookings (slot_id, user_id, status) VALUES (1, 'tim', 'cancelled')")
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23502" {
		t.Fatalf("missing request metadata = %v, want NOT NULL violation", err)
	}

	exec("ROLLBACK TO SAVEPOINT missing_key")
	exec(down)
	exec("INSERT INTO bookings (slot_id, user_id, status) VALUES (1, 'tim', 'cancelled')")

	exec("TRUNCATE bookings")
	exec(up)
}

func readMigration(t *testing.T, name string) (string, string) {
	t.Helper()
	contents, err := os.ReadFile("../../migrations/" + name)
	if err != nil {
		t.Fatal(err)
	}
	up, down, ok := strings.Cut(string(contents), "-- +goose Down")
	if !ok {
		t.Fatal("migration has no Down section")
	}
	return up, down
}

func TestBookingRequestReplay(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim", "helen")
	first := seedSlot(t, ctx, store)
	second, err := store.CreateSlot(ctx, booking.CreateSlotInput{ResourceID: first.ResourceID, StartsAt: first.EndsAt, EndsAt: first.EndsAt.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	in := booking.CreateBookingInput{SlotID: first.ID, UserID: "tim", RequestKey: "replay"}
	created, err := store.CreateBooking(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := store.CreateBooking(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	assertBooking(t, replay, created)
	changed := in
	changed.SlotID = second.ID
	if _, err := store.CreateBooking(ctx, changed); !errors.Is(err, booking.ErrIdempotencyConflict) {
		t.Fatalf("changed payload: %v", err)
	}
	changed.SlotID = 99999
	if _, err := store.CreateBooking(ctx, changed); !errors.Is(err, booking.ErrIdempotencyConflict) {
		t.Fatalf("changed nonexistent slot: %v", err)
	}
	if err := store.CancelBooking(ctx, booking.CancelBookingInput{ID: created.ID, UserID: "tim"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBooking(ctx, booking.CreateBookingInput{SlotID: first.ID, UserID: "helen", RequestKey: in.RequestKey}); err != nil {
		t.Fatalf("independent user/key: %v", err)
	}
	replay, err = store.CreateBooking(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if replay.ID != created.ID || replay.Status != booking.StatusCancelled {
		t.Fatalf("cancelled replay: %+v", replay)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("booking count = %d, want 2", count)
	}
}

func TestConcurrentBookingRequestReplay(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim")
	slot := seedSlot(t, ctx, store)
	const requests = 20
	start := make(chan struct{})
	type result struct {
		booking booking.Booking
		err     error
	}
	results := make(chan result, requests)
	for range requests {
		go func() {
			<-start
			b, err := store.CreateBooking(ctx, booking.CreateBookingInput{SlotID: slot.ID, UserID: "tim", RequestKey: "same-key"})
			results <- result{b, err}
		}()
	}
	close(start)
	id := 0
	for range requests {
		r := <-results
		if r.err != nil {
			t.Errorf("concurrent retry: %v", r.err)
			continue
		}
		if id == 0 {
			id = r.booking.ID
		}
		if r.booking.ID != id {
			t.Errorf("booking ID = %d, want %d", r.booking.ID, id)
		}
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("booking count = %d, want 1", count)
	}
}

func TestConcurrentBookingRequestChangedPayload(t *testing.T) {
	ctx, store, pool := newTestStore(t)
	seedUsers(t, ctx, store, "tim")
	first := seedSlot(t, ctx, store)
	second, err := store.CreateSlot(ctx, booking.CreateSlotInput{ResourceID: first.ResourceID, StartsAt: first.EndsAt, EndsAt: first.EndsAt.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 20)
	for i := range 20 {
		slotID := first.ID
		if i%2 == 1 {
			slotID = second.ID
		}
		go func() {
			<-start
			_, err := store.CreateBooking(ctx, booking.CreateBookingInput{SlotID: slotID, UserID: "tim", RequestKey: "same-key"})
			results <- err
		}()
	}
	close(start)
	successes, conflicts := 0, 0
	for range 20 {
		err := <-results
		if err == nil {
			successes++
		} else if errors.Is(err, booking.ErrIdempotencyConflict) {
			conflicts++
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if successes != 10 || conflicts != 10 {
		t.Fatalf("successes/conflicts = %d/%d, want 10/10", successes, conflicts)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("booking count = %d, want 1", count)
	}
}
