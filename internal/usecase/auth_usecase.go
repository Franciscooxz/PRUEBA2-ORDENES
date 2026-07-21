package usecase

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"ordersapi/internal/domain"
)

// AuthUseCase define las operaciones de autenticación.
type AuthUseCase interface {
	Register(ctx context.Context, email, password string) (token string, user *domain.User, err error)
	Login(ctx context.Context, email, password string) (token string, user *domain.User, err error)
}

type authUseCase struct {
	users  domain.UserRepository
	tokens domain.TokenService
}

func NewAuthUseCase(users domain.UserRepository, tokens domain.TokenService) AuthUseCase {
	return &authUseCase{users: users, tokens: tokens}
}

// dummyHash es un hash bcrypt válido usado para gastar el mismo tiempo de CPU
// cuando el email no existe, de modo que el login no revele por timing si un
// email está registrado.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-attack-mitigation"), bcrypt.DefaultCost)

func (uc *authUseCase) Register(ctx context.Context, email, password string) (string, *domain.User, error) {
	email = domain.NormalizeEmail(email)
	if err := domain.ValidateEmail(email); err != nil {
		return "", nil, err
	}
	if err := domain.ValidatePassword(password); err != nil {
		return "", nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	u := &domain.User{Email: email, PasswordHash: string(hash)}
	if err := uc.users.Create(ctx, u); err != nil {
		return "", nil, err // ErrEmailAlreadyUsed si el email ya existe
	}

	token, err := uc.tokens.Generate(u.ID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	email = domain.NormalizeEmail(email)

	u, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Seguridad: no revelamos si el email existe. Comparamos contra un
			// hash dummy para gastar un tiempo similar al del camino con usuario
			// (evita la enumeración de emails por timing) y devolvemos el mismo
			// error que con contraseña incorrecta.
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return "", nil, domain.ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := uc.tokens.Generate(u.ID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}
