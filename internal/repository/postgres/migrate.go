package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // registra el driver "pgx" para database/sql
	"github.com/pressly/goose/v3"

	"ordersapi/migrations"
)

// RunMigrations aplica todas las migraciones pendientes contra la base de datos.
//
// goose trabaja sobre database/sql, así que abrimos una conexión con el driver
// pgx (stdlib) usando los archivos SQL embebidos (paquete migrations). Se usa
// una conexión aparte del pool y se cierra al terminar.
func RunMigrations(databaseURL string) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("configurando el dialecto de goose: %w", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("abriendo conexión para migraciones: %w", err)
	}
	defer db.Close()

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("aplicando migraciones: %w", err)
	}
	return nil
}
