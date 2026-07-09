-- name: CreateUsageEvent :one
INSERT INTO usage_events (tenant_id, user_id, type, amount, metadata) VALUES ($1, $2, $3, $4, $5) RETURNING *;
-- name: GetUsageSummary :one
SELECT COALESCE(SUM(amount), 0)::bigint as total FROM usage_events WHERE tenant_id = $1 AND created_at >= $2 AND created_at < $3;
