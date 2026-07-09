#!/bin/bash
# Creates all migration files for the mycourses project
cd /home/z/my-project/mycourses/backend/migrations

# 0001 - extensions + users
cat > 000001_enable_extensions.up.sql << 'SQLEOF'
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    email_normalized TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    password_hash TEXT,
    auth_methods TEXT[] NOT NULL DEFAULT '{}',
    google_id TEXT, github_id TEXT, microsoft_id TEXT,
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret TEXT, mfa_encryption_key TEXT,
    recovery_codes TEXT[] NOT NULL DEFAULT '{}',
    avatar_url TEXT, bio TEXT,
    locale_preference TEXT NOT NULL DEFAULT 'en',
    theme_preference TEXT NOT NULL DEFAULT 'system',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ,
    consent_accepted_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ, last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_users_email_normalized ON users(email_normalized) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;
CREATE INDEX idx_users_github_id ON users(github_id) WHERE github_id IS NOT NULL;
CREATE INDEX idx_users_microsoft_id ON users(microsoft_id) WHERE microsoft_id IS NOT NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC);
CREATE INDEX idx_users_last_login ON users(last_login_at DESC) WHERE last_login_at IS NOT NULL;

CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW(); RETURN NEW; END; $$ LANGUAGE plpgsql;
CREATE TRIGGER trigger_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF

cat > 000001_enable_extensions.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "pgcrypto";
SQLEOF

# 0002 - plans
cat > 000002_plans.up.sql << 'SQLEOF'
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    stripe_product_id TEXT, stripe_price_id_monthly TEXT, stripe_price_id_yearly TEXT,
    monthly_price_cents BIGINT NOT NULL DEFAULT 0, yearly_price_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'usd',
    included_seats INT NOT NULL DEFAULT 1, min_seats INT NOT NULL DEFAULT 1, max_seats INT NOT NULL DEFAULT 1,
    user_limit INT, usage_credits_per_month INT NOT NULL DEFAULT 0, trial_period_days INT NOT NULL DEFAULT 0,
    entitlements JSONB NOT NULL DEFAULT '{}',
    is_system BOOLEAN NOT NULL DEFAULT FALSE, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_public BOOLEAN NOT NULL DEFAULT TRUE, sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_plans_name ON plans(name);
CREATE INDEX idx_plans_stripe_product ON plans(stripe_product_id) WHERE stripe_product_id IS NOT NULL;
CREATE INDEX idx_plans_is_active_public ON plans(is_active, is_public, sort_order);
CREATE INDEX idx_plans_entitlements ON plans USING GIN (entitlements);
CREATE TRIGGER trigger_plans_updated_at BEFORE UPDATE ON plans FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000002_plans.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_plans_updated_at ON plans;
DROP TABLE IF EXISTS plans;
SQLEOF

