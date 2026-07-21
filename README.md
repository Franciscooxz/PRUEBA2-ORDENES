# API de Órdenes · Go + GraphQL + PostgreSQL + JWT

API de gestión de órdenes de compra. Maneja usuarios, productos y órdenes con sus
ítems, con **autenticación JWT stateless**, **creación de órdenes transaccional**
(valida y descuenta stock en una sola transacción) y **DataLoader** para resolver
el problema N+1.

> Prueba técnica 2 — Wind Consulting.

---

## Stack

| Herramienta | Uso |
|---|---|
| **Go 1.26** | Lenguaje |
| **[gqlgen](https://gqlgen.com/)** | Servidor GraphQL *schema-first* |
| **PostgreSQL** vía **[pgx/v5](https://github.com/jackc/pgx)** | Base de datos |
| **[goose](https://github.com/pressly/goose)** | Migraciones (embebidas en el binario) |
| **[golang-jwt/v5](https://github.com/golang-jwt/jwt)** | Autenticación JWT |
| **bcrypt** | Hash de contraseñas |
| **[dataloadgen](https://github.com/vikstrous/dataloadgen)** | DataLoader (N+1) |

---

## Cómo correr

### Opción A — Docker (recomendada)

Levanta la API y PostgreSQL juntos. La app **aplica migraciones y siembra productos
automáticamente** al arrancar.

```bash
docker compose up --build
```

- **Playground GraphQL**: `http://localhost:8080/`
- **Endpoint**: `http://localhost:8080/query`

### Opción B — Local (PostgreSQL propio)

```bash
# 1. Copia el ejemplo de variables y ajusta la contraseña
cp .env.example .env

# 2. Crea la base de datos (una vez)
createdb wind_orders

# 3. Levanta la app (aplica migraciones + seed y arranca el servidor)
go run ./cmd
```

### Variables de entorno

| Variable | Obligatoria | Descripción |
|---|---|---|
| `DATABASE_URL` | ✅ | Cadena de conexión a PostgreSQL |
| `JWT_SECRET` | ✅ | Secreto para firmar los JWT |
| `JWT_EXPIRATION` | ❌ | Expiración del token (ej. `24h`, `30m`). Por defecto 24h |
| `PORT` | ❌ | Puerto del servidor (por defecto 8080) |

---

## Arquitectura (Clean Architecture, 4 capas)

La dependencia siempre apunta **hacia el dominio**.

```
┌──────────────────────────────────────────────────────────────┐
│  DELIVERY  (internal/delivery/graphql)                          │
│  Resolvers + middleware de auth + DataLoaders + mapeo de errores│
└────────────────────────────┬───────────────────────────────────┘
                             │ usa
┌────────────────────────────▼───────────────────────────────────┐
│  USE CASE  (internal/usecase)                                    │
│  Auth, productos, órdenes. Orquesta la transacción.              │
└────────────────────────────┬───────────────────────────────────┘
                             │ depende de (interfaces)
┌────────────────────────────▼───────────────────────────────────┐
│  DOMAIN  (internal/domain)                                       │
│  Entidades · interfaces de repositorio · TokenService · TxManager│
│  · errores tipados · validaciones. Go puro.                      │
└────────────────────────────△───────────────────────────────────┘
                             │ implementan
┌────────────────────────────┴───────────────────────────────────┐
│  ADAPTERS/REPOSITORY  (internal/repository/postgres, internal/auth)│
│  Repositorios pgx · TxManager · servicio JWT.                     │
└──────────────────────────────────────────────────────────────────┘
```

### Decisión de diseño: manejo de transacciones

Crear una orden debe ser **atómico**: validar el stock, descontarlo e insertar la
orden con sus ítems tienen que ocurrir juntos o no ocurrir. La solución es un
**`TxManager`** (interface en el dominio) cuyo método `Do(ctx, fn)` inicia una
transacción, la coloca en el `context` y ejecuta `fn`. Cada repositorio obtiene su
ejecutor de consultas con un helper `querierFrom(ctx)`: si hay una transacción en
el `context`, la usa; si no, usa el pool. Así el caso de uso escribe
`txManager.Do(ctx, func(ctx){ descontar stock; crear orden })` y todo corre en la
**misma transacción sin que el use case conozca nunca `*pgx.Tx`**. Si `fn` devuelve
error, se hace *rollback*; si no, *commit*.

---

## Estructura del proyecto

```
.
├── cmd/
│   ├── main.go                  # servidor (composition root)
│   └── migrate/main.go          # comando de migraciones
├── internal/
│   ├── domain/                  # entidades, interfaces, errores, validaciones
│   ├── usecase/                 # auth, product, order (+ tests)
│   ├── auth/                    # servicio JWT (implementa domain.TokenService)
│   ├── repository/postgres/     # repos pgx, TxManager, migraciones, seed (+ tests integración)
│   └── delivery/graphql/
│       ├── resolver.go, schema.resolvers.go, mapper.go
│       ├── middleware/auth.go   # extrae el JWT al context
│       ├── dataloader/          # loaders (N+1)
│       ├── generated/, model/   # código generado por gqlgen
├── migrations/                  # SQL de goose (embebido)
├── graph/schema.graphqls
├── docker-compose.yml, Dockerfile
├── .golangci.yml
└── README.md
```

---

## API GraphQL

`register` y `login` son públicas. **El resto exige** el header
`Authorization: Bearer <token>`.

### Autenticación

```graphql
mutation {
  register(input: { email: "ana@example.com", password: "password123" }) {
    token
    user { id email }
  }
}
```
```graphql
mutation {
  login(input: { email: "ana@example.com", password: "password123" }) {
    token
  }
}
```

Luego se envía el token en el header:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### Productos (listado paginado con filtro)

```graphql
query {
  products(filter: { name: "teclado", minPrice: 10, maxPrice: 100 }, page: 1, pageSize: 20) {
    total
    items { id name price stock }
  }
}
```

### Órdenes

```graphql
# Crear una orden (valida y descuenta stock en una transacción)
mutation {
  createOrder(input: { items: [{ productId: "ID_PRODUCTO", quantity: 2 }] }) {
    id
    total
    status
    items { quantity unitPrice product { name } }
    user { email }
  }
}
```
```graphql
# Mis órdenes / detalle / cancelar
query    { myOrders(page: 1, pageSize: 10) { total items { id status total } } }
query    { order(id: "ID") { id status items { product { name } quantity } } }
mutation { cancelOrder(id: "ID") { id status } }   # restaura el stock
```

### Manejo de errores tipado

Los errores de dominio se traducen a errores GraphQL con un `code` estable en
`extensions`:

| Código | Situación |
|---|---|
| `UNAUTHENTICATED` | Falta token válido, o credenciales inválidas en login |
| `FORBIDDEN` | La orden no pertenece al usuario |
| `NOT_FOUND` | Usuario/producto/orden inexistente |
| `BAD_USER_INPUT` | Validación (email, contraseña, cantidad, orden vacía) |
| `CONFLICT` | Email ya registrado, stock insuficiente, orden no cancelable |
| `INTERNAL_SERVER_ERROR` | Error inesperado (sin exponer detalle) |

---

## DataLoader (N+1)

Al listar N órdenes con sus ítems, resolver el producto de cada ítem y el usuario
de cada orden dispararía N consultas. Los **DataLoaders** (`internal/delivery/graphql/dataloader`)
agrupan todas las cargas del mismo tipo dentro de una petición en **una sola
consulta** (`SELECT ... WHERE id = ANY($1)`). Los loaders se crean **nuevos por
cada petición** (vía middleware), para que su caché no se comparta entre usuarios.

---

## Tests

```bash
# Unitarios (dominio + use cases con mocks) — no requieren base de datos
go test ./...

# Con los tests de integración de los repositorios (requieren PostgreSQL):
# En Windows PowerShell:
$env:TEST_DATABASE_URL="postgres://postgres:PASSWORD@localhost:5432/wind_orders?sslmode=disable"; go test ./...
# En bash:
TEST_DATABASE_URL="postgres://..." go test ./...
```

Cobertura: **dominio/validaciones** (email, contraseña, estado de orden),
**use cases** (auth, órdenes, productos, con mocks) y **2 tests de integración**
sobre los repositorios (incluye la verificación del *rollback* de la transacción).
Si `TEST_DATABASE_URL` no está definida, los de integración se omiten.

Lint: `golangci-lint run ./...` (config en `.golangci.yml`).

---

## Migraciones

Se aplican automáticamente al arrancar la app. Para aplicarlas manualmente:

```bash
go run ./cmd/migrate
```

---

## Requisitos cubiertos

**Obligatorios**
- [x] Go, gqlgen, PostgreSQL (pgx)
- [x] Clean Architecture con 4 capas
- [x] JWT en un servicio detrás de una interface (no hardcodeado)
- [x] Creación de orden + descuento de stock en **una sola transacción**, sin exponer `*pgx.Tx` al use case
- [x] Migraciones con goose
- [x] `docker-compose.yml` que levanta app + Postgres
- [x] 16 tests: use cases, dominio/validaciones y 2 de integración
- [x] Errores de dominio tipados mapeados a `code` + `message` en GraphQL

**Deseables**
- [x] Expiración del token configurable por variable de entorno
- [x] `golangci-lint` con configuración en el repo
- [x] Seed de productos iniciales al arrancar
- [x] DataLoader para el N+1
