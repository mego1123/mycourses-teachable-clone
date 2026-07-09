-- name: GetCouponByID :one
SELECT * FROM course_coupons WHERE id = $1;
-- name: CreateCoupon :one
INSERT INTO course_coupons (tenant_id, code, discount_type, discount_value, currency, course_id, expires_at, usage_limit)
VALUES ($1, UPPER($2), $3, $4, $5, $6, $7, $8) RETURNING *;
-- name: DeleteCoupon :exec
DELETE FROM course_coupons WHERE id = $1;
-- name: IncrementCouponUsage :exec
UPDATE course_coupons SET used_count = used_count + 1 WHERE id = $1;
-- name: ListCouponsByTenant :many
SELECT * FROM course_coupons WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: ValidateCoupon :one
SELECT * FROM course_coupons WHERE tenant_id = $1 AND code = UPPER($2) AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW()) AND (usage_limit IS NULL OR used_count < usage_limit);
