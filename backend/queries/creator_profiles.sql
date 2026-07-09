-- name: GetCreatorProfileByTenant :one
SELECT * FROM creator_profiles WHERE tenant_id = $1;
-- name: UpsertCreatorProfile :one
INSERT INTO creator_profiles (tenant_id, bio, website_url, social_links, avatar_url, banner_url)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id) DO UPDATE SET bio = EXCLUDED.bio, website_url = EXCLUDED.website_url, social_links = EXCLUDED.social_links, avatar_url = EXCLUDED.avatar_url, banner_url = EXCLUDED.banner_url, updated_at = NOW()
RETURNING *;
