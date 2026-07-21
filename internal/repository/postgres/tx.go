package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ordersapi/internal/domain"
)

// querier abstrae lo común entre el pool y una transacción: ambos saben ejecutar
// consultas. Así los repositorios escriben el mismo código dentro o fuera de una
// transacción.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// txKey es la clave (privada) bajo la que se guarda la transacción en el context.
type txKey struct{}

// TxManager implementa domain.TxManager sobre un pool de pgx.
type TxManager struct {
	pool *pgxpool.Pool
}

var _ domain.TxManager = (*TxManager)(nil)

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// Do ejecuta fn dentro de una transacción. Coloca la transacción en el context;
// los repositorios la detectan con querierFrom y operan dentro de ella. Si fn
// devuelve error se hace rollback; si no, commit. Así "crear orden + descontar
// stock" es atómico sin que el use case conozca *pgx.Tx.
func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	// Si ya hay una transacción en el context, la reutilizamos en lugar de abrir
	// otra: así una llamada anidada a Do se une a la transacción existente en
	// vez de crear una independiente (que rompería la atomicidad).
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback de seguridad ante panics; tras un Commit exitoso es un no-op.
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// querierFrom devuelve la transacción activa del context, o el pool si no hay
// ninguna. Es lo que permite que un repositorio funcione igual dentro y fuera de
// una transacción.
func querierFrom(ctx context.Context, pool *pgxpool.Pool) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// isUniqueViolation indica si el error es una violación de restricción UNIQUE
// (código SQLSTATE 23505 de PostgreSQL).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
