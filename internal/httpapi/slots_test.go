package httpapi_test

import (
	"book-a-nook/internal/booking"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestListSlotsPagination(t *testing.T) {
	handler, slots := seedAPI(t)
	after := 0
	for _, want := range slots {
		rec := request(handler, http.MethodGet, fmt.Sprintf("/v1/slots?resource_id=nook-1&limit=1&after=%d", after), "", "")
		var page []booking.Slot
		if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusOK || len(page) != 1 || page[0].ID != want.ID {
			t.Fatalf("page: %d %s", rec.Code, rec.Body)
		}
		after = page[0].ID
	}
	rec := request(handler, http.MethodGet, fmt.Sprintf("/v1/slots?resource_id=nook-1&after=%d", after), "", "")
	if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("empty page: %d %s", rec.Code, rec.Body)
	}
}

func TestListSlotsInvalidQuery(t *testing.T) {
	handler, _ := seedAPI(t)
	for _, query := range []string{"", "resource_id=nook-1&limit=-1", "resource_id=nook-1&limit=101", "resource_id=nook-1&limit=x", "resource_id=nook-1&after=-1", "resource_id=nook-1&after=x", "resource_id=nook-1&after=", "resource_id=nook-1&after=2147483648"} {
		rec := request(handler, http.MethodGet, "/v1/slots?"+query, "", "")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("query %s: %d %s", query, rec.Code, rec.Body)
		}
	}
	rec := request(handler, http.MethodGet, "/v1/slots?resource_id=missing", "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing resource: %d", rec.Code)
	}
}
