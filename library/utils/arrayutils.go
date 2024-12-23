package utils

import "strings"

// StringArrayContains checks contains string in string array.
func StringArrayContains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// StringArrayContainsForwardMatch checks contains string in string array with forward match.
func StringArrayContainsForwardMatch(arr []string, str string) bool {
	for _, v := range arr {
		if strings.Index(v, str) == 0 {
			return true
		}
	}
	return false
}

// IntArrayContains checks contains int in int array.
func IntArrayContains(arr []int, num int) bool {
	for _, v := range arr {
		if v == num {
			return true
		}
	}

	return false
}
