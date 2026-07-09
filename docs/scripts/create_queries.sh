#!/bin/bash
cd /home/z/my-project/mycourses/backend/queries

cat > users.sql << 'EOF'
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;
-- name: GetUserByEmailNormalized :one
SELECT * FROM users WHERE email_normalized = $1 AND deleted_at IS NULL;
-- name: CreateUser :one
INSERT INTO users (email, email_normalized, display_name, password_hash, locale_preference, theme_preference, email_verified, is_active, is_admin, consent_accepted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;
-- name: UpdateUserProfile :one
UPDATE users SET display_name = $2, avatar_url = $3, bio = $4, locale_preference = $5, theme_preference = $6
WHERE id = $1 AND deleted_at IS NULL RETURNING *;
-- name: VerifyUserEmail :exec
UPDATE users SET email_verified = TRUE, email_verified_at = NOW() WHERE id = $1;
-- name: UpdateLastLogin :exec
UPDATE users SET last_login_at = NOW() WHERE id = $1;
-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW(), is_active = FALSE WHERE id = $1;
-- name: ListUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;
-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;
-- name: AdminListUsers :many
SELECT * FROM users WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR email_normalized ILIKE '%' || $1 || '%') ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: AdminCountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR email_normalized ILIKE '%' || $1 || '%');
EOF

