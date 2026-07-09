CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('video','pdf','image')),
    title TEXT NOT NULL DEFAULT '',
    cf_stream_id TEXT, r2_key TEXT, r2_url TEXT,
    size_bytes BIGINT NOT NULL DEFAULT 0, mime_type TEXT NOT NULL DEFAULT '',
    duration_sec INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'processing' CHECK (status IN ('processing','ready','failed','deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_media_tenant_kind ON media_assets(tenant_id, kind, created_at DESC);
CREATE INDEX idx_media_cf_stream ON media_assets(cf_stream_id) WHERE cf_stream_id IS NOT NULL;
CREATE INDEX idx_media_status ON media_assets(status);
CREATE TRIGGER trigger_media_assets_updated_at BEFORE UPDATE ON media_assets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
