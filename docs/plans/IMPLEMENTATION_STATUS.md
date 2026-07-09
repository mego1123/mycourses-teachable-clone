# Implementation Status — Honest Assessment

> This document tracks what was actually implemented vs. what the v4 plan specified.
> Generated after all development work was completed.

---

## Phase-by-Phase Assessment

### Phase 0: Postgres Migration — ✅ MOSTLY DONE

**What was done:**
- ✅ 29 migrations created and tested (16 core + 13 course-specific)
- ✅ 18 sqlc query files → 21 generated Go files (4,470+ LOC)
- ✅ `internal/db/postgres.go` with connection pool, migration runner, transaction helper
- ✅ pg_cron cleanup jobs (10 scheduled jobs for TTL tables)
- ✅ Postgres full-text search with pg_trgm
- ✅ E2E tests using embedded-postgres (all passing)

**What was NOT done:**
- ❌ Existing lastsaas MongoDB handlers were NOT rewritten to use Postgres (1,583 MongoDB call sites across 78 files). The existing auth, billing, branding, admin, etc. handlers still use MongoDB. Only the NEW course-specific handlers use Postgres.
- ❌ MongoDB driver was NOT removed from go.mod
- ❌ `internal/db/mongodb.go` and `internal/db/schema.go` were NOT deleted
- ❌ Auth middleware was NOT migrated to Postgres (still uses MongoDB `*models.User`)

**Impact**: The app runs in a hybrid mode — MongoDB for existing lastsaas functionality, Postgres for course platform. The `getUserIDFromContext` function in course handlers returns `nil` because the auth middleware stores MongoDB user objects, not Postgres UUIDs. This means **creator/learner endpoints can't identify the authenticated user** until auth is migrated or a bridge is built.

---

### Phase 1: Course Domain Models — ✅ DONE

**What was done:**
- ✅ 13 course tables created (courses, sections, lessons, media_assets, enrollments, course_progress, course_coupons, reviews, payouts, custom_domains, certificates, creator_profiles, processed_stripe_events)
- ✅ All tables have proper indexes, constraints, FK relationships
- ✅ sqlc queries for all tables
- ✅ Entitlements support (max_courses, custom_domain_enabled, commission_rate_bps)

**What was NOT done:**
- ❌ Creator plan seeding (Creator Free/Pro/Business) was NOT implemented — plans exist in the schema but aren't seeded
- ❌ Configstore defaults (commission_rate, payout_schedule, etc.) were NOT seeded

---

### Phase 2: Creator Studio Backend — ✅ MOSTLY DONE

**What was done:**
- ✅ CourseHandler (CRUD + publish/unpublish + storefront + marketplace)
- ✅ SectionHandler (CRUD)
- ✅ LessonHandler (CRUD)
- ✅ EnrollmentHandler (checkout + list)
- ✅ CourseProgressHandler (upsert + auto-certificate)
- ✅ CouponHandler (CRUD + validate)
- ✅ ReviewHandler (create + list + hide)
- ✅ CustomDomainHandler (CRUD + Cloudflare integration)
- ✅ CertificateHandler (list + verify)
- ✅ StorefrontHandler (checkout session creation)
- ✅ ConnectHandler (onboarding + status)
- ✅ PayoutHandler (list + request)
- ✅ CourseWebhookHandler (9 event types, idempotent)
- ✅ SEOHandler (sitemap + robots.txt)

**What was NOT done:**
- ❌ MediaHandler (Cloudflare Stream direct upload URL generation) — the `media_assets` table and queries exist but no API endpoint creates them. Video lessons have no way to upload video files.
- ❌ Auth middleware not wired — handlers can't identify the authenticated user (returns `nil`)
- ❌ RequireEntitlement checks not implemented in handlers (no max_courses enforcement)
- ❌ Rate limits not applied to new storefront endpoints

---

### Phase 3: Creator Studio Frontend — ✅ DONE

**What was done:**
- ✅ CreatorLayout with sidebar nav (Dashboard, Courses, Sales, Payouts, Coupons, Domain, Reviews)
- ✅ CreatorDashboardPage (revenue/enrollments/courses stats + recent courses)
- ✅ CoursesPage (course table + create form + publish/unpublish + delete)
- ✅ CourseEditorPage (tabbed: Details, Curriculum, Pricing — with section + lesson management)
- ✅ CouponsPage (CRUD with percent/fixed, usage limit)
- ✅ SalesPage (revenue stats + transaction table)
- ✅ CustomDomainPage (add domain + DNS instructions + status badges)
- ✅ ReviewsPage (review list + hide button)
- ✅ ConnectOnboardingPage (Stripe Connect onboarding + status polling)
- ✅ PayoutsPage (payout history + request payout)
- ✅ Routes wired into App.tsx with lazy loading

