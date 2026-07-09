CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', slug TEXT NOT NULL,
    price_cents BIGINT NOT NULL DEFAULT 0, currency TEXT NOT NULL DEFAULT 'usd',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
    thumbnail_url TEXT, intro_video_url TEXT, category TEXT,
    drip_enabled BOOLEAN NOT NULL DEFAULT FALSE, published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_courses_tenant_slug ON courses(tenant_id, slug);
CREATE INDEX idx_courses_tenant_status ON courses(tenant_id, status, created_at DESC);
CREATE INDEX idx_courses_marketplace ON courses(status, published_at DESC) WHERE status = 'published';
CREATE INDEX idx_courses_category ON courses(category, status) WHERE status = 'published';
CREATE INDEX idx_courses_title_trgm ON courses USING GIN (title gin_trgm_ops);
CREATE INDEX idx_courses_search ON courses USING GIN (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'')));
CREATE TRIGGER trigger_courses_updated_at BEFORE UPDATE ON courses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
