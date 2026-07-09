CREATE TABLE tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'user',
    invited_by UUID REFERENCES users(id), invited_at TIMESTAMPTZ, accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_memberships_tenant_user ON tenant_memberships(tenant_id, user_id);
CREATE INDEX idx_memberships_user ON tenant_memberships(user_id);
CREATE INDEX idx_memberships_tenant_role ON tenant_memberships(tenant_id, role);
CREATE INDEX idx_memberships_role ON tenant_memberships(role);
ALTER TABLE tenant_memberships ADD CONSTRAINT chk_memberships_role CHECK (role IN ('owner', 'admin', 'creator', 'student', 'user'));
CREATE TRIGGER trigger_memberships_updated_at BEFORE UPDATE ON tenant_memberships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
