-- +goose Up

CREATE TABLE batch_stats(
    id uuid primary key,
    sym_id uuid not null,
    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    open DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL,
    high DOUBLE PRECISION NOT NULL,
    low DOUBLE PRECISION NOT NULL,
    volume DOUBLE PRECISION NOT NULL,
    period_minutes  INT NOT NULL,
    FOREIGN KEY(sym_id) REFERENCES symbols(id) ON DELETE CASCADE
);
-- Index on sym_id for quick lookups of batches for a specific symbol
CREATE INDEX idx_batch_stats_sym_id ON batch_stats (sym_id);


-- Index on sym_id for quick lookups of batches for a specific symbol
CREATE INDEX idx_batch_stats_period ON batch_stats (period_minutes);
-- Composite index on sym_id and time range for efficient filtering
CREATE INDEX idx_batch_stats_sym_time ON batch_stats (sym_id, start_time, end_time);

-- Index on time range to support time-based queries across all symbols
CREATE INDEX idx_batch_stats_time_range ON batch_stats (start_time, end_time);


-- +goose Down
DROP TABLE batch_stats;
DROP INDEX IF EXISTS idx_batch_stats_period;
DROP INDEX IF EXISTS idx_batch_stats_sym_id;
DROP INDEX IF EXISTS idx_batch_stats_sym_time;
DROP INDEX IF EXISTS idx_batch_stats_time_range;
