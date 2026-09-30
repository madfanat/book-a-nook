package booking

import "errors"

var (
	// Resources
	ErrInvalidResourceID = errors.New("invalid resource ID")
	ErrResourceNotFound  = errors.New("resource not found")
	ErrResourceExists    = errors.New("resource is already created")

	// Users
	ErrInvalidUserID = errors.New("invalid user ID")
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user is already created")

	// Slots
	ErrInvalidSlotID    = errors.New("invalid slot ID")
	ErrSlotNotFound     = errors.New("slot not found")
	ErrInvalidTimeRange = errors.New("invalid time range")
	ErrSlotOverlap      = errors.New("slots overlap")

	// Bookings
	ErrInvalidBookingID     = errors.New("invalid booking ID")
	ErrBookingNotFound      = errors.New("booking not found")
	ErrInvalidBookingStatus = errors.New("invalid status")
	ErrSlotAlreadyBooked    = errors.New("slot is already booked")

	// Other
	ErrInvalidLimit = errors.New("invalid limit")
)
