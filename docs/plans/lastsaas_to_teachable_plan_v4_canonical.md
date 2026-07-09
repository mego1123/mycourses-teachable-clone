# LastSaaS → Teachable-Style Course-Selling SaaS
## Production-Ready Build Plan (v4 — Canonical)

> **This is the single canonical document. It supersedes v1, v2, v2 addendum, and v3.**
>
> Source: 38 Q&A sessions with codewiki.google's AI wiki for `jonradoff/lastsaas`.

---

## Executive Summary

Fork `jonradoff/lastsaas` (Go 1.25 + React 19 + Tailwind 4 SaaS boilerplate) and transform it into a multitenant course-selling platform like Teachable, where each creator gets their own branded storefront at `creatorslug.mycourses.com` (wildcard SSL, free), learners enroll and watch video courses, and the platform takes a commission on each sale via Stripe Connect.

**Stack decisions (final):**
- **Database**: PostgreSQL only (migrate from MongoDB in Phase 0, before any course code)
- **Video/CDN**: Cloudflare Stream + R2 + Cloudflare CDN (free tier covers MVP)
- **Creator storefronts**: Wildcard subdomains `*.mycourses.com` (free, instant) → custom domains via Cloudflare for SaaS later
- **Payments**: Stripe + Stripe Connect (Express accounts, destination charges with `application_fee_amount`)
- **Hosting**: Fly.io

**Timeline**: 15–17 weeks (3–4 week Postgres migration + 12–13 week course platform build).
**Cost**: $0 infra during development, ~$25/month at launch, ~$200–400/month at 1000 creators.

---

## Part 1 — What LastSaaS Gives You for Free

| Capability | Location in lastsaas | Notes |
|---|---|---|
| Auth: password + magic link + OAuth + TOTP MFA + recovery codes + session revoke | `backend/internal/auth/*`, `api/handlers/auth.go`, `frontend/src/contexts/AuthContext.tsx` | bcrypt + common-password blacklist + timing-attack protection + refresh-token rotation |
| Security middleware | `middleware/security.go` (CSP, HSTS, X-Frame-Options, X-Content-Type-Options) | CORS via `rs/cors` with allow-lists |
| Distributed rate limiting | `middleware/ratelimit.go` | Per-IP via `GetClientIP(r)`, per-user via JWT |
| RBAC + membership | `middleware/rbac.go`, `models/membership.go` (`RoleOwner/Admin/User`) | Extend with `RoleCreator/Student` |
| Tenant isolation | `middleware/tenant.go`, every query filters by `tenantId` | Migrate to Postgres Row-Level Security |
| Entitlement gating | `middleware/tenant.go::RequireEntitlement` | Use for `max_courses`, `custom_domain_enabled` |
| Stripe billing | `internal/stripe/stripe.go`, `api/handlers/{billing.go, bundles.go, usage.go, promotions.go, webhook.go}` | Extend for Stripe Connect |
| Webhook event system | `internal/events/emitter.go`, `internal/webhooks/{dispatcher.go, crypto.go}` (HMAC-SHA256), `models/event_definition.go` | Extend with course events |
| White-label branding engine | `models/branding.go`, `api/handlers/branding.go`, `frontend/src/contexts/BrandingContext.tsx`, `components/BrandingThemeInjector.tsx` | Logos, colors, fonts, custom CSS/HTML, favicon, analytics |
| Custom public pages per tenant | `BrandingHandler.GetPublicPage` / `CreatePage` | Creator can publish `/page/:slug` |
| Email via Resend | `internal/email/resend.go` | Extend with course templates |
| Observability | `internal/health/`, `internal/metrics/`, `internal/telemetry/`, `internal/syslog/`, `internal/datadog/` | Admin dashboard at `/admin/health` |
| System config | `internal/config/config.go` (YAML + env), `internal/configstore/` (DB-backed dynamic config) | Add `commission_rate`, `custom_domains.enabled` |
| Admin console | `frontend/src/pages/admin/*` (20+ pages), `api/handlers/admin.go` | Extend for creators/payouts |
| CLI | `backend/cmd/lastsaas/main.go` | Extend for course/payout ops |
| Testing | `internal/testutil/testutil.go`, per-handler `*_test.go`, `frontend/e2e/*.spec.ts` (Playwright) | Follow existing conventions |

