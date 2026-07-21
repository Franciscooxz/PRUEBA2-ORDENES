// Package migrations embebe los archivos .sql de migración en el binario, para
// que goose los aplique sin depender de la ruta del sistema de archivos.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