# 0003 - tenants
cat > 000003_tenants.up.sql << 'SQLEOF'
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL, slug TEXT NOT NULL,
    plan_id UUID REFERENCES plans(id) ON DELETE SET NULL,
    stripe_customer_id TEXT, stripe_subscription_id TEXT,
    billing_status TEXT NOT NULL DEFAULT 'active', billing_interval TEXT,
    trial_ends_at TIMESTAMPTZ, billing_waiver BOOLEAN NOT NULL DEFAULT FALSE, billing_waiver_reason TEXT,
    stripe_connect_account_id TEXT, stripe_connect_status TEXT NOT NULL DEFAULT 'pending',
    commission_rate_bps INT NOT NULL DEFAULT 1000, payout_schedule TEXT NOT NULL DEFAULT 'weekly',
    is_root BOOLEAN NOT NULL DEFAULT FALSE, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_tenants_slug ON tenants(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_stripe_customer ON tenants(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE INDEX idx_tenants_stripe_subscription ON tenants(stripe_subscription_id) WHERE stripe_subscription_id IS NOT NULL;
CREATE INDEX idx_tenants_stripe_connect ON tenants(stripe_connect_account_id) WHERE stripe_connect_account_id IS NOT NULL;
CREATE INDEX idx_tenants_billing_status ON tenants(billing_status);
CREATE INDEX idx_tenants_is_root ON tenants(is_root) WHERE is_root = TRUE;
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);
CREATE TRIGGER trigger_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000003_tenants.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_tenants_updated_at ON tenants;
DROP TABLE IF EXISTS tenants;
SQLEOF

# 0004 - memberships
cat > 000004_tenant_memberships.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000004_tenant_memberships.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_memberships_updated_at ON tenant_memberships;
DROP TABLE IF EXISTS tenant_memberships;
SQLEOF

# 0005 - financial_transactions
cat > 000005_financial_transactions.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000005_financial_transactions.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS financial_transactions;
SQLEOF

# 0006 - refresh_tokens
cat > 000006_refresh_tokens.up.sql << 'SQLEOF'
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL, family_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL, revoked_at TIMESTAMPTZ, revoked_reason TEXT,
    replaced_by UUID REFERENCES refresh_tokens(id), user_agent TEXT, ip_address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id, revoked_at);
CREATE INDEX idx_refresh_tokens_family ON refresh_tokens(family_id);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
SQLEOF
cat > 000006_refresh_tokens.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS refresh_tokens;
SQLEOF

# 0007 - verification_tokens, oauth_states, revoked_tokens
cat > 000007_verification_oauth_revoked.up.sql << 'SQLEOF'
CREATE TABLE verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL, type TEXT NOT NULL CHECK (type IN ('email_verification','password_reset','magic_link','email_change')),
    expires_at TIMESTAMPTZ NOT NULL, used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_verification_tokens_hash ON verification_tokens(token_hash);
CREATE INDEX idx_verification_tokens_user_type ON verification_tokens(user_id, type);
CREATE INDEX idx_verification_tokens_expires ON verification_tokens(expires_at) WHERE used_at IS NULL;

CREATE TABLE oauth_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state TEXT NOT NULL, provider TEXT NOT NULL CHECK (provider IN ('google','github','microsoft')),
    redirect_url TEXT, expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_oauth_states_state ON oauth_states(state);
CREATE INDEX idx_oauth_states_expires ON oauth_states(expires_at);

CREATE TABLE revoked_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jti TEXT NOT NULL, user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), expires_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX idx_revoked_tokens_jti ON revoked_tokens(jti);
CREATE INDEX idx_revoked_tokens_expires ON revoked_tokens(expires_at);
SQLEOF
cat > 000007_verification_oauth_revoked.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS revoked_tokens;
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS verification_tokens;
SQLEOF

# 0008 - api_keys, webhooks, webhook_deliveries
cat > 000008_api_keys_webhooks.up.sql << 'SQLEOF'
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL, key_hash TEXT NOT NULL, key_preview TEXT NOT NULL,
    authority TEXT NOT NULL CHECK (authority IN ('admin','user')),
    created_by UUID REFERENCES users(id), last_used_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_tenant ON api_keys(tenant_id, is_active);
CREATE TRIGGER trigger_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL, url TEXT NOT NULL, secret TEXT,
    events TEXT[] NOT NULL DEFAULT '{}', is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_webhooks_tenant ON webhooks(tenant_id, is_active);
CREATE INDEX idx_webhooks_events ON webhooks USING GIN (events);
CREATE TRIGGER trigger_webhooks_updated_at BEFORE UPDATE ON webhooks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL, payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','delivered','failed')),
    attempts INT NOT NULL DEFAULT 0, response_code INT, response_body TEXT,
    next_attempt_at TIMESTAMPTZ, delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id, created_at DESC);
