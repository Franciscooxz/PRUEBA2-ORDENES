package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"ordersapi/internal/domain"
	"ordersapi/internal/repository/postgres"
)

// setupTestDB conecta a la base de TEST_DATABASE_URL, aplica migraciones y
// limpia las tablas. Si la variable no está definida, se omiten estos tests
// (para no fallar en entornos sin base de datos).
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL no definida; se omiten los tests de integración")
	}
	if err := postgres.RunMigrations(dsn); err != nil {
		t.Fatalf("migraciones: %v", err)
	}
	pool, err := postgres.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE order_items, orders, products, users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func TestIntegration_ProductRepository(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	repo := postgres.NewProductRepository(pool)

	var id string
	if err := pool.QueryRow(ctx,
		`INSERT INTO products (name, price, stock) VALUES ('Teclado', 50, 3) RETURNING id::text`,
	).Scan(&id); err != nil {
		t.Fatal(err)
	}

	// FindByID
	p, err := repo.FindByID(ctx, id)
	if err != nil || p.Name != "Teclado" || p.Stock != 3 {
		t.Fatalf("FindByID = %+v, err=%v", p, err)
	}

	// DecrementStock atómico: pedir más de lo disponible -> ErrInsufficientStock
	// y el stock NO cambia.
	if err := repo.DecrementStock(ctx, id, 10); !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("DecrementStock(10) = %v, se esperaba ErrInsufficientStock", err)
	}
	if p, _ = repo.FindByID(ctx, id); p.Stock != 3 {
		t.Errorf("el stock cambió a %d tras un descuento fallido", p.Stock)
	}

	// Descuento válido: 3 - 2 = 1.
	if err := repo.DecrementStock(ctx, id, 2); err != nil {
		t.Fatal(err)
	}
	if p, _ = repo.FindByID(ctx, id); p.Stock != 1 {
		t.Errorf("stock = %d, se esperaba 1", p.Stock)
	}

	// List con filtro por nombre.
	name := "Tecl"
	items, total, err := repo.List(ctx, domain.ProductFilter{Name: &name}, 0, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Errorf("List: total=%d items=%d err=%v", total, len(items), err)
	}
}

func TestIntegration_OrderRepository_Transaction(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	products := postgres.NewProductRepository(pool)
	orders := postgres.NewOrderRepository(pool)
	users := postgres.NewUserRepository(pool)
	txm := postgres.NewTxManager(pool)

	// Seed: usuario y producto (stock 5).
	u := &domain.User{Email: "int@test.com", PasswordHash: "h"}
	if err := users.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	var prodID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO products (name, price, stock) VALUES ('P', 10, 5) RETURNING id::text`,
	).Scan(&prodID); err != nil {
		t.Fatal(err)
	}

	// Transacción exitosa: descontar stock + crear orden.
	order := &domain.Order{
		UserID: u.ID, Status: domain.OrderStatusPending, Total: 20,
		Items: []domain.OrderItem{{ProductID: prodID, Quantity: 2, UnitPrice: 10}},
	}
	if err := txm.Do(ctx, func(ctx context.Context) error {
		if err := products.DecrementStock(ctx, prodID, 2); err != nil {
			return err
		}
		return orders.Create(ctx, order)
	}); err != nil {
		t.Fatalf("transacción: %v", err)
	}

	// FindByID carga la orden con sus ítems.
	got, err := orders.FindByID(ctx, order.ID)
	if err != nil || len(got.Items) != 1 || got.Items[0].Quantity != 2 {
		t.Fatalf("FindByID = %+v, err=%v", got, err)
	}
	if p, _ := products.FindByID(ctx, prodID); p.Stock != 3 {
		t.Errorf("stock = %d, se esperaba 3", p.Stock)
	}

	// Transacción que falla: debe hacer ROLLBACK (el stock no queda descontado).
	sentinel := errors.New("fallo simulado")
	err = txm.Do(ctx, func(ctx context.Context) error {
		if err := products.DecrementStock(ctx, prodID, 1); err != nil {
			return err
		}
		return sentinel // fuerza el rollback
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("error = %v, se esperaba el fallo simulado", err)
	}
	if p, _ := products.FindByID(ctx, prodID); p.Stock != 3 {
		t.Errorf("rollback falló: stock = %d, se esperaba que siguiera en 3", p.Stock)
	}
}
