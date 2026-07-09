# LastSaaS → Teachable-Style Course-Selling SaaS
## Production-Ready Build Plan (v2)

> Source: 26 Q&A sessions with codewiki.google's AI wiki for `jonradoff/lastsaas`.
> Raw answers preserved at `/home/z/my-project/download/wiki_answers/01_*.json` through `26_*.json`.
> Companion document: `lastsaas_to_teachable_plan.md` (v1 — concept mapping & high-level architecture).

This v2 plan adds everything v1 omitted: security, testing, deployment, observability, migrations, webhooks/events, refunds/disputes, GDPR, i18n, video streaming, certificates, drip content, marketplace discovery, scaling, CI/CD, and a complete file checklist.

---

## Part 1 — Production Foundations LastSaaS Already Gives You

| Capability | Concrete location in lastsaas | Production-grade? |
|---|---|---|
| **Auth: password + magic link + OAuth (Google/GitHub/Microsoft) + TOTP MFA + recovery codes + session revoke** | `backend/internal/auth/{password.go, totp.go, jwt.go, *_oauth.go}`, `api/handlers/auth.go` | ✅ Yes — bcrypt cost-configurable, common-password blacklist, min length 10, timing-attack protection via `DummyCompare`, refresh-token rotation with replay detection |
| **Security middleware** | `middleware/security.go` (CSP, HSTS, X-Frame-Options DENY, X-Content-Type-Options nosniff, X-XSS-Protection), `rs/cors` with explicit allow-lists | ✅ Yes |
| **Distributed rate limiting** | `middleware/ratelimit.go` (`NewDistributedRateLimiter(database)`), per-IP via `GetClientIP(r)`, per-user via JWT claim, stored in MongoDB so it works across instances | ✅ Yes — extend with new limit types for storefront |
| **Body size limit + panic recovery** | `middleware/bodylimit.go`, `middleware/recovery.go` | ✅ Yes |
| **RBAC** | `middleware/rbac.go`, `middleware/auth.go`, `models/membership.go` (`RoleOwner/Admin/User`) | ✅ Yes — extend with `RoleCreator/Student` |
| **Tenant isolation** | `middleware/tenant.go`, every DB query filters by `tenantId` | ✅ Yes |
| **Entitlement gating** | `middleware/tenant.go::RequireEntitlement` checks `plan.Entitlements` map | ✅ Yes — use for `max_courses`, `custom_domain_enabled` etc. |
| **Billing: Stripe Checkout, Customer Portal, webhooks, promotions/coupons, credit bundles, usage metering** | `internal/stripe/stripe.go`, `api/handlers/{billing.go, bundles.go, usage.go, promotions.go, webhook.go}` | ✅ Yes — extend for Stripe Connect |
| **Webhook event system** | `internal/events/emitter.go`, `internal/webhooks/{dispatcher.go, crypto.go}` (HMAC-SHA256 signing), `models/event_definition.go` (19 built-in event types) | ✅ Yes — extend with course events |
| **White-label branding engine** | `models/branding.go`, `api/handlers/branding.go` (CRUD + asset upload via GridFS + custom pages), `frontend/src/contexts/BrandingContext.tsx`, `components/BrandingThemeInjector.tsx` | ✅ Yes — already does logos, colors, fonts, custom CSS, custom HTML head injection, favicon, analytics snippets |
| **Email via Resend** | `internal/email/resend.go` | ✅ Yes — extend with course templates |
| **Observability** | `internal/health/{health.go, query.go, integrations.go}`, `internal/metrics/metrics.go`, `middleware/metrics.go`, `internal/telemetry/service.go`, `internal/syslog/syslog.go`, `internal/datadog/client.go` | ✅ Yes — admin dashboard at `/admin/health` |
| **System config (static + dynamic)** | `internal/config/config.go` (YAML + env), `internal/configstore/{store.go, validate.go, seed.go}` (DB-backed dynamic config) | ✅ Yes — add `commission_rate`, `custom_domains.enabled` etc. via configstore |
| **DB schema management** | `internal/db/schema.go` (JSON schema validators + IndexModel creation at startup), `internal/db/mongodb.go` | ✅ Yes — zero-downtime: app creates indexes on boot |
| **Bootstrap** | `api/handlers/bootstrap.go` creates root tenant + admin user on first run | ✅ Yes |
| **Admin console** | `frontend/src/pages/admin/*` (20+ pages), `api/handlers/admin.go` | ✅ Yes — extend for creators/payouts |
| **CLI** | `backend/cmd/lastsaas/main.go` (user management, tenant management, config) | ✅ Yes — extend for course/payout ops |
| **Testing** | `internal/testutil/testutil.go` (`MustConnectTestDB`, `CleanupCollections`), per-handler `*_test.go` files (billing_test.go, auth_test.go, tenant_test.go, etc.), `frontend/e2e/*.spec.ts` (Playwright), `frontend/vitest.config.ts` | ✅ Yes — follow existing conventions |
| **CI/CD** | `.github/workflows/ci.yml` (lint, test, build, codecov), `backend/Makefile` (test, test-integration, test-e2e targets), `eslint.config.js`, `VERSION` + `VERSIONS.md` | ✅ Yes |
| **i18n** | Not present | ❌ Must add (see §10) |
| **a11y** | Basic Tailwind patterns; no formal WAI-ARIA audit | ⚠️ Must verify (see §10) |

---

## Part 2 — Complete File Checklist

Every new file to create, organized by layer. Each entry shows what it contains and what existing file to model after.

### 2.1 Backend Models (`backend/internal/models/`)

| New file | Contains | Model after |
|---|---|---|
| `course.go` | `Course` struct + status constants (`CourseStatusDraft/Published/Archived`) | `models/plan.go` (similar shape: ID, TenantID, title, price, status) |
| `section.go` | `Section` struct (CourseID, Title, Order, DripOffsetDays) | `models/announcement.go` (parent-child + ordering) |
| `lesson.go` | `Lesson` struct (SectionID, Type, Content, MediaAssetID, IsPreview, DurationSec) | `models/announcement.go` |
| `media_asset.go` | `MediaAsset` struct (TenantID, Kind, StorageURL, SizeBytes, MimeType) + index on `tenantId` | `models/branding.go::BrandingAsset` (already models binary assets) |
| `enrollment.go` | `Enrollment` struct (CourseID, TenantID, UserID, Status, PricePaidCents, CouponID, EnrollmentDate) + compound index `(userId, courseId)` unique | `models/membership.go` (User+Tenant link with status) |
| `course_progress.go` | `CourseProgress` struct (EnrollmentID, LessonID, UserID, Completed, VideoPositionSec) + compound index `(enrollmentId, lessonId)` unique | `models/usage_event.go` (event-style records per user) |
| `course_coupon.go` | `CourseCoupon` struct (TenantID, Code, DiscountType, DiscountValue, CourseID, ExpiresAt, UsageLimit, UsedCount) + unique index `(tenantId, code)` | `models/credit_bundle.go` |
| `review.go` | `Review` struct (CourseID, UserID, Rating, Comment) + compound index `(courseId, userId)` unique | `models/announcement.go` |
| `payout.go` | `Payout` struct (TenantID, StripeTransferID, AmountCents, Currency, Status, InitiatedAt, CompletedAt) | `models/billing.go::FinancialTransaction` |
| `custom_domain.go` | `CustomDomain` struct (Domain, TenantID, Status, DNSVerified, SSLStatus) + unique index on `domain` | `models/api_key.go` (per-tenant resource with status) |
| `certificate.go` | `Certificate` struct (EnrollmentID, UserID, CourseID, CertificateNumber, IssuedAt, VerificationToken) + unique index on `certificateNumber` | `models/invitation.go` (token-based resource) |
| `creator_profile.go` | `CreatorProfile` struct (UserID/TenantID, Bio, WebsiteURL, SocialLinks) | `models/user.go` |

### 2.2 Backend Stripe Connect (`backend/internal/stripe/`)

| New file | Contains | Model after |
|---|---|---|
| `connect.go` | `OnboardCreator(tenantID, returnURL, refreshURL) → account_link.URL`, `CreateCourseCheckout(course, creatorAcctID, commissionBps, learnerCustomerID, couponID, successURL, cancelURL) → checkout.Session`, `InitiatePayout(creatorAcctID, amountCents, currency) → transfer.ID`, `ReverseTransfer(transferID, amountCents) → reversal.ID` | `internal/stripe/stripe.go::CreateCheckoutSession` |
| `connect_test.go` | Mock Stripe API via `httptest.NewServer`, test onboarding/checkout/payout flows | `stripe_test.go::setupMockStripe` |

### 2.3 Backend Handlers (`backend/internal/api/handlers/`)

