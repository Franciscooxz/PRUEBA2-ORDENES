package domain

import (
	"net/mail"
	"strings"
)

const minPasswordLength = 8

// NormalizeEmail deja el email en minúsculas y sin espacios alrededor.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateEmail comprueba que sea una dirección de correo simple y válida
// (rechaza formatos como "Nombre <a@b.com>").
func ValidateEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword exige una longitud mínima.
func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return ErrWeakPassword
	}
	return nil
}
