package domain_test

import (
	"errors"
	"testing"

	"ordersapi/internal/domain"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"válido", "user@example.com", false},
		{"válido con subdominio", "a.b@mail.co.uk", false},
		{"sin arroba", "userexample.com", true},
		{"vacío", "", true},
		{"con nombre", "Nombre <a@b.com>", true},
		{"solo dominio", "@example.com", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateEmail(tc.email)
			switch {
			case tc.wantErr && !errors.Is(err, domain.ErrInvalidEmail):
				t.Errorf("ValidateEmail(%q) = %v, se esperaba ErrInvalidEmail", tc.email, err)
			case !tc.wantErr && err != nil:
				t.Errorf("ValidateEmail(%q) = %v, se esperaba nil", tc.email, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"suficiente", "password123", false},
		{"justo 8", "12345678", false},
		{"corta", "corta", true},
		{"vacía", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidatePassword(tc.password)
			switch {
			case tc.wantErr && !errors.Is(err, domain.ErrWeakPassword):
				t.Errorf("ValidatePassword(%q) = %v, se esperaba ErrWeakPassword", tc.password, err)
			case !tc.wantErr && err != nil:
				t.Errorf("ValidatePassword(%q) = %v, se esperaba nil", tc.password, err)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := domain.NormalizeEmail("  USER@Example.COM  "); got != "user@example.com" {
		t.Errorf("NormalizeEmail = %q, se esperaba %q", got, "user@example.com")
	}
}
