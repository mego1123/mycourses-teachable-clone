-- name: GetPlanByID :one
SELECT * FROM plans WHERE id = $1;
-- name: ListActivePublicPlans :many
SELECT * FROM plans WHERE is_active = TRUE AND is_public = TRUE ORDER BY sort_order ASC, monthly_price_cents ASC;
-- name: CreatePlan :one
INSERT INTO plans (name, description, monthly_price_cents, yearly_price_cents, currency, included_seats, min_seats, max_seats, usage_credits_per_month, trial_period_days, entitlements, is_system, is_active, is_public, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING *;
-- name: UpsertPlan :one
INSERT INTO plans (name, description, monthly_price_cents, yearly_price_cents, currency, included_seats, min_seats, max_seats, usage_credits_per_month, trial_period_days, entitlements, is_system, is_active, is_public, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (name) DO UPDATE SET monthly_price_cents = EXCLUDED.monthly_price_cents, entitlements = EXCLUDED.entitlements, is_active = EXCLUDED.is_active, sort_order = EXCLUDED.sort_order, updated_at = NOW()
RETURNING *;
