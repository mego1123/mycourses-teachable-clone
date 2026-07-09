-- name: ListMessagesByUser :many
SELECT * FROM messages WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: CreateMessage :one
INSERT INTO messages (user_id, tenant_id, type, title, body, metadata) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
-- name: MarkMessageRead :exec
UPDATE messages SET is_read = TRUE WHERE id = $1 AND user_id = $2;
-- name: CountUnreadMessages :one
SELECT COUNT(*) FROM messages WHERE user_id = $1 AND is_read = FALSE;
