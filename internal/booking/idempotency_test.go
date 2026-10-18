package booking

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestCreateBookingRequestKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want error
	}{
		{"missing", "", ErrInvalidRequestKey},
		{"blank", " \t\n", ErrInvalidRequestKey},
		{"maximum length", strings.Repeat("a", 128), nil},
		{"too long", strings.Repeat("a", 129), ErrInvalidRequestKey},
		{"length in bytes", strings.Repeat("é", 65), ErrInvalidRequestKey},
		{"reserved", "__legacy_booking__:42", ErrInvalidRequestKey},
		{"opaque", " Request-A ", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := CreateBookingInput{SlotID: 7, UserID: "alice", RequestKey: tt.key}
			if err := in.Validate(); !errors.Is(err, tt.want) {
				t.Fatalf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateBookingRequestHash(t *testing.T) {
	var first, retry CreateBookingInput
	if err := json.Unmarshal([]byte(`{"slot_id":7,"user_id":"alice"}`), &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{ "user_id": "alice", "slot_id": 7 }`), &retry); err != nil {
		t.Fatal(err)
	}
	first.RequestKey = "first"
	retry.RequestKey = "second"
	if got := first.RequestHash(); got != retry.RequestHash() || len(got) != 64 {
		t.Fatalf("equivalent payloads have different or invalid hashes: %q, %q", got, retry.RequestHash())
	}
	retry.SlotID = 8
	if first.RequestHash() == retry.RequestHash() {
		t.Fatal("different slots have the same hash")
	}
}

func TestCreateBookingRequestKeyExcludedFromJSON(t *testing.T) {
	var in CreateBookingInput
	if err := json.Unmarshal([]byte(`{"slot_id":7,"user_id":"tim","RequestKey":"forged","request_key":"forged"}`), &in); err != nil {
		t.Fatal(err)
	}

	if in.RequestKey != "" {
		t.Fatal("JSON supplied a request key")
	}

	in.RequestKey = "private-key"
	encoded, err := json.Marshal(in)

	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(encoded), "private-key") {
		t.Fatalf("request key exposed in JSON: %s", encoded)
	}
}
