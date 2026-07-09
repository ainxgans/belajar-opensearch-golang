CREATE TABLE IF NOT EXISTS products (
  id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  sku          TEXT UNIQUE NOT NULL,
  name         TEXT NOT NULL,
  description  TEXT NOT NULL DEFAULT '',
  brand        TEXT NOT NULL,
  category     TEXT NOT NULL,
  price        NUMERIC(12,2) NOT NULL,
  stock        INT NOT NULL DEFAULT 0,
  rating       NUMERIC(2,1) NOT NULL DEFAULT 0,
  tags         TEXT[] NOT NULL DEFAULT '{}',
  attributes   JSONB NOT NULL DEFAULT '[]',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE products REPLICA IDENTITY FULL;
