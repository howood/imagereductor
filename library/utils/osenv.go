package utils

import (
	"os"
	"strconv"
)

//GetOsEnv is get os env with default value
func GetOsEnv(key, defaultdata string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultdata
}

//GetOsEnvInt is get os env with default value converting to int
func GetOsEnvInt(key string, defaultdata int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intval, err := strconv.Atoi(value); err == nil {
			return intval
		}
	}
	return defaultdata
}
