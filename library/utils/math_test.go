package utils_test

import (
	"testing"

	"github.com/howood/imagereductor/library/utils"
)

func Test_RoundFloat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    float64
		decimals int
		expected float64
	}{
		{"round to 2 decimals", 1.23456, 2, 1.23},
		{"round up to 2 decimals", 1.235, 2, 1.24},
		{"round to 0 decimals", 1.5, 0, 2},
		{"round negative number", -1.4, 0, -1},
		{"already rounded", 2.5, 1, 2.5},
		{"zero", 0, 3, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := utils.RoundFloat(tc.input, tc.decimals)
			if got != tc.expected {
				t.Fatalf("RoundFloat(%v, %d) = %v, want %v", tc.input, tc.decimals, got, tc.expected)
			}
		})
	}
}

func Test_InRanged(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{"value in range", 5, 1, 10, 5},
		{"value over max", 15, 1, 10, 10},
		{"value below min", -5, 1, 10, 1},
		{"value equals max", 10, 1, 10, 10},
		{"value equals min", 1, 1, 10, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := utils.InRanged(tc.value, tc.min, tc.max)
			if got != tc.expected {
				t.Fatalf("InRanged(%v, %v, %v) = %v, want %v", tc.value, tc.min, tc.max, got, tc.expected)
			}
		})
	}
}
