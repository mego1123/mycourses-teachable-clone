CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    stripe_transfer_id TEXT, stripe_payout_id TEXT,
    amount_cents BIGINT NOT NULL, currency TEXT NOT NULL DEFAULT 'usd',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','failed','cancelled')),
    failure_reason TEXT,
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payouts_tenant_date ON payouts(tenant_id, initiated_at DESC);
CREATE INDEX idx_payouts_status ON payouts(status, initiated_at DESC);
CREATE INDEX idx_payouts_stripe_transfer ON payouts(stripe_transfer_id) WHERE stripe_transfer_id IS NOT NULL;
CREATE INDEX idx_payouts_stripe_payout ON payouts(stripe_payout_id) WHERE stripe_payout_id IS NOT NULL;
CREATE TRIGGER trigger_payouts_updated_at BEFORE UPDATE ON payouts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