---

## Part 2 — Stack Decisions (Final)

### 2.1 Database: PostgreSQL only

**Decision**: Migrate from MongoDB to PostgreSQL in Phase 0, before any course code is written.

**Rationale**: Migrating a small codebase (~30 models) costs 3–4 weeks. Migrating later costs 6–8 weeks. The DB access layer is centralized in `internal/db/mongodb.go`.

**Tooling**:
- `pgx/v5` — PostgreSQL driver
- `sqlc` — generates type-safe Go from SQL queries (compile-time query checking)
- `golang-migrate` — versioned SQL migration files
- `github.com/google/uuid` — replaces `primitive.ObjectID`

**Migration edge cases (resolved via wiki Q34)**:
- Aggregation pipelines → SQL `GROUP BY` + `CASE WHEN` + `JOIN`
- TTL indexes on 13 collections → `pg_cron` jobs
- Webhook dispatcher atomicity → `SELECT ... FOR UPDATE SKIP LOCKED`
- JSON Schema validators → `CHECK` constraints + `JSONB`
- Leader election → `SELECT FOR UPDATE` on `leader_locks` table

**No hybrid DB. One database, one backup, one mental model.**

### 2.2 Video/CDN: Cloudflare (free tier)

**Decision**: Cloudflare Stream (video) + Cloudflare R2 (assets) + Cloudflare CDN (HTML/JS/CSS).

**Rationale**: Cloudflare's free tier covers the first 6–12 months at $0.

**Free tier limits**:
- CDN bandwidth: unlimited free
- R2 storage: 10GB free, no egress fees to CF CDN
- Stream upload: 100 minutes/month free
- Stream delivery: 10,000 minutes/month free
- Workers: 100k requests/day free

**When to switch to Bunny**: When video delivery exceeds 10,000 min/month regularly.

### 2.3 Creator storefronts: Wildcard subdomains → custom domains

**Phase 1 (MVP)**: Wildcard subdomain `*.mycourses.com`
- Cloudflare issues one wildcard SSL cert covering unlimited subdomains
- Creator signs up with slug `john` → instantly gets `john.mycourses.com`
- Cost: $0

**Phase 2 (Growth)**: Add custom domain support via Cloudflare for SaaS
- Creator adds `academy.johndoe.com` → CNAME to `mycourses.com` → Cloudflare auto-issues SSL
- Cost: $0.10/domain/month + $5/month base

**Routing logic** (in `CustomDomainTenantMiddleware`):
```go
host := r.Host
if strings.HasSuffix(host, "."+platformDomain) {
    // Wildcard subdomain: john.mycourses.com → tenant slug "john"
    slug := strings.TrimSuffix(host, "."+platformDomain)
    tenant, _ := db.GetTenantBySlug(ctx, slug)
} else if host != platformDomain {
    // Custom domain: academy.johndoe.com → lookup in custom_domains table
    tenant, _ := db.GetTenantByCustomDomain(ctx, host)
}
```

### 2.4 Payments: Stripe + Stripe Connect

**Decision**: Stripe Connect with Express accounts + destination charges.

