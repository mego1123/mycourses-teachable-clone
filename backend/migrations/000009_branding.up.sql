CREATE TABLE branding_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_branding_config_gin ON branding_configs USING GIN (config);
CREATE TRIGGER trigger_branding_configs_updated_at BEFORE UPDATE ON branding_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE branding_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL, type TEXT NOT NULL CHECK (type IN ('logo','favicon','media','custom')),
    mime_type TEXT NOT NULL, size_bytes BIGINT NOT NULL, data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_branding_assets_tenant ON branding_assets(tenant_id, type);

CREATE TABLE custom_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL, title TEXT NOT NULL, html_body TEXT NOT NULL DEFAULT '',
    meta_description TEXT, is_published BOOLEAN NOT NULL DEFAULT FALSE, sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_custom_pages_tenant_slug ON custom_pages(tenant_id, slug);
CREATE INDEX idx_custom_pages_published ON custom_pages(tenant_id, is_published, sort_order);
CREATE TRIGGER trigger_custom_pages_updated_at BEFORE UPDATE ON custom_pages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
