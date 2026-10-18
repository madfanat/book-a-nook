package httpapi

import (
	"book-a-nook/internal/booking"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, booking.ErrInvalidResourceID),
		errors.Is(err, booking.ErrInvalidUserID),
		errors.Is(err, booking.ErrInvalidSlotID),
		errors.Is(err, booking.ErrInvalidBookingID),
		errors.Is(err, booking.ErrInvalidRequestKey),
		errors.Is(err, booking.ErrInvalidTimeRange),
		errors.Is(err, booking.ErrInvalidLimit),
		errors.Is(err, booking.ErrInvalidCursor):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, booking.ErrResourceNotFound),
		errors.Is(err, booking.ErrUserNotFound),
		errors.Is(err, booking.ErrSlotNotFound),
		errors.Is(err, booking.ErrBookingNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, booking.ErrResourceExists),
		errors.Is(err, booking.ErrUserExists),
		errors.Is(err, booking.ErrSlotAlreadyBooked),
		errors.Is(err, booking.ErrIdempotencyConflict),
		errors.Is(err, booking.ErrSlotOverlap):
		writeError(w, http.StatusConflict, err.Error())

	default:
		slog.Error("handle request", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
