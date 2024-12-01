-- +goose Up


CREATE TABLE exchanges (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL UNIQUE,
    code TEXT NOT NULL UNIQUE, 
    currency TEXT NOT NULL,
    country TEXT NOT NULL,
    iso2 TEXT NOT NULL,
    iso3 TEXT NOT NULL,
    op_mic TEXT NOT NULL
);

CREATE TABLE symbols(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    exchange_name TEXT NOT NULL,
    exchange_id uuid not null,
    currency TEXT NOT NULL,
    type TEXT NOT NULL,
    foreign key(exchange_id) REFERENCES exchanges(id)
);
CREATE TABLE crypto_tick (
    id UUID PRIMARY KEY,
    sym_id UUID NOT NULL,
    price NUMERIC(18, 8) NOT NULL,
    time BIGINT NOT NULL,          -- Unix timestamp
    volume NUMERIC(18, 8) NOT NULL,
    daily_change NUMERIC(18, 8) NOT NULL,
    daily_diff NUMERIC(18, 8) NOT NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sym_id) REFERENCES symbols(id) ON DELETE CASCADE
);
CREATE INDEX idx_crypto_tick_time ON crypto_tick (time);

-- For queries that might frequently look up both symbol and time together
CREATE INDEX idx_crypto_tick_sym_time ON crypto_tick (sym_id, time);

-- +goose Down
DROP INDEX IF EXISTS idx_crypto_tick_time;
DROP INDEX IF EXISTS idx_crypto_tick_sym_time;
DROP TABLE crypto_tick;
DROP TABLE symbols;
DROP TABLE exchanges;
