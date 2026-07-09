-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;
-- name: GetUserByEmailNormalized :one
SELECT * FROM users WHERE email_normalized = $1 AND deleted_at IS NULL;
-- name: CreateUser :one
INSERT INTO users (email, email_normalized, display_name, password_hash, locale_preference, theme_preference, email_verified, is_active, is_admin, consent_accepted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;
-- name: UpdateUserProfile :one
UPDATE users SET display_name = $2, avatar_url = $3, bio = $4, locale_preference = $5, theme_preference = $6
WHERE id = $1 AND deleted_at IS NULL RETURNING *;
-- name: VerifyUserEmail :exec
UPDATE users SET email_verified = TRUE, email_verified_at = NOW() WHERE id = $1;
-- name: UpdateLastLogin :exec
UPDATE users SET last_login_at = NOW() WHERE id = $1;
-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW(), is_active = FALSE WHERE id = $1;
-- name: ListUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;
-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;
-- name: AdminListUsers :many
SELECT * FROM users WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR email_normalized ILIKE '%' || $1 || '%') ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: AdminCountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR email_normalized ILIKE '%' || $1 || '%');
