-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1 AND deleted_at IS NULL;
-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1 AND deleted_at IS NULL;
-- name: GetRootTenant :one
SELECT * FROM tenants WHERE is_root = TRUE AND deleted_at IS NULL LIMIT 1;
-- name: CreateTenant :one
INSERT INTO tenants (name, slug, plan_id, billing_status, stripe_connect_status, commission_rate_bps, payout_schedule, is_root, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: UpdateTenantStripeConnect :exec
UPDATE tenants SET stripe_connect_account_id = $2, stripe_connect_status = $3 WHERE id = $1;
-- name: UpdateTenantCommissionRate :exec
UPDATE tenants SET commission_rate_bps = $2 WHERE id = $1;
-- name: SoftDeleteTenant :exec
UPDATE tenants SET deleted_at = NOW(), is_active = FALSE WHERE id = $1;
-- name: ListTenants :many
SELECT * FROM tenants WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR name ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%') AND ($2::text IS NULL OR $2 = '' OR billing_status = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4;
-- name: CountTenants :one
SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR name ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%') AND ($2::text IS NULL OR $2 = '' OR billing_status = $2);
-- name: CountActiveTenants :one
SELECT COUNT(*) FROM tenants WHERE billing_status = 'active' AND deleted_at IS NULL;