| New file | Endpoints | Model after |
|---|---|---|
| `course.go` | `POST/GET/PUT/DELETE /api/creator/courses`, `GET /api/storefront/courses`, `GET /api/storefront/courses/:slug` | `handlers/branding.go` (rich CRUD + public/admin split) |
| `section.go` | `POST/GET/PUT/DELETE /api/creator/courses/:courseId/sections` | `handlers/announcements.go` |
| `lesson.go` | `POST/GET/PUT/DELETE /api/creator/sections/:sectionId/lessons`, `GET /api/learner/lessons/:id` (enrollment-gated) | `handlers/announcements.go` |
| `media.go` | `POST /api/creator/media/upload` (presigned PUT URL), `GET /api/creator/media/:id`, `GET /api/learner/media/:id` (signed playback URL, enrollment-gated) | `handlers/branding.go::UploadAsset/ServeAsset` |
| `storefront.go` | `GET /api/storefront/branding` (already exists), `GET /api/storefront/courses`, `GET /api/storefront/courses/:slug`, `GET /api/storefront/courses/:id/reviews`, `POST /api/storefront/coupons/validate`, `POST /api/storefront/checkout/:courseSlug` | `handlers/branding.go::GetPublicPage` |
| `enrollment.go` | `POST /api/learner/enrollments/:courseId` (initiates checkout), `GET /api/learner/enrollments` (list mine), `GET /api/creator/enrollments` (list sales), `POST /api/creator/enrollments/:id/refund` | `handlers/billing.go` |
| `course_progress.go` | `POST /api/learner/progress/:lessonId` (upsert), `GET /api/learner/progress/course/:courseId` | `handlers/usage.go::RecordUsage` (atomic upsert pattern) |
| `coupon.go` | `POST/GET/PUT/DELETE /api/creator/coupons` | `handlers/promotions.go` (already creates Stripe coupons; this is the per-course analogue) |
| `review.go` | `POST /api/learner/reviews`, `GET /api/storefront/courses/:id/reviews`, `PUT /api/learner/reviews/:id` | `handlers/announcements.go` |
| `payout.go` | `GET /api/creator/payouts`, `POST /api/creator/payouts/request`, `GET /api/admin/payouts` | `handlers/billing.go::AdminListTransactions` |
| `custom_domain.go` | `POST /api/creator/custom-domains`, `GET /api/creator/custom-domains`, `GET /api/creator/custom-domains/:id/status`, `DELETE /api/creator/custom-domains/:id`, `GET /internal/caddy/allowed?domain=...` (called by Caddy on-demand TLS) | `handlers/apikeys.go` (per-tenant CRUD) |
| `certificate.go` | `GET /api/learner/certificates`, `GET /api/learner/certificates/:id/download` (PDF), `GET /api/public/certificates/verify/:token` (public verification page) | `handlers/branding.go::ServeAsset` |
| `connect.go` | `GET /api/creator/connect/onboarding` (returns Stripe AccountLink URL), `GET /api/creator/connect/status`, `POST /api/creator/connect/refresh` | `handlers/billing.go::CreateCheckoutSession` |
| `course_admin.go` | `GET /api/admin/courses` (cross-tenant), `GET /api/admin/creators` (list all creator tenants), `PUT /api/admin/creators/:id/commission` | `handlers/admin.go::ListUsers` |

Extend existing files:

| Existing file | Change |
|---|---|
| `models/membership.go` | Add `RoleCreator`, `RoleStudent` constants |
| `models/billing.go` | Add new `TransactionType` constants: `TransactionCoursePurchase`, `TransactionConnectTransferReversal`, `TransactionDisputeWithdrawal`, `TransactionDisputeReinstatement`, `TransactionCreatorPayout`, `TransactionPlatformCommission` |
| `models/tenant.go` | Add fields: `StripeConnectAccountID string`, `StripeConnectStatus string`, `CommissionRateBps int`, `PayoutSchedule string`, `Slug string` (unique) |
| `models/plan.go` | Add entitlements: `max_courses`, `max_video_storage_mb`, `max_video_bandwidth_gb_month`, `custom_domain_enabled`, `commission_rate_bps_override` |
| `models/usage_event.go` | Add `UsageType` constants: `video_storage_mb_hours`, `video_bandwidth_gb`, `course_published_count` |
| `api/handlers/webhook.go` | Add handlers for: `transfer.reversed`, `charge.dispute.funds_withdrawn`, `charge.dispute.funds_reinstated`, `account.updated` (Connect onboarding status), `payout.paid`, `payout.failed`. Extend `handleCheckoutCompleted` to detect course purchases (via metadata) and create `Enrollment` + `FinancialTransaction(TransactionCoursePurchase)` + emit `enrollment.created` event |
| `internal/events/emitter.go` | Emit new events: `course.published`, `course.unpublished`, `enrollment.created`, `enrollment.completed`, `lesson.completed`, `payout.paid`, `payout.failed`, `certificate.issued`, `custom_domain.activated` |
| `cmd/server/main.go` | Register all new routes + new middleware (`CustomDomainTenantMiddleware`, `RequireEnrollment`, `RequireDripEligibility`). Insert `CustomDomainTenantMiddleware` BEFORE `NewTenantMiddleware` in the chain |
| `internal/db/schema.go` | Add JSON schema validators + indexes for all 12 new collections |

### 2.4 Backend Middleware (`backend/internal/middleware/`)

| New file | Contains | Model after |
|---|---|---|
| `custom_domain.go` | `CustomDomainTenantMiddleware`: reads `Host` header, looks up `custom_domains` collection, injects `TenantID` into context. Runs BEFORE `NewTenantMiddleware`. Falls through to header-based resolution for platform domain | `middleware/tenant.go::TenantMiddleware` |
| `require_enrollment.go` | `RequireEnrollment(courseIDParam string)`: fetches `courseId` from URL, checks `enrollments` collection for active record matching `(userId, courseId)`. Returns 403 if missing | `middleware/tenant.go::RequireActiveBilling` and `RequireEntitlement` |
| `drip.go` | `RequireDripEligibility(lessonIDParam string)`: fetches lesson + section, computes `enrollmentDate + section.DripOffsetDays`, returns 403 if `time.Now()` is before that | `middleware/tenant.go::RequireEntitlement` (same pattern: context check + return 403) |
| `ratelimit.go` | Add constants: `StorefrontBrowsingLimit`, `StorefrontSearchLimit`, `CheckoutInitiationLimit`, `CertificateVerifyLimit`. Apply to corresponding routes in `main.go` | existing limit constants (`LoginAttemptLimit`, `TokenRefreshLimit`) |

### 2.5 Backend Config & Migrations

| New file / change | Contains |
|---|---|
| `backend/config/dev.example.yaml` (extend) | Add: `s3.bucket`, `s3.region`, `s3.access_key_id`, `s3.secret_access_key`, `cloudfront.distribution_id` (optional), `stripe.connect.enabled`, `stripe.connect.client_id`, `caddy.on_demand_tls_ask_url` |
| `backend/config/prod.example.yaml` (extend) | Same as above |
| `internal/configstore/seed.go` (extend) | Seed default dynamic config: `course.default_commission_rate_bps = 1000` (10%), `course.payout_schedule = "weekly"`, `custom_domains.enabled = true`, `marketplace.discovery_enabled = true` |
| `internal/db/schema.go` (extend) | For each new collection: (1) call `db.CreateCollection` with `Validator` JSON schema, (2) call `db.Collection().Indexes().CreateMany` with `IndexModel` slice. All idempotent — runs on every boot |
| `internal/planstore/seed.go` (extend) | Add creator plans: "Creator Free" (1 course, 100MB storage, no custom domain, 20% commission), "Creator Pro" (10 courses, 50GB storage, custom domain, 10% commission), "Creator Business" (unlimited courses, 500GB storage, custom domain, 5% commission) |

### 2.6 Frontend Pages (`frontend/src/pages/`)

| New file | Purpose | Model after |
|---|---|---|
| `public/storefront/StorefrontHomePage.tsx` | Creator's branded homepage at custom domain | `pages/public/LandingPage.tsx` |
| `public/storefront/CourseListPage.tsx` | All published courses by creator | `pages/app/DashboardPage.tsx` (card grid) |
| `public/storefront/CourseDetailPage.tsx` | Course landing page with curriculum, price, reviews, instructor bio, JSON-LD for SEO | `pages/public/CustomPage.tsx` |
| `public/storefront/CheckoutPage.tsx` | Coupon entry + redirect to Stripe Checkout | `pages/app/BuyCreditsPage.tsx` |
| `public/storefront/CartPage.tsx` | Optional: cart for multi-course checkout | — |
| `public/MarketplacePage.tsx` | Platform-wide marketplace (lists courses across all creators) | `pages/public/LandingPage.tsx` |
| `public/CertificateVerifyPage.tsx` | Public certificate verification by token | `pages/public/CustomPage.tsx` |
| `app/creator/CreatorDashboardPage.tsx` | Revenue, enrollments, top courses | `pages/app/DashboardPage.tsx` |
| `app/creator/CoursesPage.tsx` | Course list + create new | `pages/app/TeamPage.tsx` (list+invite pattern) |
| `app/creator/CourseEditorPage.tsx` | Tabbed editor: details, sections, lessons, media, pricing, coupons, drip, certificates | `pages/admin/BrandingPage.tsx` (tabbed editor pattern) |
| `app/creator/LessonEditorPage.tsx` | Upload video/PDF, write content, set preview/drip | `pages/admin/ConfigPage.tsx` |
| `app/creator/SalesPage.tsx` | Sales transactions, refunds | `pages/app/settings/BillingTab.tsx` |
| `app/creator/PayoutsPage.tsx` | Payout history, request payout, Stripe Connect onboarding | `pages/app/settings/BillingTab.tsx` |
| `app/creator/CouponsPage.tsx` | CRUD course coupons | `pages/admin/PromotionsPage.tsx` (direct analogue) |
| `app/creator/CustomDomainPage.tsx` | Add domain, show DNS instructions, poll status | `pages/app/settings/SecurityTab.tsx` |
| `app/creator/ConnectOnboardingPage.tsx` | Stripe Connect onboarding redirect | `pages/auth/AuthCallbackPage.tsx` |
| `app/creator/ReviewsPage.tsx` | Moderate reviews | `pages/admin/MessagesPage.tsx` |
| `app/learner/MyCoursesPage.tsx` | All enrollments across all creators | `pages/app/DashboardPage.tsx` |
| `app/learner/CoursePlayerPage.tsx` | Video player + lesson sidebar + progress tracking | — (new) |
| `app/learner/CourseOverviewPage.tsx` | Course dashboard: progress, certificate, Q&A | `pages/app/DashboardPage.tsx` |
| `app/learner/BillingHistoryPage.tsx` | Purchase history across creators | `pages/app/settings/BillingTab.tsx` |
| `app/learner/CertificatesPage.tsx` | Downloadable certificates | `pages/app/PlanPage.tsx` |
| `app/learner/ProfilePage.tsx` | Display name, avatar, notification preferences | `pages/app/settings/ProfileTab.tsx` |
| `admin/CreatorsPage.tsx` | Platform admin: list/search/suspend creators | `pages/admin/TenantsPage.tsx` (direct analogue) |
| `admin/MarketplaceTransactionsPage.tsx` | All sales across platform, filter by creator/date | `pages/admin/FinancialPage.tsx` |
| `admin/PayoutsAdminPage.tsx` | View/process creator payouts | `pages/admin/FinancialPage.tsx` |
| `admin/MarketplaceConfigPage.tsx` | Commission rates, featured courses, categories | `pages/admin/ConfigPage.tsx` |

