package domain

import "errors"

// Errores de dominio tipados (sentinel errors). Las capas superiores los
// identifican con errors.Is y los mapean a un código de error GraphQL.
var (
	// Usuarios / auth
	ErrUserNotFound       = errors.New("usuario no encontrado")
	ErrEmailAlreadyUsed   = errors.New("el email ya está registrado")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrInvalidEmail       = errors.New("email inválido")
	ErrWeakPassword       = errors.New("la contraseña debe tener al menos 8 caracteres")

	// Productos
	ErrProductNotFound   = errors.New("producto no encontrado")
	ErrInsufficientStock = errors.New("stock insuficiente")

	// Órdenes
	ErrOrderNotFound      = errors.New("orden no encontrada")
	ErrOrderNotOwned      = errors.New("la orden no pertenece al usuario")
	ErrOrderNotCancelable = errors.New("solo se pueden cancelar órdenes en estado PENDING")
	ErrEmptyOrder         = errors.New("la orden debe tener al menos un ítem")
	ErrInvalidQuantity    = errors.New("la cantidad debe ser mayor que 0")
)
