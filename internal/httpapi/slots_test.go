package httpapi_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/httpapi"
	"book-a-nook/internal/memory"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCreateSlotValidRequest(t *testing.T) {
	ctx := context.Background()
	service := booking.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	resource, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: "nook-1"})
	if err != nil {
		t.Fatalf("setup: CreateResource() error: %v", err)
	}

	startsAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	endsAt := startsAt.Add(time.Hour)

	response := request(
		handler,
		http.MethodPost,
		"/slots",
		"application/json",
		`{
			"resource_id": " nook-1 ",
			"starts_at": "2026-11-07T12:00:00Z",
			"ends_at": "2026-11-07T13:00:00Z"
		}`,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}

	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var slot booking.Slot
	if err := json.Unmarshal(response.Body.Bytes(), &slot); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if slot.ID <= 0 {
		t.Errorf("ID = %d, want a positive ID", slot.ID)
	}
	if slot.ResourceID != resource.ID {
		t.Errorf("ResourceID = %q, want %q", slot.ResourceID, resource.ID)
	}
	if !slot.StartsAt.Equal(startsAt) {
		t.Errorf("StartsAt = %v, want %v", slot.StartsAt, startsAt)
	}
	if !slot.EndsAt.Equal(endsAt) {
		t.Errorf("EndsAt = %v, want %v", slot.EndsAt, endsAt)
	}
	if slot.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreateSlotInvalidRequest(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantError   error
	}{
		{
			name:       "missing ContentType",
			body:       `{}`,
			wantStatus: httpapi.ErrInvalidContentType.Status(),
			wantError:  httpapi.ErrInvalidContentType,
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
			body:        `{"resource_id":`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "invalid ResourceID type",
			contentType: "application/json",
			body:        `{"resource_id":123}`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "invalid time range",
			contentType: "application/json",
			body: `{
				"resource_id":"nook-1",
				"starts_at":"tomorrow",
				"ends_at":"2026-11-07T13:00:00Z"
			}`,
			wantStatus: httpapi.ErrInvalidBody.Status(),
			wantError:  httpapi.ErrInvalidBody,
		},
		{
			name:        "multiple JSON values",
			contentType: "application/json",
			body:        `{"resource_id":"nook-1"} {}`,
			wantStatus:  httpapi.ErrMultipleJSONValues.Status(),
			wantError:   httpapi.ErrMultipleJSONValues,
		},
		{
			name:        "trailing garbage",
			contentType: "application/json",
			body:        `{"resource_id":"nook-1"} garbage`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "missing ResourceID",
			contentType: "application/json",
			body: `{
				"starts_at":"2026-11-07T12:00:00Z",
				"ends_at":"2026-11-07T13:00:00Z"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  booking.ErrInvalidResourceID,
		},
		{
			name:        "whitespace ResourceID",
			contentType: "application/json",
			body: `{
				"resource_id":" ",
				"starts_at":"2026-11-07T12:00:00Z",
				"ends_at":"2026-11-07T13:00:00Z"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  booking.ErrInvalidResourceID,
		},
		{
			name:        "non-existent resource",
			contentType: "application/json",
			body: `{
				"resource_id":"nook-2",
				"starts_at":"2026-11-07T12:00:00Z",
				"ends_at":"2026-11-07T13:00:00Z"
			}`,
			wantStatus: http.StatusNotFound,
			wantError:  booking.ErrResourceNotFound,
		},
		{
			name:        "missing starts_at",
			contentType: "application/json",
			body: `{
				"resource_id":"nook-1",
				"ends_at":"2026-11-07T13:00:00Z"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  booking.ErrInvalidTimeRange,
		},
		{
			name:        "missing ends_at",
			contentType: "application/json",
			body: `{
				"resource_id":"nook-1",
				"starts_at":"2026-11-07T12:00:00Z"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  booking.ErrInvalidTimeRange,
		},
		{
			name:        "EndsAt before StartsAt",
			contentType: "application/json",
			body: `{
				"resource_id":"nook-1",
				"starts_at":"2026-11-07T13:00:00Z",
				"ends_at":"2026-11-07T12:00:00Z"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  booking.ErrInvalidTimeRange,
		},
		{
			name:        "EndsAt equals StartsAt",
			contentType: "application/json",
			body: `{
				"resource_id":"nook-1",
				"starts_at":"2026-11-07T12:00:00Z",
				"ends_at":"2026-11-07T12:00:00Z"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  booking.ErrInvalidTimeRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := booking.NewService(memory.NewStore())
			handler := httpapi.NewHandler(service)

			_, err := service.CreateResource(
				context.Background(),
				booking.CreateResourceInput{ID: "nook-1"},
			)
			if err != nil {
				t.Fatalf("setup: CreateResource() error: %v", err)
			}

			response := request(
				handler,
				http.MethodPost,
				"/slots",
				tt.contentType,
				tt.body,
			)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d",
					response.Code, tt.wantStatus)
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
				t.Errorf("error = %q, want %q",
					body.Error, tt.wantError.Error())
			}
		})
	}
}