CREATE INDEX idx_webhook_deliveries_pending ON webhook_deliveries(status, next_attempt_at) WHERE status IN ('pending','failed');
SQLEOF
cat > 000008_api_keys_webhooks.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_webhooks_updated_at ON webhooks;
DROP TRIGGER IF EXISTS trigger_api_keys_updated_at ON api_keys;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS api_keys;
SQLEOF

# 0009 - branding
cat > 000009_branding.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000009_branding.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_custom_pages_updated_at ON custom_pages;
DROP TRIGGER IF EXISTS trigger_branding_configs_updated_at ON branding_configs;
DROP TABLE IF EXISTS custom_pages;
DROP TABLE IF EXISTS branding_assets;
DROP TABLE IF EXISTS branding_configs;
SQLEOF

# 0010 - invitations, audit_log, system_logs
cat > 000010_invitations_audit_logs.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000010_invitations_audit_logs.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS system_logs;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS invitations;
SQLEOF

# 0011 - events, telemetry, usage
cat > 000011_events_telemetry_usage.up.sql << 'SQLEOF'
CREATE TABLE event_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'custom', is_active BOOLEAN NOT NULL DEFAULT TRUE, is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_event_definitions_name ON event_definitions(name);
CREATE INDEX idx_event_definitions_active ON event_definitions(is_active, category);
CREATE TRIGGER trigger_event_definitions_updated_at BEFORE UPDATE ON event_definitions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE telemetry_events (
    id UUID DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id), anonymous_id TEXT,
    event_type TEXT NOT NULL, event_name TEXT NOT NULL,
    properties JSONB NOT NULL DEFAULT '{}', session_id TEXT, ip_address TEXT, user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);
CREATE TABLE telemetry_events_2026_07 PARTITION OF telemetry_events FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE telemetry_events_2026_08 PARTITION OF telemetry_events FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE INDEX idx_telemetry_events_user ON telemetry_events(user_id, created_at DESC);
CREATE INDEX idx_telemetry_events_type ON telemetry_events(event_type, created_at DESC);
CREATE INDEX idx_telemetry_events_name ON telemetry_events(event_name, created_at DESC);

