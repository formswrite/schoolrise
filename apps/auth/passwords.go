package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const passwordHashCost = bcrypt.DefaultCost + 2

var (
	ErrEmptyPassword    = errors.New("auth: empty password")
	ErrPasswordMismatch = errors.New("auth: password mismatch")
)

func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), passwordHashCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func VerifyPassword(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if err == nil {
		return nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPasswordMismatch
	}

	return err
}
