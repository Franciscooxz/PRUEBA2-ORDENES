package usecase

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Paginated es un resultado paginado de la capa de aplicación (genérico para
// reutilizarlo con productos, órdenes, etc.).
type Paginated[T any] struct {
	Items    []T
	Total    int
	Page     int
	PageSize int
}

// normalizePagination aplica valores por defecto y límites, y calcula el offset.
// page < 1 -> 1; pageSize fuera de rango se ajusta a [1, maxPageSize].
func normalizePagination(page, pageSize int) (offset, limit, normPage, normPageSize int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return (page - 1) * pageSize, pageSize, page, pageSize
}
