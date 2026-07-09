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
