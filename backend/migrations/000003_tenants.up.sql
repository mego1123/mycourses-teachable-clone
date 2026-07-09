CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL, slug TEXT NOT NULL,
    plan_id UUID REFERENCES plans(id) ON DELETE SET NULL,
    stripe_customer_id TEXT, stripe_subscription_id TEXT,
    billing_status TEXT NOT NULL DEFAULT 'active', billing_interval TEXT,
    trial_ends_at TIMESTAMPTZ, billing_waiver BOOLEAN NOT NULL DEFAULT FALSE, billing_waiver_reason TEXT,
    stripe_connect_account_id TEXT, stripe_connect_status TEXT NOT NULL DEFAULT 'pending',
    commission_rate_bps INT NOT NULL DEFAULT 1000, payout_schedule TEXT NOT NULL DEFAULT 'weekly',
    is_root BOOLEAN NOT NULL DEFAULT FALSE, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_tenants_slug ON tenants(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_stripe_customer ON tenants(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE INDEX idx_tenants_stripe_subscription ON tenants(stripe_subscription_id) WHERE stripe_subscription_id IS NOT NULL;
CREATE INDEX idx_tenants_stripe_connect ON tenants(stripe_connect_account_id) WHERE stripe_connect_account_id IS NOT NULL;
CREATE INDEX idx_tenants_billing_status ON tenants(billing_status);
CREATE INDEX idx_tenants_is_root ON tenants(is_root) WHERE is_root = TRUE;
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);
CREATE TRIGGER trigger_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
