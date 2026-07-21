package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ordersapi/internal/domain"
)

// OrderRepository implementa domain.OrderRepository sobre PostgreSQL.
type OrderRepository struct {
	pool *pgxpool.Pool
}

var _ domain.OrderRepository = (*OrderRepository)(nil)

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// rowScanner es lo común entre pgx.Row (QueryRow) y pgx.Rows (Query): ambos
// saben escanear una fila.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(s rowScanner) (*domain.Order, error) {
	var o domain.Order
	var status string
	if err := s.Scan(&o.ID, &o.UserID, &o.Total, &status, &o.CreatedAt); err != nil {
		return nil, err
	}
	o.Status = domain.OrderStatus(status)
	return &o, nil
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	q := querierFrom(ctx, r.pool)
	err := q.QueryRow(ctx,
		`INSERT INTO orders (user_id, total, status)
		 VALUES ($1::uuid, $2, $3::order_status)
		 RETURNING id::text, created_at`,
		o.UserID, o.Total, string(o.Status),
	).Scan(&o.ID, &o.CreatedAt)
	if err != nil {
		return err
	}
	for _, it := range o.Items {
		if _, err := q.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			 VALUES ($1::uuid, $2::uuid, $3, $4)`,
			o.ID, it.ProductID, it.Quantity, it.UnitPrice,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	q := querierFrom(ctx, r.pool)
	o, err := scanOrder(q.QueryRow(ctx,
		`SELECT id::text, user_id::text, total::float8, status::text, created_at
		 FROM orders WHERE id = $1::uuid`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	itemsByOrder, err := r.itemsForOrders(ctx, q, []string{o.ID})
	if err != nil {
		return nil, err
	}
	o.Items = itemsByOrder[o.ID]
	return o, nil
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, int, error) {
	q := querierFrom(ctx, r.pool)

	var total int
	if err := q.QueryRow(ctx,
		`SELECT count(*) FROM orders WHERE user_id = $1::uuid`, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := q.Query(ctx,
		`SELECT id::text, user_id::text, total::float8, status::text, created_at
		 FROM orders WHERE user_id = $1::uuid ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0, limit)
	ids := make([]string, 0, limit)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
		ids = append(ids, o.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// Cerramos rows para liberar la conexión ANTES de la siguiente consulta.
	// Si no, mantener el result-set abierto mientras itemsForOrders pide otra
	// conexión puede agotar el pool y bloquearse bajo concurrencia. Close es
	// idempotente, así que el defer posterior queda como no-op.
	rows.Close()

	// Carga todos los ítems de la página en una sola consulta (evita N+1 aquí).
	itemsByOrder, err := r.itemsForOrders(ctx, q, ids)
	if err != nil {
		return nil, 0, err
	}
	for _, o := range orders {
		o.Items = itemsByOrder[o.ID]
	}
	return orders, total, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	q := querierFrom(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE orders SET status = $2::order_status WHERE id = $1::uuid`,
		id, string(status),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// itemsForOrders carga los ítems de varias órdenes en una sola consulta y los
// agrupa por order_id.
func (r *OrderRepository) itemsForOrders(ctx context.Context, q querier, orderIDs []string) (map[string][]domain.OrderItem, error) {
	result := make(map[string][]domain.OrderItem, len(orderIDs))
	if len(orderIDs) == 0 {
		return result, nil
	}
	rows, err := q.Query(ctx,
		`SELECT order_id::text, product_id::text, quantity, unit_price::float8
		 FROM order_items WHERE order_id = ANY($1::uuid[]) ORDER BY id`,
		orderIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var orderID string
		var it domain.OrderItem
		if err := rows.Scan(&orderID, &it.ProductID, &it.Quantity, &it.UnitPrice); err != nil {
			return nil, err
		}
		result[orderID] = append(result[orderID], it)
	}
	return result, rows.Err()
}
