package account

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72

	SessionLifetime = 7 * 24 * time.Hour
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash []byte    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}

type Credentials struct {
	Email    string
	Password string
}

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrEmailTaken         = fmt.Errorf("email already registered")
	ErrUnauthorized       = fmt.Errorf("unauthorized")
	ErrInvalidInput       = fmt.Errorf("invalid input")
)

func NormalizeEmail(value string) (string, error) {

	value = strings.ToLower(strings.TrimSpace(value))

	if len(value) > 254 {
		return "", ErrInvalidInput
	}

	address, err := mail.ParseAddress(value)

	if err != nil || address.Address != value || !strings.Contains(value, "@") {
		return "", ErrInvalidInput
	}

	return value, nil
}

func ValidatePassword(value string) error {

	if utf8.RuneCountInString(value) < MinPasswordLength || len(value) > MaxPasswordLength {
		return ErrInvalidInput
	}

	return nil
}