---

### Phase 4: Storefront & Learner Frontend — ✅ MOSTLY DONE

**What was done:**
- ✅ StorefrontHomePage (course grid + hero)
- ✅ CourseDetailPage (header, curriculum, reviews, JSON-LD, checkout CTA)
- ✅ CheckoutPage (coupon entry + Stripe Checkout redirect)
- ✅ CertificateVerifyPage (public verification by token)
- ✅ MyCoursesPage (enrollments + certificates)
- ✅ CoursePlayerPage (HTML5 video player + lesson sidebar + progress tracking)
- ✅ useCourseProgress hook (15-second heartbeat, auto-complete at 90%)
- ✅ Routes wired into App.tsx

**What was NOT done:**
- ❌ Video player has empty `src=""` — not wired to actual Cloudflare Stream signed URLs (because MediaHandler doesn't exist)
- ❌ No actual video playback possible yet
- ❌ CoursePlayerPage's progress sidebar uses `sectionLessons` state that loads asynchronously but doesn't show loading state

---

### Phase 5: Stripe Connect — ✅ MOSTLY DONE

**What was done:**
- ✅ `internal/stripe/connect.go` service (OnboardCreator, CreateCourseCheckout, InitiatePayout, ReverseTransfer, GetAccountStatus, VerifyWebhookSignature)
- ✅ ConnectHandler (onboarding URL + live status from Stripe)
- ✅ PayoutHandler (list + request with available balance calculation)
- ✅ CourseWebhookHandler (9 event types: checkout.session.completed, charge.refunded, transfer.reversed, charge.dispute.created/funds_withdrawn/funds_reinstated, account.updated, payout.paid, payout.failed)
- ✅ Webhook idempotency via `processed_stripe_events` table
- ✅ Frontend ConnectOnboardingPage (redirects to Stripe, polls status)
- ✅ Frontend PayoutsPage (payout history + request payout button)
- ✅ StorefrontHandler.CreateCheckout (creates Stripe Checkout Session with destination charges)
- ✅ Frontend CheckoutPage redirects to Stripe Checkout URL

**What was NOT done:**
- ❌ Auth not wired — `CreateCheckout` can't get the learner's user ID (returns 401)
- ❌ No Stripe webhook signature verification test (requires real Stripe CLI)
- ❌ No refund initiation from creator dashboard (webhook handler exists but no UI to trigger refunds)

---

### Phase 6: Custom Domains — ✅ DONE

**What was done:**
- ✅ `internal/cloudflare/client.go` (CreateCustomHostname, GetCustomHostname, DeleteCustomHostname)
- ✅ CustomDomainHandler fully integrated with Cloudflare API
- ✅ CustomDomainTenantMiddleware (wildcard subdomain + custom domain resolution)
- ✅ Frontend CustomDomainPage with DNS instructions + status polling
- ✅ `GetStatus` endpoint for live Cloudflare status sync
- ✅ Routes wired (`/api/creator/custom-domains/{id}/status`)

**What was NOT done:**
- ❌ Background worker to poll pending domains (DNS verification happens on-demand via `GetStatus`, not proactively)
- ❌ Cloudflare for SaaS not actually tested (requires real CF credentials)

---

### Phase 7: Certificates, Drip, Reviews — ✅ MOSTLY DONE

**What was done:**
- ✅ Certificate auto-issuance on course completion (in CourseProgressHandler — checks if all lessons complete, creates certificate)
- ✅ CertificateHandler (list mine + verify by token)
- ✅ CertificateVerifyPage (public verification with styled certificate display)
- ✅ RequireDripEligibility middleware (checks section.drip_offset_days against enrollment date)
- ✅ ReviewHandler (create with enrollment check, list public with average rating, hide)

**What was NOT done:**
- ❌ No PDF generation for certificates (verification page shows HTML, no download)
- ❌ RequireDripEligibility middleware not wired to routes (exists but not applied)
- ❌ RequireEnrollment middleware not wired to routes (exists but not applied, and returns uuid.Nil because auth not migrated)
- ❌ Drip UI in CoursePlayerPage not implemented (no "Available in X days" display for locked lessons)

---

### Phase 8: Marketplace & SEO — ✅ DONE

**What was done:**
- ✅ MarketplacePage (cross-tenant course listing with Postgres FTS)
- ✅ JSON-LD Course schema.org structured data on CourseDetailPage
- ✅ SEOHandler with `/sitemap.xml` endpoint (lists all published courses)
- ✅ `/robots.txt` endpoint (allows crawlers, disallows /api, /studio, /learn)

---

### Phase 9: Email, Notifications, i18n — ⚠️ PARTIAL

**What was done:**
- ✅ `internal/email/course_email.go` service with 4 HTML templates
- ✅ Templates: enrollment_created, certificate_issued, payout_paid, refund_processed
- ✅ Resend integration (sends via Resend API)

**What was NOT done:**
- ❌ Email service NOT wired to handlers (enrollment creation doesn't send email, certificate issuance doesn't send email, payout doesn't send email)
- ❌ No i18n (react-i18next not installed, no locale files, no locale switching)
- ❌ No notification preferences (User.locale_preference field exists in DB but not used)
- ❌ No RTL support
- ❌ No notification toast system

---

### Phase 10: Admin Console — ❌ NOT DONE

**What was NOT done:**
- ❌ No CreatorsPage (platform admin can't manage creators)
- ❌ No MarketplaceTransactionsPage (can't view all sales)
- ❌ No PayoutsAdminPage (can't view/process creator payouts)
- ❌ No MarketplaceConfigPage (can't configure commission rates)
- ❌ Existing admin console unchanged

---

### Phase 11: Security, Testing, Polish — ⚠️ PARTIAL

**What was done:**
- ✅ RequireEnrollment middleware (exists, but not wired)
- ✅ RequireDripEligibility middleware (exists, but not wired)
- ✅ Webhook idempotency (processed_stripe_events table)
- ✅ 6 E2E tests passing

**What was NOT done:**
- ❌ Postgres Row-Level Security policies NOT created
- ❌ Rate limits NOT applied to storefront endpoints
- ❌ Auth middleware NOT integrated with Postgres (critical gap)
- ❌ No golangci-lint or eslint runs on new code
- ❌ No a11y audit (axe-core)
- ❌ No load testing
- ❌ No pen testing

---

### Phase 12: Launch Readiness — ❌ NOT DONE

- ❌ No deployment to Fly.io
- ❌ No monitoring/alerting setup
- ❌ No runbook
- ❌ No beta testing
- ❌ No production Stripe/Cloudflare configuration

---

## Summary: What Works vs. What Doesn't

### ✅ What works end-to-end (tested):
1. Database layer: all 29 migrations apply, all CRUD operations work against real Postgres
2. Course CRUD: create, update, publish, unpublish, list (via API tests)
3. Section + Lesson creation (via API tests)
4. Enrollment + progress tracking + completion percentage (via API tests)
5. Certificate creation + verification by token (via API tests)
6. Stripe event idempotency (via API tests)
7. Marketplace search with FTS (via API tests)
8. Custom domain creation + tenant resolution (via API tests)
9. Frontend builds and renders all 15 pages
10. Creator studio UI: courses, coupons, sales, domain, reviews, connect, payouts
11. Storefront UI: course browsing, detail pages, checkout flow
12. Learner UI: my courses, course player (video element exists, progress tracking works)

### ❌ What doesn't work (critical gaps):
1. **Authentication**: Course handlers can't identify the logged-in user because auth middleware uses MongoDB, not Postgres. This means enrollment, progress, certificates, payouts all return 401.
2. **Video upload + playback**: No MediaHandler exists. The video `<source>` is empty. No way to upload or play videos.
3. **Email sending**: Templates exist but aren't triggered by any handler.
4. **Enforcement middleware**: RequireEnrollment and RequireDripEligibility exist but aren't applied to any routes.
5. **Plan seeding**: Creator plans (Free/Pro/Business) aren't seeded.
6. **i18n**: No internationalization at all.
7. **Admin extensions**: No new admin pages.
8. **Security hardening**: No RLS, no rate limits, no a11y audit.

### 🔧 What needs to be done to make it production-ready:
1. **Bridge auth**: Either migrate auth to Postgres OR build a bridge that maps MongoDB user IDs to Postgres UUIDs. This is the #1 blocker.
2. **MediaHandler**: Implement Cloudflare Stream direct upload URL generation + signed playback URL serving.
3. **Wire enforcement middleware**: Apply RequireEnrollment to `/api/learner/lessons/:id` and RequireDripEligibility to drip-gated lessons.
4. **Wire email triggers**: Call email service from enrollment handler, certificate handler, payout handler.
5. **Seed plans + config**: Create the 3 creator plans + 7 config defaults.
6. **Deploy + configure**: Set up Fly.io, Cloudflare, Stripe in production mode.
