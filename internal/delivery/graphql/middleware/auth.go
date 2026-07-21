// Package middleware contiene middlewares HTTP de la capa de delivery.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"ordersapi/internal/domain"
)

type contextKey struct{}

// userIDKey es la clave (privada) bajo la que se guarda el userID en el context.
var userIDKey contextKey

// Auth extrae el token de "Authorization: Bearer <token>". Si es válido, coloca
// el userID en el context. NO rechaza las peticiones sin token: eso lo deciden
// los resolvers (así register/login pueden ser públicos y el resto exigir sesión).
func Auth(tokens domain.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
				if userID, err := tokens.Verify(strings.TrimSpace(raw)); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext devuelve el userID autenticado, o "" si no hay sesión.
func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}
