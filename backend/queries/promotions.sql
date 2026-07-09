-- name: ListPromotions :many
SELECT * FROM promotions WHERE is_active = TRUE ORDER BY created_at DESC;
-- name: CreatePromotion :one
INSERT INTO promotions (name, stripe_coupon_id, discount_type, discount_value, currency, applies_to, is_active, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;
-- name: UpdatePromotion :one
UPDATE promotions SET name = $2, is_active = $3 WHERE id = $1 RETURNING *;
-- name: DeactivatePromotion :exec
UPDATE promotions SET is_active = FALSE WHERE id = $1;
