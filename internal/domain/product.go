package domain

import "context"

// Product es un producto del catálogo.
type Product struct {
	ID    string
	Name  string
	Price float64
	Stock int
}

// ProductFilter son los criterios opcionales del listado paginado. Un puntero en
// nil significa "sin filtrar por ese campo".
type ProductFilter struct {
	Name     *string
	MinPrice *float64
	MaxPrice *float64
}

// ProductRepository define la persistencia de productos.
//
// DecrementStock/IncrementStock existen para poder ajustar el stock dentro de la
// transacción de creación/cancelación de una orden: la transacción viaja por el
// context (ver TxManager), de modo que el use case no maneja *sql.Tx.
type ProductRepository interface {
	// FindByID devuelve ErrProductNotFound si no existe.
	FindByID(ctx context.Context, id string) (*Product, error)
	// FindByIDs devuelve los productos pedidos indexados por ID (carga por lotes
	// para el DataLoader).
	FindByIDs(ctx context.Context, ids []string) (map[string]*Product, error)
	// List devuelve una página de productos y el total que cumple el filtro.
	List(ctx context.Context, filter ProductFilter, offset, limit int) (items []*Product, total int, err error)
	// DecrementStock resta cantidad al stock; devuelve ErrInsufficientStock si
	// no hay suficiente. Debe ser atómico respecto a lecturas concurrentes.
	DecrementStock(ctx context.Context, productID string, quantity int) error
	// IncrementStock devuelve stock al inventario (al cancelar una orden).
	IncrementStock(ctx context.Context, productID string, quantity int) error
}
