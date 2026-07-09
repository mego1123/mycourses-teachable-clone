-- name: GetProcessedStripeEvent :one
SELECT * FROM processed_stripe_events WHERE event_id = $1;
-- name: MarkStripeEventProcessed :exec
INSERT INTO processed_stripe_events (event_id, event_type) VALUES ($1, $2) ON CONFLICT (event_id) DO NOTHING;
