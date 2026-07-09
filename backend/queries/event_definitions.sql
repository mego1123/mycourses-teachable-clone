-- name: ListEventDefinitions :many
SELECT * FROM event_definitions WHERE ($1::boolean IS NULL OR is_active = $1) ORDER BY category, name;
-- name: CreateEventDefinition :one
INSERT INTO event_definitions (name, description, category, is_active, is_system) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description, is_active = EXCLUDED.is_active, updated_at = NOW() RETURNING *;
-- name: UpdateEventDefinition :one
UPDATE event_definitions SET description = $2, is_active = $3 WHERE id = $1 RETURNING *;
-- name: DeleteEventDefinition :exec
DELETE FROM event_definitions WHERE id = $1 AND is_system = FALSE;
