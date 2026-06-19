package helper

import "testing"

func TestCalculatePercentage(t *testing.T) {
	cases := []struct {
		name     string
		current  float64
		previous float64
		want     float64
	}{
		{"previous zero, current positive => 100", 100, 0, 100},
		{"previous zero, current zero => 0", 0, 0, 0},
		{"growth", 150, 100, 50},
		{"decline", 50, 100, -50},
		{"current zero, previous positive => -100", 0, 100, -100},
		{"no change", 100, 100, 0},
		{"rounds to one decimal", 105, 90, 16.7},   // 16.666... -> 16.7
		{"rounds half down to one decimal", 130, 120, 8.3}, // 8.333... -> 8.3
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculatePercentage(tc.current, tc.previous)
			if got != tc.want {
				t.Fatalf("CalculatePercentage(%v, %v) = %v, want %v", tc.current, tc.previous, got, tc.want)
			}
		})
	}
}
