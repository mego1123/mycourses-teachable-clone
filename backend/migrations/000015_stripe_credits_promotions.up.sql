CREATE TABLE stripe_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_type TEXT NOT NULL CHECK (internal_type IN ('plan','credit_bundle')),
    internal_id UUID NOT NULL, stripe_product_id TEXT NOT NULL, stripe_price_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_stripe_mappings_product ON stripe_mappings(stripe_product_id);
CREATE INDEX idx_stripe_mappings_internal ON stripe_mappings(internal_type, internal_id);

CREATE TABLE credit_bundles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    credits INT NOT NULL, price_cents BIGINT NOT NULL, currency TEXT NOT NULL DEFAULT 'usd',
    stripe_product_id TEXT, stripe_price_id TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE, sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_credit_bundles_tenant ON credit_bundles(tenant_id, is_active, sort_order);
CREATE TRIGGER trigger_credit_bundles_updated_at BEFORE UPDATE ON credit_bundles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL, stripe_coupon_id TEXT NOT NULL,
    discount_type TEXT NOT NULL CHECK (discount_type IN ('percent','fixed')), discount_value INT NOT NULL,
    currency TEXT, applies_to TEXT NOT NULL CHECK (applies_to IN ('all','plans','credit_bundles')),
    eligible_product_ids TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE, created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_promotions_active ON promotions(is_active);
CREATE TRIGGER trigger_promotions_updated_at BEFORE UPDATE ON promotions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE promotion_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    code TEXT NOT NULL, stripe_promotion_code_id TEXT NOT NULL,
    max_redemptions INT, times_redeemed INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_promotion_codes_code ON promotion_codes(code);
CREATE INDEX idx_promotion_codes_promotion ON promotion_codes(promotion_id);
