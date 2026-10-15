package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
	"strconv"
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

func (handler *Handler) listSlots(w http.ResponseWriter, r *http.Request) {
	input := booking.ListSlotsInput{}

	query := r.URL.Query()

	input.ResourceID = query.Get("resource_id")

	if limit := query.Get("limit"); limit != "" {
		limit, err := strconv.Atoi(limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, booking.ErrInvalidLimit.Error())
			return
		}

		input.Limit = limit
	}

	slots, err := handler.service.ListSlots(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, slots)
}
