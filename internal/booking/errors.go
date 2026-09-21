package booking

import "errors"

var (
	ErrInvalidStatus     = errors.New("invalid status")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrInvalidSlotID     = errors.New("invalid slot ID")
	ErrInvalidResourceID = errors.New("invalid resource ID")
	ErrInvalidUserID     = errors.New("invalid user ID")
)
