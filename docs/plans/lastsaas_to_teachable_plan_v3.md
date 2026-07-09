# Final Stack Decisions (v3) — Supersedes Prior Addendums

> This document updates and partially supersedes `lastsaas_to_teachable_plan_v2_addendum.md`.
> Source wiki Q&As: #27, #28, #29, #30, #31 (migration approach), #32 (Postgres schema).

---

## Decision 1: Database — Migrate to PostgreSQL NOW (before writing course code)

### You're right — migrating now is cheaper than migrating later

This is one of the most well-established principles in software engineering: **refactor cost grows super-linearly with codebase size.**

| Scenario | Codebase size at migration | Migration effort |
|---|---|---|
| Migrate NOW (before adding course code) | lastsaas only: ~30 models, ~50 handlers | **3–4 weeks** |
| Migrate LATER (after course code) | lastsaas + course code: ~45 models, ~75 handlers | **6–8 weeks** + risk of disrupting paying customers |

Migrating now also means **all your new course code is written for Postgres from day one** — no "Mongo first, refactor later" technical debt.

### What the wiki confirmed about the migration

The DB access in lastsaas is **centralized and clean** — this is the key enabler:

```go
// internal/db/mongodb.go — current structure
type MongoDB struct {
    Client   *mongo.Client
    Database *mongo.Database
}

// Methods return *mongo.Collection — handlers call operations directly
func (db *MongoDB) Users() *mongo.Collection { return db.Database.Collection("users") }
func (db *MongoDB) Tenants() *mongo.Collection { ... }
// ... ~20 such methods

// Handler usage pattern (no repository layer)
user, err := h.db.Users().FindOne(ctx, bson.M{"email": email}).Decode(&user)
```

**No repository abstraction exists today.** Handlers call MongoDB operations directly. This means migration touches every handler, but the pattern is mechanical.

### The migration plan (concrete, file-by-file)

#### Week 1: Foundation

1. **Add Postgres driver + tooling to `go.mod`:**
   ```go
   github.com/jackc/pgx/v5              // Postgres driver (better than database/sql)
   github.com/sqlc-dev/sqlc              // Generate type-safe Go from SQL
   github.com/golang-migrate/migrate/v4  // Schema migrations
   github.com/google/uuid                // Replace primitive.ObjectID
   ```

2. **Replace `internal/db/mongodb.go` with `internal/db/postgres.go`:**
   ```go
   type DB struct {
       Pool *pgxpool.Pool
       Queries *sqlc.Queries  // generated from .sql files
   }
   
   // Repository-style methods (one per entity)
   func (db *DB) Users() *sqlc.Queries { return db.Queries }
   // ... or use the generated Queries directly
   ```

3. **Set up `sqlc` config (`backend/sqlc.yaml`):**
   ```yaml
   version: "2"
   sql:
     - engine: "postgresql"
       queries: "queries/"
       schema: "migrations/"
       gen:
         go:
           package: "db"
           out: "internal/db/gen"
           sql_package: "pgx/v5"
           emit_json_tags: true
           emit_pointers_for_null_types: true
   ```

4. **Create migration files** (`backend/migrations/`):
   - `000001_create_users.up.sql` / `.down.sql`
   - `000002_create_tenants.up.sql` / `.down.sql`
   - `000003_create_memberships.up.sql` / `.down.sql`
   - ... one per existing collection (the wiki gave us 30+ collection names)
   - Each migration is idempotent and reversible

#### Week 2: Schema + models

