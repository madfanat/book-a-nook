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
			want:   ErrInvalidBookingStatus,
		},
		{
			name:   "invalid status",
			status: "invalid",
			want:   ErrInvalidBookingStatus,
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

func TestCreateResourceInputValidate(t *testing.T) {
	tests := []struct {
		name  string
		input CreateResourceInput
		want  error
	}{
		{
			name: "valid",
			input: CreateResourceInput{
				ID: "nook-1",
			},
			want: nil,
		},
		{
			name: "empty ID",
			input: CreateResourceInput{
				ID: "",
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
				ResourceID: "nook-1",
				StartsAt:   now,
				EndsAt:     now.Add(time.Hour),
			},
			want: nil,
		},
		{
			name: "zero StartsAt",
			input: CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   time.Time{},
				EndsAt:     now.Add(time.Hour),
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "zero EndsAt",
			input: CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now.Add(time.Hour),
				EndsAt:     time.Time{},
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "EndsAt before StartsAt",
			input: CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now.Add(time.Hour),
				EndsAt:     now,
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "EndsAt equals StartsAt",
			input: CreateSlotInput{
				ResourceID: "nook-1",
				StartsAt:   now,
				EndsAt:     now,
			},
			want: ErrInvalidTimeRange,
		},
		{
			name: "empty ResourceID",
			input: CreateSlotInput{
				ResourceID: "",
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
				UserID: "tim",
			},
			want: nil,
		},
		{
			name: "zero SlotID",
			input: CreateBookingInput{
				SlotID: 0,
				UserID: "tim",
			},
			want: ErrInvalidSlotID,
		},
		{
			name: "negative SlotID",
			input: CreateBookingInput{
				SlotID: -1,
				UserID: "tim",
			},
			want: ErrInvalidSlotID,
		},
		{
			name: "empty UserID",
			input: CreateBookingInput{
				SlotID: 1,
				UserID: "",
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
