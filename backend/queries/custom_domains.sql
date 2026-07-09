-- name: GetCustomDomainByID :one
SELECT * FROM custom_domains WHERE id = $1;
-- name: GetTenantByCustomDomain :one
SELECT t.* FROM tenants t JOIN custom_domains cd ON cd.tenant_id = t.id WHERE cd.domain = $1 AND cd.status = 'active' AND t.deleted_at IS NULL;
-- name: CreateCustomDomain :one
INSERT INTO custom_domains (domain, tenant_id, status, verification_records) VALUES ($1, $2, 'pending', $3) RETURNING *;
-- name: UpdateCustomDomainStatus :exec
UPDATE custom_domains SET status = $2, dns_verified = $3, ssl_status = $4, cf_hostname_id = $5 WHERE id = $1;
-- name: DeleteCustomDomain :exec
DELETE FROM custom_domains WHERE id = $1;
-- name: ListCustomDomainsByTenant :many
SELECT * FROM custom_domains WHERE tenant_id = $1 ORDER BY created_at DESC;
-- name: ListPendingCustomDomains :many
SELECT * FROM custom_domains WHERE status = 'pending' ORDER BY created_at ASC;
