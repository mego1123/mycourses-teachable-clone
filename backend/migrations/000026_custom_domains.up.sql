CREATE TABLE custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain TEXT NOT NULL, tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','active','failed')),
    dns_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ssl_status TEXT NOT NULL DEFAULT 'pending' CHECK (ssl_status IN ('pending','active','failed')),
    cf_hostname_id TEXT, verification_records JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX idx_custom_domains_tenant ON custom_domains(tenant_id, status);
CREATE INDEX idx_custom_domains_status ON custom_domains(status, dns_verified, ssl_status);
CREATE TRIGGER trigger_custom_domains_updated_at BEFORE UPDATE ON custom_domains FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