### 2.7 Frontend Contexts, Hooks, API Client

| New file / change | Purpose |
|---|---|
| `frontend/src/contexts/EnrollmentContext.tsx` | Provides learner's enrollments, prefetch on learner routes |
| `frontend/src/contexts/CreatorContext.tsx` | Wraps `TenantContext` for creator routes; exposes `isCreator()`, `creatorTenant`, `stripeConnectStatus` |
| `frontend/src/contexts/NotificationContext.tsx` | Toasts for course/enrollment/payout events (subscribes to existing telemetry stream if present) |
| `frontend/src/hooks/useCourseProgress.ts` | 15-second heartbeat to `POST /api/learner/progress/:lessonId` with current video position |
| `frontend/src/hooks/useStorefrontBranding.ts` | Wraps `BrandingContext` for storefront routes — verifies tenant resolved correctly |
| `frontend/src/hooks/useDripEligible.ts` | Pre-fetch lesson drip status for player UI |
| `frontend/src/api/client.ts` (extend) | Add namespaces: `client.courses`, `client.sections`, `client.lessons`, `client.enrollments`, `client.progress`, `client.coupons`, `client.reviews`, `client.payouts`, `client.customDomains`, `client.certificates`, `client.connect`, `client.storefront`, `client.media` |
| `frontend/src/types/index.ts` (extend) | Add TypeScript interfaces for all new entities |
| `frontend/src/components/VideoPlayer.tsx` | Wraps `react-player` or `video.js`; emits `onProgress` to `useCourseProgress` |
| `frontend/src/components/CourseCard.tsx` | Reusable card for marketplace + storefront + learner dashboard |
| `frontend/src/components/Certificate.tsx` | Renders certificate as HTML, with "Download PDF" via `html2canvas` + `jsPDF` |
| `frontend/src/components/CouponInput.tsx` | Coupon entry with validation |
| `frontend/src/App.tsx` (extend) | Add all new route groups (see §4 of v1 plan) |
| `frontend/src/components/BrandingThemeInjector.tsx` (no change) | Already does everything needed |

### 2.8 Frontend Tests

| New file | Purpose |
|---|---|
| `frontend/e2e/storefront.spec.ts` | Browse courses, view detail, checkout redirect |
| `frontend/e2e/learner.spec.ts` | Enroll, watch lesson, complete course, get certificate |
| `frontend/e2e/creator.spec.ts` | Create course, upload video, publish, view sale, request payout |
| `frontend/e2e/custom-domain.spec.ts` | Add custom domain, verify DNS, see it serve branded storefront (mocked via Host header) |
| `frontend/src/components/__tests__/VideoPlayer.test.tsx` | Unit test progress emission |

### 2.9 Backend Tests

| New file | Purpose |
|---|---|
| `backend/internal/api/handlers/course_test.go` | CRUD + tenant isolation + storefront visibility |
| `backend/internal/api/handlers/enrollment_test.go` | Checkout flow, webhook → enrollment, refund flow |
| `backend/internal/api/handlers/course_progress_test.go` | Upsert + completion flag |
| `backend/internal/api/handlers/payout_test.go` | Payout creation, Stripe Connect integration with mock |
| `backend/internal/api/handlers/custom_domain_test.go` | Add domain, DNS verify, Caddy `ask` endpoint |
| `backend/internal/api/handlers/coupon_test.go` | CRUD + validation + usage limit enforcement |
| `backend/internal/api/handlers/review_test.go` | Only enrolled learners can review; one review per user per course |
| `backend/internal/api/handlers/certificate_test.go` | Auto-issue on completion, verification token uniqueness |
| `backend/internal/api/handlers/storefront_test.go` | Public endpoints work without auth, respect custom domain middleware |
| `backend/internal/api/handlers/connect_test.go` | Onboarding URL generation, status webhook handling |
| `backend/internal/middleware/custom_domain_test.go` | Host header resolution, fallback to platform domain |
| `backend/internal/middleware/require_enrollment_test.go` | 403 when not enrolled, 200 when enrolled |
| `backend/internal/middleware/drip_test.go` | 403 before drip date, 200 after |
| `backend/internal/stripe/connect_test.go` | Mock Stripe API for onboarding/checkout/payout/refund/transfer reversal |

### 2.10 Infrastructure

| New file / change | Purpose |
|---|---|
| `Caddyfile` (new or extend existing) | Configure `on_demand_tls { ask http://localhost:8080/internal/caddy/allowed }` + reverse proxy to backend. Caddy auto-issues Let's Encrypt certs for any approved custom domain |
| `Dockerfile` (extend) | Ensure Caddy is in the final image with the new Caddyfile |
| `fly.toml` (extend) | Add env vars for S3, Stripe Connect, custom domain config |
| `.github/workflows/ci.yml` (extend) | Run new tests; add Stripe Connect webhook signature verification test job |
| `scripts/seed-creator-plans.sql` (or Go) | One-time script to seed Creator Free/Pro/Business plans via `planstore.Seed` |

---

## Part 3 — Security & Abuse Protection (Production Hardening)

### 3.1 What you inherit (no work needed)

- ✅ CORS allow-list from `cfg.Frontend.URL`
- ✅ Security headers (CSP, HSTS, X-Frame-Options DENY, X-Content-Type-Options nosniff, X-XSS-Protection) via `middleware.SecurityHeaders`
- ✅ Body size limit + panic recovery
- ✅ bcrypt password hashing with cost factor, common-password blacklist, min length 10, complexity rules (uppercase, lowercase, number, special char)
- ✅ Timing-attack protection on login via `DummyCompare`
- ✅ TOTP MFA with encrypted-at-rest secrets + recovery codes
- ✅ Refresh-token rotation with replay-detection (entire token family invalidated on replay)
- ✅ Session listing + revocation; password change revokes all sessions
- ✅ Distributed rate limiter backed by MongoDB (works across instances)
- ✅ Per-IP rate limit on `auth/register`, `auth/login`, `auth/forgot-password`, `auth/verify-email`
- ✅ Per-user rate limit on telemetry endpoints
- ✅ NoSQL injection protection (MongoDB driver parameterization + struct-tag validation)
- ✅ API key hashing (only `KeyHash` stored; `KeyPreview` shown in UI for identification)

### 3.2 New security work for the course platform

| Concern | Implementation |
|---|---|
| **Storefront abuse (browsing bots, scraping)** | Add `StorefrontBrowsingLimit` (e.g. 200 req/min per IP via `middleware.GetClientIP`). Apply to `/api/storefront/*` routes in `main.go` |
| **Checkout abuse (card testing)** | Add `CheckoutInitiationLimit` (e.g. 10/min per IP). Log all checkout attempts to `syslog` for fraud detection |
| **Coupon brute-force** | Add `CouponValidateLimit` (e.g. 30/min per IP). After 5 failed attempts on a single coupon code from one IP, temporarily block that IP+code combo |
| **Certificate verification abuse** | Add `CertificateVerifyLimit` (e.g. 60/min per IP) on `/api/public/certificates/verify/:token` |
| **Cross-tenant content access** | New `RequireEnrollment` middleware: fetches `courseId` from URL, checks `enrollments` collection for `(userId, courseId, status=active)`. Returns 403 (NOT 404 — avoid leaking course existence) if missing. Apply to all `/api/learner/lessons/:id` and `/api/learner/media/:id` routes |
| **Video hotlinking** | Signed S3/CloudFront URLs with 10-minute TTL, issued only after `RequireEnrollment` passes. Add `Referer` check on media endpoint (reject if Referer doesn't match platform or creator's custom domain) |
| **Drip content bypass** | New `RequireDripEligibility` middleware: checks `enrollmentDate + section.DripOffsetDays <= now`. Returns 403 if too early |
| **Custom domain spoofing** | `CustomDomainTenantMiddleware` only resolves domains with `status=active` AND `dns_verified=true`. Pending/failed domains fall through to platform branding. Caddy's `ask` endpoint independently verifies before issuing certs |
| **Creator API keys** | Extend `APIKey` model with `TenantID` field. Creator-issued keys have `Authority=APIKeyAuthorityUser` and are scoped to their tenant. Validate in middleware that the API key's `TenantID` matches the request's resolved tenant |
| **Stripe Connect key isolation** | Creator's `StripeConnectAccountID` is read-only after onboarding. Only platform admin can update it. Webhook signature verification uses the platform's webhook secret (Connect events are signed by Stripe and verifiable with platform secret) |
| **Marketplace commission fraud** | All `FinancialTransaction` records are immutable after creation. Platform admin can audit but not edit. Commission rate changes only apply to future sales, never retroactively |
| **PII in URLs** | Certificate verification uses opaque random tokens (32 bytes hex), NOT user IDs or enrollment IDs. Certificate download URLs are signed |
| **CSRF on storefront checkout** | Stripe Checkout handles CSRF internally. For any non-Stripe POST on storefront (e.g. coupon validation), require `Origin` header to match the resolved tenant's custom domain or the platform domain |

### 3.3 Production security checklist before launch

- [ ] Audit all new handlers with `go vet` + `golangci-lint`
- [ ] Run `npx eslint` on all new frontend files
- [ ] Add Playwright e2e test that verifies a learner from tenant A cannot access tenant B's paid content via direct API call
- [ ] Add Playwright e2e test that verifies a learner cannot access a drip-gated lesson before its release date
- [ ] Load-test storefront endpoints with `k6` or `vegeta` to verify rate limiter holds
- [ ] Pen-test the checkout flow with Stripe's test cards (declined, insufficient funds, 3DS)
- [ ] Verify CSP doesn't block video player (HLS playback from CDN may need `media-src` directive update in `middleware/security.go`)
- [ ] Enable Stripe Radar rules for marketplace fraud
- [ ] Configure Stripe Connect to require KYC for all creators before first payout

---

## Part 4 — Testing Strategy

### 4.1 Backend testing conventions (follow exactly)

LastSaaS uses standard Go testing with these helpers in `internal/testutil/testutil.go`:

