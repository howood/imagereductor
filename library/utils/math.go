package utils

import (
	"math"
)

func RoundFloat(input float64, decimals int) float64 {
	var pow float64 = 1
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(math.Floor((input*pow)+0.5)) / pow
}