CREATE TABLE usage_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id), type TEXT NOT NULL, amount BIGINT NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_usage_events_tenant_type ON usage_events(tenant_id, type, created_at DESC);
CREATE INDEX idx_usage_events_user ON usage_events(user_id, created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_usage_events_created ON usage_events(created_at DESC);
SQLEOF
cat > 000011_events_telemetry_usage.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS usage_events;
DROP TABLE IF EXISTS telemetry_events;
DROP TABLE IF EXISTS event_definitions;
SQLEOF

# 0012 - metrics, leader_locks
cat > 000012_metrics_leader_locks.up.sql << 'SQLEOF'
CREATE TABLE system_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpu_usage FLOAT, memory_usage FLOAT, disk_usage FLOAT, goroutines INT,
    http_requests JSONB, mongodb_stats JSONB, go_runtime JSONB, integrations JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_system_metrics_time ON system_metrics(timestamp DESC);

CREATE TABLE daily_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL, dau INT NOT NULL DEFAULT 0, wau INT NOT NULL DEFAULT 0, mau INT NOT NULL DEFAULT 0,
    new_users INT NOT NULL DEFAULT 0, new_tenants INT NOT NULL DEFAULT 0, active_tenants INT NOT NULL DEFAULT 0,
    arr_cents BIGINT NOT NULL DEFAULT 0, mrr_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_daily_metrics_date ON daily_metrics(date);

CREATE TABLE leader_locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lock_name TEXT NOT NULL, holder_id TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL, last_renewed TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_leader_locks_name ON leader_locks(lock_name);
CREATE INDEX idx_leader_locks_expires ON leader_locks(expires_at);
SQLEOF
cat > 000012_metrics_leader_locks.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS leader_locks;
DROP TABLE IF EXISTS daily_metrics;
DROP TABLE IF EXISTS system_metrics;
SQLEOF

# 0013 - messages, announcements, config_vars
cat > 000013_messages_announcements_config.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000013_messages_announcements_config.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_config_vars_updated_at ON config_vars;
DROP TRIGGER IF EXISTS trigger_announcements_updated_at ON announcements;
DROP TABLE IF EXISTS config_vars;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS messages;
SQLEOF

# 0014 - sso, webauthn, auth_codes, impersonation
cat > 000014_sso_webauthn_auth_impersonation.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000014_sso_webauthn_auth_impersonation.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_sso_connections_updated_at ON sso_connections;
DROP TABLE IF EXISTS impersonation_logs;
DROP TABLE IF EXISTS auth_codes;
DROP TABLE IF EXISTS webauthn_sessions;
DROP TABLE IF EXISTS webauthn_credentials;
DROP TABLE IF EXISTS sso_connections;
SQLEOF

# 0015 - stripe_mappings, credit_bundles, promotions
cat > 000015_stripe_credits_promotions.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000015_stripe_credits_promotions.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_promotions_updated_at ON promotions;
DROP TRIGGER IF EXISTS trigger_credit_bundles_updated_at ON credit_bundles;
DROP TABLE IF EXISTS promotion_codes;
DROP TABLE IF EXISTS promotions;
DROP TABLE IF EXISTS credit_bundles;
DROP TABLE IF EXISTS stripe_mappings;
SQLEOF

# 0016 - pg_cron jobs
cat > 000016_pg_cron_jobs.up.sql << 'SQLEOF'
DO $$BEGIN CREATE EXTENSION IF NOT EXISTS pg_cron; EXCEPTION WHEN OTHERS THEN NULL; END$$;
CREATE OR REPLACE FUNCTION _schedule_cleanup_jobs() RETURNS void AS $$
DECLARE
    jobs text[] := ARRAY[
        'delete-expired-refresh-tokens|0 * * * *|DELETE FROM refresh_tokens WHERE expires_at < NOW()',
        'delete-expired-verification-tokens|0 * * * *|DELETE FROM verification_tokens WHERE expires_at < NOW()',
        'delete-expired-oauth-states|0 * * * *|DELETE FROM oauth_states WHERE expires_at < NOW()',
        'delete-expired-revoked-tokens|0 * * * *|DELETE FROM revoked_tokens WHERE expires_at < NOW()',
        'delete-expired-invitations|0 0 * * *|DELETE FROM invitations WHERE expires_at < NOW() AND status = ''pending''',
        'delete-expired-webauthn-sessions|0 * * * *|DELETE FROM webauthn_sessions WHERE expires_at < NOW()',
        'delete-expired-auth-codes|0 * * * *|DELETE FROM auth_codes WHERE expires_at < NOW()',
        'delete-old-audit-logs|0 0 * * *|DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL ''90 days''',
        'delete-old-system-metrics|0 0 * * *|DELETE FROM system_metrics WHERE timestamp < NOW() - INTERVAL ''30 days''',
        'delete-old-webhook-deliveries|0 0 * * *|DELETE FROM webhook_deliveries WHERE status = ''delivered'' AND created_at < NOW() - INTERVAL ''30 days'''
    ];
    job text; parts text[];
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
        FOREACH job IN ARRAY jobs LOOP
            parts := string_to_array(job, '|');
            BEGIN PERFORM cron.schedule(parts[1], parts[2], parts[3]); EXCEPTION WHEN OTHERS THEN NULL; END;
        END LOOP;
    END IF;
END;
$$ LANGUAGE plpgsql;
SELECT _schedule_cleanup_jobs();
DROP FUNCTION _schedule_cleanup_jobs();
SQLEOF
cat > 000016_pg_cron_jobs.down.sql << 'SQLEOF'
-- pg_cron jobs are not explicitly unscheduled (safe to leave)
SQLEOF

echo "Migrations 0001-0016 created"
ls *.up.sql | wc -l