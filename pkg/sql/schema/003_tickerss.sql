-- +goose Up


DROP TABLE IF EXISTS us_trade_tick;
CREATE TABLE us_trade_tick(
    id UUID PRIMARY KEY,
    sym_id UUID NOT NULL,
    price NUMERIC(18, 8) NOT NULL,
    time BIGINT NOT NULL,          -- Unix timestamp
    conditions TEXT NOT NULL,
    volume NUMERIC(18,8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sym_id) REFERENCES symbols(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS forex_tick;
CREATE TABLE forex_tick(
    id UUID PRIMARY KEY,
    sym_id UUID NOT NULL,
    ask_price NUMERIC(18, 8) NOT NULL,
    bid_price NUMERIC(18,8) NOT NULL,
    time BIGINT NOT NULL,          -- Unix timestamp
    daily_change NUMERIC(18, 8) NOT NULL,
    daily_diff NUMERIC(18, 8) NOT NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sym_id) REFERENCES symbols(id) ON DELETE CASCADE
);
-- For queries that might frequently look up both symbol and time together
CREATE INDEX idx_forex_tick_sym_time ON forex_tick (sym_id, time);

-- +goose Down

DROP INDEX IF EXISTS idx_forex_tick_time;
DROP INDEX IF EXISTS idx_forex_tick_sym_time;
DROP TABLE IF EXISTS forex_tick;
DROP TABLE IF EXISTS us_trade_tick;

