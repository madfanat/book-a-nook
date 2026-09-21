package booking

import (
	"errors"
	"testing"
	"time"
)

func TestStatusValidate(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   error
	}{
		{
			name:   "valid active",
			status: StatusActive,
			want:   nil,
		},
		{
			name:   "valid cancelled",
			status: StatusCancelled,
			want:   nil,
		},
		{
			name:   "empty status",
			status: "",
			want:   ErrInvalidStatus,
		},
		{
			name:   "invalid status",
			status: "invalid",
			want:   ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Validate()
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateSlotInputValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input CreateSlotInput
		want  error
	}{
		{
			name: "valid",
			input: CreateSlotInput{
				ResourceID: 1,
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			},
			want: nil,
		},
		{
			name: "zero StartsAt",
			input: CreateSlotInput{
				ResourceID: 1,
				StartsAt:   time.Time{},
				EndsAt:     now.Add(time.Hour),
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "zero EndsAt",
			input: CreateSlotInput{
				ResourceID: 1,
				StartsAt:   now.Add(time.Hour),
				EndsAt:     time.Time{},
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "EndsAt before StartsAt",
			input: CreateSlotInput{
				ResourceID: 1,
				StartsAt:   now.Add(time.Hour),
				EndsAt:     now,
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "EndsAt equals StartsAt",
			input: CreateSlotInput{
				ResourceID: 1,
				StartsAt:   now,
				EndsAt:     now,
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "zero ResourceID",
			input: CreateSlotInput{
				ResourceID: 0,
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			},
			want: ErrInvalidResourceID,
		},
		{
			name: "negative ResourceID",
			input: CreateSlotInput{
				ResourceID: -1,
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			},
			want: ErrInvalidResourceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateBookingInputValidate(t *testing.T) {
	tests := []struct {
		name  string
		input CreateBookingInput
		want  error
	}{
		{
			name: "valid",
			input: CreateBookingInput{
				SlotID: 1,
				UserID: 1,
			},
			want: nil,
		},
		{
			name: "zero SlotID",
			input: CreateBookingInput{
				SlotID: 0,
				UserID: 1,
			},
			want: ErrInvalidSlotID,
		},
		{
			name: "negative SlotID",
			input: CreateBookingInput{
				SlotID: -1,
				UserID: 1,
			},
			want: ErrInvalidSlotID,
		},
		{
			name: "zero UserID",
			input: CreateBookingInput{
				SlotID: 1,
				UserID: 0,
			},
			want: ErrInvalidUserID,
		},
		{
			name: "negative UserID",
			input: CreateBookingInput{
				SlotID: 1,
				UserID: -1,
			},
			want: ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}
