package httpapi_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/httpapi"
	"book-a-nook/internal/memory"
	"net/http"
	"net/http/httptest"
	"strings"
)

func newTestHandler() http.Handler {
	service := booking.NewService(memory.NewStore())
	return httpapi.NewHandler(service)
}

func request(
	handler http.Handler,
	method, path, contentType, body string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}