```go
// Setup pattern (from syslog_test.go)
func TestCourseHandler_Create(t *testing.T) {
    db := testutil.MustConnectTestDB(t)        // real MongoDB (testcontainers or local)
    defer testutil.CleanupCollections(t, db)

    // Seed user + tenant + membership using testhelpers
    user, tenant, token := testutil.SeedAuthenticatedUser(t, db, models.RoleCreator)

    // Mock Stripe if needed
    mockStripe := setupMockStripe(t)   // pattern from stripe_test.go
    defer mockStripe.Close()

    handler := NewCourseHandler(db, mockStripe.Client)
    req := httptest.NewRequest("POST", "/api/creator/courses", body)
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("X-Tenant-ID", tenant.ID.Hex())

    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    require.Equal(t, http.StatusCreated, rr.Code)
    // ... assert on response body, DB state, event emission
}
```

Conventions:
- One `_test.go` per handler file, in the same package
- Use `testify/require` + `testify/assert`
- Use `t.Run` for sub-tests
- Test both happy path AND error cases (validation, auth, tenant isolation, rate limit)
- For Stripe: always mock via `httptest.NewServer` (see `setupMockStripe` in `stripe_test.go`)
- For MongoDB: real connection (testcontainers or local), not in-memory

### 4.2 Frontend testing conventions

- **Unit tests**: `vitest` (config in `vitest.config.ts`). Co-locate as `*.test.tsx` next to components
- **E2E**: Playwright, specs in `frontend/e2e/*.spec.ts`. Config in `playwright.config.ts`. Existing specs: `smoke.spec.ts`, `auth.spec.ts`, `admin.spec.ts`, `navigation.spec.ts`
- For new features, add an e2e spec that covers the full user journey (creator creates course → learner enrolls → learner completes → creator sees payout)

### 4.3 Required test coverage for production