5. **Write the full Postgres schema** (wiki Q32 gave us the templates). Key decisions:

   ```sql
   -- Replace primitive.ObjectID with UUID everywhere
   CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- for gen_random_uuid()
   
   -- Users
   CREATE TABLE users (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       email TEXT UNIQUE NOT NULL,
       name TEXT,
       password_hash TEXT,
       mfa_enabled BOOLEAN DEFAULT FALSE,
       mfa_secret TEXT,
       auth_methods TEXT[] DEFAULT '{}',  -- embedded array → Postgres array
       is_active BOOLEAN DEFAULT TRUE,
       is_admin BOOLEAN DEFAULT FALSE,
       created_at TIMESTAMPTZ DEFAULT NOW(),
       updated_at TIMESTAMPTZ DEFAULT NOW()
   );
   
   -- Tenants (creator schools)
   CREATE TABLE tenants (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name TEXT NOT NULL,
       slug TEXT UNIQUE NOT NULL,           -- new: for subdomain routing
       plan_id UUID REFERENCES plans(id),
       stripe_customer_id TEXT,
       stripe_connect_account_id TEXT,      -- new: for Stripe Connect
       commission_rate_bps INT DEFAULT 1000, -- new: 10% default
       billing_status TEXT DEFAULT 'active',
       is_root BOOLEAN DEFAULT FALSE,
       created_at TIMESTAMPTZ DEFAULT NOW(),
       updated_at TIMESTAMPTZ DEFAULT NOW()
   );
   
   -- Memberships
   CREATE TABLE tenant_memberships (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
       role TEXT NOT NULL CHECK (role IN ('owner','admin','creator','student','user')),
       created_at TIMESTAMPTZ DEFAULT NOW(),
       UNIQUE(tenant_id, user_id)
   );
   
   -- Branding config (JSONB for flexible nested structure)
   CREATE TABLE branding_configs (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
       config JSONB NOT NULL DEFAULT '{}',  -- NavItems, colors, fonts, custom HTML all go here
       created_at TIMESTAMPTZ DEFAULT NOW(),
       updated_at TIMESTAMPTZ DEFAULT NOW()
   );
   CREATE INDEX idx_branding_tenant ON branding_configs(tenant_id);
   CREATE INDEX idx_branding_config ON branding_configs USING GIN (config);
   
   -- Financial transactions (ACID critical!)
   CREATE TABLE financial_transactions (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       tenant_id UUID REFERENCES tenants(id),
       user_id UUID REFERENCES users(id),
       type TEXT NOT NULL,  -- 'subscription','credit_purchase','course_purchase','refund','payout','commission','dispute_withdrawal','dispute_reinstatement','transfer_reversal'
       amount_cents BIGINT NOT NULL,  -- negative for debits
       currency TEXT DEFAULT 'usd',
       stripe_charge_id TEXT,
       stripe_transfer_id TEXT,
       stripe_session_id TEXT,
       description TEXT,
       invoice_number TEXT UNIQUE,
       metadata JSONB DEFAULT '{}',  -- course_id, learner_id, etc.
       created_at TIMESTAMPTZ DEFAULT NOW()
   );
   CREATE INDEX idx_financial_tenant_type ON financial_transactions(tenant_id, type, created_at DESC);
   CREATE INDEX idx_financial_user ON financial_transactions(user_id, created_at DESC);
   
   -- ... continue for all ~20 existing tables
   ```

6. **Rewrite all model structs** in `internal/models/`:
   - Replace `primitive.ObjectID` → `uuid.UUID`
   - Replace `bson:"..."` tags → `json:"..."` tags (sqlc uses JSON tags)
   - Replace `bson.M`/`bson.D` → remove (sqlc generates typed structs)
   - Keep `time.Time` (works with pgx)
   - Replace embedded arrays: `AuthMethods []string` works as `TEXT[]` in Postgres
   - `BrandingConfig` struct → use `json.RawMessage` or a typed struct that serializes to JSONB

#### Week 3: Query layer + handlers

7. **Write SQL queries** (`backend/queries/*.sql`) — sqlc generates Go from these:
   ```sql
   -- queries/users.sql
   -- name: GetUserByEmail :one
   SELECT * FROM users WHERE email = $1;
   
   -- name: CreateUser :one
   INSERT INTO users (email, name, password_hash, auth_methods)
   VALUES ($1, $2, $3, $4)
   RETURNING *;
   
   -- name: UpdateUser :exec
   UPDATE users SET name = $2, updated_at = NOW() WHERE id = $1;
   
   -- queries/tenants.sql
   -- name: GetTenantByID :one
   SELECT * FROM tenants WHERE id = $1;
   
   -- name: GetTenantBySlug :one
   SELECT * FROM tenants WHERE slug = $1;
   
   -- name: GetTenantByCustomDomain :one
   SELECT t.* FROM tenants t
   JOIN custom_domains cd ON cd.tenant_id = t.id
   WHERE cd.domain = $1 AND cd.status = 'active';
   ```

