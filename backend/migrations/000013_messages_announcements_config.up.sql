CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    type TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT FALSE, metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_messages_user ON messages(user_id, is_read, created_at DESC);
CREATE INDEX idx_messages_tenant ON messages(tenant_id, created_at DESC) WHERE tenant_id IS NOT NULL;

CREATE TABLE announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    title TEXT NOT NULL, body TEXT NOT NULL DEFAULT '',
    is_published BOOLEAN NOT NULL DEFAULT FALSE, published_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_announcements_tenant ON announcements(tenant_id, is_published, published_at DESC);
CREATE TRIGGER trigger_announcements_updated_at BEFORE UPDATE ON announcements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE config_vars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT NOT NULL, value JSONB NOT NULL, description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general', is_system BOOLEAN NOT NULL DEFAULT FALSE, is_readonly BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_config_vars_key ON config_vars(key);
CREATE INDEX idx_config_vars_category ON config_vars(category);
CREATE TRIGGER trigger_config_vars_updated_at BEFORE UPDATE ON config_vars FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
