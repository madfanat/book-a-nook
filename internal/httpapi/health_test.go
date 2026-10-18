package httpapi_test

import (
	"book-a-nook/internal/booking"
	"book-a-nook/internal/httpapi"
	"book-a-nook/internal/memory"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestHealth(t *testing.T) {
	for _, available := range []bool{true, false} {
		calls := 0
		handler := httpapi.NewHandler(booking.NewService(memory.NewStore()), nil, func(ctx context.Context) error {
			calls++
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > time.Second {
				t.Fatal("database probe is not bounded")
			}
			if !available {
				return errors.New("database unavailable")
			}
			return nil
		})
		live := request(handler, "GET", "/livez", "", "", http.Header{"Authorization": {""}})
		if live.Code != http.StatusOK || calls != 0 {
			t.Fatalf("livez: %d, probes %d", live.Code, calls)
		}
		ready := request(handler, "GET", "/readyz", "", "", http.Header{"Authorization": {""}})
		want := http.StatusOK
		if !available {
			want = http.StatusServiceUnavailable
		}
		if ready.Code != want || calls != 1 {
			t.Fatalf("readyz: %d, probes %d", ready.Code, calls)
		}
	}
}
