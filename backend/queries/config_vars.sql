-- name: GetConfigVar :one
SELECT * FROM config_vars WHERE key = $1;
-- name: ListConfigVars :many
SELECT * FROM config_vars WHERE ($1::text IS NULL OR $1 = '' OR category = $1) ORDER BY category, key;
-- name: UpsertConfigVar :one
INSERT INTO config_vars (key, value, description, category, is_system, is_readonly) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, description = EXCLUDED.description, updated_at = NOW() RETURNING *;
-- name: DeleteConfigVar :exec
DELETE FROM config_vars WHERE key = $1 AND is_readonly = FALSE;
