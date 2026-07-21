package graphql

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"ordersapi/internal/domain"
	"ordersapi/internal/usecase"
)

// Resolver es la raíz de inyección de dependencias de la capa GraphQL. Depende de
// las interfaces de los casos de uso, no de implementaciones concretas.
type Resolver struct {
	AuthUC    usecase.AuthUseCase
	ProductUC usecase.ProductUseCase
	OrderUC   usecase.OrderUseCase
	// Users se usa en el field resolver Order.user (en la fase de DataLoader se
	// reemplaza por un loader que agrupa las cargas).
	Users domain.UserRepository
}
