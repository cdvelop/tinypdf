package fontManager

import "math"

func round(f float64) int {
	if f < 0 {
		return -int(math.Floor(-f + 0.5))
	}
	return int(math.Floor(f + 0.5))
}
