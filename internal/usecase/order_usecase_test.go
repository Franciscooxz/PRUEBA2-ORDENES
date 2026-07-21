package usecase_test

import (
	"context"
	"errors"
	"testing"

	"ordersapi/internal/domain"
	"ordersapi/internal/usecase"
)

func TestOrder_Create_Success(t *testing.T) {
	product := &domain.Product{ID: "p1", Name: "X", Price: 10, Stock: 5}
	var decremented int
	products := &mockProductRepo{
		findByIDFn:       func(_ context.Context, _ string) (*domain.Product, error) { return product, nil },
		decrementStockFn: func(_ context.Context, _ string, qty int) error { decremented = qty; return nil },
	}
	var saved *domain.Order
	orders := &mockOrderRepo{
		createFn: func(_ context.Context, o *domain.Order) error {
			o.ID = "o1"
			saved = o
			return nil
		},
	}
	uc := usecase.NewOrderUseCase(orders, products, fakeTxManager{})

	o, err := uc.Create(context.Background(), "u1", []usecase.OrderItemInput{{ProductID: "p1", Quantity: 2}})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if o.Total != 20 {
		t.Errorf("total = %v, se esperaba 20 (10 x 2)", o.Total)
	}
	if o.Status != domain.OrderStatusPending {
		t.Errorf("status = %v, se esperaba PENDING", o.Status)
	}
	if len(o.Items) != 1 || o.Items[0].UnitPrice != 10 {
		t.Errorf("ítems mal construidos: %+v", o.Items)
	}
	if decremented != 2 {
		t.Errorf("se descontó %d, se esperaba 2", decremented)
	}
	if saved == nil {
		t.Error("no se persistió la orden")
	}
}

func TestOrder_Create_EmptyOrder(t *testing.T) {
	uc := usecase.NewOrderUseCase(&mockOrderRepo{}, &mockProductRepo{}, fakeTxManager{})
	_, err := uc.Create(context.Background(), "u1", nil)
	if !errors.Is(err, domain.ErrEmptyOrder) {
		t.Errorf("error = %v, se esperaba ErrEmptyOrder", err)
	}
}

func TestOrder_Create_InsufficientStock(t *testing.T) {
	products := &mockProductRepo{
		findByIDFn: func(_ context.Context, _ string) (*domain.Product, error) {
			return &domain.Product{ID: "p1", Price: 10, Stock: 1}, nil
		},
		decrementStockFn: func(_ context.Context, _ string, _ int) error { return domain.ErrInsufficientStock },
	}
	orders := &mockOrderRepo{
		createFn: func(_ context.Context, _ *domain.Order) error {
			t.Fatal("no debe crear la orden si falta stock")
			return nil
		},
	}
	uc := usecase.NewOrderUseCase(orders, products, fakeTxManager{})

	_, err := uc.Create(context.Background(), "u1", []usecase.OrderItemInput{{ProductID: "p1", Quantity: 5}})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("error = %v, se esperaba ErrInsufficientStock", err)
	}
}

func TestOrder_Cancel_NotOwned(t *testing.T) {
	orders := &mockOrderRepo{
		findByIDFn: func(_ context.Context, _ string) (*domain.Order, error) {
			return &domain.Order{ID: "o1", UserID: "otro", Status: domain.OrderStatusPending}, nil
		},
	}
	uc := usecase.NewOrderUseCase(orders, &mockProductRepo{}, fakeTxManager{})

	_, err := uc.Cancel(context.Background(), "u1", "o1")
	if !errors.Is(err, domain.ErrOrderNotOwned) {
		t.Errorf("error = %v, se esperaba ErrOrderNotOwned", err)
	}
}

func TestOrder_Cancel_NotPending(t *testing.T) {
	orders := &mockOrderRepo{
		findByIDFn: func(_ context.Context, _ string) (*domain.Order, error) {
			return &domain.Order{ID: "o1", UserID: "u1", Status: domain.OrderStatusCancelled}, nil
		},
	}
	uc := usecase.NewOrderUseCase(orders, &mockProductRepo{}, fakeTxManager{})

	_, err := uc.Cancel(context.Background(), "u1", "o1")
	if !errors.Is(err, domain.ErrOrderNotCancelable) {
		t.Errorf("error = %v, se esperaba ErrOrderNotCancelable", err)
	}
}