**Flow**:
1. Creator clicks "Connect Stripe" → backend creates Express account → redirects to Stripe onboarding
2. Learner buys course → backend creates Checkout Session with `payment_intent_data.application_fee_amount` (commission) + `transfer_data.destination` (creator's connected account)
3. Stripe splits funds automatically: platform gets commission, creator gets remainder
4. Stripe handles payout schedule to creator's bank
5. Webhooks update `Payout` records + emit events

**Refund/dispute handling** (per wiki Q18):
- Full refund: `stripe.Refund.New` → `charge.refunded` + `transfer.reversed` webhooks
- Chargeback: `charge.dispute.created` → `charge.dispute.funds_withdrawn` → `TransactionDisputeWithdrawal`
- Won chargeback: `charge.dispute.funds_reinstated` → `TransactionDisputeReinstatement`

**Creator bears chargeback liability** (their course, their customer).

### 2.5 Hosting: Fly.io

lastsaas is already configured for Fly.io. Keep this. Stateless backend scales horizontally on Fly.

---

## Part 3 — Complete File Checklist

### 3.1 Backend Models (`backend/internal/models/`)

| New file | Contains |
|---|---|
| `course.go` | `Course` struct + status constants (UUID PK, FK to tenants) |
| `section.go` | `Section` struct (CourseID, Title, Order, DripOffsetDays) |
| `lesson.go` | `Lesson` struct (SectionID, Type, Content, MediaAssetID, IsPreview, DurationSec) |
| `media_asset.go` | `MediaAsset` struct (TenantID, Kind, CFStreamID, R2Key, Status) |
| `enrollment.go` | `Enrollment` struct (UNIQUE course_id + user_id) |
| `course_progress.go` | `CourseProgress` struct (UNIQUE enrollment_id + lesson_id) |
| `course_coupon.go` | `CourseCoupon` struct (UNIQUE tenant_id + code) |
| `review.go` | `Review` struct (UNIQUE course_id + user_id, rating 1-5) |
| `payout.go` | `Payout` struct (TenantID, StripeTransferID, AmountCents, Status) |
| `custom_domain.go` | `CustomDomain` struct (UNIQUE domain) |
| `certificate.go` | `Certificate` struct (UNIQUE certificate_number, UNIQUE verification_token) |
| `creator_profile.go` | `CreatorProfile` struct (Bio, WebsiteURL, SocialLinks) |

**Extend existing models**:
- `models/membership.go` → add `RoleCreator`, `RoleStudent`
- `models/tenant.go` → add `Slug`, `StripeConnectAccountID`, `CommissionRateBps`, `PayoutSchedule`
- `models/plan.go` → add entitlements: `max_courses`, `max_video_storage_mb`, `custom_domain_enabled`
- `models/user.go` → add `LocalePreference` (for i18n)
- `models/billing.go` → add transaction types: `TransactionCoursePurchase`, `TransactionConnectTransferReversal`, `TransactionDisputeWithdrawal`, `TransactionDisputeReinstatement`, `TransactionCreatorPayout`, `TransactionPlatformCommission`

### 3.2 Backend Handlers (`backend/internal/api/handlers/`)

| New file | Endpoints |
|---|---|
| `course.go` | CRUD + publish/unpublish + storefront listing + marketplace |
| `section.go` | CRUD under `/api/creator/courses/:courseId/sections` |
| `lesson.go` | CRUD + content access (enrollment-gated) |
| `media.go` | Upload (CF Stream direct upload URL) + serve (signed playback URL) |
| `storefront.go` | Public course listing, detail, checkout session creation |
| `enrollment.go` | Checkout, list mine, list by creator, refund |
| `course_progress.go` | Upsert progress, get by course |
| `coupon.go` | CRUD + validate |
| `review.go` | Create (enrolled only), list public, hide |
| `payout.go` | List, request payout |
| `custom_domain.go` | CRUD + status + Cloudflare provisioning |
| `certificate.go` | List mine, download PDF, verify by token |
| `connect.go` | Stripe Connect onboarding + status |
| `course_webhook.go` | Stripe webhook handler (9 event types, idempotent) |
| `seo.go` | sitemap.xml, robots.txt |

**Extend existing**:
- `api/handlers/webhook.go` → add Connect event handlers

### 3.3 Backend Middleware (`backend/internal/middleware/`)

| New file | Contains |
|---|---|
| `custom_domain.go` | `CustomDomainTenantMiddleware`: resolves tenant from wildcard subdomain OR custom domain |
| `require_enrollment.go` | `RequireEnrollment`: checks enrollments table, allows preview lessons |
| `require_enrollment.go` | `RequireDripEligibility`: checks section.drip_offset_days against enrollment date |

### 3.4 Backend Services

| New file | Contains |
|---|---|
| `internal/stripe/connect.go` | `OnboardCreator`, `CreateCourseCheckout`, `InitiatePayout`, `ReverseTransfer`, `GetAccountStatus`, `VerifyWebhookSignature` |
| `internal/cloudflare/client.go` | `CreateCustomHostname`, `GetCustomHostname`, `DeleteCustomHostname` |
| `internal/email/course_email.go` | `SendEnrollmentEmail`, `SendCertificateEmail`, `SendPayoutEmail`, `SendRefundEmail` |

### 3.5 Backend Migrations (29 total)

**Phase 0 migrations (0001-0016)**: Migrate existing lastsaas collections
- 0001: users + extensions (pgcrypto, pg_trgm)
- 0002: plans (with JSONB entitlements)
- 0003: tenants (with slug, stripe_connect_account_id, commission_rate_bps)
- 0004: tenant_memberships (with creator/student roles)
- 0005: financial_transactions (with course_purchase, commission, payout types)
- 0006: refresh_tokens
- 0007: verification_tokens + oauth_states + revoked_tokens
- 0008: api_keys + webhooks + webhook_deliveries
- 0009: branding_configs + branding_assets + custom_pages
- 0010: invitations + audit_log + system_logs
- 0011: event_definitions + telemetry_events (partitioned) + usage_events
- 0012: system_metrics + daily_metrics + leader_locks
- 0013: messages + announcements + config_vars
- 0014: sso_connections + webauthn_credentials/sessions + auth_codes + impersonation_logs
- 0015: stripe_mappings + credit_bundles + promotions + promotion_codes
- 0016: pg_cron cleanup jobs (10 scheduled jobs)

**Phase 1 migrations (0017-0029)**: Course-specific tables
- 0017: courses (with FTS index + trigram)
- 0018: sections (with drip_offset_days)
- 0019: media_assets (CF Stream + R2)
- 0020: lessons (with is_preview)
- 0021: enrollments (UNIQUE course_id + user_id)
- 0022: course_progress (UNIQUE enrollment_id + lesson_id)
- 0023: course_coupons (UNIQUE tenant_id + code)
- 0024: reviews (UNIQUE course_id + user_id, rating 1-5)
- 0025: payouts
- 0026: custom_domains
- 0027: certificates (UNIQUE certificate_number + verification_token)
- 0028: creator_profiles
- 0029: processed_stripe_events (webhook idempotency)

### 3.6 Frontend Pages (15 total)

**Creator Studio (9 pages)**:
- `CreatorDashboardPage.tsx` — Revenue, enrollments, top courses
- `CoursesPage.tsx` — Course list + create + publish/unpublish
- `CourseEditorPage.tsx` — Tabbed editor (details, curriculum, pricing)
- `CouponsPage.tsx` — CRUD coupons
- `SalesPage.tsx` — Revenue stats + transaction table
- `CustomDomainPage.tsx` — Add domain, DNS instructions, status
- `ReviewsPage.tsx` — Review list with hide
- `ConnectOnboardingPage.tsx` — Stripe Connect onboarding + status
- `PayoutsPage.tsx` — Payout history + request payout

**Storefront (4 pages)**:
- `StorefrontHomePage.tsx` — Course grid
- `CourseDetailPage.tsx` — Course detail with JSON-LD, curriculum, reviews
- `CheckoutPage.tsx` — Coupon entry + Stripe Checkout redirect
- `CertificateVerifyPage.tsx` — Public certificate verification

**Learner (2 pages)**:
- `MyCoursesPage.tsx` — Enrollments + certificates
- `CoursePlayerPage.tsx` — Video player + lesson sidebar + progress tracking

### 3.7 Frontend Components & Hooks

- `CreatorLayout.tsx` — Sidebar nav + top bar
- `useCourseProgress.ts` — 15-second heartbeat to backend, auto-complete at 90%

### 3.8 Frontend API Client

- `api/clientBase.ts` — Axios instance with auth + refresh interceptors
- `api/courseApi.ts` — 10 API namespaces (course, section, lesson, enrollment, progress, coupon, review, payout, customDomain, certificate)
- `types/course.ts` — TypeScript interfaces for all course entities

---

## Part 4 — Postgres Migration (Phase 0, 3–4 weeks)

### Week 1: Foundation
- Add pgx + sqlc + golang-migrate to go.mod
- Replace `internal/db/mongodb.go` with `internal/db/postgres.go`
- Set up `sqlc.yaml` config

### Week 2: Schema + models
- Write 16 migration files for existing tables
- Write 16 query files for sqlc codegen
- Run `sqlc generate` to produce typed Go code
- Rewrite all model structs: `primitive.ObjectID` → `uuid.UUID`

### Week 3: Query layer + handlers
- Rewrite each handler's DB calls (mechanical)
- Replace leader election with `SELECT FOR UPDATE`
- Replace TTL indexes with `pg_cron` jobs
- Replace aggregation pipelines with SQL

### Week 4: Testing + cleanup
- Update `testutil/testutil.go` to use Postgres (testcontainers or embedded-postgres)
- Run full test suite, fix failures
- Remove MongoDB driver from `go.mod`
- Delete `internal/db/mongodb.go` and `internal/db/schema.go`

**Exit criteria**: All existing lastsaas tests pass against Postgres. No MongoDB references.

### Postgres schema highlights

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";     -- typo-tolerant search

-- Users (UUID PK, auth_methods TEXT[], locale_preference for i18n)
-- Plans (entitlements JSONB, GIN index)
-- Tenants (slug UNIQUE, stripe_connect_account_id, commission_rate_bps)
-- Memberships (CHECK role IN owner/admin/creator/student/user)
-- Financial transactions (CHECK type IN 10 types, metadata JSONB)
-- Refresh tokens (TTL via pg_cron)
-- Telemetry events (partitioned by month)
-- Leader locks (for distributed metric collection)

-- Course tables
-- Courses (FTS index + trigram, UNIQUE tenant_id + slug)
-- Sections (drip_offset_days)
-- Lessons (is_preview, FK to media_assets)
-- Enrollments (UNIQUE course_id + user_id)
-- Course progress (UNIQUE enrollment_id + lesson_id)
-- Course coupons (UNIQUE tenant_id + code)
-- Reviews (UNIQUE course_id + user_id, rating 1-5 CHECK)
-- Payouts (stripe_transfer_id, status CHECK)
-- Custom domains (UNIQUE domain, cf_hostname_id)
-- Certificates (UNIQUE certificate_number + verification_token)
-- Creator profiles (social_links JSONB)
-- Processed Stripe events (idempotency, PRIMARY KEY event_id)
```

---

## Part 5 — Security & Abuse Protection

### 5.1 What you inherit (no work needed)

- ✅ Security headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options, X-XSS-Protection)
- ✅ CORS allow-list
- ✅ Body size limit + panic recovery
- ✅ bcrypt password hashing + complexity rules + common-password blacklist
- ✅ TOTP MFA with encrypted secrets + recovery codes
- ✅ Refresh-token rotation with replay detection
- ✅ Session listing + revocation
- ✅ API key hashing

### 5.2 New security work

| Concern | Implementation |
|---|---|
| Storefront abuse | Rate limits: `StorefrontBrowsingLimit` (200/min/IP), `CheckoutInitiationLimit` (10/min/IP) |
| Cross-tenant content access | `RequireEnrollment` middleware + Postgres Row-Level Security |
| Video hotlinking | Cloudflare Stream signed JWTs (10-min TTL) + Referer check |
| Drip bypass | `RequireDripEligibility` middleware |
| Custom domain spoofing | Only resolve domains with `status=active AND dns_verified=true` |
| Stripe Connect key isolation | `StripeConnectAccountID` read-only after onboarding |
| PII in URLs | Certificate verification uses 32-byte hex random tokens |
| Webhook idempotency | `processed_stripe_events` table, ON CONFLICT DO NOTHING |

---

## Part 6 — Video Player, Search, Email

### 6.1 Video player

**Decision**: HTML5 `<video>` element for MVP, upgrade to video.js when needed.

**Implementation**:
- `<video>` element with `controls`, `playsInline`
- `onPlay` → start 15-second heartbeat via `useCourseProgress` hook
- `onPause` → stop heartbeat
- `onEnded` → report completed=true
- Source: Cloudflare Stream signed URL (set from backend)

### 6.2 Search

**Decision**: Postgres full-text search with `pg_trgm` for MVP.

```sql
-- FTS index on courses
CREATE INDEX idx_courses_search ON courses USING GIN (
    to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,''))
);

