package event

import (
	"testing"
	"time"
)

func TestCalcTotalHours(t *testing.T) {
	day := func(h, m int) time.Time {
		return time.Date(2026, time.June, 1, h, m, 0, 0, time.UTC)
	}

	cases := []struct {
		name     string
		checkin  time.Time
		checkout time.Time
		want     float64
	}{
		{"08:00 -> 12:30 = 4.50", day(8, 0), day(12, 30), 4.5},
		{"08:15 -> 17:45 = 9.50", day(8, 15), day(17, 45), 9.5},
		{"09:00 -> 18:00 = 9.00", day(9, 0), day(18, 0), 9},
		{"08:00 -> 17:30 = 9.50", day(8, 0), day(17, 30), 9.5},
		{"same time = 0.00", day(10, 0), day(10, 0), 0},
		{"rounds to 2 decimals (10 min = 0.17)", day(9, 0), day(9, 10), 0.17},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := calcTotalHours(tc.checkin, tc.checkout)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("calcTotalHours(%v, %v) = %v, want %v", tc.checkin, tc.checkout, got, tc.want)
			}
		})
	}
}

func TestCalcTotalHours_CheckoutBeforeCheckin(t *testing.T) {
	checkin := time.Date(2026, time.June, 1, 17, 0, 0, 0, time.UTC)
	checkout := time.Date(2026, time.June, 1, 8, 0, 0, 0, time.UTC)

	if _, err := calcTotalHours(checkin, checkout); err == nil {
		t.Fatalf("expected error when checkout is before checkin, got nil")
	}
}
