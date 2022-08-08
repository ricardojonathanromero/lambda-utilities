package utils

import "os"

// GetEnv returns a string value/*
func GetEnv(name, value string) string {
	if val, ok := os.LookupEnv(name); ok {
		return val
	}

	return value
}
