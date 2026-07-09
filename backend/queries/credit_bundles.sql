-- name: ListCreditBundlesByTenant :many
SELECT * FROM credit_bundles WHERE tenant_id = $1 AND is_active = TRUE ORDER BY sort_order ASC;
-- name: ListCreditBundlesPublic :many
SELECT * FROM credit_bundles WHERE is_active = TRUE ORDER BY sort_order ASC;
-- name: CreateCreditBundle :one
INSERT INTO credit_bundles (tenant_id, name, description, credits, price_cents, currency, is_active, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;
-- name: UpdateCreditBundle :one
UPDATE credit_bundles SET name = $2, description = $3, credits = $4, price_cents = $5, is_active = $6, sort_order = $7 WHERE id = $1 RETURNING *;
-- name: DeleteCreditBundle :exec
DELETE FROM credit_bundles WHERE id = $1;
