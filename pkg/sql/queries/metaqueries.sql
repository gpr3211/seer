-- name: CreateTicker :one
INSERT INTO symbols(id,code,name,exchange_name,exchange_id,currency,type)
VALUES($1,$2,$3,$4,$5,$6,$7) -- 
    RETURNING *;

-- name: CreateExchange :one
INSERT INTO exchanges(id,name,code,country,currency,iso2,iso3,op_mic)
VALUES($1, $2, $3, $4, $5, $6, $7,$8)
    RETURNING *;

-- name: FetchExchanges :many
select * from exchanges;
-- name: CheckIfExists :one
select name from exchanges where name = $1;

-- name: GetExchangeId :one
select id from exchanges where  code= $1;

-- name: GetTickerId :one
SELECT id from symbols where code = $1;
-- name: GetTickerExchangeId :one
select exchange_id from symbols where code= $1;


