-- name: GetEnrollmentByID :one
SELECT * FROM enrollments WHERE id = $1;
-- name: GetEnrollmentByCourseAndUser :one
SELECT * FROM enrollments WHERE course_id = $1 AND user_id = $2;
-- name: GetEnrollmentByStripeSessionID :one
SELECT * FROM enrollments WHERE stripe_session_id = $1;
-- name: CreateEnrollment :one
INSERT INTO enrollments (course_id, tenant_id, user_id, status, price_paid_cents, currency, coupon_id, stripe_session_id, stripe_charge_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: CompleteEnrollment :exec
UPDATE enrollments SET status = 'completed', completed_at = NOW() WHERE id = $1;
-- name: RefundEnrollment :exec
UPDATE enrollments SET status = 'refunded', refunded_at = NOW(), refund_reason = $2 WHERE id = $1;
-- name: ListEnrollmentsByUser :many
SELECT * FROM enrollments WHERE user_id = $1 AND status = ANY($2::text[]) ORDER BY enrolled_at DESC LIMIT $3 OFFSET $4;
-- name: ListEnrollmentsByTenant :many
SELECT * FROM enrollments WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR status = $2) ORDER BY enrolled_at DESC LIMIT $3 OFFSET $4;
-- name: CountEnrollmentsByCourse :one
SELECT COUNT(*) FROM enrollments WHERE course_id = $1 AND status = 'active';
