package httpapi

import (
	"book-a-nook/internal/booking"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("writeJSON error: %v", err)
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
		errors.Is(err, booking.ErrInvalidTimeRange):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, booking.ErrResourceNotFound),
		errors.Is(err, booking.ErrUserNotFound),
		errors.Is(err, booking.ErrSlotNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, booking.ErrResourceExists),
		errors.Is(err, booking.ErrUserExists),
		errors.Is(err, booking.ErrSlotAlreadyBooked),
		errors.Is(err, booking.ErrBookingOverlap):
		writeError(w, http.StatusConflict, err.Error())

	default:
		log.Printf("service error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
