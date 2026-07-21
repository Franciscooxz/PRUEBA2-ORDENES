package graphql

import (
	"context"
	"errors"
	"log"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"ordersapi/internal/delivery/graphql/middleware"
	"ordersapi/internal/domain"
)

// requireUser devuelve el userID autenticado o un error UNAUTHENTICATED si no hay
// sesión. Lo usan los resolvers que exigen token.
func requireUser(ctx context.Context) (string, error) {
	userID := middleware.UserIDFromContext(ctx)
	if userID == "" {
		return "", &gqlerror.Error{
			Message:    "autenticación requerida",
			Extensions: map[string]any{"code": "UNAUTHENTICATED"},
		}
	}
	return userID, nil
}

// intOr devuelve *p, o def si p es nil (para argumentos opcionales de paginación).
func intOr(p *int, def int) int {
	if p != nil {
		return *p
	}
	return def
}

// toGraphQLError traduce un error de dominio a *gqlerror.Error con una extensión
// `code` estable. Los errores no reconocidos no exponen su detalle al cliente.
func toGraphQLError(err error) error {
	if err == nil {
		return nil
	}

	var code string
	switch {
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrOrderNotFound):
		code = "NOT_FOUND"
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrEmptyOrder):
		code = "BAD_USER_INPUT"
	case errors.Is(err, domain.ErrInvalidCredentials):
		code = "UNAUTHENTICATED"
	case errors.Is(err, domain.ErrOrderNotOwned):
		code = "FORBIDDEN"
	case errors.Is(err, domain.ErrEmailAlreadyUsed),
		errors.Is(err, domain.ErrInsufficientStock),
		errors.Is(err, domain.ErrOrderNotCancelable):
		code = "CONFLICT"
	default:
		// Error inesperado: se registra y se devuelve un mensaje genérico.
		log.Printf("error interno no controlado: %v", err)
		return &gqlerror.Error{
			Message:    "error interno del servidor",
			Extensions: map[string]any{"code": "INTERNAL_SERVER_ERROR"},
		}
	}

	return &gqlerror.Error{
		Message:    err.Error(),
		Extensions: map[string]any{"code": code},
	}
}