8. **Rewrite each handler** — mechanical replacement:
   ```go
   // BEFORE (Mongo)
   var user models.User
   err := h.db.Users().FindOne(ctx, bson.M{"email": email}).Decode(&user)
   
   // AFTER (Postgres via sqlc)
   user, err := h.db.Queries.GetUserByEmail(ctx, email)
   ```

9. **Replace leader election** (`internal/metrics/metrics.go`):
   ```go
   // BEFORE: FindOneAndUpdate with upsert + $or filter
   err := db.DailyMetrics().FindOneAndUpdate(ctx, filter, update, opts).Decode(&lock)
   
   // AFTER: SELECT FOR UPDATE in a transaction
   tx, _ := db.Pool.Begin(ctx)
   defer tx.Rollback(ctx)
   
   var lock LeaderLock
   err := tx.QueryRow(ctx,
       `SELECT holder_id, expires_at FROM leader_locks WHERE lock_name = 'daily_metrics' FOR UPDATE`,
   ).Scan(&lock.HolderID, &lock.ExpiresAt)
   
   if errors.Is(err, pgx.ErrNoRows) || lock.ExpiresAt.Before(time.Now()) {
       _, err = tx.Exec(ctx,
           `INSERT INTO leader_locks (lock_name, holder_id, expires_at) 
            VALUES ('daily_metrics', $1, $2)
            ON CONFLICT (lock_name) DO UPDATE SET holder_id = $1, expires_at = $2`,
           instanceID, time.Now().Add(5*time.Minute))
   }
   tx.Commit(ctx)
   ```

10. **Replace TTL indexes** with `pg_cron` jobs or partition-by-date:
    ```sql
    -- For refresh_tokens, audit_log, telemetry_events etc.
    -- Option A: pg_cron job
    SELECT cron.schedule('cleanup-refresh-tokens', '0 * * * *', 
        'DELETE FROM refresh_tokens WHERE expires_at < NOW()');
    
    -- Option B: Partition by date (better for high-volume tables)
    CREATE TABLE telemetry_events (
        id UUID DEFAULT gen_random_uuid(),
        user_id UUID,
        event_type TEXT,
        payload JSONB,
        created_at TIMESTAMPTZ DEFAULT NOW()
    ) PARTITION BY RANGE (created_at);
    
    CREATE TABLE telemetry_events_2026_07 PARTITION OF telemetry_events
        FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
    -- Drop old partitions instead of DELETE
    ```

#### Week 4: Testing + cleanup

11. **Run the existing test suite** — lastsaas has comprehensive tests (`billing_test.go`, `auth_test.go`, `tenant_test.go`, `isolation_test.go`, etc.). These now test against Postgres via testcontainers. Fix failures.

12. **Remove MongoDB driver** from `go.mod`. Delete `internal/db/mongodb.go` and `internal/db/schema.go`.

13. **Update config** (`internal/config/config.go`):
    ```yaml
    database:
      url: "postgres://user:pass@localhost:5432/lastsaas?sslmode=disable"
      max_conns: 25
      min_conns: 5
    ```

14. **Update Dockerfile** — replace MongoDB with Postgres in the dev/test environment.

### What stays in MongoDB (optional — you can migrate everything)

For simplicity, **migrate everything to Postgres**. The one collection worth considering keeping in MongoDB (or moving to a dedicated time-series store) is `telemetry_events` — it's write-heavy, append-only, and high-volume. But Postgres with partitioning handles this fine for the first 10k+ learners. Don't prematurely split your data layer.

**Recommendation: 100% Postgres. One database. One mental model. One backup strategy.**

### Migration risk mitigation

