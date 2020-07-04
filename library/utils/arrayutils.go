package utils

import "strings"

func StringArrayContains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func StringArrayContainsForwardMatch(arr []string, str string) bool {
	for _, v := range arr {
		if strings.Index(v, str) == 0 {
			return true
		}
	}
	return false
}

func IntArrayContains(arr []int, num int) bool {
	for _, v := range arr {
		if v == num {
			return true
		}
	}
	return false
}
