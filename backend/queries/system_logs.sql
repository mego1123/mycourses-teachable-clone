-- name: ListSystemLogs :many
SELECT * FROM system_logs WHERE ($1::text IS NULL OR $1 = '' OR level = $1) AND ($2::text IS NULL OR $2 = '' OR category = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4;
-- name: CountSystemLogs :one
SELECT COUNT(*) FROM system_logs WHERE ($1::text IS NULL OR $1 = '' OR level = $1) AND ($2::text IS NULL OR $2 = '' OR category = $2);
-- name: CreateSystemLog :exec
INSERT INTO system_logs (level, category, message, details, tenant_id, request_id) VALUES ($1, $2, $3, $4, $5, $6);
-- name: GetSeverityCounts :one
SELECT COUNT(*) FILTER (WHERE level = 'error') as error_count, COUNT(*) FILTER (WHERE level = 'warn') as warn_count, COUNT(*) FILTER (WHERE level = 'info') as info_count FROM system_logs WHERE created_at >= $1;
