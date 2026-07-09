-- name: GetLessonByID :one
SELECT * FROM lessons WHERE id = $1;
-- name: CreateLesson :one
INSERT INTO lessons (section_id, course_id, title, type, content, media_asset_id, sort_order, is_preview, duration_sec)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: UpdateLesson :one
UPDATE lessons SET title = $2, type = $3, content = $4, media_asset_id = $5, sort_order = $6, is_preview = $7, duration_sec = $8 WHERE id = $1 RETURNING *;
-- name: DeleteLesson :exec
DELETE FROM lessons WHERE id = $1;
-- name: ListLessonsBySection :many
SELECT * FROM lessons WHERE section_id = $1 ORDER BY sort_order ASC;
-- name: ListLessonsByCourse :many
SELECT * FROM lessons WHERE course_id = $1 ORDER BY sort_order ASC;
-- name: ListPreviewLessonsByCourse :many
SELECT * FROM lessons WHERE course_id = $1 AND is_preview = TRUE ORDER BY sort_order ASC;
-- name: CountLessonsByCourse :one
SELECT COUNT(*) FROM lessons WHERE course_id = $1;
