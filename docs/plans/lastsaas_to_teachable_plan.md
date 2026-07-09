# Turning LastSaaS into a Teachable-Style Multitenant Course-Selling SaaS

> Source: Synthesis of 8 Q&A sessions with the **codewiki.google** AI wiki for `jonradoff/lastsaas`.
> Raw wiki answers are preserved alongside this file in `/home/z/my-project/download/wiki_answers/`.

---

## 0. What LastSaaS already gives you for free

Before designing new things, here's what you do **not** have to build:

| Capability | Where it lives in the repo | How good is the fit? |
|---|---|---|
| Multi-tenant data isolation | `backend/internal/middleware/tenant.go`, `models/tenant.go`, `db/schema.go` | ✅ Excellent — this is the foundation. Each creator = one tenant. |
| Per-tenant white-label branding | `models/branding.go`, `api/handlers/branding.go`, `frontend/src/contexts/BrandingContext.tsx`, `components/BrandingThemeInjector.tsx` | ✅ Excellent — already does logos, colors, fonts, custom CSS, custom public pages, favicon, analytics injection. This IS the "branded storefront" engine. |
| Custom public pages per tenant | `BrandingHandler.GetPublicPage` / `ListPublicPages` / `CreatePage` / `UpdatePage` | ✅ Already supported — a creator can publish `/page/:slug` on their storefront today. |
| Auth (password, magic link, OAuth Google/GitHub/Microsoft, MFA/TOTP, sessions, JWT rotation) | `backend/internal/auth/*`, `api/handlers/auth.go`, `frontend/src/contexts/AuthContext.tsx` | ✅ Production-ready — reuse for both creators and learners. |
| RBAC + membership | `models/membership.go`, `middleware/rbac.go`, `middleware/auth.go` | ✅ Roles are `owner/admin/user` today; you extend, not replace. |
| Stripe billing (Checkout, Customer Portal, webhooks) | `backend/internal/stripe/stripe.go`, `api/handlers/billing.go`, `api/handlers/webhook.go` | ✅ Good for the **platform charging creators** subscription fees. Needs Stripe Connect extension for marketplace payouts. |
| Plans, entitlements, credits | `models/plan.go`, `models/credit_bundle.go`, `models/usage_event.go`, `api/handlers/plans.go`, `api/handlers/bundles.go`, `api/handlers/usage.go` | ✅ Use to gate creator features (max courses, custom-domain entitlement) and meter usage (video minutes, storage). |
| Admin console | `frontend/src/pages/admin/*` | ✅ Reuse to manage creators, payouts, marketplace settings. |
| Outgoing webhooks + event emitter | `backend/internal/events/emitter.go`, `internal/webhooks/dispatcher.go` | ✅ Use to emit `course.published`, `enrollment.created`, `payout.paid` events. |
| Resend email | `backend/internal/email/resend.go` | ✅ Use for enrollment confirmations, payout notifications. |

**Stack:** Go 1.25 + gorilla/mux backend, MongoDB (Atlas), React 19 + Vite 7 + Tailwind 4 frontend, JWT auth, deployed via multi-stage Dockerfile (Caddy included) on Fly.io.

**Frontend contexts already wiring per-tenant UI:** `AuthContext`, `TenantContext`, `BrandingContext`, `ThemeContext`. API client (`frontend/src/api/client.ts`) automatically injects `X-Tenant-ID` and bearer token.

---

## 1. Concept mapping — LastSaaS → Teachable

The single most important decision: **a LastSaaS `Tenant` IS a Teachable "creator school".** Everything else hangs off that.

| Teachable concept | LastSaaS concept | Action |
|---|---|---|
| Creator / School | `Tenant` | **Reuse as-is.** Add fields: `Slug`, `StripeConnectAccountID`, `CommissionRateBps`, `PayoutSchedule`. |
| Creator's team (instructors, admins) | `TenantMembership` with `owner/admin/user` | **Reuse as-is.** |
| Learner | `User` | **Reuse as-is.** A learner is just a `User` who has `Enrollment` rows. |
| Learner access to a school | `TenantMembership` with new role `RoleStudent` **OR** implicit via `Enrollment` | **Extend.** See §3 below. |
| Branding (logo, colors, fonts, custom pages) | `BrandingConfig` + `BrandingAsset` + `CustomPage` | **Reuse as-is.** One per tenant (school). |
| Custom domain (e.g. `academy.johndoe.com`) | New `CustomDomain` model | **Add.** See §5. |
| Plan / tier the creator pays the platform | `Plan` + `Subscription` | **Reuse + extend** with creator entitlements (`max_courses`, `custom_domain_enabled`, `video_storage_mb`, `commission_rate_bps`). |
| Course | New `Course` model | **Add.** Belongs to a `Tenant`. |
| Section / module | New `Section` model | **Add.** Belongs to `Course`. |
| Lesson (video/PDF/text/quiz) | New `Lesson` model + `MediaAsset` | **Add.** Belongs to `Section`. |
| Enrollment | New `Enrollment` model | **Add.** Links `User` (learner) + `Course` + `Tenant`. |
| Lesson progress | New `CourseProgress` model | **Add.** Per-lesson completion + video position. |
| Coupon | New `CourseCoupon` model | **Add.** Per-tenant, optionally per-course. |
| Review / rating | New `Review` model | **Add.** |
| Payout to creator | New `Payout` model + Stripe Connect `Transfer` | **Add.** |
| Sale / order | Extend `FinancialTransaction` | **Extend** with new types `TransactionCoursePurchase`, `TransactionPlatformCommission`, `TransactionCreatorPayout` and link to `CourseID`/`CreatorTenantID`/`LearnerUserID`. |
| Webhook events | Existing event emitter | **Extend** with course/enrollment/payout events. |

