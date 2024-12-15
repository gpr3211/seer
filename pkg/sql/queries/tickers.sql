-- name: AddCryptoTick :one
INSERT INTO crypto_tick(id,sym_id,price,time,volume,daily_change,daily_diff)
VALUES($1, $2, $3, $4, $5, $6, $7)
    RETURNING *;

-- name: AddForexTick :one
INSERT INTO forex_tick(id, sym_id, ask_price,bid_price, time, daily_change, daily_diff)
VALUES($1,$2,$3,$4,$5,$6,$7)
    RETURNING *;

-- name: AddUsTick :one
INSERT INTO us_trade_tick(id, sym_id, price, time, conditions, volume)
VALUES($1,$2,$3,$4,$5,$6)
RETURNING *;

