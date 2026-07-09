-- name: ListPublishedAnnouncements :many
SELECT * FROM announcements WHERE tenant_id = $1 AND is_published = TRUE ORDER BY published_at DESC LIMIT $2 OFFSET $3;
-- name: ListAllAnnouncements :many
SELECT * FROM announcements WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: CreateAnnouncement :one
INSERT INTO announcements (tenant_id, title, body, is_published, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING *;
-- name: UpdateAnnouncement :one
UPDATE announcements SET title = $2, body = $3, is_published = $4, published_at = CASE WHEN $4 = TRUE AND published_at IS NULL THEN NOW() ELSE published_at END WHERE id = $1 RETURNING *;
-- name: DeleteAnnouncement :exec
DELETE FROM announcements WHERE id = $1;
