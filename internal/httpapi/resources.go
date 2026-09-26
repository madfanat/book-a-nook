package httpapi

import (
	"book-a-nook/internal/booking"
	"encoding/json"
	"io"
	"mime"
	"net/http"
)

func (handler *Handler) createResource(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType,
			"Content-Type must be application/json")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var input booking.CreateResourceInput
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			writeError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		} else {
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
		}
		return
	}

	resource, err := handler.service.CreateResource(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resource)
}
