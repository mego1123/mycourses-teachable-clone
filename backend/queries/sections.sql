-- name: GetSectionByID :one
SELECT * FROM sections WHERE id = $1;
-- name: CreateSection :one
INSERT INTO sections (course_id, title, description, sort_order, drip_offset_days) VALUES ($1, $2, $3, $4, $5) RETURNING *;
-- name: UpdateSection :one
UPDATE sections SET title = $2, description = $3, sort_order = $4, drip_offset_days = $5 WHERE id = $1 RETURNING *;
-- name: DeleteSection :exec
DELETE FROM sections WHERE id = $1;
-- name: ListSectionsByCourse :many
SELECT * FROM sections WHERE course_id = $1 ORDER BY sort_order ASC;
