-- +goose Up
-- Esquema inicial: usuarios, productos, órdenes y sus ítems.

-- Los montos usan NUMERIC (exacto para dinero); las fechas TIMESTAMPTZ (instante
-- absoluto, correcto para un sistema con posibles zonas horarias distintas).

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name  TEXT          NOT NULL,
    price NUMERIC(12,2) NOT NULL CHECK (price > 0),
    stock INT           NOT NULL CHECK (stock >= 0)
);

-- Estado de la orden como ENUM: conjunto fijo y pequeño, con integridad garantizada.
CREATE TYPE order_status AS ENUM ('PENDING', 'CONFIRMED', 'CANCELLED');

CREATE TABLE orders (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID          NOT NULL REFERENCES users(id),
    total      NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total >= 0),
    status     order_status  NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user ON orders(user_id);

CREATE TABLE order_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID          NOT NULL REFERENCES products(id),
    quantity   INT           NOT NULL CHECK (quantity > 0),
    -- Precio del producto al momento de la compra (foto histórica).
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0)
);
CREATE INDEX idx_order_items_order ON order_items(order_id);

-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
