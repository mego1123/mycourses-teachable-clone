-- name: GetBrandingConfigByTenant :one
SELECT * FROM branding_configs WHERE tenant_id = $1;
-- name: UpsertBrandingConfig :one
INSERT INTO branding_configs (tenant_id, config) VALUES ($1, $2)
ON CONFLICT (tenant_id) DO UPDATE SET config = EXCLUDED.config, updated_at = NOW() RETURNING *;
-- name: CreateBrandingAsset :one
INSERT INTO branding_assets (tenant_id, name, type, mime_type, size_bytes, data) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
-- name: GetBrandingAssetByID :one
SELECT * FROM branding_assets WHERE id = $1;
-- name: ListBrandingAssetsByTenant :many
SELECT * FROM branding_assets WHERE tenant_id = $1 ORDER BY created_at DESC;
-- name: DeleteBrandingAsset :exec
DELETE FROM branding_assets WHERE id = $1 AND tenant_id = $2;
-- name: GetCustomPageBySlug :one
SELECT * FROM custom_pages WHERE tenant_id = $1 AND slug = $2 AND is_published = TRUE;
-- name: ListCustomPagesByTenant :many
SELECT * FROM custom_pages WHERE tenant_id = $1 AND is_published = TRUE ORDER BY sort_order ASC;
-- name: ListAllCustomPages :many
SELECT * FROM custom_pages WHERE tenant_id = $1 ORDER BY sort_order ASC;
-- name: CreateCustomPage :one
INSERT INTO custom_pages (tenant_id, slug, title, html_body, meta_description, is_published, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;
-- name: UpdateCustomPage :one
UPDATE custom_pages SET slug = $2, title = $3, html_body = $4, meta_description = $5, is_published = $6, sort_order = $7 WHERE id = $1 RETURNING *;
-- name: DeleteCustomPage :exec
DELETE FROM custom_pages WHERE id = $1 AND tenant_id = $2;
