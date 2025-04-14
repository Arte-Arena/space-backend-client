package utils

import (
	"strconv"
)

func ParseIntOrDefault(s string, defaultValue int) (int, error) {
	if s == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue, err
	}

	return value, nil
}
