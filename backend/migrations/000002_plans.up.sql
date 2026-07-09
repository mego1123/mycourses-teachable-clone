CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    stripe_product_id TEXT, stripe_price_id_monthly TEXT, stripe_price_id_yearly TEXT,
    monthly_price_cents BIGINT NOT NULL DEFAULT 0, yearly_price_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'usd',
    included_seats INT NOT NULL DEFAULT 1, min_seats INT NOT NULL DEFAULT 1, max_seats INT NOT NULL DEFAULT 1,
    user_limit INT, usage_credits_per_month INT NOT NULL DEFAULT 0, trial_period_days INT NOT NULL DEFAULT 0,
    entitlements JSONB NOT NULL DEFAULT '{}',
    is_system BOOLEAN NOT NULL DEFAULT FALSE, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_public BOOLEAN NOT NULL DEFAULT TRUE, sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_plans_name ON plans(name);
CREATE INDEX idx_plans_stripe_product ON plans(stripe_product_id) WHERE stripe_product_id IS NOT NULL;
CREATE INDEX idx_plans_is_active_public ON plans(is_active, is_public, sort_order);
CREATE INDEX idx_plans_entitlements ON plans USING GIN (entitlements);
CREATE TRIGGER trigger_plans_updated_at BEFORE UPDATE ON plans FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
