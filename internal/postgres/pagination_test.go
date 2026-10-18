package postgres_test

import (
	"book-a-nook/internal/booking"
	"testing"
	"time"
)

func TestListSlotsPagination(t *testing.T) {
	ctx, store, _ := newTestStore(t)
	seedResources(t, ctx, store, "first", "second", "empty")
	start := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	var ids []int
	for i := range 105 {
		for _, resource := range []string{"first", "second"} {
			slot, err := store.CreateSlot(ctx, booking.CreateSlotInput{ResourceID: resource, StartsAt: start.Add(time.Duration(i) * time.Hour), EndsAt: start.Add(time.Duration(i+1) * time.Hour)})
			if err != nil {
				t.Fatal(err)
			}
			if resource == "first" {
				ids = append(ids, slot.ID)
			}
		}
	}
	after, seen := 0, 0
	for {
		page, err := store.ListSlots(ctx, booking.ListSlotsInput{ResourceID: "first", Limit: 100, After: after})
		if err != nil {
			t.Fatal(err)
		}
		if len(page) == 0 {
			break
		}
		if len(page) > 100 {
			t.Fatal("unbounded page")
		}
		for _, slot := range page {
			if seen >= len(ids) || slot.ID != ids[seen] || slot.ResourceID != "first" {
				t.Fatalf("unexpected slot %+v at offset %d", slot, seen)
			}
			seen++
		}
		after = page[len(page)-1].ID
	}
	if seen != len(ids) {
		t.Fatalf("listed %d of %d slots", seen, len(ids))
	}
	page, err := store.ListSlots(ctx, booking.ListSlotsInput{ResourceID: "empty", Limit: 100})
	if err != nil || page == nil || len(page) != 0 {
		t.Fatalf("empty resource: %v %v", page, err)
	}
}
