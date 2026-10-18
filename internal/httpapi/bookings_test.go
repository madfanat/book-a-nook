package httpapi_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/memory"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func seedAPI(t *testing.T) (http.Handler, []booking.Slot) {
	t.Helper()
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	if _, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"tim", "helen"} {
		if _, err := service.CreateUser(ctx, booking.CreateUserInput{ID: id}); err != nil {
			t.Fatal(err)
		}
	}
	start := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	var slots []booking.Slot
	for i := range 3 {
		slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{ResourceID: "nook-1", StartsAt: start.Add(time.Duration(i) * time.Hour), EndsAt: start.Add(time.Duration(i+1) * time.Hour)})
		if err != nil {
			t.Fatal(err)
		}
		slots = append(slots, slot)
	}
	return testHandler(service), slots
}

func TestBookingLifecycle(t *testing.T) {
	handler, slots := seedAPI(t)
	payload := fmt.Sprintf(`{"slot_id":%d}`, slots[0].ID)
	created := requestBooking(handler, "application/json", payload)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body)
	}
	var b booking.Booking
	if err := json.Unmarshal(created.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if b.UserID != "tim" || b.SlotID != slots[0].ID || b.Status != booking.StatusActive {
		t.Fatalf("booking = %+v", b)
	}
	location := fmt.Sprintf("/v1/bookings/%d", b.ID)
	if created.Header().Get("Location") != location {
		t.Fatal("missing booking location")
	}
	retry := requestBooking(handler, "application/json", payload)
	if retry.Code != http.StatusCreated || retry.Body.String() != created.Body.String() {
		t.Fatalf("replay: %d %s", retry.Code, retry.Body)
	}
	changed := requestBooking(handler, "application/json", fmt.Sprintf(`{"slot_id":%d}`, slots[1].ID))
	if changed.Code != http.StatusConflict {
		t.Fatalf("changed payload: %d %s", changed.Code, changed.Body)
	}
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		wrongOwner := request(handler, method, location+"?user_id=tim", "", "", http.Header{"Authorization": {"Bearer helen-token"}})
		if wrongOwner.Code != http.StatusNotFound {
			t.Fatalf("wrong owner %s: %d", method, wrongOwner.Code)
		}
	}
	got := request(handler, http.MethodGet, location, "", "")
	if got.Code != http.StatusOK || got.Body.String() != created.Body.String() {
		t.Fatalf("get: %d %s", got.Code, got.Body)
	}
	for range 2 {
		cancelled := request(handler, http.MethodDelete, location, "", "")
		if cancelled.Code != http.StatusNoContent || cancelled.Body.Len() != 0 {
			t.Fatalf("cancel: %d %s", cancelled.Code, cancelled.Body)
		}
	}
	replacement := request(handler, http.MethodPost, "/v1/bookings", "application/json", payload, http.Header{"Idempotency-Key": {"new-key"}, "Authorization": {"Bearer helen-token"}})
	if replacement.Code != http.StatusCreated {
		t.Fatalf("rebook: %d %s", replacement.Code, replacement.Body)
	}
	replay := requestBooking(handler, "application/json", payload)
	var cancelled booking.Booking
	if err := json.Unmarshal(replay.Body.Bytes(), &cancelled); err != nil {
		t.Fatal(err)
	}
	if replay.Code != http.StatusCreated || cancelled.ID != b.ID || cancelled.Status != booking.StatusCancelled {
		t.Fatalf("cancelled replay: %d %s", replay.Code, replay.Body)
	}
}

func TestBookingInvalidRequests(t *testing.T) {
	for _, tc := range []struct {
		name, contentType, body string
		status                  int
	}{
		{"missing content type", "", `{}`, 415},
		{"wrong content type", "text/plain", `{}`, 415},
		{"empty", "application/json", "", 400},
		{"malformed", "application/json", `{"slot_id":`, 400},
		{"wrong type", "application/json", `{"slot_id":"one"}`, 400},
		{"multiple values", "application/json", `{} {}`, 400},
		{"trailing garbage", "application/json", `{} garbage`, 400},
		{"missing slot", "application/json", `{}`, 400},
		{"null", "application/json", `null`, 400},
		{"zero slot", "application/json", `{"slot_id":0}`, 400},
		{"negative slot", "application/json", `{"slot_id":-1}`, 400},
		{"slot outside database range", "application/json", `{"slot_id":2147483648}`, 400},
		{"forged owner", "application/json", `{"slot_id":1,"user_id":"helen"}`, 400},
		{"forged key", "application/json", `{"slot_id":1,"request_key":"forged"}`, 400},
		{"unknown field", "application/json", `{"slot_id":1,"other":true}`, 400},
		{"oversized value", "application/json", `{"slot_id":"` + strings.Repeat("1", 16384) + `"}`, 413},
		{"oversized trailing whitespace", "application/json", `{"slot_id":1}` + strings.Repeat(" ", 16384), 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := requestBooking(newTestHandler(), tc.contentType, tc.body)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if rec.Header().Get("Content-Type") != "application/json" {
				t.Fatal("missing JSON content type")
			}
		})
	}
}

func TestBookingMissingAndConflicting(t *testing.T) {
	handler, slots := seedAPI(t)
	missing := requestBooking(handler, "application/json", `{"slot_id":999}`)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing slot: %d", missing.Code)
	}
	payload := fmt.Sprintf(`{"slot_id":%d}`, slots[0].ID)
	if rec := requestBooking(handler, "application/json", payload); rec.Code != http.StatusCreated {
		t.Fatal(rec.Body)
	}
	rec := request(handler, http.MethodPost, "/v1/bookings", "application/json", payload, http.Header{"Idempotency-Key": {"another-key"}})
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), booking.ErrSlotAlreadyBooked.Error()) {
		t.Fatalf("occupied slot: %d %s", rec.Code, rec.Body)
	}
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		for _, tc := range []struct {
			id     string
			status int
		}{{"999", 404}, {"abc", 400}, {"0", 400}, {"-1", 400}, {"2147483648", 400}} {
			rec := request(handler, method, "/v1/bookings/"+tc.id, "", "")
			if rec.Code != tc.status {
				t.Fatalf("%s %s: %d %s", method, tc.id, rec.Code, rec.Body)
			}
		}
	}
}

func TestAuthentication(t *testing.T) {
	handler, _ := seedAPI(t)
	for _, methodPath := range []struct{ method, path string }{
		{"GET", "/v1/slots?resource_id=nook-1"}, {"POST", "/v1/bookings"}, {"GET", "/v1/bookings/1"}, {"DELETE", "/v1/bookings/1"},
	} {
		for _, authorization := range []string{"", "Bearer", "Bearer unknown", "Basic tim-token", "Bearer tim-token extra"} {
			rec := request(handler, methodPath.method, methodPath.path, "application/json", `{}`, http.Header{"Authorization": {authorization}})
			if rec.Code != http.StatusUnauthorized || rec.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("%s %s auth %q: %d", methodPath.method, methodPath.path, authorization, rec.Code)
			}
		}
	}
	for _, path := range []string{"/users", "/resources", "/slots", "/bookings", "/v1/users", "/v1/resources"} {
		rec := request(handler, http.MethodPost, path, "application/json", `{}`)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("setup route %s remains public: %d", path, rec.Code)
		}
	}
}