---

## 2. Backend — new models, handlers, and routes

### 2.1 New models (create under `backend/internal/models/`)

```go
// course.go
type Course struct {
    ID            primitive.ObjectID `bson:"_id"`
    TenantID      primitive.ObjectID `bson:"tenant_id"`       // creator's tenant
    Title         string             `bson:"title"`
    Description   string             `bson:"description"`
    PriceCents    int64              `bson:"price_cents"`
    Currency      string             `bson:"currency"`         // "usd"
    Status        string             `bson:"status"`           // draft|published|archived
    ThumbnailURL  string             `bson:"thumbnail_url"`
    IntroVideoURL string             `bson:"intro_video_url"`
    Slug          string             `bson:"slug"`             // unique within tenant
    CategoryID    *primitive.ObjectID `bson:"category_id,omitempty"`
    DripSchedule  *DripConfig        `bson:"drip_schedule,omitempty"`
    CreatedAt     time.Time          `bson:"created_at"`
    UpdatedAt     time.Time          `bson:"updated_at"`
}

// section.go
type Section struct {
    ID          primitive.ObjectID `bson:"_id"`
    CourseID    primitive.ObjectID `bson:"course_id"`
    Title       string             `bson:"title"`
    Description string             `bson:"description"`
    Order       int                `bson:"order"`
}

// lesson.go
type Lesson struct {
    ID          primitive.ObjectID `bson:"_id"`
    SectionID   primitive.ObjectID `bson:"section_id"`
    CourseID    primitive.ObjectID `bson:"course_id"`
    Title       string             `bson:"title"`
    Type        string             `bson:"type"`              // video|text|pdf|quiz
    Content     string             `bson:"content"`           // HTML/markdown for text
    MediaAssetID *primitive.ObjectID `bson:"media_asset_id,omitempty"`
    Order       int                `bson:"order"`
    IsPreview   bool               `bson:"is_preview"`        // free preview lesson
    DurationSec int                `bson:"duration_sec"`
}

// media_asset.go — model after BrandingAsset
type MediaAsset struct {
    ID         primitive.ObjectID `bson:"_id"`
    TenantID   primitive.ObjectID `bson:"tenant_id"`
    Kind       string             `bson:"kind"`               // video|pdf|image
    StorageURL string             `bson:"storage_url"`        // S3/CDN URL or signed URL
    SizeBytes  int64              `bson:"size_bytes"`
    MimeType   string             `bson:"mime_type"`
    CreatedAt  time.Time          `bson:"created_at"`
}

// enrollment.go
type Enrollment struct {
    ID             primitive.ObjectID `bson:"_id"`
    CourseID       primitive.ObjectID `bson:"course_id"`
    TenantID       primitive.ObjectID `bson:"tenant_id"`      // creator
    UserID         primitive.ObjectID `bson:"user_id"`        // learner
    Status         string             `bson:"status"`         // active|completed|refunded
    PricePaidCents int64              `bson:"price_paid_cents"`
    CouponID       *primitive.ObjectID `bson:"coupon_id,omitempty"`
    EnrollmentDate time.Time          `bson:"enrollment_date"`
    CompletionDate *time.Time         `bson:"completion_date,omitempty"`
    ExpiresAt      *time.Time         `bson:"expires_at,omitempty"`
}

// course_progress.go
type CourseProgress struct {
    ID            primitive.ObjectID `bson:"_id"`
    EnrollmentID  primitive.ObjectID `bson:"enrollment_id"`
    LessonID      primitive.ObjectID `bson:"lesson_id"`
    UserID        primitive.ObjectID `bson:"user_id"`
    Completed     bool               `bson:"completed"`
    CompletedAt   *time.Time         `bson:"completed_at,omitempty"`
    LastViewedAt  time.Time          `bson:"last_viewed_at"`
    VideoPositionSec int             `bson:"video_position_sec"`
}

// course_coupon.go — distinct from Stripe promotions
type CourseCoupon struct {
    ID            primitive.ObjectID `bson:"_id"`
    TenantID      primitive.ObjectID `bson:"tenant_id"`
    Code          string             `bson:"code"`
    DiscountType  string             `bson:"discount_type"` // percent|fixed
    DiscountValue int64              `bson:"discount_value"`
    CourseID      *primitive.ObjectID `bson:"course_id,omitempty"` // nil = storewide
    ExpiresAt     *time.Time          `bson:"expires_at,omitempty"`
    UsageLimit    int                 `bson:"usage_limit"`
    UsedCount     int                 `bson:"used_count"`
}

// review.go
type Review struct {
    ID        primitive.ObjectID `bson:"_id"`
    CourseID  primitive.ObjectID `bson:"course_id"`
    UserID    primitive.ObjectID `bson:"user_id"`
    Rating    int                `bson:"rating"`           // 1–5
    Comment   string             `bson:"comment"`
    CreatedAt time.Time          `bson:"created_at"`
}

// payout.go
type Payout struct {
    ID                primitive.ObjectID `bson:"_id"`
    TenantID          primitive.ObjectID `bson:"tenant_id"` // creator
    StripeTransferID  string             `bson:"stripe_transfer_id"`
    AmountCents       int64              `bson:"amount_cents"`
    Currency          string             `bson:"currency"`
    Status            string             `bson:"status"`    // pending|paid|failed
    InitiatedAt       time.Time          `bson:"initiated_at"`
    CompletedAt       *time.Time         `bson:"completed_at,omitempty"`
}

// custom_domain.go
type CustomDomain struct {
    ID        primitive.ObjectID `bson:"_id"`
    Domain    string             `bson:"domain"`   // unique index
    TenantID  primitive.ObjectID `bson:"tenant_id"`
    Status    string             `bson:"status"`   // pending|active|failed
    DNSVerified bool             `bson:"dns_verified"`
    SSLStatus string             `bson:"ssl_status"` // pending|active|failed
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

### 2.2 New handlers (under `backend/internal/api/handlers/`)

| File | Endpoints |
|---|---|
| `course.go` | `POST /api/courses` (creator), `GET /api/courses/:id`, `PUT /api/courses/:id`, `DELETE /api/courses/:id`, `GET /api/storefront/courses` (public, scoped by tenant from custom-domain middleware), `GET /api/storefront/courses/:slug` |
| `section.go` | CRUD under `/api/courses/:courseId/sections` |
| `lesson.go` | CRUD under `/api/sections/:sectionId/lessons`, `POST /api/lessons/:id/progress` |
| `enrollment.go` | `POST /api/enrollments` (creates Stripe Checkout → on success creates enrollment), `GET /api/learner/enrollments`, `GET /api/creator/enrollments` |
| `course_progress.go` | `POST /api/progress/:lessonId` (upsert), `GET /api/progress/course/:courseId` |
| `coupon.go` | CRUD under `/api/coupons`, `POST /api/storefront/coupons/validate` |
| `review.go` | `POST /api/reviews`, `GET /api/storefront/courses/:id/reviews` |
| `payout.go` | `GET /api/creator/payouts`, `POST /api/creator/payouts/request` |
| `custom_domain.go` | `POST /api/custom-domains`, `GET /api/custom-domains`, `GET /api/custom-domains/:id/status`, `DELETE /api/custom-domains/:id` |
| `media.go` | `POST /api/media/upload` (presigned URL or multipart), `GET /api/media/:id` (serve or redirect to CDN) |

### 2.3 Wire into the router (`backend/cmd/server/main.go`)

The existing pattern is: register sub-routers under `/api/...` and apply middleware chains like:

```go
// Public storefront (no auth, tenant resolved by CustomDomainTenantMiddleware or X-Tenant-ID)
storefrontRouter := apiRouter.PathPrefix("/storefront").Subrouter()
storefrontRouter.Use(middleware.NewTenantMiddleware)
storefrontRouter.HandleFunc("/courses", storefrontHandler.ListCourses).Methods("GET")
storefrontRouter.HandleFunc("/courses/{slug}", storefrontHandler.GetCourse).Methods("GET")
storefrontRouter.HandleFunc("/courses/{id}/reviews", reviewHandler.ListPublic).Methods("GET")
storefrontRouter.HandleFunc("/coupons/validate", couponHandler.Validate).Methods("POST")
storefrontRouter.HandleFunc("/checkout/{courseSlug}", enrollmentHandler.CreateCheckout).Methods("POST")
storefrontRouter.HandleFunc("/branding", brandingHandler.GetBranding).Methods("GET")   // already exists

