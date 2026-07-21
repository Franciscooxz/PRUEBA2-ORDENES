package graphql

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import "ordersapi/internal/usecase"

// Resolver es la raíz de inyección de dependencias de la capa GraphQL. Depende de
// las interfaces de los casos de uso, no de implementaciones concretas.
// Las relaciones (Order.user, OrderItem.product) se resuelven vía DataLoader,
// inyectado por middleware en el context.
type Resolver struct {
	AuthUC    usecase.AuthUseCase
	ProductUC usecase.ProductUseCase
	OrderUC   usecase.OrderUseCase
}
