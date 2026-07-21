package domain

import (
	"context"
	"time"
)

// OrderStatus es el estado del ciclo de vida de una orden.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// OrderItem es una línea de la orden. Guarda unitPrice como foto del precio al
// momento de la compra (no se recalcula si el producto cambia de precio después).
type OrderItem struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

// Order es una orden de compra de un usuario.
type Order struct {
	ID        string
	UserID    string
	Items     []OrderItem
	Total     float64
	Status    OrderStatus
	CreatedAt time.Time
}

// OrderRepository define la persistencia de órdenes (y sus ítems).
type OrderRepository interface {
	// Create inserta la orden y sus ítems.
	Create(ctx context.Context, o *Order) error
	// FindByID devuelve ErrOrderNotFound si no existe.
	FindByID(ctx context.Context, id string) (*Order, error)
	// ListByUser devuelve una página de órdenes del usuario y el total.
	ListByUser(ctx context.Context, userID string, offset, limit int) (items []*Order, total int, err error)
	// UpdateStatus cambia el estado de una orden.
	UpdateStatus(ctx context.Context, id string, status OrderStatus) error
}

// TxManager ejecuta fn dentro de una transacción de base de datos.
//
// Es la pieza que permite que "crear la orden + descontar el stock" ocurra de
// forma atómica sin que el use case conozca *sql.Tx: la transacción se inicia
// aquí, se coloca en el context, y los repositorios la detectan desde el context
// para operar dentro de ella. Si fn devuelve error, se hace rollback; si no,
// commit.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
