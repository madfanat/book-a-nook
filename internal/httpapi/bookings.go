package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
	"strconv"
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

func (handler *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, booking.ErrInvalidBookingID.Error())
		return
	}

	input := booking.GetBookingInput{
		ID:     id,
		UserID: r.URL.Query().Get("user_id"),
	}

	found, err := handler.service.GetBooking(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, found)
}
