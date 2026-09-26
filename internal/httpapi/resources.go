package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
)

func (handler *Handler) createResource(w http.ResponseWriter, r *http.Request) {
	input, reqErr := readJSON[booking.CreateResourceInput](r)
	if reqErr != nil {
		writeError(w, reqErr.status, reqErr.message)
		return
	}

	resource, err := handler.service.CreateResource(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resource)
}
