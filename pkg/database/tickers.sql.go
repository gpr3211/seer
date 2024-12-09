// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: tickers.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const addCryptoTick = `-- name: AddCryptoTick :one
INSERT INTO crypto_tick(id,sym_id,price,time,volume,daily_change,daily_diff)
VALUES($1, $2, $3, $4, $5, $6, $7)
    RETURNING id, sym_id, price, time, volume, daily_change, daily_diff, created_at
`

type AddCryptoTickParams struct {
	ID          uuid.UUID
	SymID       uuid.UUID
	Price       string
	Time        int64
	Volume      string
	DailyChange string
	DailyDiff   string
}

func (q *Queries) AddCryptoTick(ctx context.Context, arg AddCryptoTickParams) (CryptoTick, error) {
	row := q.db.QueryRowContext(ctx, addCryptoTick,
		arg.ID,
		arg.SymID,
		arg.Price,
		arg.Time,
		arg.Volume,
		arg.DailyChange,
		arg.DailyDiff,
	)
	var i CryptoTick
	err := row.Scan(
		&i.ID,
		&i.SymID,
		&i.Price,
		&i.Time,
		&i.Volume,
		&i.DailyChange,
		&i.DailyDiff,
		&i.CreatedAt,
	)
	return i, err
}
