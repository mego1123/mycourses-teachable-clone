-- name: GetMembership :one
SELECT * FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2;
-- name: ListMembershipsByTenant :many
SELECT * FROM tenant_memberships WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: CountMembershipsByTenant :one
SELECT COUNT(*) FROM tenant_memberships WHERE tenant_id = $1;
-- name: ListMembershipsByUser :many
SELECT * FROM tenant_memberships WHERE user_id = $1 ORDER BY created_at DESC;
-- name: CreateMembership :one
INSERT INTO tenant_memberships (tenant_id, user_id, role) VALUES ($1, $2, $3) RETURNING *;
-- name: UpdateMembershipRole :exec
UPDATE tenant_memberships SET role = $3 WHERE tenant_id = $1 AND user_id = $2;
-- name: DeleteMembership :exec
DELETE FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2;
-- name: CountMembersByRole :one
SELECT COUNT(*) FROM tenant_memberships WHERE tenant_id = $1 AND role = $2;
