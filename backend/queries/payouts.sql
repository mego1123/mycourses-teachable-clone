-- name: GetPayoutByID :one
SELECT * FROM payouts WHERE id = $1;
-- name: GetPayoutByStripeTransferID :one
SELECT * FROM payouts WHERE stripe_transfer_id = $1;
-- name: CreatePayout :one
INSERT INTO payouts (tenant_id, stripe_transfer_id, stripe_payout_id, amount_cents, currency, status)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
-- name: UpdatePayoutStatus :exec
UPDATE payouts SET status = $2, failure_reason = $3, completed_at = CASE WHEN $2 = 'paid' THEN NOW() ELSE completed_at END WHERE id = $1;
-- name: ListPayoutsByTenant :many
SELECT * FROM payouts WHERE tenant_id = $1 ORDER BY initiated_at DESC LIMIT $2 OFFSET $3;
-- name: SumPayoutsByTenant :one
SELECT COALESCE(SUM(amount_cents), 0)::bigint as total_cents FROM payouts WHERE tenant_id = $1 AND status = 'paid';