| Area | Required tests |
|---|---|
| Course CRUD | Create, update, delete, publish/unpublish, list by creator, list by storefront (only published), tenant isolation (creator A cannot see creator B's drafts) |
| Enrollment | Checkout → webhook → enrollment created, duplicate enrollment prevention, refund → enrollment revoked, list-by-learner (cross-tenant), list-by-creator |
| Course progress | Upsert on lesson view, completion flag set when video reaches 90%, completion idempotent, certificate auto-issued when all lessons complete |
| Coupon | Create, validate (valid code, expired code, usage-limit-reached, wrong tenant), apply at checkout (price reduction), used_count increment, refund restores used_count |
| Review | Only enrolled learners can review, one review per user per course, rating validation (1-5), creator can hide (not delete) reviews |
| Payout | Stripe Connect onboarding flow (mock), payout request → Stripe Transfer created → webhook updates status, failed payout handling |
| Custom domain | Add domain, DNS lookup verification (mock), Caddy `ask` endpoint returns 200 for active / 404 for unknown, custom domain middleware resolves correct tenant, fallback to platform domain works |
| Stripe Connect webhooks | `transfer.reversed` creates `TransactionConnectTransferReversal`, `charge.dispute.funds_withdrawn` creates `TransactionDisputeWithdrawal`, `charge.dispute.funds_reinstated` creates `TransactionDisputeReinstatement`, `account.updated` updates `Tenant.StripeConnectStatus` |
| Drip | Lesson accessible immediately if no drip config, lesson 403 before drip date, lesson 200 after drip date, drip respects section-level config |
| Certificate | Auto-issued on course completion, verification token is unique, public verification page works without auth, PDF download contains correct data |
| Security | Cross-tenant access denied (403), unauthenticated storefront works, unauthenticated learner endpoints 401, rate limit returns 429 after threshold |

---

## Part 5 — Deployment & Configuration

### 5.1 New environment variables (extend `internal/config/config.go`)

```yaml
# config/prod.yaml additions
s3:
  bucket: "my-courses-prod"
  region: "us-east-1"
  access_key_id: "${S3_ACCESS_KEY_ID}"      # from env
  secret_access_key: "${S3_SECRET_ACCESS_KEY}"
  cdn_base_url: "https://cdn.mycourses.com"  # optional CloudFront

stripe:
  connect:
    enabled: true
    client_id: "${STRIPE_CONNECT_CLIENT_ID}"  # for OAuth onboarding
    webhook_secret: "${STRIPE_CONNECT_WEBHOOK_SECRET}"

custom_domains:
  enabled: true
  platform_domain: "mycourses.com"
  caddy_ask_endpoint: "http://localhost:8080/internal/caddy/allowed"

media:
  max_video_size_mb: 2048
  allowed_video_types: ["video/mp4", "video/quicktime"]
  allowed_doc_types: ["application/pdf"]
  signed_url_ttl_minutes: 10

marketplace:
  discovery_enabled: true
  default_commission_rate_bps: 1000   # 10%
```

### 5.2 New dynamic config (extend `internal/configstore/seed.go`)

These can be changed at runtime via the admin console without restart:

```go
// configstore/seed.go additions
{
    Key: "course.default_commission_rate_bps",
    Value: 1000,
    Description: "Default platform commission in basis points (1000 = 10%)",
    Category: "marketplace",
},
{
    Key: "course.payout_schedule",
    Value: "weekly",  // daily | weekly | monthly | manual
    Description: "When creator payouts are initiated",
    Category: "marketplace",
},
{
    Key: "custom_domains.enabled",
    Value: true,
    Description: "Allow creators to add custom domains",
    Category: "custom_domains",
},
{
    Key: "marketplace.discovery_enabled",
    Value: true,
    Description: "Show courses in the public marketplace",
    Category: "marketplace",
},
{
    Key: "course.refund_window_days",
    Value: 30,
    Description: "How many days after purchase a learner can request a refund",
    Category: "course",
},
```

### 5.3 Dockerfile changes

The existing multi-stage Dockerfile already builds the Go backend, builds the React frontend, and packages everything with Caddy. The only addition: copy the new `Caddyfile` with `on_demand_tls` configuration.

```dockerfile
# In the final stage
COPY Caddyfile /etc/caddy/Caddyfile
```

```caddyfile
# Caddyfile
{
    on_demand_tls {
        ask http://localhost:8080/internal/caddy/allowed
    }
    email admin@mycourses.com
}

# Platform domain with explicit cert
https://mycourses.com, https://*.mycourses.com {
    reverse_proxy localhost:8080
}

# Custom domains (on-demand TLS)
https:// {
    tls {
        on_demand
    }
    reverse_proxy localhost:8080
}
```

### 5.4 fly.toml additions

```toml
[env]
S3_BUCKET = "my-courses-prod"
STRIPE_CONNECT_ENABLED = "true"
CUSTOM_DOMAINS_ENABLED = "true"

[experimental]
# Caddy needs to bind 443 for on-demand TLS
```

### 5.5 Bootstrap flow (extend `api/handlers/bootstrap.go`)

The existing bootstrap creates a root tenant + admin user. Extend it to also:
1. Seed the 3 creator plans (Free/Pro/Business) via `planstore.Seed`
2. Seed the dynamic config defaults via `configstore.Seed`
3. Create the platform's own `BrandingConfig` (the marketplace homepage branding)

### 5.6 Zero-downtime migration strategy

For each new collection:

1. Add the model struct + JSON schema validator + indexes to `internal/db/schema.go`
2. On next deploy, the app calls `db.CreateCollection` (idempotent — skips if exists) with the validator, then `Indexes().CreateMany` (idempotent — skips existing indexes)
3. For indexes on existing collections (e.g. adding `tenantId` index to a new sub-collection), same approach — `CreateMany` is idempotent
4. **Critical**: never use `collMod` to change a validator that would reject existing documents. Always make new fields optional in the validator
5. For data backfill (e.g. setting `Tenant.Slug` from `Tenant.Name` for all existing tenants), add a one-time migration function called from `cmd/lastsaas/main.go` as `lastsaas migrate add-tenant-slugs`

---

## Part 6 — Observability

### 6.1 Reuse the existing observability stack

| Layer | What exists | How to extend |
|---|---|---|
| **Health checks** (`internal/health/`) | `/api/health` endpoint, integrations panel in admin UI | Add health checks for: S3 connectivity, Stripe API reachability, Caddy on-demand TLS endpoint |
| **Metrics** (`internal/metrics/`, `middleware/metrics.go`) | Prometheus-style counters, request duration histograms | Add metrics: `course_published_total`, `enrollment_created_total{tenant_id}`, `payout_paid_total{tenant_id}`, `video_bytes_served_total{tenant_id}`, `custom_domain_active_total` |
| **Telemetry** (`internal/telemetry/service.go`) | Product analytics, custom event tracking via `POST /api/telemetry/events` | Emit custom events: `course.viewed` (storefront), `course.published`, `enrollment.created`, `lesson.completed`, `payout.requested`, `certificate.downloaded`. These appear in the admin "Custom Event Explorer" automatically |
| **System log** (`internal/syslog/`) | Structured log storage in MongoDB, queryable via admin UI | Log all course lifecycle events, all payout events, all custom domain provisioning events with `tenantId` for filtering |
| **Datadog** (`internal/datadog/client.go`) | Optional Datadog integration | Forward the new metrics above to Datadog |

### 6.2 Admin dashboard additions

| Existing admin page | New sections to add |
|---|---|
| `pages/admin/HealthPage.tsx` | S3 storage health, Stripe Connect health, Caddy TLS provisioning queue |
| `pages/admin/DashboardPage.tsx` | Platform-wide metrics: total courses, total enrollments (today/30d/all-time), GMV, total payouts, active creators |
| `pages/admin/FinancialPage.tsx` | Marketplace transactions view, commission revenue, dispute tracking |
| `pages/admin/PMPage.tsx` (product analytics) | New event types automatically appear in Custom Event Explorer |

### 6.3 Creator-facing analytics

In `pages/app/creator/CreatorDashboardPage.tsx`:
- Revenue today / 30d / all-time
- Enrollments today / 30d
- Top courses by revenue
- Top courses by enrollment
- Conversion rate (storefront views → checkouts → enrollments)
- Payout history with status
- Video storage used vs. plan limit
- Video bandwidth used this month vs. plan limit

All data sourced from existing `FinancialTransaction`, `Enrollment`, and `UsageEvent` collections with appropriate `tenantId` filter.

---

## Part 7 — Webhooks & Events

### 7.1 Existing event system

LastSaaS has a robust event emitter + webhook dispatcher:

- `internal/events/emitter.go` — synchronous in-process event emission
- `internal/webhooks/dispatcher.go` — async webhook delivery with retries
- `internal/webhooks/crypto.go` — HMAC-SHA256 signing of webhook payloads
- `models/event_definition.go` — admin-configurable event definitions (19 built-in)
- `api/handlers/event_definitions.go` — admin CRUD for event types
- `api/handlers/webhooks.go` — admin CRUD for webhook endpoints (per-tenant)

### 7.2 New event types to register

Register these via `models.EventDefinition` entries (seeded in `internal/configstore/seed.go` or admin UI):

| Event key | Emitted when | Payload |
|---|---|---|
| `course.created` | Creator saves a new course draft | `courseId`, `tenantId`, `title` |
| `course.published` | Creator publishes a course | `courseId`, `tenantId`, `title`, `priceCents` |
| `course.unpublished` | Creator unpublishes | `courseId`, `tenantId` |
| `enrollment.created` | Learner completes checkout | `enrollmentId`, `courseId`, `learnerUserId`, `tenantId`, `pricePaidCents` |
| `enrollment.completed` | Learner completes all lessons | `enrollmentId`, `courseId`, `learnerUserId`, `tenantId` |
| `lesson.completed` | Learner completes a single lesson | `enrollmentId`, `lessonId`, `learnerUserId` |
| `payout.paid` | Stripe Transfer succeeds | `payoutId`, `tenantId`, `amountCents`, `stripeTransferId` |
| `payout.failed` | Stripe Transfer fails | `payoutId`, `tenantId`, `failureReason` |
| `certificate.issued` | Certificate auto-generated | `certificateId`, `enrollmentId`, `learnerUserId`, `courseId` |
| `custom_domain.activated` | Custom domain passes DNS + SSL checks | `domain`, `tenantId` |
| `custom_domain.failed` | Custom domain provisioning fails | `domain`, `tenantId`, `failureReason` |
| `refund.processed` | Course purchase refunded | `enrollmentId`, `amountCents`, `reason` |
| `review.posted` | Learner posts a review | `reviewId`, `courseId`, `rating` |

### 7.3 Emission pattern (follow existing convention)

```go
// In course handler, after publishing
err := h.eventEmitter.Emit(ctx, events.Event{
    Type:    "course.published",
    TenantID: tenant.ID,
    Payload: map[string]any{
        "courseId":    course.ID.Hex(),
        "title":       course.Title,
        "priceCents":  course.PriceCents,
    },
})
```

The emitter dispatches to all registered webhooks for that tenant (with HMAC signing), records to `syslog`, and forwards to telemetry. **No new infrastructure needed.**

---

## Part 8 — Stripe Connect Marketplace (Detailed)

### 8.1 Onboarding flow

1. Creator clicks "Connect Stripe" in `/studio/payouts`
2. Frontend calls `GET /api/creator/connect/onboarding`
3. Backend calls `stripe.Account.New({Type: "express", Country: "US", Metadata: {tenant_id: ...}})` → stores `acct_...` in `Tenant.StripeConnectAccountID`
4. Backend calls `stripe.AccountLink.New({Account: acctID, Type: "account_onboarding", ReturnURL: ..., RefreshURL: ...})` → returns URL
5. Frontend redirects to Stripe-hosted onboarding
6. Creator completes KYC, bank account, etc.
7. Stripe sends `account.updated` webhook → backend updates `Tenant.StripeConnectStatus = "active"`
8. Creator can now receive payouts

### 8.2 Course purchase flow

1. Learner clicks "Buy" on `/courses/:slug`
2. Frontend calls `POST /api/storefront/checkout/:courseSlug` with optional `couponCode`
3. Backend validates coupon (if provided), computes final price
4. Backend creates `stripe.Checkout.Session` with:
   - `Mode: "payment"`
   - `LineItems: [{PriceData: {Currency, UnitAmount: finalPrice, ProductData: {Name: courseTitle}}}]`
   - `PaymentIntentData.ApplicationFeeAmount: finalPrice * commissionBps / 10000`
   - `PaymentIntentData.TransferData.Destination: creatorStripeConnectAccountID`
   - `ClientReferenceID: courseID + "|" + learnerUserID` (for webhook correlation)
   - `Metadata: {course_id, learner_user_id, tenant_id, coupon_id}`
   - `SuccessURL: /learn/courses/:courseId?status=success`
   - `CancelURL: /courses/:slug?status=cancelled`
5. Backend returns `session.URL` to frontend → frontend redirects
6. Learner pays on Stripe-hosted page
7. Stripe sends `checkout.session.completed` webhook
8. Backend handler (`handleCheckoutCompleted` extended):
   - Parses `metadata.course_id`, `metadata.learner_user_id`, `metadata.tenant_id`
   - Creates `Enrollment` record (status: active, price_paid_cents: session.AmountTotal, coupon_id from metadata)
   - Creates `FinancialTransaction` (type: `TransactionCoursePurchase`, amount: session.AmountTotal, linked to course + learner + creator tenant)
   - Creates `FinancialTransaction` (type: `TransactionPlatformCommission`, amount: session.AmountTotal * commissionBps / 10000)
   - Emits `enrollment.created` event (triggers email to learner + webhook to creator)
   - Increments `CourseCoupon.UsedCount` if coupon was used

### 8.3 Payout flow

**Option A: Stripe-managed (recommended)** — Set the creator's Express account to auto-payout daily/weekly/monthly. Stripe handles the schedule. You just listen for `payout.paid` / `payout.failed` webhooks and record `Payout` records for the creator's dashboard.

**Option B: Platform-managed** — On a schedule (cron job via `cmd/lastsaas/main.go` or a worker), query all creators with positive balance, call `stripe.Transfer.New({Amount, Currency, Destination: creatorAcctID})`, record `Payout` record with `status=pending`, update to `status=paid` when `payout.paid` webhook arrives.

### 8.4 Refund & dispute handling

| Scenario | Stripe API call | Webhook events fired | FinancialTransaction records created | Enrollment update |
|---|---|---|---|---|
| **Full refund** (platform-initiated) | `stripe.Refund.New({Charge: originalChargeID})` | `charge.refunded` (existing handler) + `transfer.reversed` (NEW handler) | `TransactionRefund` (negative, platform side) + `TransactionConnectTransferReversal` (negative, creator side) | Status → `refunded` |
| **Partial refund** | `stripe.Refund.New({Charge, Amount: partialAmount})` | Same as above, with partial amounts | Same types, partial amounts | Status stays `active` (or `partial_refund` if you want to track) |
| **Chargeback initiated** | (no action — Stripe notifies) | `charge.dispute.created` (existing) | Optionally `TransactionDisputePending` | Status → `disputed` |
| **Chargeback funds withdrawn** | (no action) | `charge.dispute.funds_withdrawn` (NEW) | `TransactionDisputeWithdrawal` (negative, against whoever bears liability — usually the creator) | No change |
| **Chargeback won, funds reinstated** | (no action) | `charge.dispute.funds_reinstated` (NEW) + `charge.dispute.closed` (existing) | `TransactionDisputeReinstatement` (positive) | Status → `active` |
| **Chargeback lost** | (no action) | `charge.dispute.closed` (existing) | (no new transaction — `TransactionDisputeWithdrawal` already recorded the loss) | Status → `chargeback_lost` |

**Liability decision**: By default, the creator bears chargeback liability (their course, their customer). Debit the creator's balance via `TransactionDisputeWithdrawal` linked to their tenant. If the creator disputes and wins, credit via `TransactionDisputeReinstatement`. Document this clearly in the creator TOS.

---

## Part 9 — Video Streaming & Content Storage

### 9.1 Architecture

```
Creator uploads video
  ↓
Frontend requests presigned PUT URL from /api/creator/media/upload
  ↓
Backend generates S3 presigned PUT URL (15-min TTL)
  ↓
Frontend uploads directly to S3 (bypasses backend — no bandwidth cost)
  ↓
Frontend notifies backend /api/creator/media/:id/complete
  ↓
Backend triggers transcoding (optional — see below)
  ↓
Backend records MediaAsset row with StorageURL
  ↓
Learner requests /api/learner/lessons/:id
  ↓
RequireEnrollment middleware → 403 if not enrolled
  ↓
Backend generates presigned GET URL (10-min TTL) for the HLS manifest or MP4
  ↓
Frontend VideoPlayer plays it, useCourseProgress heartbeat every 15s
```

### 9.2 Transcoding decision

| Option | Pros | Cons |
|---|---|---|
| **No transcoding** (serve original MP4) | Simplest, no infra | No adaptive bitrate, large files for slow connections, no DRM |
| **S3 + CloudFront with HLS** (use AWS MediaConvert or Mux/Cloudflare Stream) | Adaptive bitrate, global CDN, signed URLs | Adds transcoding cost + latency (minutes per video) |
| **Mux/Cloudflare Stream** (managed) | Best DX, handles transcoding + CDN + signed playback in one API | Per-video cost + vendor lock-in |

**Recommendation for MVP**: No transcoding. Accept MP4 uploads, serve via S3 presigned URLs. For v2, migrate to Mux or Cloudflare Stream.

### 9.3 Storage cost metering (reuse `UsageEvent`)

When a creator uploads a video, record:

```go
h.usageHandler.RecordUsage(ctx, usage.UsageRecord{
    TenantID: tenant.ID,
    Type:     "video_storage_mb_hours",
    Amount:   fileSizeBytes / 1024 / 1024,  // MB
    Metadata: map[string]any{"media_asset_id": asset.ID.Hex()},
})
```

When a learner watches, record:

```go
h.usageHandler.RecordUsage(ctx, usage.UsageRecord{
    TenantID: creatorTenant.ID,
    Type:     "video_bandwidth_gb",
    Amount:   bytesServed / 1024 / 1024 / 1024,
    Metadata: map[string]any{"lesson_id": lessonID, "learner_user_id": learnerID},
})
```

The existing `UsageHandler.RecordUsage` does atomic credit deduction. If a creator's plan includes 50GB storage and 1TB bandwidth/month, overage either blocks uploads or charges overage fees (via `CreditBundle` purchase).

### 9.4 Access gating

| Lesson type | Access control |
|---|---|
| Free preview (`IsPreview=true`) | Public presigned URL, no auth required. CDN-cacheable. |
| Paid lesson | `RequireEnrollment` middleware → 10-min presigned URL. URL is single-use per session (frontend must re-request if video is paused > 10 min). |
| Drip-gated lesson | `RequireDripEligibility` middleware in addition to enrollment check |

### 9.5 DRM (optional, v2)

For premium content protection beyond signed URLs, consider:
- **CloudFront signed URLs with custom policy** (restrict by IP + expiry)
- **AWS MediaConvert + Apple FairPlay / Widevine** (full DRM, complex)
- **Mux DRM** (managed, easiest)

For MVP, signed URLs with 10-min TTL + enrollment check is sufficient. Add a `Referer` check on S3 bucket policy to prevent hotlinking.

---

## Part 10 — i18n, Accessibility, Email, Notifications

### 10.1 i18n (must add — not in lastsaas)

LastSaaS has no i18n. For a course platform serving multiple regions, add `react-i18next`:

```bash
cd frontend && npm install react-i18next i18next
```

```
frontend/src/
├── i18n/
│   ├── config.ts          # i18next init
│   ├── locales/
│   │   ├── en.json
│   │   ├── es.json
│   │   ├── fr.json
│   │   ├── de.json
│   │   ├── pt.json
│   │   └── ...
├── hooks/
│   └── useLocale.ts       # reads from user profile + browser
```

Namespace translations by feature: `common`, `storefront`, `creator`, `learner`, `admin`, `auth`, `billing`.

Add `Locale` field to `User` model. Default to browser locale on signup, user can change in settings. `BrandingThemeInjector` already injects `<html lang="...">` — extend to read from user locale.

### 10.2 Accessibility

- Use the existing `components/ui/*` primitives (Button, Input, Card, Select, Textarea, Modal, Badge, Alert) — they follow basic Tailwind patterns
- Audit with `@axe-core/playwright` in e2e tests
- For the video player: add closed-caption track support (WebVTT), keyboard shortcuts (play/pause, seek, mute), ARIA labels for all controls
- For the course player sidebar: keyboard-navigable lesson list, ARIA `current="page"` on active lesson
- Verify color contrast meets WCAG AA (4.5:1) for all branded themes — add a contrast checker to the branding settings page that warns creators if their chosen colors fail

### 10.3 RTL support

Tailwind CSS 4 supports RTL via `dir="rtl"` on `<html>` + logical properties (`ms-`, `me-`, `ps-`, `pe-`). Add `dir` attribute based on locale. Test with Arabic/Hebrew locales.

### 10.4 Email templates (extend `internal/email/resend.go`)

LastSaaS uses Resend with inline HTML string templates. Add new templates as functions:

```go
// internal/email/templates.go (new file)
func CourseEnrollmentEmail(courseTitle, creatorName, learnerName, courseURL string) string {
    return fmt.Sprintf(`<html>...branded HTML with course title, link, receipt...</html>`)
}

func PayoutPaidEmail(creatorName, amountFormatted, payoutDate string) string { ... }
func LessonCompletedEmail(courseTitle, lessonTitle, learnerName, nextLessonURL string) string { ... }
func CertificateIssuedEmail(courseTitle, learnerName, certificateURL, verificationToken string) string { ... }
func CoursePublishedEmail(creatorName, courseTitle, storefrontURL string) string { ... }
func AbandonedCartEmail(learnerName, courseTitle, checkoutURL string) string { ... }
func RefundProcessedEmail(courseTitle, amountFormatted, learnerName string) string { ... }
```

### 10.5 Notification preferences

Add `NotificationPreferences` struct to `User` model:

```go
type NotificationPreferences struct {
    EmailCourseEnrollment bool `bson:"email_course_enrollment" json:"email_course_enrollment"`
    EmailLessonCompleted  bool `bson:"email_lesson_completed" json:"email_lesson_completed"`
    EmailCertificateIssued bool `bson:"email_certificate_issued" json:"email_certificate_issued"`
    EmailPayoutPaid       bool `bson:"email_payout_paid" json:"email_payout_paid"`
    EmailCoursePublished  bool `bson:"email_course_published" json:"email_course_published"`
    EmailMarketing        bool `bson:"email_marketing" json:"email_marketing"`
}
```

Defaults: all `true` except `EmailMarketing`. Learner can edit in `/learn/settings`. Creator can edit in `/studio/settings`.

Before sending any email, check the relevant preference. Centralize in a `notificationService.Send(notificationType, user, data)` that respects preferences + renders the template + calls Resend.

---

## Part 11 — Marketplace Discovery, SEO, Certificates, Drip

### 11.1 Marketplace discovery page

Teachable has a public marketplace. Build it as a new public route at `/marketplace` (or `/discover`):

| Concern | Implementation |
|---|---|
| Cross-tenant course index | Query `courses` collection with `{status: "published"}` (no `tenantId` filter). Add compound index `{status: 1, published_at: -1}` for fast listing |
| Search | Use MongoDB text index on `courses` collection: `db.courses.createIndex({title: "text", description: "text"})`. Query via `{$text: {$search: query}}` |
| Filter by category | Add `Category` field to `Course`. Index on `{status: 1, category: 1, published_at: -1}` |
| Filter by price (free/paid) | Derive from `PriceCents` at query time |
| Filter by rating | Maintain `rating_average` and `rating_count` denormalized fields on `Course`, updated on review creation. Index on `{status: 1, rating_average: -1}` |
| Pagination | Use the existing convention (see §13 of v1 plan): `?page=1&limit=20`, response shape `{items, total, page, limit}` |
| Card display | Show creator name, course title, thumbnail, price, rating, lesson count. Creator branding (logo, color) is shown on the card as a small badge. Marketplace page itself uses platform branding. |
| Click-through | Link to the creator's storefront: `https://[creator-custom-domain-or-subdomain]/courses/:slug`. If creator has no custom domain, use `https://[platform-domain]/s/:creatorSlug/courses/:slug` |
| Keeping index fresh | Courses are denormalized — no separate index to maintain. The `rating_average`/`rating_count` fields are updated atomically on review creation via `FindOneAndUpdate` with `$inc` and `$set` |

### 11.2 SEO

| Concern | Implementation |
|---|---|
| Clean URLs | `/courses/:slug` (slug unique per tenant), marketplace at `/marketplace`, creator storefront at `/s/:creatorSlug` or custom domain root |
| Per-course meta tags | `BrandingThemeInjector` already injects `<head>` content. Extend to accept per-page overrides (title, description, OG image, JSON-LD). On `CourseDetailPage`, fetch course + inject `Course` schema.org JSON-LD |
| Sitemaps | New endpoint `GET /sitemap.xml` that lists all published courses across all tenants. Cache for 1 hour. Submit to Google Search Console |
| Server-side rendering | LastSaaS is a SPA (Vite). For SEO, either (a) add `react-router-dom` data loaders + a lightweight prerender step, or (b) accept that Google can render JS (modern Googlebot does). For MVP, (b) is fine — ensure meta tags are in `index.html` template and updated via `BrandingThemeInjector` |
| Canonical URLs | When a course is accessible via both custom domain and platform subdomain, set `<link rel="canonical">` to the custom domain URL |

### 11.3 Certificates

| Concern | Implementation |
|---|---|
| Auto-issuance | In `course_progress` handler, when last lesson of a course is marked complete: create `Certificate` record with `CertificateNumber` (e.g. `CERT-2026-{seq}`) + `VerificationToken` (32-byte hex random). Emit `certificate.issued` event → triggers email |
| PDF generation | Use `chromedp` (headless Chrome) or `gofpdf` to render certificate HTML → PDF. Template includes course title, learner name, creator name, issue date, certificate number, verification URL. Store PDF in S3, serve via signed URL |
| Verification page | Public route `/certificates/verify/:token` — fetches `Certificate` by `VerificationToken`, shows learner name, course title, issue date, "Verified ✓" badge. No auth required. |
| Sharing | Learner can share verification URL on LinkedIn. Add "Add to LinkedIn Profile" button using LinkedIn's ACAP API |

### 11.4 Drip content

| Concern | Implementation |
|---|---|
| Configuration | Add `DripOffsetDays int` to `Section` (e.g. 0 = immediate, 7 = 7 days after enrollment). Optional: per-lesson offset. |
| Enforcement | `RequireDripEligibility` middleware: fetch lesson → section → compute `enrollmentDate + section.DripOffsetDays`. If `time.Now() < requiredDate`, return 403 with `{"available_at": requiredDate}` |
| Frontend | `CoursePlayerPage` checks drip status for each lesson. Locked lessons show a lock icon + "Available in X days" tooltip. Lesson list shows progression. |
| Edge case | If a creator changes `DripOffsetDays` after enrollments exist, the new value applies prospectively (doesn't unlock content earlier than originally promised, doesn't delay already-unlocked content). Track `drip_unlocked_at` per `CourseProgress` row once unlocked. |

---

## Part 12 — GDPR / Data Privacy

### 12.1 What lastsaas already gives you

- `api/handlers/auth.go::ExportData` — exports user's data as JSON (extend for course data)
- `api/handlers/auth.go::DeleteAccount` — deletes user account (extend to cascade)
- Session/refresh token management with explicit revocation
- Telemetry events can be anonymized by setting `userId = null`

### 12.2 New work for course platform

| GDPR right | Implementation |
|---|---|
| **Right to access** (data export) | Extend `ExportData` to include: `enrollments`, `course_progress`, `reviews`, `certificates`, `purchases` (FinancialTransactions where learner is involved). Return as JSON download |
| **Right to be forgotten** | Extend `DeleteAccount` to: (1) anonymize `enrollments` (set `userId = null`, keep for creator's revenue records — legitimate interest), (2) anonymize `course_progress` (set `userId = null`), (3) delete `reviews` (or anonymize author), (4) revoke `certificates` (mark as `learner_deleted`, keep verification page showing "Certificate revoked" — prevents fraud), (5) delete `custom_domains` if user was a creator, (6) anonymize `FinancialTransaction.UserID` |
| **Right to rectification** | Standard profile edit page already exists |
| **Data minimization** | Don't collect more than needed. For learners: email, name, payment method (via Stripe — not stored locally). For creators: additional KYC data via Stripe Connect (not stored locally) |
| **Consent** | Add a consent checkbox at signup: "I agree to the Terms of Service and Privacy Policy". Store `consent_accepted_at` timestamp on `User` |
| **Data retention** | Auto-delete `course_progress` for refunded enrollments after 90 days. Keep `FinancialTransaction` records for 7 years (tax/legal requirement) |
| **Cross-border transfer** | If using S3, configure bucket region to match user's locale (EU users → eu-west-1, etc.). Stripe handles this for payment data |
| **Cookie consent** | Add a cookie consent banner on the marketplace page. LastSaaS uses cookies for auth + analytics — categorize them in the banner |

### 12.3 Creator data deletion

When a creator deletes their school (tenant):

1. Mark all `courses` as `archived`
2. Notify all enrolled learners (email) that the school is closing in 30 days
3. After 30 days: anonymize `enrollments`, `course_progress`, `reviews`. Delete `courses`, `sections`, `lessons`, `media_assets`. Revoke `certificates` (mark as `school_closed`).
4. Keep `FinancialTransaction` records for tax compliance (7 years)
5. Delete `custom_domains` (release the domains)
6. Mark `tenant` as `deleted` (soft delete — keep for audit trail)

---

## Part 13 — Scaling to 10k Creators + 100k Learners

### 13.1 Statelessness of the backend

LastSaaS backend is **stateless** (good for horizontal scaling):
- Sessions are JWT-based (no server-side session store)
- Rate limiter uses MongoDB (shared across instances)
- Event emitter is in-process but events are persisted to MongoDB before dispatch
- Webhook dispatcher reads from MongoDB queue (works across instances)

**Bottleneck**: Webhook dispatcher. At 10k creators each with 1-5 webhook endpoints, dispatch latency could degrade. **Mitigation**: Migrate webhook dispatch to a dedicated worker (separate process reading from a MongoDB queue, or move to Redis + BullMQ for v2).

### 13.2 Critical MongoDB indexes for scale

```javascript
// Courses
db.courses.createIndex({tenant_id: 1, status: 1, published_at: -1})  // storefront listing
db.courses.createIndex({status: 1, published_at: -1})                 // marketplace
db.courses.createIndex({tenant_id: 1, slug: 1}, {unique: true})
db.courses.createIndex({title: "text", description: "text"})          // search

// Enrollments
db.enrollments.createIndex({user_id: 1, status: 1})                   // learner dashboard
db.enrollments.createIndex({tenant_id: 1, enrollment_date: -1})       // creator sales
db.enrollments.createIndex({user_id: 1, course_id: 1}, {unique: true}) // prevent duplicates

// Course progress
db.course_progress.createIndex({enrollment_id: 1, lesson_id: 1}, {unique: true})
db.course_progress.createIndex({user_id: 1, last_viewed_at: -1})

// Custom domains
db.custom_domains.createIndex({domain: 1}, {unique: true})
db.custom_domains.createIndex({tenant_id: 1, status: 1})

// Media assets
db.media_assets.createIndex({tenant_id: 1, created_at: -1})
db.media_assets.createIndex({kind: 1, tenant_id: 1})

// Financial transactions
db.financial_transactions.createIndex({tenant_id: 1, type: 1, created_at: -1})
db.financial_transactions.createIndex({user_id: 1, type: 1, created_at: -1})

// Payouts
db.payouts.createIndex({tenant_id: 1, status: 1, initiated_at: -1})

// Certificates
db.certificates.createIndex({verification_token: 1}, {unique: true})
db.certificates.createIndex({user_id: 1, course_id: 1}, {unique: true})
db.certificates.createIndex({certificate_number: 1}, {unique: true})
```

### 13.3 Infrastructure changes for scale

| Component | MVP (lastsaas default) | 10k creators / 100k learners |
|---|---|---|
| **MongoDB** | Single Atlas M10 instance | M30+ with replica set. Add read replicas for analytics queries (admin dashboard, creator dashboard). Consider sharding `course_progress` by `user_id` (largest collection) |
| **Backend instances** | 1 Fly.io machine | 3-5 machines behind Fly load balancer. Stateless, so horizontal scaling is trivial |
| **Frontend** | Served by Go backend (SPA) | Move to CloudFront/Cloudflare Pages. Backend serves only API |
| **Media storage** | S3 with presigned URLs | S3 + CloudFront CDN for playback. Consider Cloudflare Stream or Mux for transcoding + DRM at scale |
| **Webhook dispatch** | In-process (works for MVP) | Dedicated worker process. Migrate to Redis + BullMQ or AWS SQS for reliable at-scale dispatch |
| **Branding assets** | MongoDB GridFS (current lastsaas) | Migrate to S3 (GridFS doesn't scale for binary delivery) |
| **Search** | MongoDB text index | Elasticsearch or Algolia for faceted search, typo tolerance, relevance tuning |
| **Rate limiter** | MongoDB-backed (current) | Redis-backed for lower latency at scale |
| **Email** | Resend (current) | Resend handles scale. Add SendGrid as backup. Use webhooks to track delivery/bounces |
| **Caddy on-demand TLS** | Single Caddy instance | Caddy cluster with shared cert store (Redis backend) or move to Cloudflare for SaaS (managed custom-domain TLS) |

### 13.4 Cost profile (rough estimates at 10k creators / 100k learners)

| Item | Monthly cost estimate |
|---|---|
| Fly.io (5 machines, 2GB each) | ~$250 |
| MongoDB Atlas M30 + 2 read replicas | ~$600 |
| S3 storage (50TB video, 1TB DB) | ~$1,200 |
| CloudFront bandwidth (50TB/mo) | ~$2,500 |
| Resend email (1M emails/mo) | ~$80 |
| Stripe fees (3% + $0.30 per transaction) | Variable (passed to learner or split with creator) |
| Domain/DNS management | ~$50 |

**Total infra: ~$4,500/mo** at 100k learners. Revenue from 10k creators × $29/mo avg = $290k/mo → healthy margin.

---

## Part 14 — CI/CD & Quality Gates

### 14.1 Extend the existing CI (`.github/workflows/ci.yml`)

```yaml
# Add to existing ci.yml
jobs:
  backend-test:
    runs-on: ubuntu-latest
    services:
      mongodb:
        image: mongo:7
        ports: ["27017:27017"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - run: cd backend && make test-integration
      - run: cd backend && make test
      - uses: codecov/codecov-action@v4
        with: { file: ./backend/coverage.out }

  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd frontend && npm ci && npm test
      - run: cd frontend && npm run lint
      - run: cd frontend && npm run build

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker compose up -d
      - run: cd frontend && npx playwright test
      - uses: actions/upload-artifact@v4
        if: failure()
        with: { name: playwright-report, path: frontend/playwright-report/ }

  stripe-webhook-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: cd backend && go test ./internal/stripe/... -run TestConnectWebhookSignature
```

### 14.2 New Makefile targets (`backend/Makefile`)

```makefile
test-course:
	go test ./internal/api/handlers/ -run "TestCourse|TestEnrollment|TestPayout|TestCustomDomain"

test-connect:
	go test ./internal/stripe/ -run "TestConnect"

lint:
	golangci-lint run ./...

migrate-add-tenant-slugs:
	go run ./cmd/lastsaas migrate add-tenant-slugs
```

### 14.3 Pre-merge checklist (enforce via CODEOWNERS + branch protection)

- [ ] All new handlers have `_test.go` files with happy path + error path + tenant isolation test
- [ ] All new frontend pages have at least one Playwright e2e covering the happy path
- [ ] `golangci-lint` passes with zero warnings
- [ ] `eslint` passes with zero warnings
- [ ] No secrets in code (use `gitleaks` pre-commit hook)
- [ ] OpenAPI spec updated (auto-generated from handler annotations)
- [ ] If new env vars added: `config/dev.example.yaml` and `config/prod.example.yaml` updated
- [ ] If new DB collections: `internal/db/schema.go` updated with validator + indexes
- [ ] If new event types: registered in `configstore/seed.go`
- [ ] If new email templates: `internal/email/templates.go` updated + test send to self
- [ ] If new Stripe webhook events: handler added + test with Stripe CLI (`stripe listen --forward-to localhost:8080/api/webhooks`)

---

## Part 15 — Revised Implementation Roadmap

### Phase 0 — Setup (3 days)
- Fork lastsaas. Rename project. Update `VERSION`, `VERSIONS.md`, `README.md`.
- Add new env vars to `config/dev.example.yaml` + `config/prod.example.yaml`
- Set up S3 bucket + IAM user for media storage
- Set up Stripe Connect (create platform account, get client ID)
- Configure Caddyfile with `on_demand_tls`

### Phase 1 — Domain Models & Migrations (1 week)
- Add 12 new model files in `backend/internal/models/`
- Extend `internal/db/schema.go` with validators + indexes
- Extend `models/membership.go` with `RoleCreator`, `RoleStudent`
- Extend `models/tenant.go` with `StripeConnectAccountID`, `CommissionRateBps`, `Slug`
- Extend `models/billing.go` with new `TransactionType` constants
- Extend `models/plan.go` with creator entitlements
- Extend `internal/planstore/seed.go` with Creator Free/Pro/Business plans
- Write migration script for existing tenants (set `Slug` from `Name`)

### Phase 2 — Creator Studio Backend (1.5 weeks)
- Implement `course.go`, `section.go`, `lesson.go`, `media.go` handlers
- Implement `coupon.go`, `review.go` handlers
- Wire routes in `cmd/server/main.go` under `/api/creator/*` with `RequireRole(RoleCreator, RoleOwner, RoleAdmin)`
- Implement S3 presigned URL generation in `media.go`
- Write tests for all handlers following the `billing_test.go` pattern
- Implement `media.go` upload via presigned URL + `complete` notification

### Phase 3 — Creator Studio Frontend (1.5 weeks)
- Build `CourseEditorPage` (tabbed: details, curriculum, pricing, coupons, drip, certificates)
- Build `LessonEditorPage` with video upload (presigned URL flow)
- Build `CoursesPage`, `CouponsPage`, `ReviewsPage`
- Add `CreatorContext` + `client.courses/sections/lessons/coupons/reviews/media` API namespaces
- Add Playwright e2e: `creator.spec.ts`

### Phase 4 — Storefront & Learner Experience (2 weeks)
- Implement `storefront.go` handler (public, no auth, tenant resolved by `CustomDomainTenantMiddleware`)
- Implement `enrollment.go` handler (initiate checkout, list mine, list sales)
- Implement `course_progress.go` handler
- Implement `RequireEnrollment` middleware
- Build `StorefrontHomePage`, `CourseListPage`, `CourseDetailPage`, `CheckoutPage`
- Build `CoursePlayerPage` with `VideoPlayer` + `useCourseProgress` hook
- Build `MyCoursesPage`, `CourseOverviewPage`
- Add Playwright e2e: `storefront.spec.ts`, `learner.spec.ts`

### Phase 5 — Stripe Connect (1.5 weeks)
- Implement `internal/stripe/connect.go`
- Implement `connect.go` handler (onboarding, status, refresh)
- Extend `webhook.go::handleCheckoutCompleted` to detect course purchases + create enrollment + transactions + emit event
- Add new webhook handlers: `transfer.reversed`, `charge.dispute.funds_withdrawn`, `charge.dispute.funds_reinstated`, `account.updated`, `payout.paid`, `payout.failed`
- Implement `payout.go` handler
- Build `ConnectOnboardingPage`, `PayoutsPage`, `SalesPage`
- Add Playwright e2e: full purchase → payout flow with Stripe test mode

### Phase 6 — Custom Domains (1 week)
- Implement `custom_domain.go` middleware
- Implement `custom_domain.go` handler (CRUD + status + DNS verify)
- Implement `caddy.go` handler (`/internal/caddy/allowed` endpoint)
- Configure Caddyfile with `on_demand_tls`
- Build `CustomDomainPage` with DNS instructions + status polling
- Add background worker (goroutine) to poll pending domains' DNS
- Add Playwright e2e: `custom-domain.spec.ts`

### Phase 7 — Certificates, Drip, Reviews (1 week)
- Implement `certificate.go` handler + PDF generation
- Implement `RequireDripEligibility` middleware
- Add `DripOffsetDays` to `Section` model
- Build `CertificateVerifyPage`, `CertificatesPage`
- Build drip UI in `CoursePlayerPage` (locked lessons, "Available in X days")
- Auto-issue certificates on course completion (extend `course_progress` handler)

### Phase 8 — Marketplace & SEO (1 week)
- Build `MarketplacePage` with search + filters
- Implement search via MongoDB text index
- Add JSON-LD `Course` schema to `CourseDetailPage`
- Implement `/sitemap.xml` endpoint
- Add canonical URL handling for custom domain vs subdomain

### Phase 9 — Email, Notifications, i18n (1 week)
- Add `react-i18next` + initial locale files (en, es, fr, de, pt)
- Add `Locale` field to `User` model
- Implement `notificationService` with preference checks
- Add 8 email templates
- Build notification preferences UI in learner + creator settings

### Phase 10 — Admin Console (1 week)
- Build `CreatorsPage`, `MarketplaceTransactionsPage`, `PayoutsAdminPage`, `MarketplaceConfigPage`
- Extend admin dashboard with platform-wide metrics
- Add commission rate configuration (per-tenant override)

### Phase 11 — Security, Testing, Polish (1.5 weeks)
- Add new rate limit types + apply to storefront routes
- Audit all new handlers for tenant isolation
- Run `golangci-lint`, `eslint`, `@axe-core/playwright`
- Load test with `k6` (storefront browsing, checkout, video playback)
- Pen-test checkout flow with Stripe test cards
- Update CSP to allow video playback from CDN
- Enable Stripe Radar rules
- Configure Stripe Connect KYC requirements

### Phase 12 — Launch Readiness (1 week)
- Run full e2e suite in CI
- Set up monitoring (Datadog or Prometheus + Grafana)
- Set up alerting (Stripe webhook failures, payout failures, custom domain SSL failures)
- Write runbook for common incidents (refunds, dispute responses, custom domain DNS issues)
- Soft launch with 5-10 beta creators
- Fix beta feedback
- Public launch

**Total estimated timeline: 13-15 weeks** for a team of 1-2 engineers.

---

## Part 16 — Quick Reference: All 26 Wiki Q&A

| # | File | Topic |
|---|---|---|
| 01 | `01_foundations.json` | Architecture & tech stack |
| 02 | `02_multitenancy_branding.json` | Multi-tenancy + branding deep-dive |
| 03 | `03_billing_credits_stripe.json` | Billing, plans, credits, Stripe |
| 04 | `04_auth_rbac.json` | Auth, users, RBAC, impersonation |
| 05 | `05_frontend_ext.json` | Frontend routes, contexts, branding injection |
| 06 | `06_teachable_mapping.json` | Concrete Teachable transformation mapping |
| 07 | `07_custom_domains.json` | Custom domain request lifecycle |
| 08 | `08_content_storage.json` | Media/content storage patterns |
| 09 | `09_security_ratelimit.json` | Security middleware, rate limiting, API keys, MFA, tenant isolation |
| 10 | `10_testing_patterns.json` | Test conventions, testutil, MongoDB setup, handler tests |
| 11 | `11_deploy_config.json` | Config system, configstore, Dockerfile, fly.toml, bootstrap |
| 12 | `12_observability.json` | Health, metrics, telemetry, syslog, Datadog |
| 13 | `13_pagination_errors.json` | Pagination, error responses, validation, OpenAPI |
| 14 | `14_events_webhooks.json` | Event emitter, webhook dispatcher, signing, event definitions |
| 15 | `15_admin_console.json` | Admin pages, admin API, AdminRoute, configstore management |
| 16 | `16_migrations_seeding.json` | DB schema, seeding, indexes, zero-downtime migration pattern |
| 17 | `17_video_streaming.json` | BrandingHandler asset patterns, S3 presigned URLs, HLS, usage metering |
| 18 | `18_refunds_disputes.json` | Stripe webhook events, FinancialTransaction reversals, Connect marketplace refund/chargeback flows |
| 19 | `19_data_privacy_gdpr.json` | User deletion cascade, data export, telemetry anonymization, GDPR |
| 20 | `20_email_notifications.json` | Resend service, email templates, notification triggers |
| 21 | `21_marketplace_discovery.json` | Cross-tenant course index, search, SEO, schema.org, sitemaps |
| 22 | `22_certificates_drip.json` | Certificate model + PDF + verification, drip content gating pattern |
| 23 | `23_cost_scaling.json` | Statelessness, horizontal scaling, indexes for scale, infra changes at 10k/100k |
| 24 | `24_cicd_quality.json` | CI workflow, linting, versioning, Makefile, manifest files |
| 25 | `25_i18n_a11y.json` | i18n absence, react-i18next recommendation, a11y patterns, RTL |
| 26 | `26_file_checklist.json` | Comprehensive file checklist for the fork |

All files at `/home/z/my-project/download/wiki_answers/`.

---

## How to ask more follow-up questions

The wiki query scripts are persisted at:
- `/home/z/my-project/scripts/ask_one.mjs` — threads full history (use for short follow-up chains)
- `/home/z/my-project/scripts/ask_one_short.mjs` — threads last 2 Q&A pairs (use when history gets long)
- `/home/z/my-project/scripts/ask_one_clean.mjs` — no history (use when safety filter triggers)

Run any of them with:

```bash
node /home/z/my-project/scripts/ask_one_clean.mjs 27_your_id "Your question about lastsaas here"
```

The answer saves to `/home/z/my-project/download/wiki_answers/27_your_id.json`.

---

## Final Production Readiness Checklist

Before going live, verify each item:

### Backend
- [ ] All 12 new models implemented with validation tags
- [ ] All 14 new handlers implemented with tests
- [ ] All 3 new middleware implemented with tests
- [ ] `internal/db/schema.go` has validators + indexes for all new collections
- [ ] Stripe Connect onboarding + checkout + payout + refund flows working in test mode
- [ ] All new webhook handlers tested with Stripe CLI
- [ ] Custom domain middleware + Caddy on-demand TLS working
- [ ] S3 presigned URL upload + playback working
- [ ] Rate limits applied to all storefront endpoints
- [ ] `RequireEnrollment` middleware on all paid content endpoints
- [ ] `RequireDripEligibility` middleware on drip-gated lessons
- [ ] Event emitter fires for all 13 new event types
- [ ] Email templates + notification preferences working
- [ ] OpenAPI spec auto-generated and accurate

### Frontend
- [ ] All 28 new pages implemented
- [ ] All 4 new contexts/hooks implemented
- [ ] `client.ts` extended with all new API namespaces
- [ ] `App.tsx` routes wired
- [ ] Video player with progress tracking working
- [ ] Certificate PDF generation + verification working
- [ ] i18n with at least 5 locales
- [ ] a11y audit passes (axe-core)
- [ ] Playwright e2e covers creator flow + learner flow + custom domain flow

### Infrastructure
- [ ] Caddyfile configured with on-demand TLS
- [ ] Dockerfile builds with new Caddyfile
- [ ] fly.toml has all new env vars
- [ ] S3 bucket + IAM configured
- [ ] Stripe Connect platform account configured
- [ ] MongoDB indexes created in production
- [ ] Monitoring + alerting configured
- [ ] Backup strategy in place (MongoDB Atlas automated backups)

### Business / Legal
- [ ] Terms of Service updated for marketplace (creator + learner)
- [ ] Privacy Policy updated for course data + GDPR
- [ ] Creator agreement includes commission rate + payout schedule + chargeback liability
- [ ] Cookie consent banner implemented
- [ ] Data retention policy documented
- [ ] Stripe Connect KYC requirements enforced

### Launch
- [ ] Beta test with 5-10 creators
- [ ] Load test passed (1000 concurrent storefront visitors)
- [ ] Pen test passed (no critical vulnerabilities)
- [ ] Runbook documented
- [ ] On-call rotation established
- [ ] Status page configured