-- Trigram index for typo tolerance
CREATE INDEX idx_courses_title_trgm ON courses USING GIN (title gin_trgm_ops);
```

### 6.3 Email

**Decision**: Go `html/template` with file-based templates, 4 templates:
- `enrollment_created.html`
- `certificate_issued.html`
- `payout_paid.html`
- `refund_processed.html`

---

## Part 7 — Stripe Connect Marketplace

### 7.1 Onboarding flow
1. Creator clicks "Connect Stripe" in `/studio/connect`
2. Backend: `stripe.Account.New({Type: "express"})` → stores `acct_...`
3. Backend: `stripe.AccountLink.New(...)` → returns URL
4. Frontend redirects to Stripe-hosted onboarding
5. Stripe sends `account.updated` webhook → backend updates status

### 7.2 Course purchase flow
1. Learner clicks "Buy" on `/courses/:slug`
2. Frontend: `POST /api/storefront/checkout` with courseId + couponCode
3. Backend validates coupon, creates `stripe.Checkout.Session` with:
   - `PaymentIntentData.ApplicationFeeAmount` (commission)
   - `PaymentIntentData.TransferData.Destination` (creator's connected account)
   - `Metadata: {course_id, learner_user_id, tenant_id, type: "course_purchase"}`
4. Frontend redirects to `session.URL`
5. Stripe sends `checkout.session.completed` webhook
6. Backend (idempotent): creates enrollment + `course_purchase` + `platform_commission` transactions

### 7.3 Payout flow
- Stripe-managed: auto-payout daily/weekly/monthly to creator's bank
- Platform-managed: `POST /api/creator/payouts/request` → `stripe.Transfer.New`

### 7.4 Webhook idempotency
```sql
CREATE TABLE processed_stripe_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ DEFAULT NOW()
);
```
Every webhook handler checks this table before processing.

---

## Part 8 — Custom Domain Flow

### 8.1 MVP: Wildcard subdomains (free, instant)
- DNS: `*.mycourses.com` → A record (Cloudflare proxied)
- SSL: Cloudflare auto-issues wildcard cert
- Creator signs up with slug "john" → instantly gets `https://john.mycourses.com`