// Creator-only (auth + role)
creatorRouter := apiRouter.PathPrefix("/creator").Subrouter()
creatorRouter.Use(middleware.NewAuthMiddleware)
creatorRouter.Use(middleware.NewTenantMiddleware)
creatorRouter.Use(middleware.RequireRole(models.RoleOwner, models.RoleAdmin))
creatorRouter.HandleFunc("/courses", courseHandler.Create).Methods("POST")
creatorRouter.HandleFunc("/courses/{id}", courseHandler.Update).Methods("PUT")
creatorRouter.HandleFunc("/payouts", payoutHandler.List).Methods("GET")
creatorRouter.HandleFunc("/custom-domains", customDomainHandler.Create).Methods("POST")

// Learner-only (auth + has-enrollment check)
learnerRouter := apiRouter.PathPrefix("/learner").Subrouter()
learnerRouter.Use(middleware.NewAuthMiddleware)
learnerRouter.HandleFunc("/enrollments", enrollmentHandler.ListMine).Methods("GET")
learnerRouter.HandleFunc("/progress/{lessonId}", progressHandler.Update).Methods("POST")
```

### 2.4 Stripe Connect integration

**Reusing the existing Stripe wrapper** at `backend/internal/stripe/stripe.go`, add:

1. **Onboarding** — when a creator upgrades to a paid plan, call Stripe's `AccountLink` API to start Connect onboarding (Express account). Store the resulting `acct_...` ID in `Tenant.StripeConnectAccountID`.
2. **Course purchase** — in `enrollmentHandler.CreateCheckout`, build a `Checkout.Session` with `payment_intent_data.application_fee_amount` set to `price * commissionRateBps / 10000` and `transfer_data.destination = creatorStripeConnectAccountID`. Stripe splits the funds automatically.
3. **Payouts** — either let Stripe auto-payout to the creator's bank (default Express behavior) or initiate manual `Transfer` objects on a schedule and record them in the `Payout` collection.
4. **Webhooks** — extend `api/handlers/webhook.go` (today it handles `invoice.paid`, `customer.subscription.*`, `checkout.session.completed`). Add Connect handlers for:
   - `account.updated` → set `Tenant.StripeConnectStatus`
   - `transfer.created` / `payout.paid` / `payout.failed` → upsert `Payout` records
   - `checkout.session.completed` (with `transfer_data`) → create `Enrollment` + `FinancialTransaction` (type `TransactionCoursePurchase`) + emit `enrollment.created` event

---

## 3. Roles and access model

Today: `models/membership.go` defines `RoleOwner`, `RoleAdmin`, `RoleUser`. Extend it:

```go
const (
    RoleOwner   MemberRole = "owner"
    RoleAdmin   MemberRole = "admin"
    RoleCreator MemberRole = "creator"  // instructor with course-publish rights
    RoleStudent MemberRole = "student"  // learner (optional — see below)
    RoleUser    MemberRole = "user"
)
```

**Key design decision: where do learners live?**

The wiki flagged this ambiguity. Two options:

- **Option A (recommended): Learners are not members of the creator's tenant.** They are global `User`s. Their relationship to a creator is purely through `Enrollment.TenantID`. This avoids N memberships per learner (one per creator they bought from) and matches Teachable/Thinkific behavior, where learners can buy from many schools with one account.
- **Option B: Learners get `RoleStudent` membership in each creator's tenant.** Heavier, but enables per-school learner dashboards and lets the existing `RequireRole` middleware protect learner endpoints with no new code.

**Pick A.** Add a new `RequireEnrollment(courseID)` middleware that checks the `enrollments` collection instead of `tenant_memberships`. This keeps creator-tenant membership clean (only instructors/staff) and lets the platform-scale to learners buying from many creators.

For platform staff auditing a creator's school, keep using `RequireRole(RoleOwner, RoleAdmin)` on the existing `Tenant` membership.

---

## 4. Frontend — adding the storefront + learner app + creator studio

### 4.1 New page tree under `frontend/src/pages/`

```
pages/
├── public/
│   ├── LandingPage.tsx              (exists — platform homepage)
│   ├── storefront/
│   │   ├── StorefrontHomePage.tsx   (creator's branded homepage at custom domain)
│   │   ├── CourseListPage.tsx       (all published courses by creator)
│   │   ├── CourseDetailPage.tsx     (course landing page with curriculum, price, reviews)
│   │   ├── CheckoutPage.tsx         (Stripe Checkout, coupon entry)
│   │   ├── CustomPage.tsx           (exists — reuse for /page/:slug)
├── app/
│   ├── DashboardPage.tsx            (exists)
│   ├── creator/                     (NEW — creator studio)
│   │   ├── CreatorDashboardPage.tsx
│   │   ├── CoursesPage.tsx          (list/create courses)
│   │   ├── CourseEditorPage.tsx     (sections, lessons, media upload)
│   │   ├── SalesPage.tsx
│   │   ├── PayoutsPage.tsx
│   │   ├── CouponsPage.tsx
│   │   ├── CustomDomainPage.tsx
│   ├── learner/                     (NEW — learner dashboard)
│   │   ├── MyCoursesPage.tsx
│   │   ├── CoursePlayerPage.tsx     (video player, lesson nav, progress tracking)
│   │   ├── BillingHistoryPage.tsx
│   ├── settings/                    (exists)
├── admin/                           (exists — platform admin)
```

### 4.2 Wire storefront branding

The existing `BrandingContext` already loads branding for the active tenant. The change for custom domains: instead of resolving the tenant from the user's JWT or `X-Tenant-ID` header, resolve it from the **Host header via the new `CustomDomainTenantMiddleware`**.

- The frontend doesn't need to know which tenant it's serving. The backend middleware injects the tenant context, and `GET /api/branding` returns the right `BrandingConfig` automatically.
- `BrandingThemeInjector` already injects CSS variables, fonts, logos, favicons, and custom HTML — no changes needed.
- For platform routes (e.g. `app.teachable-clone.com/admin`), branding resolves to the platform's own tenant.

### 4.3 New contexts/hooks

- `CreatorContext` — wraps `TenantContext` for pages under `/app/creator/*`; exposes `isCreator()`, `creatorTenantID`.
- `EnrollmentContext` — for `/app/learner/*`; exposes the learner's active enrollments.
- `useCourseProgress(courseId)` — hook that subscribes to the player's `timeupdate` events and `POST`s to `/api/progress/:lessonId` every 15s.

### 4.4 API client additions

Extend `frontend/src/api/client.ts` (axios instance) with namespace methods:

```ts
client.courses = {
  listStorefront: () => client.get('/api/storefront/courses'),
  get: (slug) => client.get(`/api/storefront/courses/${slug}`),
  create: (data) => client.post('/api/creator/courses', data),
  update: (id, data) => client.put(`/api/creator/courses/${id}`, data),
  // ...
}
client.enrollments = { /* ... */ }
client.progress = { update: (lessonId, payload) => client.post(`/api/learner/progress/${lessonId}`, payload) }
client.payouts = { /* ... */ }
client.customDomains = { /* ... */ }
```

The existing interceptor already attaches `X-Tenant-ID` and bearer token — no change.

### 4.5 Routing in `frontend/src/App.tsx`

Add the new route groups. Public storefront routes live at the **root** of the SPA so a creator's custom domain serves them by default:

```tsx
<Routes>
  {/* Storefront (served on custom domain OR platform domain with tenant resolved by middleware) */}
  <Route path="/" element={<StorefrontHomePage/>} />
  <Route path="/courses" element={<CourseListPage/>} />
  <Route path="/courses/:slug" element={<CourseDetailPage/>} />
  <Route path="/checkout/:courseSlug" element={<CheckoutPage/>} />
  <Route path="/page/:slug" element={<CustomPage/>} /> {/* exists */}

  {/* Auth */}
  <Route path="/auth/*" element={<AuthRoutes/>} />

  {/* Learner app */}
  <Route path="/learn/*" element={<ProtectedRoute><LearnerLayout/></ProtectedRoute>}>
    <Route path="my-courses" element={<MyCoursesPage/>} />
    <Route path="course/:courseId/:lessonId" element={<CoursePlayerPage/>} />
  </Route>

  {/* Creator studio */}
  <Route path="/studio/*" element={<ProtectedRoute><CreatorLayout/></ProtectedRoute>}>
    <Route path="dashboard" element={<CreatorDashboardPage/>} />
    <Route path="courses" element={<CoursesPage/>} />
    <Route path="courses/:id/edit" element={<CourseEditorPage/>} />
    <Route path="sales" element={<SalesPage/>} />
    <Route path="payouts" element={<PayoutsPage/>} />
    <Route path="coupons" element={<CouponsPage/>} />
    <Route path="domain" element={<CustomDomainPage/>} />
  </Route>

  {/* Platform admin (existing) */}
  <Route path="/admin/*" element={<AdminRoute><AdminLayout/></AdminRoute>} />
