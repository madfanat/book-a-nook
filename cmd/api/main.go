package main

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/postgres"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("set DATABASE_URL")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	store := postgres.NewStore(pool)

	resource, err := store.CreateResource(
		ctx,
		booking.CreateResourceInput{ID: "hell"},
	)
	if err != nil {
		return err
	}

	user, err := store.CreateUser(
		ctx,
		booking.CreateUserInput{ID: "tim"},
	)
	if err != nil {
		return err
	}

	slot, err := store.CreateSlot(
		ctx,
		booking.CreateSlotInput{
			ResourceID: "hell",
			StartsAt:   time.Now().Add(time.Hour),
			EndsAt:     time.Now().Add(2 * time.Hour)},
	)
	if err != nil {
		return err
	}

	slots, err := store.ListSlots(ctx, booking.ListSlotsInput{ResourceID: "hell", Limit: 10})
	if err != nil {
		return err
	}

	b, err := store.CreateBooking(
		ctx,
		booking.CreateBookingInput{
			SlotID: 1,
			UserID: "tim"},
	)
	if err != nil {
		return err
	}

	err = store.CancelBooking(
		ctx,
		booking.CancelBookingInput{
			ID:     1,
			UserID: "tim"},
	)
	if err != nil {
		return err
	}

	got, err := store.GetBooking(
		ctx,
		booking.GetBookingInput{
			ID:     1,
			UserID: "tim"},
	)

	newBooking, err := store.CreateBooking(
		ctx,
		booking.CreateBookingInput{
			SlotID: 1,
			UserID: "tim"},
	)
	if err != nil {
		return err
	}

	fmt.Printf("Resource: %+v\n", resource)
	fmt.Printf("User: %+v\n", user)
	fmt.Printf("Slot: %+v\n", slot)
	fmt.Printf("Slots: %+v\n", slots)
	fmt.Printf("Booking: %+v\n", b)
	fmt.Printf("Cancelled booking: %+v\n", got)
	fmt.Printf("New booking: %+v\n", newBooking)
	return nil
}