### 8.2 Phase 2: Custom domains via Cloudflare for SaaS
1. Creator enters `academy.johndoe.com`
2. Backend: `POST /zones/{zone}/custom_hostnames` → Cloudflare auto-issues SSL
3. Creator points CNAME: `academy.johndoe.com` → `mycourses.com`
4. Backend polls Cloudflare API; updates `custom_domains.status = 'active'`

### 8.3 SSL/TLS
- Wildcard cert doesn't cover apex (`mycourses.com`) — Cloudflare handles separately
- Wildcard doesn't cover multi-level subdomains (`app.john.mycourses.com`)
- Cloudflare for SaaS handles per-domain SSL automatically

---

## Part 9 — Observability

Reuse existing stack:
- Health checks: add Cloudflare API + Stripe API checks
- Metrics: add `course_published_total`, `enrollment_created_total`, `payout_paid_total`
- Telemetry: emit `course.viewed`, `enrollment.created`, `lesson.completed`, `certificate.downloaded`
- Materialized view for creator revenue (refreshed hourly via pg_cron)

---

## Part 10 — Webhooks & Events

New event types to register:
- `course.created`, `course.published`
- `enrollment.created`, `enrollment.completed`
- `lesson.completed`
- `payout.paid`, `payout.failed`
- `certificate.issued`
- `custom_domain.activated`
- `refund.processed`
- `review.posted`

