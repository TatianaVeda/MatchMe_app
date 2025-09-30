package utils

import (
	"errors"
	"regexp"
	"unicode"

	"github.com/sirupsen/logrus"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if the email address is valid using a regular expression.
func ValidateEmail(email string) error {
	// Use a regex to validate standard email format
	if !emailRegex.MatchString(email) {
		logrus.Debugf("ValidateEmail: invalid email format %s", email)
		return errors.New("invalid email format")
	}
	return nil
}

// ValidatePassword checks password complexity (length, letters, and digits).
func ValidatePassword(password string) error {
	// Enforce minimum length
	if len(password) < 8 {
		logrus.Debug("ValidatePassword: password is too short, must be at least 8 chars")
		return errors.New("password must be at least 8 characters")
	}
	var hasLetter, hasDigit bool
	// Check for at least one letter and one digit
	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		logrus.Debug("ValidatePassword: password does not contain required characters like letters and numbers")
		return errors.New("password must contain both letters and numbers")
	}
	return nil
}

// ValidateStringLength checks that a string does not exceed the maximum length.
func ValidateStringLength(str string, max int) error {
	// Compare string length to the allowed maximum
	if len(str) > max {
		logrus.Debugf("ValidateStringLength: string exceeds maximum length (%d > %d)", len(str), max)
		return errors.New("value is too long")
	}
	return nil
}
