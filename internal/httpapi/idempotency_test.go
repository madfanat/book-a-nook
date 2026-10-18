package httpapi_test

import (
	"book-a-nook/internal/booking"
	"context"
	"net/http"
	"strings"
	"testing"
)

type captureBookingStore struct {
	booking.Store
	input booking.CreateBookingInput
}

func (s *captureBookingStore) CreateBooking(_ context.Context, in booking.CreateBookingInput) (booking.Booking, error) {
	s.input = in
	return booking.Booking{ID: 1, SlotID: in.SlotID, UserID: in.UserID, Status: booking.StatusActive}, nil
}

func TestBookingRequestKeyFromHeader(t *testing.T) {
	store := &captureBookingStore{}
	handler := testHandler(booking.NewService(store))
	response := request(handler, http.MethodPost, "/v1/bookings", "application/json",
		`{"slot_id":7}`,
		http.Header{"Idempotency-Key": {"Request-A"}})
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body)
	}
	if store.input.UserID != "tim" {
		t.Fatalf("UserID = %q, want authenticated principal", store.input.UserID)
	}
	if store.input.RequestKey != "Request-A" {
		t.Fatalf("RequestKey = %q, want exact header value", store.input.RequestKey)
	}
}

func TestBookingInvalidRequestKey(t *testing.T) {
	for _, key := range []string{"", "   ", strings.Repeat("a", 129), "__legacy_booking__:1"} {
		t.Run(key, func(t *testing.T) {
			response := request(newTestHandler(), http.MethodPost, "/v1/bookings", "application/json",
				`{"slot_id":7}`,
				http.Header{"Idempotency-Key": {key}})
			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), booking.ErrInvalidRequestKey.Error()) {
				t.Fatalf("status = %d, body = %s; want invalid key", response.Code, response.Body)
			}
		})
	}
}
