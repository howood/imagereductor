package utils

import (
	"math"
)

// RoundFloat rounds input float64 with decimals.
func RoundFloat(input float64, decimals int) float64 {
	var pow float64 = 1
	for range decimals {
		pow *= 10
	}
	//nolint:mnd
	return float64(math.Floor((input*pow)+0.5)) / pow
}

// InRanged return value in range.
func InRanged(value, minimun, maximum float64) float64 {
	if value > maximum {
		return maximum
	}
	if value < minimun {
		return minimun
	}
	return value
}
