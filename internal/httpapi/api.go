package httpapi

import (
	"book-a-nook/internal/booking"
	"context"
	"maps"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	service *booking.Service
}

type principalKey struct{}

func NewHandler(service *booking.Service, tokens map[string]string, ping func(context.Context) error) http.Handler {
	handler := &Handler{service: service}
	api := http.NewServeMux()
	api.HandleFunc("POST /v1/bookings", handler.createBooking)
	api.HandleFunc("GET /v1/bookings/{id}", handler.getBooking)
	api.HandleFunc("DELETE /v1/bookings/{id}", handler.cancelBooking)
	api.HandleFunc("GET /v1/slots", handler.listSlots)
	tokens = maps.Clone(tokens)

	mux := http.NewServeMux()
	mux.Handle("/v1/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		userID := tokens[token]
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" || userID == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		api.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, userID)))
	}))
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if ping == nil || ping(ctx) != nil {
			writeError(w, http.StatusServiceUnavailable, "not ready")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func userID(r *http.Request) string {
	id, _ := r.Context().Value(principalKey{}).(string)
	return id
}
