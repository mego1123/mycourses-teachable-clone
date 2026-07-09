-- name: GetReviewByID :one
SELECT * FROM reviews WHERE id = $1;
-- name: CreateReview :one
INSERT INTO reviews (course_id, user_id, tenant_id, rating, comment) VALUES ($1, $2, $3, $4, $5) RETURNING *;
-- name: DeleteReview :exec
DELETE FROM reviews WHERE id = $1;
-- name: HideReview :exec
UPDATE reviews SET is_hidden = TRUE WHERE id = $1;
-- name: ListPublicReviewsByCourse :many
SELECT * FROM reviews WHERE course_id = $1 AND is_hidden = FALSE ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: GetAverageRatingByCourse :one
SELECT COALESCE(AVG(rating), 0)::float as avg_rating, COUNT(*)::int as review_count FROM reviews WHERE course_id = $1 AND is_hidden = FALSE;
