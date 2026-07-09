
-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND is_active = TRUE;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = NOW() WHERE id = $1;
