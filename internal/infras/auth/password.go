package auth

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordMinCharacters = 8
	passwordMaxCharacters = 64
	passwordMaxBytes      = 72
)

var (
	ErrPasswordLength        = errors.New("password must contain 8 to 64 characters")
	ErrPasswordTooManyBytes  = errors.New("password exceeds bcrypt 72-byte limit")
	ErrPasswordMissingLetter = errors.New("password must contain at least one English letter")
	ErrPasswordMissingDigit  = errors.New("password must contain at least one digit")
	ErrPasswordInvalidUTF8   = errors.New("password must be valid UTF-8")
	ErrPasswordMismatch      = errors.New("password does not match")
)

func ValidatePassword(password string) error {
	if !utf8.ValidString(password) {
		return ErrPasswordInvalidUTF8
	}

	characterCount := utf8.RuneCountInString(password)
	if characterCount < passwordMinCharacters || characterCount > passwordMaxCharacters {
		return ErrPasswordLength
	}
	if len([]byte(password)) > passwordMaxBytes {
		return ErrPasswordTooManyBytes
	}

	var hasLetter, hasDigit bool
	for _, character := range password {
		switch {
		case character >= 'a' && character <= 'z', character >= 'A' && character <= 'Z':
			hasLetter = true
		case character >= '0' && character <= '9':
			hasDigit = true
		}
	}
	if !hasLetter {
		return ErrPasswordMissingLetter
	}
	if !hasDigit {
		return ErrPasswordMissingDigit
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}