---

## Part 11 — GDPR

| Right | Implementation |
|---|---|
| Access | Extend `ExportData` to include enrollments, progress, reviews, certificates |
| Be forgotten | Anonymize enrollments (keep for revenue), delete reviews, revoke certificates |
| Retention | Auto-delete course_progress for refunded enrollments after 90 days. Keep financial_transactions for 7 years |

---

## Part 12 — Cost & Scaling

| Phase | Cost |
|---|---|
| Development | $0 |
| Launch (~$25/month) | Fly.io + managed Postgres + Cloudflare Free |
| 1000 creators + 10k learners (~$200-400/month) | Fly.io scale + Postgres M30 + CF for SaaS + R2 |
| 10k creators + 100k learners (~$2000-4000/month) | Fly.io cluster + Postgres cluster + CF for SaaS + R2 |

**Critical indexes**:
- `courses(tenant_id, status, created_at DESC)` — storefront listing
- `enrollments(user_id, status)` — learner dashboard
- `enrollments(tenant_id, enrolled_at DESC)` — creator sales
- `course_progress(enrollment_id, lesson_id) UNIQUE` — progress tracking
- `custom_domains(domain) UNIQUE` — domain resolution
- `financial_transactions(tenant_id, type, created_at DESC)` — revenue

---

## Part 13 — CI/CD

