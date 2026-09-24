package booking

import "errors"

var (
	// Resources
	ErrInvalidResourceID      = errors.New("invalid resource ID")
	ErrResourceNotFound       = errors.New("resource not found")
	ErrResourceAlreadyCreated = errors.New("resource is already created")

	// Users
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyCreated = errors.New("user is already created")

	// Slots
	ErrInvalidSlotID    = errors.New("invalid slot ID")
	ErrSlotNotFound     = errors.New("slot not found")
	ErrInvalidTimeRange = errors.New("invalid time range")

	// Bookings
	ErrInvalidBookingStatus = errors.New("invalid status")
	ErrSlotAlreadyBooked    = errors.New("slot is already booked")
	ErrBookingOverlap       = errors.New("slots overlap")
)
