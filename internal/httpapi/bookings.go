package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
)

func (handler *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	input, reqErr := readJSON[booking.CreateBookingInput](r)
	if reqErr != nil {
		writeError(w, reqErr.status, reqErr.message)
		return
	}

	booking, err := handler.service.CreateBooking(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, booking)
}
