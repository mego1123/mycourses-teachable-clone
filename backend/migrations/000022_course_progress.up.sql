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
