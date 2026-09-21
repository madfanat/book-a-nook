package main

import (
	"book-a-nook/internal/booking"
	"fmt"
	"log"
	"time"
)

func main() {
	start := time.Now()

	slotInput := booking.CreateSlotInput{
		ResourceID: 1,
		StartsAt:   start,
		EndsAt:     start.Add(time.Hour),
	}

	if err := slotInput.Validate(); err != nil {
		log.Fatalf("invalid slot input: %v", err)
	}

	bookingInput := booking.CreateBookingInput{
		SlotID: 1,
		UserID: 1,
	}

	if err := bookingInput.Validate(); err != nil {
		log.Fatalf("invalid booking input: %v", err)
	}

	fmt.Println("Everything is valid!")
}
