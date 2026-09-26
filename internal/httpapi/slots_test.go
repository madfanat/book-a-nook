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
