package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
	"strconv"
)

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

	if values, present := query["after"]; present {
		after, err := strconv.Atoi(values[0])
		if err != nil || after < 0 {
			writeError(w, http.StatusBadRequest, booking.ErrInvalidCursor.Error())
			return
		}
		input.After = after
	}

	slots, err := handler.service.ListSlots(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, slots)
}