func TestCreateSlotOverlap(t *testing.T) {
	start := time.Now()
	tests := []struct {
		name       string
		input      booking.CreateSlotInput
		wantStatus int
		wantError  error
	}{
		{
			name: "identical times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start,
				EndsAt:     start.Add(60 * time.Minute),
			},
			wantStatus: http.StatusConflict,
			wantError:  booking.ErrSlotOverlap,
		},
		{
			name: "overlaps start",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(-30 * time.Minute),
				EndsAt:     start.Add(30 * time.Minute),
			},
			wantStatus: http.StatusConflict,
			wantError:  booking.ErrSlotOverlap,
		},
		{
			name: "overlaps end",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(30 * time.Minute),
				EndsAt:     start.Add(90 * time.Minute),
			},
			wantStatus: http.StatusConflict,
			wantError:  booking.ErrSlotOverlap,
		},
		{
			name: "contained",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(15 * time.Minute),
				EndsAt:     start.Add(45 * time.Minute),
			},
			wantStatus: http.StatusConflict,
			wantError:  booking.ErrSlotOverlap,
		},
		{
			name: "contains",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(-60 * time.Minute),
				EndsAt:     start.Add(120 * time.Minute),
			},
			wantStatus: http.StatusConflict,
			wantError:  booking.ErrSlotOverlap,
		},
		{
			name: "adjacent before",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(-60 * time.Minute),
				EndsAt:     start,
			},
			wantStatus: http.StatusCreated,
			wantError:  nil,
		},
		{
			name: "adjacent after",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(60 * time.Minute),
				EndsAt:     start.Add(120 * time.Minute),
			},
			wantStatus: http.StatusCreated,
			wantError:  nil,
		},
		{
			name: "different times",
			input: booking.CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   start.Add(120 * time.Minute),
				EndsAt:     start.Add(180 * time.Minute),
			},
			wantStatus: http.StatusCreated,
			wantError:  nil,
		},
		{
			name: "different resource",
			input: booking.CreateSlotInput{
				ResourceID: "nook-2",
				StartsAt:   start,
				EndsAt:     start.Add(60 * time.Minute),
			},
			wantStatus: http.StatusCreated,
			wantError:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service := booking.NewService(memory.NewStore())
			for _, id := range []string{"nook-1", "nook-2"} {
				if _, err := service.CreateResource(ctx, booking.CreateResourceInput{ID: id}); err != nil {
					t.Fatalf("setup: CreateResource(): %v", err)
				}
			}
			if _, err := service.CreateSlot(ctx, booking.CreateSlotInput{
				ResourceID: "nook-1", StartsAt: start, EndsAt: start.Add(time.Hour),
			}); err != nil {
				t.Fatalf("setup: CreateSlot(): %v", err)
			}

			handler := httpapi.NewHandler(service)
			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal(): %v", err)
			}
			response := request(handler, http.MethodPost, "/slots", "application/json", string(body))
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
					t.Fatalf("Unmarshal(): %v", err)
				}
				if body.Error != tt.wantError.Error() {
					t.Errorf("error = %q, want %q", body.Error, tt.wantError.Error())
				}
				return
			}
			var got booking.Slot
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("Unmarshal(): %v", err)
			}
			if got.ID <= 0 {
				t.Errorf("ID = %d, want positive", got.ID)
			}
			if got.ResourceID != tt.input.ResourceID {
				t.Errorf("ResourceID = %q, want %q", got.ResourceID, tt.input.ResourceID)
			}
			if !got.StartsAt.Equal(tt.input.StartsAt) {
				t.Errorf("StartsAt = %v, want %v", got.StartsAt, tt.input.StartsAt)
			}
			if !got.EndsAt.Equal(tt.input.EndsAt) {
				t.Errorf("EndsAt = %v, want %v", got.EndsAt, tt.input.EndsAt)
			}
			if got.CreatedAt.IsZero() {
				t.Error("CreatedAt is zero")
			}
		})
	}
}
