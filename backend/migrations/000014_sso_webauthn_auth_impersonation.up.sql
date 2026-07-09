CREATE TABLE sso_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_type TEXT NOT NULL CHECK (provider_type IN ('saml','oidc')),
    name TEXT NOT NULL, config JSONB NOT NULL DEFAULT '{}', is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sso_connections_tenant ON sso_connections(tenant_id, is_active);
CREATE TRIGGER trigger_sso_connections_updated_at BEFORE UPDATE ON sso_connections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id TEXT NOT NULL, public_key BYTEA NOT NULL,
    attestation_type TEXT, aaguid TEXT, sign_count BIGINT NOT NULL DEFAULT 0,
    name TEXT, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_webauthn_credentials_id ON webauthn_credentials(credential_id);
CREATE INDEX idx_webauthn_credentials_user ON webauthn_credentials(user_id, is_active);

CREATE TABLE webauthn_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    challenge TEXT NOT NULL, type TEXT NOT NULL CHECK (type IN ('registration','authentication')),
    expires_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_webauthn_sessions_user ON webauthn_sessions(user_id, expires_at);
CREATE INDEX idx_webauthn_sessions_expires ON webauthn_sessions(expires_at);

CREATE TABLE auth_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL, client_id TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL, scopes TEXT[] NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ NOT NULL, used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_auth_codes_code ON auth_codes(code);
CREATE INDEX idx_auth_codes_expires ON auth_codes(expires_at) WHERE used_at IS NULL;

CREATE TABLE impersonation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id), target_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID REFERENCES tenants(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), ended_at TIMESTAMPTZ,
    ip_address TEXT, user_agent TEXT, reason TEXT
);
CREATE INDEX idx_impersonation_admin ON impersonation_logs(admin_id, started_at DESC);
CREATE INDEX idx_impersonation_target ON impersonation_logs(target_id, started_at DESC);
