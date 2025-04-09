package utils

import (
	"errors"
	"regexp"
	"unicode"

	"github.com/sirupsen/logrus"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail проверяет корректность формата email.
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		logrus.Debugf("ValidateEmail: неверный формат email %s", email)
		return errors.New("некорректный формат email")
	}
	return nil
}

// ValidatePassword проверяет, что пароль имеет минимальную длину и содержит хотя бы одну цифру и одну букву.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		logrus.Debug("ValidatePassword: пароль слишком короткий")
		return errors.New("пароль должен содержать минимум 8 символов")
	}
	var hasLetter, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		logrus.Debug("ValidatePassword: пароль не содержит необходимых символов")
		return errors.New("пароль должен содержать как буквы, так и цифры")
	}
	return nil
}

// ValidateStringLength проверяет, что строка не превышает заданную максимальную длину.
func ValidateStringLength(str string, max int) error {
	if len(str) > max {
		logrus.Debugf("ValidateStringLength: строка превышает максимальную длину (%d > %d)", len(str), max)
		return errors.New("слишком длинное значение")
	}
	return nil
}
