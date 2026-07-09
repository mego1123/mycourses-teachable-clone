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
