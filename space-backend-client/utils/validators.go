package utils

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

const (
	MaxNameLength     = 50
	MinPasswordLength = 8
	MaxPasswordLength = 50
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

func ValidateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	nameLength := utf8.RuneCountInString(name)
	if nameLength > MaxNameLength {
		return errors.New("name cannot exceed 50 characters")
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	passwordLength := utf8.RuneCountInString(password)

	if passwordLength < MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	if passwordLength > MaxPasswordLength {
		return errors.New("password cannot exceed 50 characters")
	}

	return nil
}
