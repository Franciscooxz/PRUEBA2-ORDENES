// Package config carga la configuración de la aplicación desde variables de
// entorno.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config agrupa los parámetros de arranque. Se irá ampliando por fase (JWT, etc.).
type Config struct {
	DatabaseURL string
	Port        string
}

// Load lee la configuración del entorno y valida lo obligatorio.
func Load() (Config, error) {
	// Carga .env si existe (comodidad en desarrollo local). En producción las
	// variables vienen del entorno, no hay .env, y se ignora el error.
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        getEnv("PORT", "8080"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("la variable de entorno DATABASE_URL es obligatoria")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
