CREATE TABLE sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0, drip_offset_days INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sections_course_order ON sections(course_id, sort_order);
CREATE TRIGGER trigger_sections_updated_at BEFORE UPDATE ON sections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
