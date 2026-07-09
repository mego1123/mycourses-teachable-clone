#!/bin/bash
cd /home/z/my-project/mycourses/backend/migrations

# 0017 - courses
cat > 000017_courses.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000017_courses.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_courses_updated_at ON courses;
DROP TABLE IF EXISTS courses;
SQLEOF

# 0018 - sections
cat > 000018_sections.up.sql << 'SQLEOF'
CREATE TABLE sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0, drip_offset_days INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sections_course_order ON sections(course_id, sort_order);
CREATE TRIGGER trigger_sections_updated_at BEFORE UPDATE ON sections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000018_sections.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_sections_updated_at ON sections;
DROP TABLE IF EXISTS sections;
SQLEOF

# 0019 - media_assets (before lessons which references it)
cat > 000019_media_assets.up.sql << 'SQLEOF'
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
SQLEOF
cat > 000019_media_assets.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_media_assets_updated_at ON media_assets;
DROP TABLE IF EXISTS media_assets;
SQLEOF

# 0020 - lessons
cat > 000020_lessons.up.sql << 'SQLEOF'
CREATE TABLE lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id UUID NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title TEXT NOT NULL, type TEXT NOT NULL DEFAULT 'video' CHECK (type IN ('video','text','pdf','quiz')),
    content TEXT NOT NULL DEFAULT '', media_asset_id UUID REFERENCES media_assets(id) ON DELETE SET NULL,
    sort_order INT NOT NULL DEFAULT 0, is_preview BOOLEAN NOT NULL DEFAULT FALSE, duration_sec INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_lessons_section_order ON lessons(section_id, sort_order);
CREATE INDEX idx_lessons_course ON lessons(course_id, sort_order);
CREATE INDEX idx_lessons_preview ON lessons(course_id, is_preview) WHERE is_preview = TRUE;
CREATE TRIGGER trigger_lessons_updated_at BEFORE UPDATE ON lessons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000020_lessons.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_lessons_updated_at ON lessons;
DROP TABLE IF EXISTS lessons;
SQLEOF

# 0021 - enrollments
cat > 000021_enrollments.up.sql << 'SQLEOF'
CREATE TABLE enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed','refunded','disputed','chargeback_lost')),
    price_paid_cents BIGINT NOT NULL DEFAULT 0, currency TEXT NOT NULL DEFAULT 'usd',
    coupon_id UUID, stripe_session_id TEXT, stripe_charge_id TEXT,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), completed_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ, refund_reason TEXT, expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_enrollments_course_user ON enrollments(course_id, user_id);
CREATE INDEX idx_enrollments_user_status ON enrollments(user_id, status);
CREATE INDEX idx_enrollments_tenant_date ON enrollments(tenant_id, enrolled_at DESC);
CREATE INDEX idx_enrollments_stripe_session ON enrollments(stripe_session_id) WHERE stripe_session_id IS NOT NULL;
CREATE INDEX idx_enrollments_stripe_charge ON enrollments(stripe_charge_id) WHERE stripe_charge_id IS NOT NULL;
CREATE TRIGGER trigger_enrollments_updated_at BEFORE UPDATE ON enrollments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000021_enrollments.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_enrollments_updated_at ON enrollments;
DROP TABLE IF EXISTS enrollments;
SQLEOF

# 0022 - course_progress
cat > 000022_course_progress.up.sql << 'SQLEOF'
CREATE TABLE course_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    completed BOOLEAN NOT NULL DEFAULT FALSE, completed_at TIMESTAMPTZ,
    video_position_sec INT NOT NULL DEFAULT 0, last_viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_progress_enrollment_lesson ON course_progress(enrollment_id, lesson_id);
CREATE INDEX idx_progress_user ON course_progress(user_id, last_viewed_at DESC);
CREATE INDEX idx_progress_enrollment ON course_progress(enrollment_id, completed);
CREATE TRIGGER trigger_course_progress_updated_at BEFORE UPDATE ON course_progress FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000022_course_progress.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_course_progress_updated_at ON course_progress;
DROP TABLE IF EXISTS course_progress;
SQLEOF

