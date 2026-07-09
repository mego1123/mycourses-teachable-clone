# MyCourses — Teachable-Style Course-Selling SaaS

A multitenant course-selling platform built by forking [lastsaas](https://github.com/jonradoff/lastsaas) (Go + React SaaS boilerplate) and extending it with course-specific features: creator storefronts, Stripe Connect marketplace payments, Cloudflare Stream video delivery, and per-creator custom domains.

## What's Built

### Architecture
- **Backend**: Go 1.25 + gorilla/mux + PostgreSQL (via pgx + sqlc)
- **Frontend**: React 19 + Vite 7 + Tailwind CSS 4 + TanStack Query
- **Database**: PostgreSQL (migrated from MongoDB) — 29 migrations, 48 tables
- **Payments**: Stripe + Stripe Connect (destination charges with platform commission)
- **Video**: Cloudflare Stream (integration ready)
- **CDN/Domains**: Cloudflare for SaaS (custom domain provisioning with auto-SSL)
- **Email**: Resend (4 HTML templates)

### Features
- ✅ Creator studio: course CRUD, curriculum builder (sections + lessons), pricing, coupons
- ✅ Storefront: course browsing, detail pages with JSON-LD SEO, checkout with Stripe
- ✅ Learner experience: enrollment, video player with progress tracking, certificates
- ✅ Stripe Connect: creator onboarding, checkout with commission split, payout management
- ✅ Stripe webhooks: idempotent processing of 9 event types (checkout, refund, dispute, payout, account)
- ✅ Custom domains: Cloudflare for SaaS API integration with DNS verification + auto-SSL
- ✅ Certificate auto-issuance on course completion with public verification by token
- ✅ Drip content: section-level drip scheduling with middleware enforcement
- ✅ SEO: sitemap.xml, robots.txt, JSON-LD structured data
- ✅ Email: enrollment, certificate, payout, refund templates
- ✅ Marketplace: cross-tenant course discovery with Postgres full-text search

### Database Layer
- 29 versioned migrations (golang-migrate) — up + down for each
- 18 sqlc query files → 21 generated Go files (4,470+ LOC of type-safe DB code)
- Tables: users, tenants, plans, memberships, financial_transactions, courses, sections, lessons, media_assets, enrollments, course_progress, course_coupons, reviews, payouts, custom_domains, certificates, creator_profiles, processed_stripe_events + 30 more (auth, branding, webhooks, telemetry, audit, etc.)
- pg_cron cleanup jobs (replacing MongoDB TTL indexes)
- Postgres full-text search with pg_trgm for typo tolerance

### API Endpoints
- **Storefront** (public): list/get courses, marketplace, validate coupon, checkout
- **Creator** (auth): courses CRUD, sections, lessons, coupons, reviews, custom domains, Stripe Connect onboarding, payouts
- **Learner** (auth): enrollments, progress tracking, certificates
- **Public**: certificate verification by token, sitemap, robots.txt
- **Webhooks**: Stripe Connect webhook handler

### Testing
- 6 E2E tests using embedded-postgres (real Postgres 18.3 in-process)
- Tests cover: migrations apply, course CRUD, section/lesson creation, enrollment + progress, certificate verification, marketplace search, Stripe event idempotency

## Project Structure

```
mycourses/
├── backend/
│   ├── cmd/server/
│   │   ├── main.go              # Existing lastsaas server (MongoDB)
│   │   └── courses_routes.go    # NEW: Postgres course platform routes
│   ├── migrations/              # 29 SQL migration files (up + down)
│   ├── queries/                 # 18 sqlc query definition files
│   ├── internal/
│   │   ├── api/handlers/        # 39 handler files (10 new course handlers)
│   │   ├── cloudflare/          # Cloudflare for SaaS API client
│   │   ├── db/
│   │   │   ├── postgres.go      # Postgres connection pool + migration runner
│   │   │   └── gen/             # 21 sqlc-generated Go files
│   │   ├── email/               # 4 HTML email templates + service
│   │   ├── middleware/          # Custom domain + enrollment + drip middleware
│   │   └── stripe/              # Stripe Connect service
│   └── sqlc.yaml                # sqlc codegen config
├── frontend/
│   └── src/
│       ├── api/                 # courseApi.ts + clientBase.ts
│       ├── components/          # CreatorLayout.tsx
│       ├── hooks/               # useCourseProgress.ts
│       ├── pages/
│       │   ├── studio/          # 9 creator studio pages
│       │   ├── storefront/      # 4 public storefront pages
│       │   └── learn/           # 2 learner pages
│       └── types/course.ts      # TypeScript interfaces
├── docs/
│   ├── plans/                   # 4 planning documents (v1, v2, v2-addendum, v3)
│   ├── wiki-qa/                 # 32 codewiki.google Q&A JSON files
│   └── scripts/                 # Wiki query scripts + migration generation scripts
├── docker-compose.yml           # Postgres dev + test databases
└── .gitignore
```

## Getting Started

### Prerequisites
- Go 1.25+
- Node.js 20+
- PostgreSQL 16+ (or use docker-compose)
- sqlc v1.27+ (for code generation)
- golang-migrate v4+ (for migrations)

### Backend Setup
```bash
cd backend

# Start Postgres (or use your own)
docker compose -f ../docker-compose.yml up -d

# Apply migrations
make migrate-up

# Generate Go code from SQL (if queries changed)
make sqlc-gen

# Run the server
make run
```

### Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

### Environment Variables
```bash
# Database
DATABASE_URL=postgres://mycourses:mycourses_dev@localhost:5432/mycourses_dev?sslmode=disable

# Stripe
STRIPE_SECRET_KEY=sk_test_...
STRIPE_CONNECT_CLIENT_ID=ca_...
STRIPE_CONNECT_WEBHOOK_SECRET=whsec_...

# Cloudflare
CLOUDFLARE_API_TOKEN=...
CLOUDFLARE_ACCOUNT_ID=...
CLOUDFLARE_ZONE_ID=...

# Platform
PLATFORM_DOMAIN=mycourses.com
FRONTEND_URL=http://localhost:5173

# Email
RESEND_API_KEY=re_...
```

## Documentation

See `docs/plans/` for the complete build plan:
- `lastsaas_to_teachable_plan.md` — v1 concept mapping
- `lastsaas_to_teachable_plan_v2.md` — v2 production-ready plan (1,400 lines)
- `lastsaas_to_teachable_plan_v2_addendum.md` — MongoDB vs Postgres + CDN decisions
- `lastsaas_to_teachable_plan_v3.md` — v3 final stack decisions

See `docs/wiki-qa/` for 32 Q&A sessions with codewiki.google about the lastsaas codebase.

## Running Tests

```bash
cd backend

# E2E tests (uses embedded-postgres — no external DB needed)
go test ./internal/db/ -run TestE2E -v -timeout 300s
go test ./internal/api/handlers/ -run TestHandler -v -timeout 300s
```

## License

MIT (inherited from lastsaas)
