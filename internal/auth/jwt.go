// Package auth implementa el servicio de tokens JWT (domain.TokenService).
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ordersapi/internal/domain"
)

// JWTService firma y valida JSON Web Tokens con HMAC-SHA256.
type JWTService struct {
	secret     []byte
	expiration time.Duration
}

var _ domain.TokenService = (*JWTService)(nil)

func NewJWTService(secret string, expiration time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), expiration: expiration}
}

func (s *JWTService) Generate(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID, // el userID viaja en el claim "sub"
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *JWTService) Verify(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		// Solo aceptamos HMAC: evita el ataque de confusión de algoritmo
		// (un token firmado con "none" o RS256 sería rechazado).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", errors.New("token inválido")
	}
	return claims.Subject, nil
}
