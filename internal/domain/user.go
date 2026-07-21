// Package domain contiene el núcleo del negocio: entidades e interfaces.
// No depende de GraphQL, base de datos ni frameworks; solo de Go estándar.
package domain

import (
	"context"
	"time"
)

// User es un usuario registrado. PasswordHash nunca se expone por GraphQL.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// UserRepository define la persistencia de usuarios.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	// FindByEmail devuelve ErrUserNotFound si no existe.
	FindByEmail(ctx context.Context, email string) (*User, error)
	// FindByID devuelve ErrUserNotFound si no existe.
	FindByID(ctx context.Context, id string) (*User, error)
	// FindByIDs devuelve los usuarios pedidos indexados por ID (carga por lotes
	// para el DataLoader). Los IDs inexistentes simplemente no aparecen en el mapa.
	FindByIDs(ctx context.Context, ids []string) (map[string]*User, error)
}
