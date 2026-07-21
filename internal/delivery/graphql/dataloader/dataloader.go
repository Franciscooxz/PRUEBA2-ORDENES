// Package dataloader agrupa (batch) las cargas de entidades relacionadas para
// resolver el problema N+1 en los field resolvers de GraphQL.
//
// Sin esto, listar N órdenes y pedir el producto de cada ítem dispararía N
// consultas. Con el loader, todas las cargas del mismo tipo que ocurren en una
// petición se juntan en UNA sola consulta (FindByIDs con ANY(...)).
package dataloader

import (
	"context"
	"net/http"
	"time"

	"github.com/vikstrous/dataloadgen"

	"ordersapi/internal/domain"
)

type ctxKey struct{}

var loadersKey ctxKey

// Loaders agrupa un loader por cada relación que hay que resolver.
type Loaders struct {
	UserByID    *dataloadgen.Loader[string, *domain.User]
	ProductByID *dataloadgen.Loader[string, *domain.Product]
}

func newLoaders(users domain.UserRepository, products domain.ProductRepository) *Loaders {
	return &Loaders{
		UserByID:    dataloadgen.NewLoader(userBatch(users), dataloadgen.WithWait(2*time.Millisecond)),
		ProductByID: dataloadgen.NewLoader(productBatch(products), dataloadgen.WithWait(2*time.Millisecond)),
	}
}

func userBatch(repo domain.UserRepository) func(context.Context, []string) ([]*domain.User, []error) {
	return func(ctx context.Context, ids []string) ([]*domain.User, []error) {
		byID, err := repo.FindByIDs(ctx, ids)
		return orderResults(ids, byID, err, domain.ErrUserNotFound)
	}
}

func productBatch(repo domain.ProductRepository) func(context.Context, []string) ([]*domain.Product, []error) {
	return func(ctx context.Context, ids []string) ([]*domain.Product, []error) {
		byID, err := repo.FindByIDs(ctx, ids)
		return orderResults(ids, byID, err, domain.ErrProductNotFound)
	}
}

// orderResults coloca el resultado del repositorio en el mismo orden que las
// keys, como exige dataloadgen (values[i] corresponde a ids[i]).
func orderResults[T any](ids []string, byID map[string]*T, err error, notFound error) ([]*T, []error) {
	values := make([]*T, len(ids))
	errs := make([]error, len(ids))
	if err != nil {
		for i := range errs {
			errs[i] = err
		}
		return values, errs
	}
	for i, id := range ids {
		if v, ok := byID[id]; ok {
			values[i] = v
		} else {
			errs[i] = notFound
		}
	}
	return values, errs
}

// Middleware inyecta un juego de loaders NUEVO en el context por cada petición.
// Es importante que sea por petición: el caché de un loader debe vivir solo
// durante esa petición, no compartirse entre usuarios.
func Middleware(users domain.UserRepository, products domain.ProductRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loadersKey, newLoaders(users, products))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// For devuelve los loaders del context.
func For(ctx context.Context) *Loaders {
	loaders, _ := ctx.Value(loadersKey).(*Loaders)
	return loaders
}
