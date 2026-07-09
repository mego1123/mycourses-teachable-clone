-- name: GetInvitationByToken :one
SELECT * FROM invitations WHERE token_hash = $1 AND status = 'pending' AND expires_at > NOW();
-- name: CreateInvitation :one
INSERT INTO invitations (email, tenant_id, invited_by, role, token_hash, expires_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
-- name: CancelInvitation :exec
UPDATE invitations SET status = 'revoked' WHERE id = $1 AND status = 'pending';
-- name: ListInvitationsByTenant :many
SELECT * FROM invitations WHERE tenant_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4;
