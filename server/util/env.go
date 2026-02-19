package util

import (
	"os"
	"strconv"
)

func EnvString(key, fallback string) string {
	if v := os.Getenv(key); v != "" { // Check if the environment variable is set and not empty
		return v
	}
	return fallback 
}

func EnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback 
	}
	return n
}
