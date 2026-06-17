package storage

import (
	"errors"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want time.Duration
		err  error
	}{
		{name: "minutes", in: "15m", want: 15 * time.Minute},
		{name: "hours", in: "2h", want: 2 * durHour},
		{name: "days", in: "3d", want: 3 * durDay},
		{name: "weeks", in: "4w", want: 4 * durWeek},
		{name: "years", in: "5y", want: 5 * durYear},
		{name: "invalid format", in: "abc", err: errInvalidDuration},
		{name: "unknown unit", in: "30q", err: errInvalidUnit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.in)
			if !errors.Is(err, tt.err) {
				t.Fatalf("parseDuration(%q) error = %v, want %v", tt.in, err, tt.err)
			}
			if got != tt.want {
				t.Fatalf("parseDuration(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
