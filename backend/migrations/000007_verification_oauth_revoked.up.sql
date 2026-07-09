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
