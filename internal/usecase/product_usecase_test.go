package usecase_test

import (
	"context"
	"testing"

	"ordersapi/internal/domain"
	"ordersapi/internal/usecase"
)

func TestProduct_List_NormalizaPaginacion(t *testing.T) {
	var gotOffset, gotLimit int
	products := &mockProductRepo{
		listFn: func(_ context.Context, _ domain.ProductFilter, offset, limit int) ([]*domain.Product, int, error) {
			gotOffset, gotLimit = offset, limit
			return []*domain.Product{{ID: "p1"}}, 1, nil
		},
	}
	uc := usecase.NewProductUseCase(products)

	// page=0 y pageSize=0 deben normalizarse a page=1, pageSize=20 (offset 0).
	res, err := uc.List(context.Background(), domain.ProductFilter{}, 0, 0)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if gotOffset != 0 || gotLimit != 20 {
		t.Errorf("offset/limit = %d/%d, se esperaba 0/20", gotOffset, gotLimit)
	}
	if res.Page != 1 || res.PageSize != 20 || res.Total != 1 {
		t.Errorf("resultado = page:%d size:%d total:%d", res.Page, res.PageSize, res.Total)
	}
}
