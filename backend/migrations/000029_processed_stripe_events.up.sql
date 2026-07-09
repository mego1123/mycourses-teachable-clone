CREATE TABLE processed_stripe_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_processed_stripe_type ON processed_stripe_events(event_type, processed_at DESC);
