-- name: GetMediaAssetByID :one
SELECT * FROM media_assets WHERE id = $1;
-- name: CreateMediaAsset :one
INSERT INTO media_assets (tenant_id, kind, title, cf_stream_id, r2_key, r2_url, size_bytes, mime_type, duration_sec, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;
-- name: UpdateMediaAssetStatus :exec
UPDATE media_assets SET status = $2 WHERE id = $1;
-- name: ListMediaAssetsByTenant :many
SELECT * FROM media_assets WHERE tenant_id = $1 AND status != 'deleted' ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateMediaAssetReady :exec
UPDATE media_assets SET status = 'ready', duration_sec = $2, size_bytes = $3 WHERE id = $1;
