-- name: GetWebhookByID :one
SELECT * FROM webhooks WHERE id = $1;
-- name: ListWebhooksByTenant :many
SELECT * FROM webhooks WHERE tenant_id = $1 AND is_active = TRUE ORDER BY created_at DESC;
-- name: CreateWebhook :one
INSERT INTO webhooks (tenant_id, name, url, secret, events, created_by) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
-- name: UpdateWebhook :one
UPDATE webhooks SET name = $2, url = $3, secret = $4, events = $5, is_active = $6 WHERE id = $1 RETURNING *;
-- name: DeleteWebhook :exec
DELETE FROM webhooks WHERE id = $1;
