package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "valid", password: "abc12345"},
		{name: "valid unicode", password: "密码abc123"},
		{name: "too short", password: "abc123", wantErr: ErrPasswordLength},
		{name: "too many characters", password: strings.Repeat("a", 64) + "1", wantErr: ErrPasswordLength},
		{name: "too many bytes", password: "a1" + strings.Repeat("界", 24), wantErr: ErrPasswordTooManyBytes},
		{name: "missing English letter", password: "密码123456", wantErr: ErrPasswordMissingLetter},
		{name: "missing digit", password: "abcdefgh", wantErr: ErrPasswordMissingDigit},
		{name: "invalid utf8", password: string([]byte{'a', 'b', 'c', '1', '2', '3', '4', 0xff}), wantErr: ErrPasswordInvalidUTF8},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePassword(test.password)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ValidatePassword() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	password := "abc12345"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == password {
		t.Fatal("HashPassword() returned the plaintext password")
	}
	if err := VerifyPassword(hash, password); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if err := VerifyPassword(hash, "wrong123"); !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("VerifyPassword() error = %v, want %v", err, ErrPasswordMismatch)
	}
}

func TestHashPasswordRejectsInvalidPassword(t *testing.T) {
	if _, err := HashPassword("short1"); !errors.Is(err, ErrPasswordLength) {
		t.Fatalf("HashPassword() error = %v, want %v", err, ErrPasswordLength)
	}
}

func TestVerifyPasswordRejectsInvalidHash(t *testing.T) {
	if err := VerifyPassword("not-a-bcrypt-hash", "abc12345"); err == nil {
		t.Fatal("VerifyPassword() expected an invalid hash error")
	}
}
