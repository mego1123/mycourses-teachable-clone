-- name: GetCertificateByID :one
SELECT * FROM certificates WHERE id = $1;
-- name: GetCertificateByToken :one
SELECT * FROM certificates WHERE verification_token = $1 AND status = 'active';
-- name: CreateCertificate :one
INSERT INTO certificates (enrollment_id, user_id, course_id, tenant_id, certificate_number, verification_token, learner_name, course_title, creator_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: RevokeCertificate :exec
UPDATE certificates SET status = 'revoked', revoked_reason = $2, revoked_at = NOW() WHERE id = $1;
-- name: ListCertificatesByUser :many
SELECT * FROM certificates WHERE user_id = $1 AND status = 'active' ORDER BY issued_at DESC;
-- name: GetNextCertificateNumber :one
SELECT 'CERT-' || to_char(NOW(), 'YYYY') || '-' || LPAD((COUNT(*) + 1)::text, 6, '0') as cert_number FROM certificates WHERE created_at >= date_trunc('year', NOW());

-- name: GetCertificateByEnrollment :one
SELECT * FROM certificates WHERE enrollment_id = $1;
