// Command ordersapi levanta el servidor GraphQL de la API de órdenes.
// Es el composition root: aquí se ensamblan todas las dependencias concretas.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"

	"ordersapi/internal/auth"
	"ordersapi/internal/config"
	graphqldelivery "ordersapi/internal/delivery/graphql"
	"ordersapi/internal/delivery/graphql/generated"
	"ordersapi/internal/delivery/graphql/middleware"
	"ordersapi/internal/repository/postgres"
	"ordersapi/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Aplica las migraciones al arrancar (cómodo para docker-compose y desarrollo).
	if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migraciones: %v", err)
	}

	pool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Repositorios, servicios y transacción.
	users := postgres.NewUserRepository(pool)
	products := postgres.NewProductRepository(pool)
	orders := postgres.NewOrderRepository(pool)
	txm := postgres.NewTxManager(pool)
	tokens := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration)

	// Casos de uso.
	authUC := usecase.NewAuthUseCase(users, tokens)
	productUC := usecase.NewProductUseCase(products)
	orderUC := usecase.NewOrderUseCase(orders, products, txm)

	resolver := &graphqldelivery.Resolver{
		AuthUC:    authUC,
		ProductUC: productUC,
		OrderUC:   orderUC,
		Users:     users,
	}

	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("Órdenes GraphQL", "/query"))
	// El middleware de auth envuelve el endpoint: si hay token válido, deja el
	// userID en el context para los resolvers.
	mux.Handle("/query", middleware.Auth(tokens)(srv))

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("servidor listo en http://localhost:%s/ (GraphQL en /query)", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error al iniciar el servidor: %v", err)
		}
	}()

	// Cierre ordenado ante SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("cerrando servidor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("error durante el cierre del servidor: %v", err)
	}
	log.Println("servidor detenido")
}
