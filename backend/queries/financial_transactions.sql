-- name: CreateTransaction :one
INSERT INTO financial_transactions (tenant_id, user_id, type, amount_cents, currency, description, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;
-- name: ListTransactionsByTenant :many
SELECT * FROM financial_transactions WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR type = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4;
-- name: SumTransactionsByTenantAndType :one
SELECT COALESCE(SUM(amount_cents), 0)::bigint as total_cents FROM financial_transactions WHERE tenant_id = $1 AND type = $2 AND created_at >= $3 AND created_at < $4;