# 0023 - course_coupons
cat > 000023_course_coupons.up.sql << 'SQLEOF'
CREATE TABLE course_coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code TEXT NOT NULL, discount_type TEXT NOT NULL CHECK (discount_type IN ('percent','fixed')),
    discount_value INT NOT NULL, currency TEXT NOT NULL DEFAULT 'usd',
    course_id UUID REFERENCES courses(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ, usage_limit INT, used_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_coupons_tenant_code ON course_coupons(tenant_id, code);
CREATE INDEX idx_coupons_tenant_active ON course_coupons(tenant_id, is_active);
CREATE INDEX idx_coupons_course ON course_coupons(course_id) WHERE course_id IS NOT NULL;
CREATE INDEX idx_coupons_expires ON course_coupons(expires_at) WHERE expires_at IS NOT NULL;
CREATE TRIGGER trigger_course_coupons_updated_at BEFORE UPDATE ON course_coupons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000023_course_coupons.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_course_coupons_updated_at ON course_coupons;
DROP TABLE IF EXISTS course_coupons;
SQLEOF

# 0024 - reviews
cat > 000024_reviews.up.sql << 'SQLEOF'
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT NOT NULL DEFAULT '', is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_reviews_course_user ON reviews(course_id, user_id);
CREATE INDEX idx_reviews_course ON reviews(course_id, is_hidden, created_at DESC);
CREATE INDEX idx_reviews_user ON reviews(user_id, created_at DESC);
CREATE INDEX idx_reviews_tenant ON reviews(tenant_id, created_at DESC);
CREATE TRIGGER trigger_reviews_updated_at BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000024_reviews.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_reviews_updated_at ON reviews;
DROP TABLE IF EXISTS reviews;
SQLEOF

# 0025 - payouts
cat > 000025_payouts.up.sql << 'SQLEOF'
CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    stripe_transfer_id TEXT, stripe_payout_id TEXT,
    amount_cents BIGINT NOT NULL, currency TEXT NOT NULL DEFAULT 'usd',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','failed','cancelled')),
    failure_reason TEXT,
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payouts_tenant_date ON payouts(tenant_id, initiated_at DESC);
CREATE INDEX idx_payouts_status ON payouts(status, initiated_at DESC);
CREATE INDEX idx_payouts_stripe_transfer ON payouts(stripe_transfer_id) WHERE stripe_transfer_id IS NOT NULL;
CREATE INDEX idx_payouts_stripe_payout ON payouts(stripe_payout_id) WHERE stripe_payout_id IS NOT NULL;
CREATE TRIGGER trigger_payouts_updated_at BEFORE UPDATE ON payouts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000025_payouts.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_payouts_updated_at ON payouts;
DROP TABLE IF EXISTS payouts;
SQLEOF

# 0026 - custom_domains
cat > 000026_custom_domains.up.sql << 'SQLEOF'
CREATE TABLE custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain TEXT NOT NULL, tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','active','failed')),
    dns_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ssl_status TEXT NOT NULL DEFAULT 'pending' CHECK (ssl_status IN ('pending','active','failed')),
    cf_hostname_id TEXT, verification_records JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX idx_custom_domains_tenant ON custom_domains(tenant_id, status);
CREATE INDEX idx_custom_domains_status ON custom_domains(status, dns_verified, ssl_status);
CREATE TRIGGER trigger_custom_domains_updated_at BEFORE UPDATE ON custom_domains FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000026_custom_domains.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_custom_domains_updated_at ON custom_domains;
DROP TABLE IF EXISTS custom_domains;
SQLEOF

# 0027 - certificates
cat > 000027_certificates.up.sql << 'SQLEOF'
CREATE TABLE certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    certificate_number TEXT NOT NULL, verification_token TEXT NOT NULL,
    learner_name TEXT NOT NULL, course_title TEXT NOT NULL, creator_name TEXT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','revoked')),
    revoked_reason TEXT, revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_certificates_number ON certificates(certificate_number);
CREATE UNIQUE INDEX idx_certificates_token ON certificates(verification_token);
CREATE UNIQUE INDEX idx_certificates_enrollment ON certificates(enrollment_id);
CREATE INDEX idx_certificates_user ON certificates(user_id, issued_at DESC);
CREATE INDEX idx_certificates_tenant ON certificates(tenant_id, issued_at DESC);
SQLEOF
cat > 000027_certificates.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS certificates;
SQLEOF

# 0028 - creator_profiles
cat > 000028_creator_profiles.up.sql << 'SQLEOF'
CREATE TABLE creator_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    bio TEXT NOT NULL DEFAULT '', website_url TEXT,
    social_links JSONB NOT NULL DEFAULT '{}', avatar_url TEXT, banner_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trigger_creator_profiles_updated_at BEFORE UPDATE ON creator_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
SQLEOF
cat > 000028_creator_profiles.down.sql << 'SQLEOF'
DROP TRIGGER IF EXISTS trigger_creator_profiles_updated_at ON creator_profiles;
DROP TABLE IF EXISTS creator_profiles;
SQLEOF

# 0029 - processed_stripe_events
cat > 000029_processed_stripe_events.up.sql << 'SQLEOF'
CREATE TABLE processed_stripe_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_processed_stripe_type ON processed_stripe_events(event_type, processed_at DESC);
SQLEOF
cat > 000029_processed_stripe_events.down.sql << 'SQLEOF'
DROP TABLE IF EXISTS processed_stripe_events;
SQLEOF

echo "Course migrations 0017-0029 created"
ls *.up.sql | wc -l