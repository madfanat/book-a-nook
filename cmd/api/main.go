package main

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/memory"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	now := time.Now()
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())

	resource, err := service.CreateResource(
		ctx,
		booking.CreateResourceInput{ID: "nook-1"},
	)
	if err != nil {
		log.Fatalf("setup: CreateResource() error = %v", err)
	}
	fmt.Printf("Created resource %v\n", resource.ID)

	user, err := service.CreateUser(
		ctx,
		booking.CreateUserInput{ID: "tim"},
	)
	if err != nil {
		log.Fatalf("setup: CreateUser() error = %v", err)
	}
	fmt.Printf("Created user %v\n", user.ID)

	slot, err := service.CreateSlot(
		ctx,
		booking.CreateSlotInput{
			ResourceID: resource.ID,
			StartsAt:   now,
			EndsAt:     now.Add(time.Hour),
		},
	)
	if err != nil {
		log.Fatalf("setup: CreateSlot() error = %v", err)
	}
	fmt.Printf("Created slot %d\n", slot.ID)

	booking, err := service.CreateBooking(
		ctx,
		booking.CreateBookingInput{
			SlotID: slot.ID,
			UserID: user.ID,
		},
	)
	if err != nil {
		log.Fatalf("setup: CreateBooking() error = %v", err)
	}
	fmt.Printf("Created booking %d for slot %d\n", booking.ID, booking.SlotID)
}
