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
