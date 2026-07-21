package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedProducts inserta un catálogo inicial de productos si la tabla está vacía.
// Es idempotente: si ya hay productos, no hace nada.
func SeedProducts(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM products`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO products (name, price, stock) VALUES
		('Teclado mecánico', 49.90, 50),
		('Mouse inalámbrico', 25.50, 100),
		('Monitor 27 pulgadas', 189.99, 20),
		('Auriculares Bluetooth', 79.00, 40),
		('Webcam HD', 35.00, 60)`)
	return err
}
