CREATE TABLE enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed','refunded','disputed','chargeback_lost')),
    price_paid_cents BIGINT NOT NULL DEFAULT 0, currency TEXT NOT NULL DEFAULT 'usd',
    coupon_id UUID, stripe_session_id TEXT, stripe_charge_id TEXT,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), completed_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ, refund_reason TEXT, expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_enrollments_course_user ON enrollments(course_id, user_id);
CREATE INDEX idx_enrollments_user_status ON enrollments(user_id, status);
CREATE INDEX idx_enrollments_tenant_date ON enrollments(tenant_id, enrolled_at DESC);
CREATE INDEX idx_enrollments_stripe_session ON enrollments(stripe_session_id) WHERE stripe_session_id IS NOT NULL;
CREATE INDEX idx_enrollments_stripe_charge ON enrollments(stripe_charge_id) WHERE stripe_charge_id IS NOT NULL;
CREATE TRIGGER trigger_enrollments_updated_at BEFORE UPDATE ON enrollments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
