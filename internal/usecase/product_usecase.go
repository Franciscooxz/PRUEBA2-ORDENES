package usecase

import (
	"context"

	"ordersapi/internal/domain"
)

// ProductUseCase expone la lectura de productos (los usuarios normales solo
// consultan; no crean ni editan productos).
type ProductUseCase interface {
	List(ctx context.Context, filter domain.ProductFilter, page, pageSize int) (Paginated[*domain.Product], error)
	Get(ctx context.Context, id string) (*domain.Product, error)
}

type productUseCase struct {
	products domain.ProductRepository
}

func NewProductUseCase(products domain.ProductRepository) ProductUseCase {
	return &productUseCase{products: products}
}

func (uc *productUseCase) List(ctx context.Context, filter domain.ProductFilter, page, pageSize int) (Paginated[*domain.Product], error) {
	offset, limit, p, ps := normalizePagination(page, pageSize)
	items, total, err := uc.products.List(ctx, filter, offset, limit)
	if err != nil {
		return Paginated[*domain.Product]{}, err
	}
	return Paginated[*domain.Product]{Items: items, Total: total, Page: p, PageSize: ps}, nil
}

func (uc *productUseCase) Get(ctx context.Context, id string) (*domain.Product, error) {
	return uc.products.FindByID(ctx, id)
}
