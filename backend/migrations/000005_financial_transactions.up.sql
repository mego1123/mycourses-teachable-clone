CREATE TABLE financial_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    type TEXT NOT NULL,
    amount_cents BIGINT NOT NULL, subtotal_cents BIGINT, tax_cents BIGINT, currency TEXT NOT NULL DEFAULT 'usd',
    stripe_charge_id TEXT, stripe_transfer_id TEXT, stripe_session_id TEXT, stripe_invoice_id TEXT,
    stripe_refund_id TEXT, stripe_dispute_id TEXT,
    invoice_number TEXT, description TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_financial_invoice_number ON financial_transactions(invoice_number) WHERE invoice_number IS NOT NULL;
CREATE INDEX idx_financial_tenant_type_date ON financial_transactions(tenant_id, type, created_at DESC);
CREATE INDEX idx_financial_user_date ON financial_transactions(user_id, created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_financial_stripe_charge ON financial_transactions(stripe_charge_id) WHERE stripe_charge_id IS NOT NULL;
CREATE INDEX idx_financial_stripe_transfer ON financial_transactions(stripe_transfer_id) WHERE stripe_transfer_id IS NOT NULL;
CREATE INDEX idx_financial_stripe_session ON financial_transactions(stripe_session_id) WHERE stripe_session_id IS NOT NULL;
CREATE INDEX idx_financial_type_date ON financial_transactions(type, created_at DESC);
CREATE INDEX idx_financial_metadata ON financial_transactions USING GIN (metadata);
ALTER TABLE financial_transactions ADD CONSTRAINT chk_financial_type CHECK (type IN ('subscription','credit_purchase','refund','course_purchase','platform_commission','creator_payout','connect_transfer_reversal','dispute_pending','dispute_withdrawal','dispute_reinstatement'));
