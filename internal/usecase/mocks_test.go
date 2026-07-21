package usecase_test

import (
	"context"

	"ordersapi/internal/domain"
)

// Mocks hechos a mano con campos-función: cada test configura solo el
// comportamiento que necesita. Sin librerías externas.

// --- UserRepository ---

type mockUserRepo struct {
	createFn      func(ctx context.Context, u *domain.User) error
	findByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	findByIDFn    func(ctx context.Context, id string) (*domain.User, error)
	findByIDsFn   func(ctx context.Context, ids []string) (map[string]*domain.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	return m.createFn(ctx, u)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.findByEmailFn(ctx, email)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockUserRepo) FindByIDs(ctx context.Context, ids []string) (map[string]*domain.User, error) {
	return m.findByIDsFn(ctx, ids)
}

// --- ProductRepository ---

type mockProductRepo struct {
	findByIDFn       func(ctx context.Context, id string) (*domain.Product, error)
	findByIDsFn      func(ctx context.Context, ids []string) (map[string]*domain.Product, error)
	listFn           func(ctx context.Context, f domain.ProductFilter, offset, limit int) ([]*domain.Product, int, error)
	decrementStockFn func(ctx context.Context, id string, qty int) error
	incrementStockFn func(ctx context.Context, id string, qty int) error
}

func (m *mockProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockProductRepo) FindByIDs(ctx context.Context, ids []string) (map[string]*domain.Product, error) {
	return m.findByIDsFn(ctx, ids)
}
func (m *mockProductRepo) List(ctx context.Context, f domain.ProductFilter, offset, limit int) ([]*domain.Product, int, error) {
	return m.listFn(ctx, f, offset, limit)
}
func (m *mockProductRepo) DecrementStock(ctx context.Context, id string, qty int) error {
	return m.decrementStockFn(ctx, id, qty)
}
func (m *mockProductRepo) IncrementStock(ctx context.Context, id string, qty int) error {
	return m.incrementStockFn(ctx, id, qty)
}

// --- OrderRepository ---

type mockOrderRepo struct {
	createFn       func(ctx context.Context, o *domain.Order) error
	findByIDFn     func(ctx context.Context, id string) (*domain.Order, error)
	listByUserFn   func(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, int, error)
	updateStatusFn func(ctx context.Context, id string, s domain.OrderStatus) error
}

func (m *mockOrderRepo) Create(ctx context.Context, o *domain.Order) error {
	return m.createFn(ctx, o)
}
func (m *mockOrderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockOrderRepo) ListByUser(ctx context.Context, userID string, offset, limit int) ([]*domain.Order, int, error) {
	return m.listByUserFn(ctx, userID, offset, limit)
}
func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id string, s domain.OrderStatus) error {
	return m.updateStatusFn(ctx, id, s)
}

// --- TxManager de prueba: ejecuta fn directamente (sin transacción real) ---

type fakeTxManager struct{}

func (fakeTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// --- TokenService ---

type mockTokenService struct {
	generateFn func(userID string) (string, error)
	verifyFn   func(token string) (string, error)
}

func (m *mockTokenService) Generate(userID string) (string, error) { return m.generateFn(userID) }
func (m *mockTokenService) Verify(token string) (string, error)    { return m.verifyFn(token) }