cat > plans.sql << 'EOF'
-- name: GetPlanByID :one
SELECT * FROM plans WHERE id = $1;
-- name: ListActivePublicPlans :many
SELECT * FROM plans WHERE is_active = TRUE AND is_public = TRUE ORDER BY sort_order ASC, monthly_price_cents ASC;
-- name: CreatePlan :one
INSERT INTO plans (name, description, monthly_price_cents, yearly_price_cents, currency, included_seats, min_seats, max_seats, usage_credits_per_month, trial_period_days, entitlements, is_system, is_active, is_public, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING *;
-- name: UpsertPlan :one
INSERT INTO plans (name, description, monthly_price_cents, yearly_price_cents, currency, included_seats, min_seats, max_seats, usage_credits_per_month, trial_period_days, entitlements, is_system, is_active, is_public, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (name) DO UPDATE SET monthly_price_cents = EXCLUDED.monthly_price_cents, entitlements = EXCLUDED.entitlements, is_active = EXCLUDED.is_active, sort_order = EXCLUDED.sort_order, updated_at = NOW()
RETURNING *;
EOF

cat > tenants.sql << 'EOF'
-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1 AND deleted_at IS NULL;
-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1 AND deleted_at IS NULL;
-- name: GetRootTenant :one
SELECT * FROM tenants WHERE is_root = TRUE AND deleted_at IS NULL LIMIT 1;
-- name: CreateTenant :one
INSERT INTO tenants (name, slug, plan_id, billing_status, stripe_connect_status, commission_rate_bps, payout_schedule, is_root, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: UpdateTenantStripeConnect :exec
UPDATE tenants SET stripe_connect_account_id = $2, stripe_connect_status = $3 WHERE id = $1;
-- name: UpdateTenantCommissionRate :exec
UPDATE tenants SET commission_rate_bps = $2 WHERE id = $1;
-- name: SoftDeleteTenant :exec
UPDATE tenants SET deleted_at = NOW(), is_active = FALSE WHERE id = $1;
-- name: ListTenants :many
SELECT * FROM tenants WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR name ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%') AND ($2::text IS NULL OR $2 = '' OR billing_status = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4;
-- name: CountTenants :one
SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL AND ($1::text IS NULL OR $1 = '' OR name ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%') AND ($2::text IS NULL OR $2 = '' OR billing_status = $2);
-- name: CountActiveTenants :one
SELECT COUNT(*) FROM tenants WHERE billing_status = 'active' AND deleted_at IS NULL;
EOF

cat > tenant_memberships.sql << 'EOF'
-- name: GetMembership :one
SELECT * FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2;
-- name: ListMembershipsByTenant :many
SELECT * FROM tenant_memberships WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: CountMembershipsByTenant :one
SELECT COUNT(*) FROM tenant_memberships WHERE tenant_id = $1;
-- name: ListMembershipsByUser :many
SELECT * FROM tenant_memberships WHERE user_id = $1 ORDER BY created_at DESC;
-- name: CreateMembership :one
INSERT INTO tenant_memberships (tenant_id, user_id, role) VALUES ($1, $2, $3) RETURNING *;
-- name: UpdateMembershipRole :exec
UPDATE tenant_memberships SET role = $3 WHERE tenant_id = $1 AND user_id = $2;
-- name: DeleteMembership :exec
DELETE FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2;
-- name: CountMembersByRole :one
SELECT COUNT(*) FROM tenant_memberships WHERE tenant_id = $1 AND role = $2;
EOF

cat > financial_transactions.sql << 'EOF'
-- name: CreateTransaction :one
INSERT INTO financial_transactions (tenant_id, user_id, type, amount_cents, currency, description, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;
-- name: ListTransactionsByTenant :many
SELECT * FROM financial_transactions WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR type = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4;
-- name: SumTransactionsByTenantAndType :one
SELECT COALESCE(SUM(amount_cents), 0)::bigint as total_cents FROM financial_transactions WHERE tenant_id = $1 AND type = $2 AND created_at >= $3 AND created_at < $4;
EOF

cat > courses.sql << 'EOF'
-- name: GetCourseByID :one
SELECT * FROM courses WHERE id = $1;
-- name: GetCourseBySlug :one
SELECT * FROM courses WHERE tenant_id = $1 AND slug = $2;
-- name: GetPublishedCourseBySlug :one
SELECT * FROM courses WHERE tenant_id = $1 AND slug = $2 AND status = 'published';
-- name: CreateCourse :one
INSERT INTO courses (tenant_id, title, description, slug, price_cents, currency, status, thumbnail_url, intro_video_url, category, drip_enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;
-- name: UpdateCourse :one
UPDATE courses SET title = $2, description = $3, slug = $4, price_cents = $5, currency = $6, thumbnail_url = $7, intro_video_url = $8, category = $9, drip_enabled = $10 WHERE id = $1 RETURNING *;
-- name: PublishCourse :one
UPDATE courses SET status = 'published', published_at = NOW() WHERE id = $1 RETURNING *;
-- name: UnpublishCourse :one
UPDATE courses SET status = 'draft', published_at = NULL WHERE id = $1 RETURNING *;
-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;
-- name: ListCoursesByTenant :many
SELECT * FROM courses WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR status = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4;
-- name: CountCoursesByTenant :one
SELECT COUNT(*) FROM courses WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR status = $2);
-- name: ListMarketplaceCourses :many
SELECT * FROM courses WHERE status = 'published' AND ($1::text IS NULL OR $1 = '' OR category = $1) ORDER BY published_at DESC LIMIT $2 OFFSET $3;
-- name: SearchCourses :many
SELECT * FROM courses WHERE status = 'published' AND (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'')) @@ plainto_tsquery('english', $1) OR title % $1) ORDER BY ts_rank(to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'')), plainto_tsquery('english', $1)) DESC LIMIT $2 OFFSET $3;
EOF

cat > sections.sql << 'EOF'
-- name: GetSectionByID :one
SELECT * FROM sections WHERE id = $1;
-- name: CreateSection :one
INSERT INTO sections (course_id, title, description, sort_order, drip_offset_days) VALUES ($1, $2, $3, $4, $5) RETURNING *;
-- name: UpdateSection :one
UPDATE sections SET title = $2, description = $3, sort_order = $4, drip_offset_days = $5 WHERE id = $1 RETURNING *;
-- name: DeleteSection :exec
DELETE FROM sections WHERE id = $1;
-- name: ListSectionsByCourse :many
SELECT * FROM sections WHERE course_id = $1 ORDER BY sort_order ASC;
EOF

cat > lessons.sql << 'EOF'
-- name: GetLessonByID :one
SELECT * FROM lessons WHERE id = $1;
-- name: CreateLesson :one
INSERT INTO lessons (section_id, course_id, title, type, content, media_asset_id, sort_order, is_preview, duration_sec)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: UpdateLesson :one
UPDATE lessons SET title = $2, type = $3, content = $4, media_asset_id = $5, sort_order = $6, is_preview = $7, duration_sec = $8 WHERE id = $1 RETURNING *;
-- name: DeleteLesson :exec
DELETE FROM lessons WHERE id = $1;
-- name: ListLessonsBySection :many
SELECT * FROM lessons WHERE section_id = $1 ORDER BY sort_order ASC;
-- name: ListLessonsByCourse :many
SELECT * FROM lessons WHERE course_id = $1 ORDER BY sort_order ASC;
-- name: ListPreviewLessonsByCourse :many
SELECT * FROM lessons WHERE course_id = $1 AND is_preview = TRUE ORDER BY sort_order ASC;
-- name: CountLessonsByCourse :one
SELECT COUNT(*) FROM lessons WHERE course_id = $1;
EOF

cat > media_assets.sql << 'EOF'
-- name: GetMediaAssetByID :one
SELECT * FROM media_assets WHERE id = $1;
-- name: CreateMediaAsset :one
INSERT INTO media_assets (tenant_id, kind, title, cf_stream_id, r2_key, r2_url, size_bytes, mime_type, duration_sec, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;
-- name: UpdateMediaAssetStatus :exec
UPDATE media_assets SET status = $2 WHERE id = $1;
-- name: ListMediaAssetsByTenant :many
SELECT * FROM media_assets WHERE tenant_id = $1 AND status != 'deleted' ORDER BY created_at DESC LIMIT $2 OFFSET $3;
EOF

cat > enrollments.sql << 'EOF'
-- name: GetEnrollmentByID :one
SELECT * FROM enrollments WHERE id = $1;
-- name: GetEnrollmentByCourseAndUser :one
SELECT * FROM enrollments WHERE course_id = $1 AND user_id = $2;
-- name: GetEnrollmentByStripeSessionID :one
SELECT * FROM enrollments WHERE stripe_session_id = $1;
-- name: CreateEnrollment :one
INSERT INTO enrollments (course_id, tenant_id, user_id, status, price_paid_cents, currency, coupon_id, stripe_session_id, stripe_charge_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: CompleteEnrollment :exec
UPDATE enrollments SET status = 'completed', completed_at = NOW() WHERE id = $1;
-- name: RefundEnrollment :exec
UPDATE enrollments SET status = 'refunded', refunded_at = NOW(), refund_reason = $2 WHERE id = $1;
-- name: ListEnrollmentsByUser :many
SELECT * FROM enrollments WHERE user_id = $1 AND status = ANY($2::text[]) ORDER BY enrolled_at DESC LIMIT $3 OFFSET $4;
-- name: ListEnrollmentsByTenant :many
SELECT * FROM enrollments WHERE tenant_id = $1 AND ($2::text IS NULL OR $2 = '' OR status = $2) ORDER BY enrolled_at DESC LIMIT $3 OFFSET $4;
-- name: CountEnrollmentsByCourse :one
SELECT COUNT(*) FROM enrollments WHERE course_id = $1 AND status = 'active';
EOF

cat > course_progress.sql << 'EOF'
-- name: GetProgressByEnrollmentAndLesson :one
SELECT * FROM course_progress WHERE enrollment_id = $1 AND lesson_id = $2;
-- name: UpsertProgress :one
INSERT INTO course_progress (enrollment_id, lesson_id, user_id, completed, video_position_sec, last_viewed_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (enrollment_id, lesson_id) DO UPDATE SET completed = EXCLUDED.completed, completed_at = CASE WHEN EXCLUDED.completed = TRUE THEN COALESCE(course_progress.completed_at, NOW()) ELSE NULL END, video_position_sec = EXCLUDED.video_position_sec, last_viewed_at = NOW()
RETURNING *;
-- name: ListProgressByEnrollment :many
SELECT * FROM course_progress WHERE enrollment_id = $1 ORDER BY last_viewed_at DESC;
-- name: CountCompletedLessonsByEnrollment :one
SELECT COUNT(*) FROM course_progress WHERE enrollment_id = $1 AND completed = TRUE;
-- name: GetCourseCompletionPercentage :one
SELECT CASE WHEN COUNT(*) = 0 THEN 0 ELSE (COUNT(cp.id) * 100 / COUNT(*))::int END as percentage FROM lessons LEFT JOIN course_progress cp ON cp.lesson_id = lessons.id AND cp.enrollment_id = $1 AND cp.completed = TRUE WHERE lessons.course_id = (SELECT course_id FROM enrollments WHERE id = $1);
EOF

cat > course_coupons.sql << 'EOF'
-- name: GetCouponByID :one
SELECT * FROM course_coupons WHERE id = $1;
-- name: CreateCoupon :one
INSERT INTO course_coupons (tenant_id, code, discount_type, discount_value, currency, course_id, expires_at, usage_limit)
VALUES ($1, UPPER($2), $3, $4, $5, $6, $7, $8) RETURNING *;
-- name: DeleteCoupon :exec
DELETE FROM course_coupons WHERE id = $1;
-- name: IncrementCouponUsage :exec
UPDATE course_coupons SET used_count = used_count + 1 WHERE id = $1;
-- name: ListCouponsByTenant :many
SELECT * FROM course_coupons WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: ValidateCoupon :one
SELECT * FROM course_coupons WHERE tenant_id = $1 AND code = UPPER($2) AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW()) AND (usage_limit IS NULL OR used_count < usage_limit);
EOF

cat > reviews.sql << 'EOF'
-- name: GetReviewByID :one
SELECT * FROM reviews WHERE id = $1;
-- name: CreateReview :one
INSERT INTO reviews (course_id, user_id, tenant_id, rating, comment) VALUES ($1, $2, $3, $4, $5) RETURNING *;
-- name: DeleteReview :exec
DELETE FROM reviews WHERE id = $1;
-- name: HideReview :exec
UPDATE reviews SET is_hidden = TRUE WHERE id = $1;
-- name: ListPublicReviewsByCourse :many
SELECT * FROM reviews WHERE course_id = $1 AND is_hidden = FALSE ORDER BY created_at DESC LIMIT $2 OFFSET $3;
-- name: GetAverageRatingByCourse :one
SELECT COALESCE(AVG(rating), 0)::float as avg_rating, COUNT(*)::int as review_count FROM reviews WHERE course_id = $1 AND is_hidden = FALSE;
EOF

cat > payouts.sql << 'EOF'
-- name: GetPayoutByID :one
SELECT * FROM payouts WHERE id = $1;
-- name: GetPayoutByStripeTransferID :one
SELECT * FROM payouts WHERE stripe_transfer_id = $1;
-- name: CreatePayout :one
INSERT INTO payouts (tenant_id, stripe_transfer_id, stripe_payout_id, amount_cents, currency, status)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
-- name: UpdatePayoutStatus :exec
UPDATE payouts SET status = $2, failure_reason = $3, completed_at = CASE WHEN $2 = 'paid' THEN NOW() ELSE completed_at END WHERE id = $1;
-- name: ListPayoutsByTenant :many
SELECT * FROM payouts WHERE tenant_id = $1 ORDER BY initiated_at DESC LIMIT $2 OFFSET $3;
-- name: SumPayoutsByTenant :one
SELECT COALESCE(SUM(amount_cents), 0)::bigint as total_cents FROM payouts WHERE tenant_id = $1 AND status = 'paid';
EOF

cat > custom_domains.sql << 'EOF'
-- name: GetCustomDomainByID :one
SELECT * FROM custom_domains WHERE id = $1;
-- name: GetTenantByCustomDomain :one
SELECT t.* FROM tenants t JOIN custom_domains cd ON cd.tenant_id = t.id WHERE cd.domain = $1 AND cd.status = 'active' AND t.deleted_at IS NULL;
-- name: CreateCustomDomain :one
INSERT INTO custom_domains (domain, tenant_id, status, verification_records) VALUES ($1, $2, 'pending', $3) RETURNING *;
-- name: UpdateCustomDomainStatus :exec
UPDATE custom_domains SET status = $2, dns_verified = $3, ssl_status = $4, cf_hostname_id = $5 WHERE id = $1;
-- name: DeleteCustomDomain :exec
DELETE FROM custom_domains WHERE id = $1;
-- name: ListCustomDomainsByTenant :many
SELECT * FROM custom_domains WHERE tenant_id = $1 ORDER BY created_at DESC;
-- name: ListPendingCustomDomains :many
SELECT * FROM custom_domains WHERE status = 'pending' ORDER BY created_at ASC;
EOF

cat > certificates.sql << 'EOF'
-- name: GetCertificateByID :one
SELECT * FROM certificates WHERE id = $1;
-- name: GetCertificateByToken :one
SELECT * FROM certificates WHERE verification_token = $1 AND status = 'active';
-- name: CreateCertificate :one
INSERT INTO certificates (enrollment_id, user_id, course_id, tenant_id, certificate_number, verification_token, learner_name, course_title, creator_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: RevokeCertificate :exec
UPDATE certificates SET status = 'revoked', revoked_reason = $2, revoked_at = NOW() WHERE id = $1;
-- name: ListCertificatesByUser :many
SELECT * FROM certificates WHERE user_id = $1 AND status = 'active' ORDER BY issued_at DESC;
-- name: GetNextCertificateNumber :one
SELECT 'CERT-' || to_char(NOW(), 'YYYY') || '-' || LPAD((COUNT(*) + 1)::text, 6, '0') as cert_number FROM certificates WHERE created_at >= date_trunc('year', NOW());
EOF

cat > creator_profiles.sql << 'EOF'
-- name: GetCreatorProfileByTenant :one
SELECT * FROM creator_profiles WHERE tenant_id = $1;
-- name: UpsertCreatorProfile :one
INSERT INTO creator_profiles (tenant_id, bio, website_url, social_links, avatar_url, banner_url)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id) DO UPDATE SET bio = EXCLUDED.bio, website_url = EXCLUDED.website_url, social_links = EXCLUDED.social_links, avatar_url = EXCLUDED.avatar_url, banner_url = EXCLUDED.banner_url, updated_at = NOW()
RETURNING *;
EOF

cat > processed_stripe_events.sql << 'EOF'
-- name: GetProcessedStripeEvent :one
SELECT * FROM processed_stripe_events WHERE event_id = $1;
-- name: MarkStripeEventProcessed :exec
INSERT INTO processed_stripe_events (event_id, event_type) VALUES ($1, $2) ON CONFLICT (event_id) DO NOTHING;
EOF

echo "Query files created"
ls *.sql | wc -l