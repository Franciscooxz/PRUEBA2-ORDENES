package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ordersapi/internal/domain"
)

// ProductRepository implementa domain.ProductRepository sobre PostgreSQL.
type ProductRepository struct {
	pool *pgxpool.Pool
}

var _ domain.ProductRepository = (*ProductRepository)(nil)

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	q := querierFrom(ctx, r.pool)
	var p domain.Product
	err := q.QueryRow(ctx,
		`SELECT id::text, name, price::float8, stock FROM products WHERE id = $1::uuid`, id,
	).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) List(ctx context.Context, f domain.ProductFilter, offset, limit int) ([]*domain.Product, int, error) {
	q := querierFrom(ctx, r.pool)

	// WHERE dinámico: solo se agregan las condiciones de los filtros presentes.
	conds := []string{"TRUE"}
	args := []any{}
	n := 0
	if f.Name != nil {
		n++
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", n))
		args = append(args, "%"+*f.Name+"%")
	}
	if f.MinPrice != nil {
		n++
		conds = append(conds, fmt.Sprintf("price >= $%d", n))
		args = append(args, *f.MinPrice)
	}
	if f.MaxPrice != nil {
		n++
		conds = append(conds, fmt.Sprintf("price <= $%d", n))
		args = append(args, *f.MaxPrice)
	}
	where := strings.Join(conds, " AND ")

	var total int
	if err := q.QueryRow(ctx, "SELECT count(*) FROM products WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Paginación: los dos últimos parámetros son LIMIT y OFFSET.
	sql := fmt.Sprintf(
		`SELECT id::text, name, price::float8, stock FROM products
		 WHERE %s ORDER BY name LIMIT $%d OFFSET $%d`,
		where, n+1, n+2,
	)
	args = append(args, limit, offset)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*domain.Product, 0, limit)
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			return nil, 0, err
		}
		items = append(items, &p)
	}
	return items, total, rows.Err()
}

// DecrementStock descuenta stock de forma atómica: la condición `stock >= $2`
// garantiza que nunca quede negativo aunque haya concurrencia. Si no afecta
// ninguna fila, es porque no hay stock suficiente (el use case ya validó que el
// producto existe antes de llegar aquí).
func (r *ProductRepository) DecrementStock(ctx context.Context, productID string, quantity int) error {
	q := querierFrom(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE products SET stock = stock - $2 WHERE id = $1::uuid AND stock >= $2`,
		productID, quantity,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}

// IncrementStock devuelve stock al inventario (al cancelar una orden).
func (r *ProductRepository) IncrementStock(ctx context.Context, productID string, quantity int) error {
	q := querierFrom(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE products SET stock = stock + $2 WHERE id = $1::uuid`,
		productID, quantity,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}
