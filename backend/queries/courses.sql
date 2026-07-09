-- name: GetCourseByID :one
SELECT * FROM courses WHERE id = $1;
-- name: GetCourseBySlug :one
SELECT * FROM courses WHERE tenant_id = $1 AND slug = $2;
-- name: GetPublishedCourseBySlug :one
SELECT * FROM courses WHERE tenant_id = $1 AND slug = $2 AND status = 'published';
-- name: CreateCourse :one
INSERT INTO courses (tenant_id, title, description, slug, price_cents, currency, status, thumbnail_url, intro_video_url, category, drip_enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;
-- name: UpdateCourse :one
UPDATE courses SET title = $2, description = $3, slug = $4, price_cents = $5, currency = $6, thumbnail_url = $7, intro_video_url = $8, category = $9, drip_enabled = $10 WHERE id = $1 RETURNING *;
-- name: PublishCourse :one
UPDATE courses SET status = 'published', published_at = NOW() WHERE id = $1 RETURNING *;
-- name: UnpublishCourse :one
UPDATE courses SET status = 'draft', published_at = NULL WHERE id = $1 RETURNING *;
-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;
-- name: ListCoursesByTenant :many
SELECT * FROM courses WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR status = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4;
-- name: CountCoursesByTenant :one
SELECT COUNT(*) FROM courses WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR status = $2);
-- name: ListMarketplaceCourses :many
SELECT * FROM courses WHERE status = 'published' AND ($1::text IS NULL OR $1 = '' OR category = $1) ORDER BY published_at DESC LIMIT $2 OFFSET $3;
-- name: SearchCourses :many
SELECT * FROM courses WHERE status = 'published' AND (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'')) @@ plainto_tsquery('english', $1) OR title % $1) ORDER BY ts_rank(to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'')), plainto_tsquery('english', $1)) DESC LIMIT $2 OFFSET $3;
