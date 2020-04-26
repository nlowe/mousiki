package audio

import "math"

func RelativeDBToPercent(db float64) float64 {
	return math.Pow(10, db/20)
}
