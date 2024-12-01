-- name: AddCryptoTick :one
INSERT INTO crypto_tick(id,sym_id,price,time,volume,daily_change,daily_diff)
VALUES($1, $2, $3, $4, $5, $6, $7)
    RETURNING *;
