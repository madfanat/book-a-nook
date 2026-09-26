package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
)

func (handler *Handler) createSlot(w http.ResponseWriter, r *http.Request) {
	input, reqErr := readJSON[booking.CreateSlotInput](r)
	if reqErr != nil {
		writeError(w, reqErr.status, reqErr.message)
		return
	}

	slot, err := handler.service.CreateSlot(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, slot)
}
