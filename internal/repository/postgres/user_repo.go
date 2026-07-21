package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ordersapi/internal/domain"
)

// UserRepository implementa domain.UserRepository sobre PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

var _ domain.UserRepository = (*UserRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	q := querierFrom(ctx, r.pool)
	err := q.QueryRow(ctx,
		`INSERT INTO users (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id::text, created_at`,
		u.Email, u.PasswordHash,
	).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailAlreadyUsed
		}
		return err
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findBy(ctx, "email = $1", email)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.findBy(ctx, "id = $1::uuid", id)
}

func (r *UserRepository) FindByIDs(ctx context.Context, ids []string) (map[string]*domain.User, error) {
	result := make(map[string]*domain.User, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	q := querierFrom(ctx, r.pool)
	rows, err := q.Query(ctx,
		`SELECT id::text, email, password_hash, created_at FROM users WHERE id = ANY($1::uuid[])`,
		ids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
			return nil, err
		}
		result[u.ID] = &u
	}
	return result, rows.Err()
}

func (r *UserRepository) findBy(ctx context.Context, cond string, arg any) (*domain.User, error) {
	q := querierFrom(ctx, r.pool)
	var u domain.User
	err := q.QueryRow(ctx,
		`SELECT id::text, email, password_hash, created_at FROM users WHERE `+cond,
		arg,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
