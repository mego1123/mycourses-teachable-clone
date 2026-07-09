CREATE TABLE course_coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code TEXT NOT NULL, discount_type TEXT NOT NULL CHECK (discount_type IN ('percent','fixed')),
    discount_value INT NOT NULL, currency TEXT NOT NULL DEFAULT 'usd',
    course_id UUID REFERENCES courses(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ, usage_limit INT, used_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_coupons_tenant_code ON course_coupons(tenant_id, code);
CREATE INDEX idx_coupons_tenant_active ON course_coupons(tenant_id, is_active);
CREATE INDEX idx_coupons_course ON course_coupons(course_id) WHERE course_id IS NOT NULL;
CREATE INDEX idx_coupons_expires ON course_coupons(expires_at) WHERE expires_at IS NOT NULL;
CREATE TRIGGER trigger_course_coupons_updated_at BEFORE UPDATE ON course_coupons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