Extend `.github/workflows/ci.yml`:
- Backend: Postgres service container, `make migrate-up`, `make test`
- Frontend: `npm test`, `npm run lint`, `npm run build`
- E2E: `docker compose up`, `npx playwright test`
- Stripe webhook test: `go test ./internal/stripe/... -run TestConnectWebhookSignature`

---

## Part 14 — Implementation Roadmap (12 Phases)

### Phase 0 — Postgres Migration (3–4 weeks)
- Week 1: Foundation (pgx + sqlc + golang-migrate)
- Week 2: Schema + models (16 migration files)
- Week 3: Query layer + handlers (rewrite all DB calls)
- Week 4: Testing + cleanup (remove MongoDB)

### Phase 1 — Course Domain Models (1 week)
- 13 migration files for course tables
- 13 query files for sqlc
- Run `sqlc generate`
- Add `RoleCreator`, `RoleStudent` to memberships
- Seed Creator Free/Pro/Business plans

### Phase 2 — Creator Studio Backend (1.5 weeks)
- Course, Section, Lesson, Media handlers
- Coupon, Review handlers
- Wire routes under `/api/creator/*`

### Phase 3 — Creator Studio Frontend (1.5 weeks)
- CourseEditorPage (tabbed: details, curriculum, pricing)
- LessonEditorPage with video upload
- CoursesPage, CouponsPage, ReviewsPage
- CreatorLayout

### Phase 4 — Storefront & Learner Experience (2 weeks)
- StorefrontHandler (public, no auth)
- EnrollmentHandler (checkout, list)
- CourseProgressHandler
- RequireEnrollment middleware
- StorefrontHomePage, CourseDetailPage, CheckoutPage
- CoursePlayerPage with video player + progress tracking

