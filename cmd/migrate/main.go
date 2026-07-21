// Command migrate aplica las migraciones de base de datos pendientes.
// Uso: go run ./cmd/migrate  (con DATABASE_URL en el entorno o en .env).
package main

import (
	"log"

	"ordersapi/internal/config"
	"ordersapi/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("error en migraciones: %v", err)
	}
	log.Println("migraciones aplicadas correctamente")
}
