package usecase

import (
	"context"

	"ordersapi/internal/domain"
)

// OrderItemInput es un ítem solicitado al crear una orden (DTO de aplicación).
type OrderItemInput struct {
	ProductID string
	Quantity  int
}

// OrderUseCase expone las operaciones de órdenes del usuario autenticado.
type OrderUseCase interface {
	Create(ctx context.Context, userID string, items []OrderItemInput) (*domain.Order, error)
	ListByUser(ctx context.Context, userID string, page, pageSize int) (Paginated[*domain.Order], error)
	GetByID(ctx context.Context, userID, orderID string) (*domain.Order, error)
	Cancel(ctx context.Context, userID, orderID string) (*domain.Order, error)
}

type orderUseCase struct {
	orders   domain.OrderRepository
	products domain.ProductRepository
	tx       domain.TxManager
}

func NewOrderUseCase(orders domain.OrderRepository, products domain.ProductRepository, tx domain.TxManager) OrderUseCase {
	return &orderUseCase{orders: orders, products: products, tx: tx}
}

// Create arma la orden y descuenta el stock en UNA sola transacción. Como todo
// corre dentro de uc.tx.Do, si cualquier ítem falla (producto inexistente o sin
// stock) se hace rollback completo: no queda stock descontado a medias.
func (uc *orderUseCase) Create(ctx context.Context, userID string, inputs []OrderItemInput) (*domain.Order, error) {
	if len(inputs) == 0 {
		return nil, domain.ErrEmptyOrder
	}

	var order *domain.Order
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		items := make([]domain.OrderItem, 0, len(inputs))
		var total float64

		for _, in := range inputs {
			if in.Quantity <= 0 {
				return domain.ErrInvalidQuantity
			}
			p, err := uc.products.FindByID(ctx, in.ProductID)
			if err != nil {
				return err // ErrProductNotFound
			}
			// Descuento atómico: ErrInsufficientStock si no alcanza.
			if err := uc.products.DecrementStock(ctx, p.ID, in.Quantity); err != nil {
				return err
			}
			items = append(items, domain.OrderItem{
				ProductID: p.ID,
				Quantity:  in.Quantity,
				UnitPrice: p.Price, // foto del precio actual
			})
			total += p.Price * float64(in.Quantity)
		}

		order = &domain.Order{
			UserID: userID,
			Items:  items,
			Total:  total,
			Status: domain.OrderStatusPending,
		}
		return uc.orders.Create(ctx, order)
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (uc *orderUseCase) ListByUser(ctx context.Context, userID string, page, pageSize int) (Paginated[*domain.Order], error) {
	offset, limit, p, ps := normalizePagination(page, pageSize)
	items, total, err := uc.orders.ListByUser(ctx, userID, offset, limit)
	if err != nil {
		return Paginated[*domain.Order]{}, err
	}
	return Paginated[*domain.Order]{Items: items, Total: total, Page: p, PageSize: ps}, nil
}

// GetByID devuelve la orden solo si pertenece al usuario autenticado.
func (uc *orderUseCase) GetByID(ctx context.Context, userID, orderID string) (*domain.Order, error) {
	o, err := uc.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err // ErrOrderNotFound
	}
	if o.UserID != userID {
		return nil, domain.ErrOrderNotOwned
	}
	return o, nil
}

// Cancel cancela una orden PENDING propia y restaura el stock, todo en una
// transacción.
func (uc *orderUseCase) Cancel(ctx context.Context, userID, orderID string) (*domain.Order, error) {
	var result *domain.Order
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		o, err := uc.orders.FindByID(ctx, orderID)
		if err != nil {
			return err
		}
		if o.UserID != userID {
			return domain.ErrOrderNotOwned
		}
		if !o.Status.CanBeCancelled() {
			return domain.ErrOrderNotCancelable
		}

		// Devolver el stock de cada ítem al inventario.
		for _, it := range o.Items {
			if err := uc.products.IncrementStock(ctx, it.ProductID, it.Quantity); err != nil {
				return err
			}
		}
		if err := uc.orders.UpdateStatus(ctx, o.ID, domain.OrderStatusCancelled); err != nil {
			return err
		}
		o.Status = domain.OrderStatusCancelled
		result = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
