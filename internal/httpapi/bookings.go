package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
	"strconv"
)

func (handler *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	payload, reqErr := readJSON[struct {
		SlotID int `json:"slot_id"`
	}](w, r)
	if reqErr != nil {
		writeError(w, reqErr.status, reqErr.message)
		return
	}
	input := booking.CreateBookingInput{
		SlotID:     payload.SlotID,
		UserID:     userID(r),
		RequestKey: r.Header.Get("Idempotency-Key"),
	}
	created, err := handler.service.CreateBooking(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Location", "/v1/bookings/"+strconv.Itoa(created.ID))
	writeJSON(w, http.StatusCreated, created)
}

func (handler *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, booking.ErrInvalidBookingID.Error())
		return
	}
	found, err := handler.service.GetBooking(r.Context(), booking.GetBookingInput{ID: id, UserID: userID(r)})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, found)
}

func (handler *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, booking.ErrInvalidBookingID.Error())
		return
	}
	if err := handler.service.CancelBooking(r.Context(), booking.CancelBookingInput{ID: id, UserID: userID(r)}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
