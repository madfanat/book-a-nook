package httpapi_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/httpapi"
	"book-a-nook/internal/memory"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCreateBookingValidRequest(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	resource, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"})
	if err != nil {
		t.Fatalf("setup: CreateResource() error: %v", err)
	}
	user, err := service.CreateUser(ctx, booking.CreateUserInput{ID: "tim"})
	if err != nil {
		t.Fatalf("setup: CreateUser() error: %v", err)
	}
	startsAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: resource.ID,
		StartsAt:   startsAt,
		EndsAt:     startsAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot() error: %v", err)
	}

	response := request(handler, http.MethodPost, "/bookings", "application/json",
		fmt.Sprintf(`{"slot_id":%d,"user_id":" tim "}`, slot.ID))
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var got booking.Booking
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if got.ID <= 0 {
		t.Errorf("ID = %d, want a positive ID", got.ID)
	}
	if got.SlotID != slot.ID {
		t.Errorf("SlotID = %d, want %d", got.SlotID, slot.ID)
	}
	if got.UserID != user.ID {
		t.Errorf("UserID = %q, want %q", got.UserID, user.ID)
	}
	if got.Status != booking.StatusActive {
		t.Errorf("Status = %q, want %q", got.Status, booking.StatusActive)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreateBookingInvalidRequest(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantError   error
	}{
		{
			name:        "missing ContentType",
			contentType: "",
			body:        `{}`,
			wantStatus:  httpapi.ErrInvalidContentType.Status(),
			wantError:   httpapi.ErrInvalidContentType,
		},
		{
			name:        "invalid ContentType",
			contentType: "text/plain",
			body:        `{}`,
			wantStatus:  httpapi.ErrInvalidContentType.Status(),
			wantError:   httpapi.ErrInvalidContentType,
		},
		{
			name:        "empty body",
			contentType: "application/json",
			body:        "",
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "invalid body",
			contentType: "application/json",
			body:        `{"slot_id":`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "invalid SlotID type",
			contentType: "application/json",
			body:        `{"slot_id":"one","user_id":"tim"}`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "invalid UserID type",
			contentType: "application/json",
			body:        `{"slot_id":1,"user_id":123}`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "multiple JSON values",
			contentType: "application/json",
			body:        `{} {}`,
			wantStatus:  httpapi.ErrMultipleJSONValues.Status(),
			wantError:   httpapi.ErrMultipleJSONValues,
		},
		{
			name:        "trailing garbage",
			contentType: "application/json",
			body:        `{} garbage`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "missing SlotID",
			contentType: "application/json",
			body:        `{"user_id":"tim"}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidSlotID,
		},
		{
			name:        "zero SlotID",
			contentType: "application/json",
			body:        `{"slot_id":0,"user_id":"tim"}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidSlotID,
		},
		{
			name:        "negative SlotID",
			contentType: "application/json",
			body:        `{"slot_id":-1,"user_id":"tim"}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidSlotID,
		},
		{
			name:        "missing UserID",
			contentType: "application/json",
			body:        `{"slot_id":1}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidUserID,
		},
		{
			name:        "empty UserID",
			contentType: "application/json",
			body:        `{"slot_id":1,"user_id":""}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidUserID,
		},
		{
			name:        "whitespace UserID",
			contentType: "application/json",
			body:        `{"slot_id":1,"user_id":" "}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := booking.NewService(memory.NewStore())
			handler := httpapi.NewHandler(service)

			response := request(handler, http.MethodPost, "/bookings", tt.contentType, tt.body)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if body.Error != tt.wantError.Error() {
				t.Errorf("error = %q, want %q", body.Error, tt.wantError.Error())
			}
		})
	}
}

func TestCreateBookingNonExistentUser(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	resource, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"})
	if err != nil {
		t.Fatalf("setup: CreateResource() error: %v", err)
	}
	_, err = service.CreateUser(ctx, booking.CreateUserInput{ID: "tim"})
	if err != nil {
		t.Fatalf("setup: CreateUser() error: %v", err)
	}
	startsAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: resource.ID,
		StartsAt:   startsAt,
		EndsAt:     startsAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot() error: %v", err)
	}

	response := request(handler, http.MethodPost, "/bookings", "application/json",
		fmt.Sprintf(`{"slot_id":%d,"user_id":%q}`, slot.ID, "missing"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if body.Error != booking.ErrUserNotFound.Error() {
		t.Errorf("error = %q, want %q", body.Error, booking.ErrUserNotFound.Error())
	}
}

func TestCreateBookingNonExistentSlot(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	resource, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"})
	if err != nil {
		t.Fatalf("setup: CreateResource() error: %v", err)
	}
	user, err := service.CreateUser(ctx, booking.CreateUserInput{ID: "tim"})
	if err != nil {
		t.Fatalf("setup: CreateUser() error: %v", err)
	}
	startsAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: resource.ID,
		StartsAt:   startsAt,
		EndsAt:     startsAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot() error: %v", err)
	}

	response := request(handler, http.MethodPost, "/bookings", "application/json",
		fmt.Sprintf(`{"slot_id":%d,"user_id":%q}`, slot.ID+1, user.ID))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if body.Error != booking.ErrSlotNotFound.Error() {
		t.Errorf("error = %q, want %q", body.Error, booking.ErrSlotNotFound.Error())
	}
}

