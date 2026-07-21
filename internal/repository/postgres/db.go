// Package postgres contiene la conexión a la base de datos y las
// implementaciones de los repositorios sobre PostgreSQL (vía pgx).
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool crea un pool de conexiones a PostgreSQL y verifica la conectividad con
// un Ping. El llamador es responsable de cerrar el pool (pool.Close()).
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creando el pool de conexiones: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}
	return pool, nil
}