| Risk | Mitigation |
|---|---|
| Missing a `tenantId` filter during rewrite → cross-tenant data leak | Postgres Row-Level Security as defense-in-depth: `CREATE POLICY tenant_isolation ON courses USING (tenant_id = current_setting('app.current_tenant_id')::uuid);` |
| Existing tests break | Run the full test suite after each model migration; don't accumulate breakage |
| Leader election regression | Add a dedicated test that spins up 2 instances and verifies only one acquires the lock |
| Data type mismatches (ObjectID → UUID) | Write a one-time migration script that maps old ObjectIDs to new UUIDs for any existing data |
| Performance regression | Add `EXPLAIN ANALYZE` to critical queries in tests; ensure indexes match the MongoDB ones |

### Total migration effort: 3–4 weeks

After this, you have a clean Postgres codebase and can write all course-specific code (Course, Section, Lesson, Enrollment, Payout, etc.) in Postgres from day one — with ACID transactions, joins for analytics, RLS for tenant isolation, and materialized views for dashboards.

---

## Decision 2: CDN/Video — Start with Cloudflare (free tier), add Bunny later if needed

### Cloudflare free tier beats Bunny for an MVP

| Resource | Cloudflare Free | Bunny |
|---|---|---|
| CDN bandwidth | **Unlimited free** | $0.005/GB |
| Storage (R2) | **10GB free**, no egress fees | $0.01/GB/month |
| Video upload (Stream) | **100 min/month free** | $0.50/encoded min (5000 free min/month) |
| Video delivery (Stream) | **10,000 min/month free** | $0.005/GB |
| Workers (edge compute) | **100k req/day free** | Limited |
| Custom domain SSL (your domain) | **Free** | Free |
| Custom domain SSL (creator domains) | $0.10/domain/month (Cloudflare for SaaS) | Free |

**For MVP with 0–10 creators and low traffic: Cloudflare = $0/month. Bunny = ~$5–10/month.**

### Recommended MVP architecture: Cloudflare-first

```
Creator uploads video
  ↓
Frontend requests upload URL from /api/creator/media/upload
  ↓
Backend calls Cloudflare Stream API: POST /stream
  ↓
Cloudflare returns upload URL (direct-to-Stream)
  ↓
Frontend uploads MP4 directly to Cloudflare Stream
  ↓
Cloudflare auto-transcodes to HLS (240p/480p/720p/1080p)
  ↓
Backend stores MediaAsset { cf_stream_id, status: "processing" }
  ↓
Cloudflare webhook → /api/internal/cf/webhook (video.ready) → backend sets status: "ready"
  ↓
Learner requests /api/learner/lessons/:id
  ↓
RequireEnrollment middleware → 403 if not enrolled
  ↓
Backend generates Cloudflare Stream signed token (JWT with exp)
  ↓
Frontend VideoPlayer plays HLS via Cloudflare's embedded player or hls.js
```

### Custom domain strategy (phased)

| Phase | Creator domain approach | Cost |
|---|---|---|
| **MVP (0–10 creators)** | Subdomain: `john.mycourses.com` (DNS A record to your server, Cloudflare Free SSL) | $0 |
| **Growth (10–50 creators)** | Cloudflare for SaaS: creators add custom domains via API, auto-SSL | $5/month base + $0.10/domain/month |
| **Scale (50+ creators)** | Evaluate Bunny for custom domains (free per-domain) vs stick with CF for SaaS | Compare: 100 creators × $0.10 = $10/month CF vs ~$10/month Bunny delivery |

### Why Cloudflare Stream for MVP (not Bunny Stream)

1. **Free tier covers MVP**: 100 min upload + 10,000 min delivery/month is enough for 5–10 courses with 50–100 learners. You won't pay anything until you have real traction.
2. **Auto-transcoding included**: Upload MP4, get HLS with multiple bitrates. No MediaConvert setup.
3. **Token authentication built-in**: Sign JWTs with your Stream signing key, verify at edge.
4. **One vendor for everything**: CDN + storage + video + Workers (for edge logic like custom domain routing). Simpler than mixing Bunny + S3 + CloudFront.

### When to switch from Cloudflare Stream to Bunny Stream

Switch when you hit ANY of these:

| Trigger | Why switch |
|---|---|
| Video delivery exceeds 10,000 min/month regularly | Cloudflare Stream charges $0.05/1000 min beyond free tier; Bunny charges $0.005/GB which is ~10× cheaper at scale |
| You need DRM (Widevine/FairPlay) | Cloudflare Stream supports FairPlay + Widevine but only on paid plans; Bunny doesn't have DRM either — switch to Mux at this point |
| You need video analytics (drop-off rates, quality metrics) | Cloudflare Stream has basic analytics; Mux Data is best-in-class |
| You have 500+ creators with custom domains | Cloudflare for SaaS = $50/month; Bunny = $0/domain — Bunny wins on cost |

**Practical trigger**: When your monthly Cloudflare invoice exceeds $50, evaluate Bunny. Until then, Cloudflare Free is the right choice.

### The hybrid sweet spot (when you scale)

Once you have paying creators and real video traffic:

- **Cloudflare R2** for branding assets, PDFs, images (10GB free, then $0.015/GB — still cheap, no egress fees)
- **Bunny Stream** for video (cheaper delivery at scale, free custom domain SSL)
- **Cloudflare CDN** for your platform's HTML/JS/CSS (free unlimited bandwidth)
- **Bunny CDN** for video delivery pullzone (cheaper than CF Stream at scale)

This gives you the best of both: Cloudflare's free tier for the platform, Bunny's low-cost video delivery for the heavy bandwidth.

### Cloudflare-specific integration details

| Component | Implementation |
|---|---|
| **Account setup** | Create Cloudflare account, add your platform domain (e.g. `mycourses.com`) as a zone. Get API token with Stream + R2 + Workers permissions. |
| **R2 bucket** | Create R2 bucket `mycourses-assets` for branding assets + PDFs. 10GB free. |
| **Stream library** | Enable Cloudflare Stream. 100 min upload + 10k min delivery free/month. |
| **Video upload** | `POST https://api.cloudflare.com/client/v4/accounts/{acct}/stream` with `uploadURL` response. Frontend PUTs MP4 directly. |
| **Video playback** | Use Cloudflare's embedded player (`<stream src="VIDEO_ID">`) or generate signed HLS URL: `https://watch.cloudflarestream.com/{videoId}/manifest/video.m3u8?policy={base64Policy}&signature={hmac}` |
| **Signed URLs** | Generate JWT with `exp` claim, sign with Stream signing key. Verify at edge. |
| **Webhook** | Configure Stream webhook → `POST /api/internal/cf/webhook` with `video.encoded`, `video.deleted` events. Verify signature. |
| **Custom domains (later)** | Cloudflare for SaaS API: `POST /zones/{zone}/custom_hostnames` with creator's domain. Auto-issues SSL. |
| **Workers (optional)** | Use for edge logic: redirect creator custom domain to correct tenant, A/B test storefront variants, geo-routing. |

### New env vars (replaces the Bunny env vars from addendum v2)

```yaml
cloudflare:
  api_token: "${CLOUDFLARE_API_TOKEN}"
  account_id: "${CLOUDFLARE_ACCOUNT_ID}"
  zone_id: "${CLOUDFLARE_ZONE_ID}"           # for your platform domain
  r2:
    bucket: "mycourses-assets"
    access_key_id: "${R2_ACCESS_KEY_ID}"
    secret_access_key: "${R2_SECRET_ACCESS_KEY}"
    endpoint: "https://{acct_id}.r2.cloudflarestorage.com"
  stream:
    signing_key: "${CF_STREAM_SIGNING_KEY}"  # for signed video URLs
    webhook_secret: "${CF_STREAM_WEBHOOK_SECRET}"
  saas:  # only when you enable custom creator domains
    enabled: false
    fallback_origin: "app.mycourses.com"
```

### Updated MediaAsset model (Postgres version)

```sql
CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('video','pdf','image')),
    title TEXT,
    
    -- For Cloudflare Stream videos
    cf_stream_id TEXT,  -- Cloudflare Stream video ID
    
    -- For R2 assets (PDFs, images, branding)
    r2_key TEXT,        -- e.g. "tenants/{tenantId}/pdfs/{filename}"
    r2_url TEXT,        -- public URL or signed URL
    
    -- Common metadata
    size_bytes BIGINT,
    mime_type TEXT,
    duration_sec INT,   -- for videos
    status TEXT DEFAULT 'processing',  -- processing, ready, failed, deleted
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_media_tenant ON media_assets(tenant_id, created_at DESC);
CREATE INDEX idx_media_kind ON media_assets(kind, tenant_id);
```