### Phase 5 — Stripe Connect (1.5 weeks)
- `internal/stripe/connect.go`
- Connect handler (onboarding, status)
- Extend webhook handler (9 new event types)
- Idempotency via `processed_stripe_events` table
- PayoutHandler
- ConnectOnboardingPage, PayoutsPage

### Phase 6 — Wildcard Subdomains + Custom Domains (1 week)
- Cloudflare wildcard DNS + SSL
- CustomDomainTenantMiddleware (wildcard + custom domain)
- Cloudflare for SaaS API integration
- CustomDomainPage with DNS instructions + status polling

### Phase 7 — Certificates, Drip, Reviews (1 week)
- CertificateHandler + PDF generation
- RequireDripEligibility middleware
- CertificateVerifyPage, CertificatesPage
- Drip UI in CoursePlayerPage
- Auto-issue certificates on course completion

### Phase 8 — Marketplace & SEO (1 week)
- MarketplacePage with Postgres FTS + pg_trgm
- JSON-LD Course schema on CourseDetailPage
- `/sitemap.xml` endpoint
- `/robots.txt`

### Phase 9 — Email, Notifications, i18n (1.5 weeks)
- react-i18next + locale files (en, es, fr, de, pt, ar)
- EmailService with 4 templates
- NotificationContext for toasts

### Phase 10 — Admin Console (1 week)
- CreatorsPage, MarketplaceTransactionsPage, PayoutsAdminPage, MarketplaceConfigPage
- Commission rate configuration

### Phase 11 — Security, Testing, Polish (1.5 weeks)
- Rate limits on storefront endpoints
- Postgres RLS on tenant-scoped tables
- golangci-lint, eslint, axe-core
- Load test with k6
- Pen-test checkout with Stripe test cards
- Stripe Radar rules

### Phase 12 — Launch Readiness (1 week)
- Full e2e suite in CI
- Monitoring + alerting
- Runbook
- Beta test with 5-10 creators
- Public launch

**Total: 15–17 weeks**

---

## Part 15 — Final Production Readiness Checklist

### Backend
- [ ] All 29 migrations implemented
- [ ] All 15 new handlers with tests
- [ ] All 3 new middleware with tests
- [ ] Stripe Connect onboarding + checkout + payout + refund flows
- [ ] All webhook handlers tested with Stripe CLI + idempotency
- [ ] Wildcard subdomain routing + custom domain resolution
- [ ] Cloudflare Stream upload + signed playback
- [ ] Rate limits on storefront endpoints
- [ ] RequireEnrollment on all paid content
- [ ] RequireDripEligibility on drip-gated lessons
- [ ] Postgres RLS on tenant-scoped tables
- [ ] Event emitter for all 11 new event types
- [ ] Email templates (4) + notification preferences
- [ ] OpenAPI spec accurate

### Frontend
- [ ] All 15 new pages
- [ ] All contexts/hooks
- [ ] API client with all namespaces
- [ ] Video player with progress tracking
- [ ] Certificate PDF generation + verification
- [ ] i18n with 6 locales
- [ ] RTL for Arabic
- [ ] a11y audit (axe-core)
- [ ] Playwright e2e

### Infrastructure
- [ ] Cloudflare zone with wildcard DNS + SSL
- [ ] Cloudflare R2 bucket + Stream library
- [ ] Stripe Connect platform account
- [ ] Postgres with pg_cron, pg_trgm, pgcrypto
- [ ] pg_cron jobs scheduled
- [ ] Monitoring + alerting
- [ ] Backup strategy

### Business / Legal
- [ ] Terms of Service (marketplace)
- [ ] Privacy Policy (GDPR)
- [ ] Creator agreement (commission + payout + chargeback liability)
- [ ] Cookie consent
- [ ] Data retention policy
- [ ] Stripe Connect KYC enforced

---

*This is the canonical plan. All prior versions (v1, v2, v2 addendum, v3) are superseded.*
