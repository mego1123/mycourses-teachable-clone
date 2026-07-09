CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL, tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    invited_by UUID NOT NULL REFERENCES users(id),
    role TEXT NOT NULL CHECK (role IN ('owner','admin','creator','student','user')),
    token_hash TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','revoked','expired')),
    expires_at TIMESTAMPTZ NOT NULL, accepted_at TIMESTAMPTZ, accepted_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_invitations_token ON invitations(token_hash);
CREATE INDEX idx_invitations_tenant ON invitations(tenant_id, status);
CREATE INDEX idx_invitations_email ON invitations(email, status);
CREATE INDEX idx_invitations_expires ON invitations(expires_at) WHERE status = 'pending';

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users(id), tenant_id UUID REFERENCES tenants(id),
    action TEXT NOT NULL, resource_type TEXT NOT NULL, resource_id TEXT,
    details JSONB NOT NULL DEFAULT '{}', ip_address TEXT, user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_actor ON audit_log(actor_id, created_at DESC);
CREATE INDEX idx_audit_tenant ON audit_log(tenant_id, created_at DESC);
CREATE INDEX idx_audit_action ON audit_log(action, created_at DESC);
CREATE INDEX idx_audit_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_created ON audit_log(created_at DESC);

CREATE TABLE system_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level TEXT NOT NULL CHECK (level IN ('debug','info','warn','error','fatal')),
    category TEXT NOT NULL, message TEXT NOT NULL,
    details JSONB NOT NULL DEFAULT '{}', tenant_id UUID REFERENCES tenants(id),
    request_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_system_logs_level_time ON system_logs(level, created_at DESC);
CREATE INDEX idx_system_logs_category ON system_logs(category, created_at DESC);
CREATE INDEX idx_system_logs_tenant ON system_logs(tenant_id, created_at DESC) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_system_logs_request ON system_logs(request_id) WHERE request_id IS NOT NULL;
CREATE INDEX idx_system_logs_created ON system_logs(created_at DESC);
CREATE INDEX idx_system_logs_message_fts ON system_logs USING GIN (to_tsvector('english', message));
