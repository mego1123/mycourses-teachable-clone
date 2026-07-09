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
