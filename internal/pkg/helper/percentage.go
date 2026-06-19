package helper

import "math"

// CalculatePercentage returns the month-over-month growth/decline percentage of
// current relative to previous, rounded to one decimal place.
//
// Rules:
//   - previous == 0 && current  > 0  => 100
//   - previous == 0 && current == 0  => 0
//   - otherwise                      => ((current - previous) / previous) * 100
func CalculatePercentage(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}

	pct := ((current - previous) / previous) * 100
	return math.Round(pct*10) / 10
}
