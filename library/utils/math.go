package utils

import (
	"math"
)

// RoundFloat rounds input float64 with decimals
func RoundFloat(input float64, decimals int) float64 {
	var pow float64 = 1
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(math.Floor((input*pow)+0.5)) / pow
}

// InRanged return value in range
func InRanged(value, min, max float64) float64 {
	if value > max {
		return max
	}
	if value < min {
		return min
	}
	return value
}
