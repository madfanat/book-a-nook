package booking

import "errors"

var (
	ErrInvalidStatus          = errors.New("invalid status")
	ErrInvalidTimeRange       = errors.New("invalid time range")
	ErrInvalidSlotID          = errors.New("invalid slot ID")
	ErrInvalidResourceID      = errors.New("invalid resource ID")
	ErrInvalidUserID          = errors.New("invalid user ID")
	ErrUserAlreadyCreated     = errors.New("user is already created")
	ErrUserNotFound           = errors.New("user not found")
	ErrResourceAlreadyCreated = errors.New("resource is already created")
	ErrResourceNotFound       = errors.New("resource not found")
	ErrSlotNotFound           = errors.New("slot not found")
	ErrSlotAlreadyBooked      = errors.New("slot is already booked")
	ErrSlotsOverlap           = errors.New("slots overlap")
)
