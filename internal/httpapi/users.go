package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
)

func (handler *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	input, reqErr := readJSON[booking.CreateUserInput](r)
	if reqErr != nil {
		writeError(w, reqErr.status, reqErr.message)
		return
	}

	user, err := handler.service.CreateUser(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}
