package httpapi

import (
	"book-a-nook/internal/booking"
	"net/http"
)

type Handler struct {
	service *booking.Service
}

func NewHandler(service *booking.Service) http.Handler {
	handler := &Handler{service: service}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /resources", handler.createResource)
	mux.HandleFunc("POST /users", handler.createUser)
	mux.HandleFunc("POST /slots", handler.createSlot)
	mux.HandleFunc("POST /bookings", handler.createBooking)

	return mux
}

func (handler *Handler) NewHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