func TestCreateBookingAlreadyBookedSlot(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	resource, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"})
	if err != nil {
		t.Fatalf("setup: CreateResource() error: %v", err)
	}
	user, err := service.CreateUser(ctx, booking.CreateUserInput{ID: "tim"})
	if err != nil {
		t.Fatalf("setup: CreateUser() error: %v", err)
	}
	startsAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
		ResourceID: resource.ID,
		StartsAt:   startsAt,
		EndsAt:     startsAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup: CreateSlot() error: %v", err)
	}

	if _, err := service.CreateBooking(ctx, booking.CreateBookingInput{SlotID: slot.ID, UserID: user.ID}); err != nil {
		t.Fatalf("setup: CreateBooking() error: %v", err)
	}

	response := request(handler, http.MethodPost, "/bookings", "application/json",
		fmt.Sprintf(`{"slot_id":%d,"user_id":%q}`, slot.ID, user.ID))
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if body.Error != booking.ErrSlotAlreadyBooked.Error() {
		t.Errorf("error = %q, want %q", body.Error, booking.ErrSlotAlreadyBooked.Error())
	}
}

func TestCreateBookingOverlap(t *testing.T) {
	startsAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		input      booking.CreateSlotInput
		wantStatus int
		wantError  error
	}{
		{
			name:       "same resource overlapping times",
			input:      booking.CreateSlotInput{ResourceID: "nook-1", StartsAt: startsAt.Add(30 * time.Minute), EndsAt: startsAt.Add(90 * time.Minute)},
			wantStatus: http.StatusConflict,
			wantError:  booking.ErrBookingOverlap,
		},
		{
			name:       "same resource different times",
			input:      booking.CreateSlotInput{ResourceID: "nook-1", StartsAt: startsAt.Add(2 * time.Hour), EndsAt: startsAt.Add(3 * time.Hour)},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "same resource adjacent times",
			input:      booking.CreateSlotInput{ResourceID: "nook-1", StartsAt: startsAt.Add(time.Hour), EndsAt: startsAt.Add(2 * time.Hour)},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "different resource overlapping times",
			input:      booking.CreateSlotInput{ResourceID: "nook-2", StartsAt: startsAt.Add(30 * time.Minute), EndsAt: startsAt.Add(90 * time.Minute)},
			wantStatus: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())
			handler := httpapi.NewHandler(service)

			resource, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"})
			if err != nil {
				t.Fatalf("setup: CreateResource() error: %v", err)
			}
			user, err := service.CreateUser(ctx, booking.CreateUserInput{ID: "tim"})
			if err != nil {
				t.Fatalf("setup: CreateUser() error: %v", err)
			}
			startsAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
			slot, err := service.CreateSlot(ctx, booking.CreateSlotInput{
				ResourceID: resource.ID,
				StartsAt:   startsAt,
				EndsAt:     startsAt.Add(time.Hour),
			})
			if err != nil {
				t.Fatalf("setup: CreateSlot() error: %v", err)
			}

			if _, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-2"}); err != nil {
				t.Fatalf("setup: CreateResource() error: %v", err)
			}
			secondSlot, err := service.CreateSlot(ctx, tt.input)
			if err != nil {
				t.Fatalf("setup: CreateSlot() error: %v", err)
			}
			firstBooking, err := service.CreateBooking(ctx, booking.CreateBookingInput{SlotID: slot.ID, UserID: user.ID})
			if err != nil {
				t.Fatalf("setup: CreateBooking() error: %v", err)
			}
			response := request(handler, http.MethodPost, "/bookings", "application/json",
				fmt.Sprintf(`{"slot_id":%d,"user_id":%q}`, secondSlot.ID, user.ID))
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			if tt.wantError != nil {
				var body struct {
					Error string `json:"error"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatalf("Unmarshal() error: %v", err)
				}
				if body.Error != tt.wantError.Error() {
					t.Errorf("error = %q, want %q", body.Error, tt.wantError.Error())
				}
				return
			}
			var got booking.Booking
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if got.ID <= 0 {
				t.Errorf("ID = %d, want a positive ID", got.ID)
			}
			if got.ID == firstBooking.ID {
				t.Errorf("ID = %d, want a different ID", got.ID)
			}
			if got.SlotID != secondSlot.ID {
				t.Errorf("SlotID = %d, want %d", got.SlotID, secondSlot.ID)
			}
			if got.UserID != user.ID {
				t.Errorf("UserID = %q, want %q", got.UserID, user.ID)
			}
			if got.Status != booking.StatusActive {
				t.Errorf("Status = %q, want %q", got.Status, booking.StatusActive)
			}
			if got.CreatedAt.IsZero() {
				t.Error("CreatedAt is zero")
			}
		})
	}
}