### Updated cost projection (Cloudflare MVP vs Bunny MVP vs original plan)

| Scenario (MVP: 5 creators, 50 learners, 20 courses, 5GB video) | Cloudflare Free | Bunny | Original plan (S3+CloudFront) |
|---|---|---|---|
| Storage | $0 (R2 free 10GB) | $0.05 (5GB × $0.01) | $0.12 (5GB × $0.023) |
| Video delivery | $0 (under 10k min free) | $1 (200GB × $0.005) | $17 (200GB × $0.085) |
| Video encoding | $0 (under 100 min free) | $0 (under 5000 min free) | $0 (no transcoding) |
| CDN (HTML/JS/CSS) | $0 (free unlimited) | $0.50 | $4 |
| Custom domain SSL | $0 (subdomains only) | $0 | $0 (Caddy on-demand) |
| **Monthly total** | **$0** | **~$2** | **~$21** |

**For the first 6–12 months, Cloudflare saves you 100% of CDN/storage costs.**

---

## Updated Implementation Roadmap (v3)

The v2 plan had 12 phases. v3 adds a **Phase 0** for the Postgres migration and adjusts the CDN choices in later phases.

### Phase 0 — Postgres Migration (NEW, 3–4 weeks)

**Do this BEFORE any course-specific code.**

- Week 1: Add pgx + sqlc + golang-migrate to `go.mod`. Replace `internal/db/mongodb.go` with `internal/db/postgres.go`. Set up `sqlc.yaml` config.
- Week 2: Write all migration SQL files (`migrations/0001_*.sql` through `0020_*.sql`). Write all query files (`queries/*.sql`). Run `sqlc generate` to produce typed Go code. Rewrite all model structs (replace `primitive.ObjectID` with `uuid.UUID`, remove bson tags).
- Week 3: Rewrite each handler's DB calls (mechanical: `FindOne` → `GetUserByEmail`, `UpdateOne` → `UpdateUser`, etc.). Replace leader election with `SELECT FOR UPDATE`. Replace TTL indexes with `pg_cron` jobs.
- Week 4: Update config, Dockerfile, fly.toml. Run full test suite against Postgres (via testcontainers). Fix all failures. Remove MongoDB driver.

**Exit criteria**: All existing lastsaas tests pass against Postgres. No MongoDB references in codebase. App boots and works end-to-end with Postgres only.

### Phase 1 — Domain Models & Migrations (1 week, unchanged from v2)

Write the 12 new course model migrations as Postgres SQL files (not MongoDB schema.go entries). Each new table gets a migration file:
- `0021_create_courses.up.sql`
- `0022_create_sections.up.sql`
- `0023_create_lessons.up.sql`
- `0024_create_media_assets.up.sql`
- `0025_create_enrollments.up.sql`
- `0026_create_course_progress.up.sql`
- `0027_create_course_coupons.up.sql`
- `0028_create_reviews.up.sql`
- `0029_create_payouts.up.sql`
- `0030_create_custom_domains.up.sql`
- `0031_create_certificates.up.sql`
- `0032_create_creator_profiles.up.sql`

**Now you get to use real foreign keys, CHECK constraints, and ACID transactions for free.** Example:

```sql
CREATE TABLE enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('active','completed','refunded','disputed')) DEFAULT 'active',
    price_paid_cents BIGINT NOT NULL,
    coupon_id UUID REFERENCES course_coupons(id),
    enrolled_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    UNIQUE(course_id, user_id)  -- prevent duplicate enrollments
);
CREATE INDEX idx_enrollment_user ON enrollments(user_id, status);
CREATE INDEX idx_enrollment_tenant ON enrollments(tenant_id, enrolled_at DESC);
```

The `UNIQUE(course_id, user_id)` constraint replaces what MongoDB would do with a compound unique index — but now it's enforced at the DB level, not just the application level.

### Phase 2–11 — Unchanged from v2 plan, except:

- All handlers use sqlc-generated queries (type-safe, compile-time checked)
- All new tables use UUID primary keys + proper foreign keys
- `RequireEnrollment` middleware can use a single SQL query with a JOIN: `SELECT 1 FROM enrollments WHERE user_id = $1 AND course_id = $2 AND status = 'active'`
- Creator revenue dashboard uses materialized views: `CREATE MATERIALIZED VIEW creator_revenue AS SELECT tenant_id, SUM(amount_cents) as revenue FROM financial_transactions WHERE type = 'course_purchase' GROUP BY tenant_id;` — refresh daily
- Marketplace search uses Postgres FTS: `WHERE to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', $1)`
- Row-Level Security as defense-in-depth: even if a handler forgets a `WHERE tenant_id = ?`, RLS prevents cross-tenant data leakage

### Phase 4 (Storefront) — CDN update

Replace S3/Bunny references with Cloudflare:
- Video upload via Cloudflare Stream direct upload
- Video playback via Cloudflare Stream signed URLs (JWT)
- Asset storage in R2 (presigned URLs via the Cloudflare R2 S3-compatible API)
- Webhook handler at `/api/internal/cf/webhook` for Stream events

### Phase 6 (Custom Domains) — Cloudflare for SaaS

When creators want custom domains:
- Backend calls `POST /zones/{zone}/custom_hostnames` to register `academy.johndoe.com`
- Cloudflare auto-issues SSL cert
- Creator points CNAME to `mycourses.com` (or Cloudflare's fallback origin)
- No Caddy on-demand TLS needed — Cloudflare handles it

---

## Summary of v3 decisions

| Decision | v2 recommendation | v3 final recommendation | Why changed |
|---|---|---|---|
| Database | Stay with MongoDB for v1, migrate to Postgres in v2 | **Migrate to Postgres NOW (Phase 0, before course code)** | User correctly identified that migrating a small codebase is cheaper than migrating a large one. 3–4 weeks upfront saves 6–8 weeks later. |
| Video CDN | Bunny Stream + Bunny CDN | **Cloudflare Stream + R2 + CF CDN for MVP; switch to Bunny when scale demands** | Cloudflare has a genuine free tier ($0/month for MVP); Bunny has no free tier. Cloudflare Free covers the first 6–12 months at zero cost. |
| Custom domains | Caddy on-demand TLS | **Subdomains for MVP (free), Cloudflare for SaaS when creators demand custom domains** | Cloudflare for SaaS handles SSL/provisioning via API, no Caddy config needed |
| Asset storage | S3 | **Cloudflare R2** (10GB free, no egress fees to CF CDN) | Free tier + integrated with Cloudflare ecosystem |
| Schema management | MongoDB schema.go (JSON validators + indexes) | **golang-migrate** (versioned SQL migration files) + **sqlc** (type-safe Go from SQL) | Industry standard for Postgres, compile-time query checking |

## Total cost to launch (revised)

| Phase | Duration | Infra cost |
|---|---|---|
| Phase 0 (Postgres migration) | 3–4 weeks | $0 (local Postgres via Docker) |
| Phase 1–11 (course platform) | 12–13 weeks | $0 (Cloudflare Free + local Postgres) |
| Launch (Fly.io + Cloudflare + Postgres) | — | ~$25/month (Fly.io 1 machine + managed Postgres $15) |
| **Total to launch** | **15–17 weeks** | **$0 infra during development, ~$25/month at launch** |

Compare to v2 plan: 13–15 weeks but with MongoDB + Caddy + S3 + CloudFront complexity. v3 takes slightly longer upfront (3–4 week migration) but yields a cleaner, more scalable, cheaper-to-operate codebase.

## What to do next

1. **Confirm the Postgres migration decision** — this is the big one. Once you start, you're committed for 3–4 weeks.
2. If yes, I can **fork lastsaas and start Phase 0** — set up the Postgres driver, sqlc config, write the first 5 migration files (users, tenants, memberships, plans, financial_transactions) as a proof of concept.
3. Or I can **ask the wiki more questions** on any specific migration concern (e.g. how a particular handler's aggregation pipeline should translate to SQL).
