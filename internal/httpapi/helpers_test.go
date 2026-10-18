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
	headers ...http.Header,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for _, header := range headers {
		for name, values := range header {
			req.Header[name] = values
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func requestBooking(handler http.Handler, contentType, body string) *httptest.ResponseRecorder {
	return request(handler, http.MethodPost, "/bookings", contentType, body,
		http.Header{"Idempotency-Key": {"http-request"}})
}
