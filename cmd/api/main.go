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
	store := memory.NewStore()
	service := booking.NewService(store)

	resourceIn := booking.CreateResourceInput{
		ID: "nook-1",
	}
	resource, err := service.CreateResource(ctx, resourceIn)
	if err != nil {
		log.Fatalf("CreateResource: %v", err)
	}
	fmt.Printf("Created resource %v\n", resource.ID)

	userIn := booking.CreateUserInput{
		ID: "tim",
	}
	user, err := service.CreateUser(ctx, userIn)
	if err != nil {
		log.Fatalf("CreateUser: %v", err)
	}
	fmt.Printf("Created user %v\n", user.ID)

	slotIn := booking.CreateSlotInput{
		ResourceID: "nook-1",
		StartsAt:   now,
		EndsAt:     now.Add(time.Hour),
	}
	slot, err := service.CreateSlot(ctx, slotIn)
	if err != nil {
		log.Fatalf("CreateSlot: %v", err)
	}
	fmt.Printf("Created slot %d\n", slot.ID)

	bookingInput := booking.CreateBookingInput{
		SlotID: 1,
		UserID: "tim",
	}
	booking, err := service.CreateBooking(ctx, bookingInput)
	if err != nil {
		log.Fatalf("CreateBooking: %v", err)
	}
	fmt.Printf("Created booking %d for slot %d\n", booking.ID, booking.SlotID)
}
