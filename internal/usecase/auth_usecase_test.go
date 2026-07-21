package usecase_test

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"ordersapi/internal/domain"
	"ordersapi/internal/usecase"
)

func TestAuth_Register_Success(t *testing.T) {
	var created *domain.User
	repo := &mockUserRepo{
		createFn: func(_ context.Context, u *domain.User) error {
			u.ID = "user-1"
			created = u
			return nil
		},
	}
	tokens := &mockTokenService{generateFn: func(uid string) (string, error) { return "tok-" + uid, nil }}
	uc := usecase.NewAuthUseCase(repo, tokens)

	token, user, err := uc.Register(context.Background(), "  User@Example.com ", "password123")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if token != "tok-user-1" {
		t.Errorf("token = %q", token)
	}
	if user.Email != "user@example.com" {
		t.Errorf("email no normalizado: %q", user.Email)
	}
	if created == nil || created.PasswordHash == "password123" {
		t.Error("la contraseña debe guardarse hasheada, no en texto plano")
	}
}

func TestAuth_Register_WeakPassword(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *domain.User) error {
			t.Fatal("no debe crear el usuario con una contraseña débil")
			return nil
		},
	}
	uc := usecase.NewAuthUseCase(repo, &mockTokenService{})

	_, _, err := uc.Register(context.Background(), "a@b.com", "corta")
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Errorf("error = %v, se esperaba ErrWeakPassword", err)
	}
}

func TestAuth_Login_SuccessAndWrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correcta"), bcrypt.MinCost)
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: "u1", Email: email, PasswordHash: string(hash)}, nil
		},
	}
	tokens := &mockTokenService{generateFn: func(string) (string, error) { return "tok", nil }}
	uc := usecase.NewAuthUseCase(repo, tokens)

	if _, _, err := uc.Login(context.Background(), "a@b.com", "correcta"); err != nil {
		t.Errorf("login correcto falló: %v", err)
	}

	_, _, err := uc.Login(context.Background(), "a@b.com", "incorrecta")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("error = %v, se esperaba ErrInvalidCredentials", err)
	}
}

func TestAuth_Login_UserNotFound_HidesExistence(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	uc := usecase.NewAuthUseCase(repo, &mockTokenService{})

	// Aunque el usuario no exista, debe devolver ErrInvalidCredentials (no revelar
	// si el email está registrado).
	_, _, err := uc.Login(context.Background(), "noexiste@b.com", "x")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("error = %v, se esperaba ErrInvalidCredentials", err)
	}
}
