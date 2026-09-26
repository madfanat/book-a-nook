package httpapi_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/httpapi"
	"book-a-nook/internal/memory"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateResourceValidRequest(t *testing.T) {
	service := booking.NewService(memory.NewStore())
	handler := httpapi.NewHandler(service)

	response := request(
		handler,
		http.MethodPost,
		"/resources",
		"application/json",
		`{"id":" nook-1 "}`,
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

	if resource.ID != "nook-1" {
		t.Errorf("ID = %q, want nook-1", resource.ID)
	}
	if resource.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}
