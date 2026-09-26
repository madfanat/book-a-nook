package httpapi_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/httpapi"
	"book-a-nook/internal/memory"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateUserValidRequest(t *testing.T) {
	service := booking.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	response := request(
		handler,
		http.MethodPost,
		"/users",
		"application/json",
		`{"id":" tim "}`,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}

	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var resource booking.Resource
	if err := json.Unmarshal(response.Body.Bytes(), &resource); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if resource.ID != "tim" {
		t.Errorf("ID = %q, want tim", resource.ID)
	}
	if resource.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestCreateUserInvalidRequest(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantError   error
	}{
		{
			name:       "missing content type",
			body:       `{"id":"tim"}`,
			wantStatus: httpapi.ErrInvalidContentType.Status(),
			wantError:  httpapi.ErrInvalidContentType,
		},
		{
			name:        "invalid content type",
			contentType: "text/plain",
			body:        `{"id":"tim"}`,
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
			body:        `{"id":`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "invalid field type",
			contentType: "application/json",
			body:        `{"id":123}`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
		{
			name:        "whitespace ID",
			contentType: "application/json",
			body:        `{"id":" "}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidUserID,
		},
		{
			name:        "missing ID",
			contentType: "application/json",
			body:        `{}`,
			wantStatus:  http.StatusBadRequest,
			wantError:   booking.ErrInvalidUserID,
		},
		{
			name:        "multiple JSON values",
			contentType: "application/json",
			body:        `{"id":"tim"} {"id":"helen"}`,
			wantStatus:  httpapi.ErrMultipleJSONValues.Status(),
			wantError:   httpapi.ErrMultipleJSONValues,
		},
		{
			name:        "trailing garbage",
			contentType: "application/json",
			body:        `{"id":"tim"} garbage`,
			wantStatus:  httpapi.ErrInvalidBody.Status(),
			wantError:   httpapi.ErrInvalidBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := booking.NewService(memory.NewStore())
			handler := httpapi.NewHandler(service)

			response := request(
				handler,
				http.MethodPost,
				"/users",
				tt.contentType,
				tt.body,
			)

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
