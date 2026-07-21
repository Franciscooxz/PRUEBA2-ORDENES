package domain

// TokenService genera y valida tokens de autenticación. Se define como interface
// en el dominio para que el caso de uso no dependa de la librería de JWT
// concreta: la implementación se inyecta desde fuera (ver internal/auth).
type TokenService interface {
	// Generate crea un token firmado para el usuario dado.
	Generate(userID string) (string, error)
	// Verify valida un token y devuelve el userID que contiene.
	Verify(token string) (userID string, err error)
}
