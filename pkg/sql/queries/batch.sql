-- name: CreateBatchStat :one

INSERT INTO batch_stats(id,sym_id,start_time,end_time,open,close,high,low,volume,period_minutes)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
    returning *;