</Routes>
```

### 4.6 Video player & progress tracking

For the `CoursePlayerPage`, use a player library (e.g. `react-player`, `video.js`, or commercial Mux/Cloudflare Stream). On `onProgress`:

```ts
useEffect(() => {
  const id = setInterval(() => {
    api.progress.update(currentLessonId, {
      video_position_sec: playerRef.current.getCurrentTime(),
      completed: playerRef.current.getCurrentTime() / duration > 0.9,
    })
  }, 15_000)
  return () => clearInterval(id)
}, [currentLessonId])
```

The backend `progressHandler.Update` upserts into `course_progress` and — if `completed=true` — emits a `lesson.completed` event. The existing event emitter + webhook dispatcher can notify the creator's external integrations.

---

## 5. Custom domain support — full request lifecycle

This is the trickiest new infrastructure. Here's the end-to-end flow:

### 5.1 Provisioning (creator adds a domain)

1. Creator opens `/studio/domain` and enters `academy.johndoe.com`.
2. Frontend calls `POST /api/custom-domains` → backend creates a `CustomDomain` row with `status=pending`, `dns_verified=false`.
3. Backend returns DNS instructions: "Add a CNAME record: `academy.johndoe.com` → `custom.teachable-clone.com`".
4. Frontend polls `GET /api/custom-domains/:id/status`. A background worker (or on-demand check) does a DNS lookup; when the CNAME matches, it sets `dns_verified=true`.
5. **SSL/TLS provisioning** depends on hosting:
   - **Fly.io** (what lastsaas targets today): Fly auto-issues certs for any custom hostname added via the Fly API. Add a small Go service that calls `fly machines` API to register the custom hostname, then polls until the cert is issued.
   - **Caddy** (already in the repo's Dockerfile): Caddy's `on-demand TLS` issues Let's Encrypt certs for any Host header that passes an `ask` endpoint. Point Caddy's `ask` to a new backend route `GET /internal/caddy/allowed?domain=...` that returns 200 if the domain exists in `custom_domains` with `status=active`. **Recommended** — minimal infra.
6. Once DNS + SSL are good, set `CustomDomain.Status = "active"`.

### 5.2 Runtime request lifecycle

```
Browser → academy.johndoe.com
   ↓
