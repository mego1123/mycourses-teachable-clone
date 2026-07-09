CREATE TABLE creator_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    bio TEXT NOT NULL DEFAULT '', website_url TEXT,
    social_links JSONB NOT NULL DEFAULT '{}', avatar_url TEXT, banner_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trigger_creator_profiles_updated_at BEFORE UPDATE ON creator_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
