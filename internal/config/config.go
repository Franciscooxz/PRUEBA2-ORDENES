// Package config carga la configuración de la aplicación desde variables de
// entorno.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config agrupa los parámetros de arranque.
type Config struct {
	DatabaseURL   string
	Port          string
	JWTSecret     string
	JWTExpiration time.Duration
}

// Load lee la configuración del entorno (y de .env si existe) y valida lo
// obligatorio.
func Load() (Config, error) {
	// Carga .env si existe (comodidad en desarrollo local). En producción las
	// variables vienen del entorno, no hay .env, y se ignora el error.
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          getEnv("PORT", "8080"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiration: 24 * time.Hour,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("la variable de entorno DATABASE_URL es obligatoria")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("la variable de entorno JWT_SECRET es obligatoria")
	}

	// Expiración configurable (deseable): p. ej. JWT_EXPIRATION=1h, 30m, 24h.
	if v := os.Getenv("JWT_EXPIRATION"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("JWT_EXPIRATION inválido (%q): %w", v, err)
		}
		cfg.JWTExpiration = d
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