Caddy (terminates TLS via on-demand cert, reverse-proxies to backend)
   ↓
Backend HTTP server (cmd/server/main.go)
   ↓
NEW middleware: CustomDomainTenantMiddleware    ← inserted BEFORE existing NewTenantMiddleware
   - reads Host header: "academy.johndoe.com"
   - looks up CustomDomain collection → finds TenantID = <creator's tenant>
   - if found: sets ctx tenant + sets X-Tenant-ID header on the inner request
   - if not found (platform domain like "teachable-clone.com"): falls through to existing behavior
   ↓
Existing NewTenantMiddleware     ← uses X-Tenant-ID header (now set by the above)
   ↓
Existing AuthMiddleware          ← JWT to User context (skipped for public storefront routes)
   ↓
Existing RBACMiddleware          ← RequireRole for /studio/* routes
   ↓
Handler (course, branding, checkout, etc.)
   ↓
Response: branded HTML / JSON scoped to the creator's tenant
```

### 5.3 Files to touch for custom domains

| File | Change |
|---|---|
| `backend/internal/models/custom_domain.go` | **NEW** — `CustomDomain` struct. |
| `backend/internal/db/schema.go` | Add JSON schema + unique index on `domain` field. |
| `backend/internal/middleware/custom_domain.go` | **NEW** — `CustomDomainTenantMiddleware`. |
| `backend/cmd/server/main.go` | Insert `CustomDomainTenantMiddleware` at the top of the chain (before `NewTenantMiddleware`). |
| `backend/internal/api/handlers/custom_domain.go` | **NEW** — CRUD + status check + DNS verify. |
| `Dockerfile` / `Caddyfile` | Configure Caddy `on-demand_tls` with `ask` endpoint pointing to `/internal/caddy/allowed`. |
| `frontend/src/pages/app/creator/CustomDomainPage.tsx` | **NEW** — UI to add/verify domains. |
| `backend/internal/api/handlers/caddy.go` | **NEW** — `GET /internal/caddy/allowed?domain=...` returning 200/404. |

### 5.4 SSL/TLS approach with Caddy (already in the repo)

Caddy supports `on_demand_tls` mode where it asks a permission endpoint before issuing a cert:

```
# Caddyfile (sketch)
{
  on_demand_tls {
    ask http://localhost:8080/internal/caddy/allowed
  }
}

https:// {
  tls {
    on_demand
  }
  reverse_proxy localhost:8080
}
```

When a request hits an unknown host, Caddy calls the `ask` endpoint. If your backend returns 200 (because that domain is in `custom_domains` with `status=active`), Caddy issues a Let's Encrypt cert on the fly and serves the request. Zero per-creator cert management code on your side.

---

## 6. Media & course content storage

LastSaaS has no media storage today (branding assets are stored in MongoDB GridFS via `BrandingHandler.ServeAsset`). For courses you need a real blob store.

### 6.1 Recommended: S3 (or R2 / B2) + CloudFront/CDN

- Backend hands out **presigned PUT URLs** for upload and **presigned GET URLs** (or public CDN URLs) for playback.
- Add a `MediaAsset` model (§2.1) that records `StorageURL`, `SizeBytes`, `MimeType` per asset, scoped to a `TenantID`.
- The existing `BrandingHandler.UploadAsset` already uses a similar pattern — model after it.

### 6.2 Access gating

- For free preview lessons (`IsPreview=true`): public CDN URL, no auth.
- For paid lessons: short-lived signed URL (TTL ~10 min) issued only after the request passes `RequireEnrollment` middleware.
- For video streaming specifically, consider CloudFront signed URLs/cookies or signed JWT-validated URLs to prevent content leaking.

### 6.3 Usage metering (creators pay for storage/bandwidth)

Reuse the existing `UsageEvent` + `CreditBundle` system (`models/usage_event.go`, `api/handlers/usage.go`). The `UsageHandler.RecordUsage` already does atomic credit deduction. Add new `UsageEvent.Type` values:

- `video_storage_mb_hours` — meter storage
- `video_bandwidth_gb` — meter playback bandwidth
- `course_published_count` — meter number of active courses

These flow through the existing entitlements system in `Plan.Entitlements` so you can gate the creator's plan tier ("Pro plan: 100GB video storage, 1TB bandwidth/mo").

---

## 7. Stripe Connect marketplace flow — concrete code sketch

```go
// internal/stripe/connect.go (new file)

// OnboardCreator creates an Express account and returns an AccountLink URL
// for the creator to complete onboarding.
func (s *StripeService) OnboardCreator(tenantID primitive.ObjectID, returnURL, refreshURL string) (string, error) {
    acct, err := account.New(&stripe.AccountParams{
        Type:    stripe.String(string(stripe.AccountTypeExpress)),
        Country: stripe.String("US"),
        Metadata: map[string]string{
            "tenant_id": tenantID.Hex(),
        },
    })
    if err != nil {
        return "", err
    }
    // persist acct.ID to Tenant.StripeConnectAccountID
    link, err := accountlink.New(&stripe.AccountLinkParams{
        Account:    stripe.String(acct.ID),
        RefreshURL: stripe.String(refreshURL),
        ReturnURL:  stripe.String(returnURL),
        Type:       stripe.String("account_onboarding"),
    })
    if err != nil {
        return "", err
    }
    return link.URL, nil
}

// CreateCourseCheckout builds a Checkout Session that charges the learner,
// takes the platform commission, and routes the remainder to the creator.
func (s *StripeService) CreateCourseCheckout(
    course *models.Course,
    creatorStripeAccountID string,
    commissionBps int,        // e.g. 1000 = 10%
    learnerStripeCustomerID string,
    couponID *string,
    successURL, cancelURL string,
) (*stripe.CheckoutSession, error) {
    params := &stripe.CheckoutSessionParams{
        Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
        LineItems: []*stripe.CheckoutSessionLineItemParams{{
            Quantity: stripe.Int64(1),
            PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                Currency: stripe.String(course.Currency),
                ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                    Name: stripe.String(course.Title),
                },
                UnitAmount: stripe.Int64(course.PriceCents),
            },
        }},
        PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
            ApplicationFeeAmount: stripe.Int64(course.PriceCents * int64(commissionBps) / 10000),
            TransferData: &stripe.CheckoutSessionPaymentIntentDataTransferDataParams{
                Destination: stripe.String(creatorStripeAccountID),
            },
        },
        SuccessURL: stripe.String(successURL),
        CancelURL:  stripe.String(cancelURL),
    }
    if learnerStripeCustomerID != "" {
        params.Customer = stripe.String(learnerStripeCustomerID)
    }
    return session.New(params)
}
```

The webhook handler (`api/handlers/webhook.go`) then receives `checkout.session.completed`, looks up the course + learner from the session metadata, creates the `Enrollment`, records a `FinancialTransaction` of type `TransactionCoursePurchase`, and emits `enrollment.created` for the webhook dispatcher.

---

## 8. Phased implementation roadmap

### Phase 1 — Foundations (1–2 weeks)
- Fork lastsaas. Add `RoleCreator`, `RoleStudent` constants.
- Add `Course`, `Section`, `Lesson`, `MediaAsset` models + DB schema + indexes.
- Add creator-side CRUD handlers + `/studio/courses` frontend.
- Wire up S3 for media uploads (presigned URLs).

### Phase 2 — Storefront + learner experience (2–3 weeks)
- Add public storefront routes + pages under `pages/public/storefront/`.
- Add `Enrollment` + `CourseProgress` models + handlers.
- Add `RequireEnrollment` middleware.
- Build the learner dashboard and course player with progress tracking.
- Add `Review` model + public review submission for enrolled learners.

### Phase 3 — Marketplace billing (2 weeks)
- Add Stripe Connect onboarding for creators.
- Build the course checkout flow with `transfer_data` + `application_fee_amount`.
- Extend `webhook.go` with `account.updated`, `transfer.created`, `payout.paid` handlers.
- Add `Payout` model + creator payouts dashboard.
- Extend `FinancialTransaction` types.

### Phase 4 — Custom domains + per-school branding (2 weeks)
- Add `CustomDomain` model + `CustomDomainTenantMiddleware`.
- Configure Caddy `on-demand_tls` with `ask` endpoint.
- Add the creator's "Custom Domain" settings page with DNS instructions + status polling.
- Test the full lifecycle: creator adds domain → DNS verifies → Caddy issues cert → branded storefront serves correctly.

### Phase 5 — Coupons, drip, taxes (2 weeks)
- `CourseCoupon` CRUD + checkout-time validation.
- `DripConfig` on `Course` / `Section` — gate lesson visibility by `EnrollmentDate + dripDelay`.
- Stripe Tax integration for marketplace transactions (extend existing Stripe Tax usage from `internal/stripe`).

### Phase 6 — Polish & launch
- Email notifications (reuse `email.ResendService`) for enrollments, payouts, certificate issuance.
- Certificates model (optional — `Certificate` per completed enrollment).
- Analytics dashboard for creators (reuse `telemetry` + `events`).
- Public marketplace discovery page (platform homepage listing all creator schools).

**Estimated total: 10–13 weeks** for a focused team of 1–2 engineers, since LastSaaS eliminates ~60% of the boilerplate work (auth, billing, RBAC, multi-tenancy, branding, admin console).

---

## 9. What to read in the lastsaas codebase first

Before writing any new code, read these files in order — they are the seam points where every new feature plugs in:

1. `backend/internal/models/tenant.go` — the central abstraction. Your creator = a tenant.
2. `backend/internal/models/membership.go` — where you'll add `RoleCreator`/`RoleStudent`.
3. `backend/internal/middleware/tenant.go` + `middleware/auth.go` + `middleware/rbac.go` — the middleware chain you'll extend.
4. `backend/cmd/server/main.go` — where middleware and routes are wired. This is the integration map.
5. `backend/internal/api/handlers/branding.go` — the most feature-rich existing handler; model new CRUD handlers after it (validation, errors, asset serving, public vs. admin endpoints).
6. `backend/internal/api/handlers/billing.go` + `internal/stripe/stripe.go` — where Stripe Connect hooks in.
7. `frontend/src/contexts/BrandingContext.tsx` + `components/BrandingThemeInjector.tsx` — the storefront branding engine you'll reuse untouched.
8. `frontend/src/api/client.ts` — the axios instance you'll extend with new namespaces.
9. `frontend/src/App.tsx` — where you'll add the new route trees.
10. `Dockerfile` + `Caddyfile` (if present) — where on-demand TLS for custom domains gets configured.

---

## 10. Raw wiki answers

Each question's full answer from codewiki.google is preserved as JSON at:

- `/home/z/my-project/download/wiki_answers/01_foundations.json` — Architecture & tech stack
- `/home/z/my-project/download/wiki_answers/02_multitenancy_branding.json` — Multi-tenancy + branding deep-dive
- `/home/z/my-project/download/wiki_answers/03_billing_credits_stripe.json` — Billing, plans, credits, Stripe
- `/home/z/my-project/download/wiki_answers/04_auth_rbac.json` — Auth, users, RBAC, impersonation
- `/home/z/my-project/download/wiki_answers/05_frontend_ext.json` — Frontend routes, contexts, branding injection
- `/home/z/my-project/download/wiki_answers/06_teachable_mapping.json` — Concrete Teachable transformation mapping
- `/home/z/my-project/download/wiki_answers/07_custom_domains.json` — Custom domain request lifecycle
- `/home/z/my-project/download/wiki_answers/08_content_storage.json` — Media/content storage patterns

The query script is also persisted at `/home/z/my-project/scripts/ask_one.mjs` so you can ask follow-up questions later by running:

```bash
node /home/z/my-project/scripts/ask_one.mjs 09_followup "Your next question here"
```
