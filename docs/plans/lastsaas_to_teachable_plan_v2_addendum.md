# MongoDB vs PostgreSQL & Bunny vs CloudFront/Cloudflare/Mux
## Decision Addendum to the v2 Plan

> Source: Wiki Q&A #27 (MongoDB coupling), #28 (GridFS/assets), #29 (Postgres benefits), #30 (CDN — wiki doesn't cover external CDNs, so this section draws on general engineering knowledge).

---

## Question 1: MongoDB vs PostgreSQL

### TL;DR recommendation

**Stay with MongoDB for the MVP fork of lastsaas. Plan to migrate financial and relational data to PostgreSQL in v2 (post-launch) once product-market fit is proven.**

If you're starting greenfield (no lastsaas code reuse), Postgres would be the better single choice. But for forking lastsaas, the migration cost is **3–4 weeks of pure refactoring with zero new features**, which delays launch. The wiki confirmed MongoDB is woven deeply enough that this isn't a "swap the driver" change.

### What the wiki told us about MongoDB coupling

| Coupling point | Migration cost |
|---|---|
| `primitive.ObjectID` used pervasively in every model (User, Tenant, Plan, FinancialTransaction, etc.) | High — every model, every query, every API response shape changes |
| `bson.M` / `bson.D` used for all queries and updates | High — full query rewrite to SQL |
| Aggregation pipeline in `metrics.go::collectDaily` (DAU/WAU/MAU, ARR) | Medium — rewrite to SQL GROUP BY + window functions |
| Aggregation pipeline in `health/query.go::GetIntegrationCounts24h` | Medium — rewrite to SQL COUNT + GROUP BY |
| `FindOneAndUpdate` with `$inc` for atomic leader election in `metrics.go::tryAcquireOrRenew` | High — replace with `SELECT FOR UPDATE` or Postgres advisory locks |
| TTL indexes on 13 collections (refresh_tokens, audit_log, telemetry_events, etc.) | Medium — replace with `pg_cron` jobs or partition-by-date + drop partition |
| Embedded arrays (`User.AuthMethods []string`, `BrandingConfig.NavItems []NavItem`) | Medium — become junction tables or `TEXT[]`/`JSONB` |
| `BrandingAsset.Data []byte` stored directly in document (NOT GridFS — wiki clarified this) | Low — becomes `BYTEA` (but you should move to S3 anyway, see Question 2) |
| JSON Schema validators on every collection | Medium — translate to `CHECK` constraints + `JSONB` schema validation |
| Text index on `system_logs.message` | Low — Postgres FTS is strictly better |
| DB access centralized in `internal/db/mongodb.go` | **Mitigating** — refactor is contained to one package |

### Why MongoDB is fine for the MVP

1. **The hard parts of a course platform are NOT the database.** They're Stripe Connect flows, custom domain TLS, video streaming, and the creator studio UX. lastsaas already solves auth, RBAC, billing, branding, webhooks, telemetry, rate limiting, and admin console — all of which work with MongoDB today. Re-using them gets you 13 weeks to launch instead of 20+.

2. **Your new entities are well-modeled in MongoDB.** Course → Sections → Lessons is hierarchical and naturally embeds. `CourseProgress` is write-heavy append-mostly (perfect for Mongo). `UsageEvent` for video bandwidth metering is high-volume time-series data (MongoDB handles this well; Postgres would need TimescaleDB extension for equivalent performance).

3. **Financial transactions work in MongoDB if you're disciplined.** The wiki noted lastsaas already uses an atomic counter for invoice numbers. The Stripe Connect webhook handler already creates `Enrollment` + `FinancialTransaction` atomically. For the volumes of a marketplace (not a high-frequency trading system), MongoDB's single-document atomicity + idempotent webhook handlers (keyed on Stripe event ID) is sufficient. Add multi-document transactions only for the few cases that truly need them (e.g. commission split + payout record creation).

4. **Multi-tenancy via `tenantId` filter works.** Yes, Postgres Row-Level Security is more elegant — but lastsaas already has the `TenantMiddleware` that injects `tenantId` into every query context, and every existing handler follows the pattern. Your new handlers will follow the same pattern. The risk of "forgetting a `tenantId` filter" is mitigated by tests (the existing `isolation_test.go` and your new `course_test.go` should verify cross-tenant access returns 403/404).

### Why you should plan a Postgres migration for v2 (post-launch)

Once you have product-market fit (say, 1000+ paying creators), the following Postgres advantages become real:

| Postgres advantage | When it matters |
|---|---|
| **ACID multi-document transactions** | When you have complex marketplace flows: refund + transfer reversal + enrollment revocation + certificate revocation must all succeed atomically. MongoDB supports multi-doc transactions but they're slower and rarely used. |
| **Row-Level Security** | When you have 10k+ creators and want a defense-in-depth guarantee that no query can leak cross-tenant data even if a developer forgets a `WHERE tenantId = ?`. |
| **Joins for analytics** | When creator dashboards need: "top 5 courses by revenue this month, joined with review ratings, joined with learner geographic distribution." Mongo requires denormalization or multiple queries; Postgres does it in one SQL statement. |
| **Materialized views** | When you have 100k+ learners and creator revenue dashboards need sub-second response. Pre-compute daily revenue per creator in a materialized view, refresh every hour. |
| **Postgres FTS** | When marketplace search needs stemming, typo tolerance, ranking. Mongo's text index is basic; Postgres FTS with `tsvector` + GIN indexes is production-grade (still not as good as Elasticsearch/Algolia, but good enough to delay that migration). |
| **`JSONB` for branding** | Best of both worlds: relational integrity for the parent `Tenant` row + flexible JSON for `BrandingConfig`. Mongo gives you flexibility but no foreign keys. |
| **`SELECT FOR UPDATE` + `SKIP LOCKED`** | When webhook dispatch needs a job queue. Postgres can be a perfectly good job queue with `SKIP LOCKED`; MongoDB needs a separate Redis or a careful `FindOneAndUpdate`-based queue. |

### The hybrid option (worth considering for v1.5)

You don't have to pick one. A common Teachable-style architecture:

- **PostgreSQL** for the financial core: `tenants`, `users`, `memberships`, `enrollments`, `financial_transactions`, `payouts`, `course_coupons`, `certificates`. These need ACID and joins.
- **MongoDB** for the high-volume/telemetry layer: `usage_events`, `telemetry_events`, `system_logs`, `course_progress` (write-heavy), `webhook_delivery_logs`. These need scale and flexible schema.
- **S3 + CDN** for binary content: videos, PDFs, branding assets (never in DB).

This is exactly what companies like Stripe, GitHub, and Shopify do — different stores for different access patterns. The Go backend can use both `database/sql` (for Postgres) and `mongo-go-driver` (for Mongo) side-by-side; the `internal/db/` package boundary makes this clean.

### Migration path (when you decide to do it)

1. **Phase A — Read replicas + dual-write**: Stand up Postgres. Write new financial data to both Mongo and Postgres (dual-write). Read from Mongo. Verify consistency.
2. **Phase B — Read from Postgres**: Switch reads for financial endpoints to Postgres. Mongo becomes write-only for those tables.
3. **Phase C — Stop dual-write**: Cut over. Mongo retains only the telemetry/usage data.
4. **Phase D — Migrate historical data**: Backfill Postgres from Mongo for historical financial records.

Estimated effort: 4–6 weeks for an experienced team. Don't do this before product-market fit.

### Decision matrix

| If you are... | Recommendation |
|---|---|
| Forking lastsaas for a quick MVP launch (3-month horizon) | **Stay with MongoDB.** Migration cost = 3–4 weeks of zero new features. Not worth it. |
| Forking lastsaas but have 6+ months before launch and expect heavy analytics from day 1 | **Migrate to Postgres** before writing new code. Easier to migrate a small codebase than a large one. |
| Starting greenfield (not forking lastsaas) | **Use Postgres from day 1.** The relational model fits course-selling better. Use `JSONB` for branding config. |
| Already launched on Mongo and hitting scale pain | **Hybrid**: keep Mongo for telemetry, migrate financial core to Postgres via the 4-phase path above. |

---

## Question 2: Bunny CDN vs CloudFront vs Cloudflare vs Mux

### TL;DR recommendation

**For your use case (course-selling SaaS with per-creator custom domains + video streaming):**

- **MVP (Phase 1–4)**: **Bunny CDN + Bunny Stream** for video. Use Bunny's `pullzone` with edge storage for everything else. Lowest cost, best DX for custom domains, video streaming built in.
- **Alternative if you want zero vendor lock-in**: **CloudFront + S3** with Lambda@Edge for custom domain TLS.
- **Avoid for this use case**: Cloudflare Stream (vendor lock-in for video), pure Mux (expensive at scale, no CDN for non-video assets).

### Why Bunny CDN wins for this specific use case

| Requirement | Bunny | CloudFront | Cloudflare | Mux |
|---|---|---|---|---|
| **Per-creator custom domain TLS at scale (1000s of domains)** | ✅ Free SSL for unlimited custom domains via Bunny's edge certificates. Add a custom hostname in 1 API call. | ⚠️ CloudFront supports custom domains but each needs ACM cert provisioning. AWS SaaS Builder toolkit helps but it's heavier. | ✅ Cloudflare for SaaS (Custom Hostnames API) is the gold standard — but costs $0.10/domain/month. | ❌ Not a CDN for non-video assets. |
| **Video streaming (HLS, adaptive bitrate)** | ✅ Bunny Stream: uploads → auto-transcodes to HLS → serves via CDN. Token authentication built in. | ⚠️ No transcoding. You'd pair with AWS MediaConvert. | ✅ Cloudflare Stream does this, but video is locked to Cloudflare's player & API. | ✅ Mux is best-in-class for video (HLS, DRM, analytics). But pricing is per-minute-encoded + per-minute-delivered. |
| **Pricing** | ✅ ~$0.005/GB (one of the cheapest). Bunny Stream: $0.005/GB delivery + $0.50/encoded minute (first 5000 min free). | ~$0.085/GB (US/EU). Lower with CloudFront Security Bundle. | $0.05/GB (cf plan), R2 storage $0.015/GB/mo (no egress fees from R2 to Cloudflare). | Mux: $1/1k encoded minutes (SD), $1.50/1k delivered minutes. Becomes expensive at scale. |
| **Edge storage (upload once, serve globally)** | ✅ Bunny Storage Zone: $0.01/GB/mo. No egress fees between Bunny Storage and Bunny CDN. | ✅ S3 Standard: $0.023/GB/mo + egress to CloudFront is free. | ✅ R2: $0.015/GB/mo, no egress fees. | ❌ Not a storage product. |
| **Token authentication (signed URLs for paid video)** | ✅ Built-in token authentication: set a token auth key on the pullzone, generate URL signatures in Go with HMAC-SHA256. | ✅ Signed URLs / signed cookies via CloudFront API. | ✅ Cloudflare Worker validates JWT or signed URL. | ✅ Playback IDs with signing keys. |
| **Custom SSL for arbitrary creator domains** | ✅ Add a custom hostname, Bunny issues a Let's Encrypt cert automatically. No per-domain fee. | ⚠️ ACM cert per domain (free but operationally heavy at 1000+ domains). Or use CloudFront with SaaS Builder. | ✅ Cloudflare for SaaS: $0.10/domain/month + $5/month base. Best API. | ❌ |
| **DRM (Widevine, FairPlay, PlayReady)** | ❌ Not yet (Bunny Stream has token auth but not full DRM). | ⚠️ Via AWS MediaConvert + third-party DRM. Complex. | ✅ Cloudflare Stream supports FairPlay + Widevine. | ✅ Mux supports all three with signed URLs. |
| **Analytics** | ✅ Bunny dashboard: requests, bandwidth, status codes, top URLs. Decent. | ✅ CloudFront reports + CloudWatch. Powerful but complex. | ✅ Cloudflare analytics. Strong. | ✅ Mux Data: video-specific analytics (quality, completion, errors). Best in class. |
| **Vendor lock-in** | Medium — Bunny Stream's video library is portable (it's HLS), but the upload/transcode API is Bunny-specific. | Low — S3 + CloudFront is industry standard, easy to migrate. | High for Cloudflare Stream (proprietary player + API). Low for R2 + Workers. | High — Mux's value is in its managed pipeline. |
| **Documentation quality** | Good but smaller community. | Excellent (AWS). | Excellent. | Excellent. |
| **Go SDK / API client** | REST API, easy to call from Go. No official Go SDK but trivial to wrap. | AWS SDK for Go v2 (excellent). | Cloudflare Go SDK (community-supported, good). | Official Mux Go SDK. |

### Why Bunny fits the course platform use case specifically

1. **Custom domains are your biggest infra challenge.** You have 1000+ creators each wanting `academy.theirname.com`. Bunny issues free SSL for unlimited custom hostnames in a single API call. CloudFront requires ACM cert management per domain (free but operationally heavy). Cloudflare for SaaS is the only competitor that matches Bunny's DX, but it costs $0.10/domain/month = $100/month for 1000 creators.

2. **Video streaming with auto-transcoding is built in.** Bunny Stream takes an MP4 upload, transcodes to HLS with multiple bitrates, and serves via the CDN. No need to wire up AWS MediaConvert + S3 + CloudFront yourself. Mux does this too but at ~3× the cost.

3. **Token auth for paid content.** Bunny supports URL signing out of the box. Your Go backend generates an HMAC-SHA256 signature for each video URL with a TTL. No Lambda@Edge needed (unlike CloudFront).

4. **Pricing is dramatically cheaper.** For 10TB/month video delivery:
   - Bunny: $50
   - CloudFront: $850 (or $425 with Security Bundle)
   - Cloudflare: $500
   - Mux: $15,000+ (delivery + encoding)
   
   For a bootstrapped course platform, this is the difference between profitable and not.

### When you should NOT pick Bunny

| Scenario | Better choice | Why |
|---|---|---|
| You need full DRM (Widevine/FairPlay) for premium content | Mux or Cloudflare Stream | Bunny doesn't yet support full DRM |
| You're already all-in on AWS and want one vendor | CloudFront + S3 + MediaConvert | Simpler procurement |
| You expect 100TB+/month video delivery and need 99.99% SLA + multi-CDN failover | Multi-CDN setup (CloudFront + Bunny + Cloudflare via a CDN selector) | No single CDN is sufficient |
| You need edge compute (A/B testing, personalization at edge) | Cloudflare Workers or CloudFront + Lambda@Edge | Bunny's edge scripting is limited |
| Your team is unfamiliar with Bunny and you have deep AWS expertise | CloudFront | DX matters; use what your team knows |

### Recommended architecture (revised from v2 plan)

**Phase 1–4 (MVP):**

```
Creator uploads video
  ↓
Frontend requests upload URL from /api/creator/media/upload
  ↓
Backend calls Bunny Stream API: POST /videos with collection_id
  ↓
Bunny returns a video ID + direct upload URL (TTL 1 hour)
  ↓
Frontend uploads directly to Bunny
  ↓
Bunny auto-transcodes to HLS (240p/480p/720p/1080p)
  ↓
Backend stores MediaAsset { bunny_video_id, status: "processing" }
  ↓
Bunny webhook → /api/internal/bunny/webhook (video.ready) → backend sets status: "ready"
  ↓
Learner requests /api/learner/lessons/:id
  ↓
RequireEnrollment middleware → 403 if not enrolled
  ↓
Backend generates Bunny token-signed URL: https://[creator-subdomain].b-cdn.net/[video_id]/playlist.m3u8?token=[HMAC]
  ↓
Frontend VideoPlayer plays HLS via hls.js or video.js
```

**Phase 5+ (when you outgrow Bunny or need DRM):**

Migrate to Mux for video (better DRM + analytics), keep Bunny for non-video asset delivery (PDFs, images, branding assets). Or migrate to CloudFront + S3 + MediaConvert for full AWS-native stack.

### Bunny integration specifics

| Component | Implementation |
|---|---|
| **Bunny account setup** | Create Bunny account, create a Storage Zone (for non-video assets) + Stream Library (for video). Get API key + library ID. |
| **Custom creator domains on Bunny** | For each creator's custom domain: (1) Bunny API: `POST /pullzone/{id}/addCustomHostname` with the creator's domain. (2) Bunny auto-issues Let's Encrypt cert. (3) Creator points CNAME to `custom-<id>.b-cdn.net`. (4) Bunny handles SNI on incoming requests. |
| **Token auth (signed URLs)** | Set a Token Authentication Key on the pullzone. In Go, generate URL signature: `base64url(md5(secret + path + expires)) + expires`. Bunny verifies and serves. |
| **Video upload** | `POST https://video.bunnycdn.com/library/{libraryId}/videos` with `title`. Returns `{ videoId, uploadUrl }`. Frontend PUTs MP4 to `uploadUrl`. |
| **Video status webhook** | Configure Bunny webhook → `POST https://api.yourplatform.com/internal/bunny/webhook` with HMAC verification. Events: `video.created`, `video.encoded`, `video.deleted`. |
| **Playback URL** | `https://[creator-subdomain].b-cdn.net/[videoId]/playlist.m3u8?token=[signedToken]&expires=[unixTime]` |

### Cost comparison at 1000 creators + 50,000 learners + 5TB/month video delivery

| Component | Bunny | CloudFront | Cloudflare | Mux |
|---|---|---|---|---|
| Custom domain SSL (1000 domains) | $0 | $0 (ACM free, but ops cost) | $100/mo ($0.10 × 1000) | N/A |
| Video delivery (5TB) | $25 | $425 | $250 | $7,500 |
| Video encoding (assume 1000 hours uploaded/yr = ~83hr/mo) | $25 (Bunny Stream $0.50/encoded min × 5000 min free, then $0.005/min; ~83hr = 5000 min, mostly free) | $375 (MediaConvert) | $45 (Cloudflare Stream $0.05/min × 900 min beyond free tier) | $750 ($1/1k SD encoded min × 750) |
| Asset storage (1TB) | $10 | $23 | $15 | N/A |
| Non-video CDN (1TB) | $5 | $85 | $50 | N/A |
| **Monthly total** | **~$65** | **~$900** | **~$460** | **~$8,250** |

For a course platform charging creators $29–$99/month, Bunny saves you enough to fund a senior engineer.

### The honest caveats about Bunny

1. **Smaller company.** Bunny is a bootstrapped European company, not AWS/Cloudflare. If they go down or out of business, you need a backup plan. Mitigation: keep `MediaAsset.StorageURL` as an abstraction so you can swap CDNs. Always store the original MP4 in S3 as well (cheap insurance).

2. **Less mature ecosystem.** Fewer community libraries, less Stack Overflow content. The API is well-documented but you'll write more glue code yourself.

3. **No full DRM (yet).** If your creators are selling $2000 certification courses and need Hollywood-grade DRM, Bunny won't cut it. For 95% of course platforms, token auth with 10-min TTL + enrollment check is sufficient.

4. **Video analytics are basic.** Bunny Stream shows play count and bandwidth, not "30% of learners dropped at the 4-minute mark of lesson 3." Mux Data gives you that. If analytics matter for creator retention, you'd add Mux Data on top of Bunny delivery later.

### Final recommendation

| Phase | Video | Non-video CDN | Custom domain TLS |
|---|---|---|---|
| **MVP** | Bunny Stream | Bunny CDN (same pullzone) | Bunny (free, unlimited) |
| **Growth (5k creators)** | Bunny Stream + Mux Data (analytics only, no delivery) | Bunny CDN | Bunny (still) |
| **Scale (50k+ creators)** | Multi-CDN via DNS (Bunny + Cloudflare as failover) | Bunny primary, Cloudflare failover | Bunny primary, Cloudflare for SaaS failover |
| **Premium DRM tier** | Mux (for premium creators only, opt-in feature) | Bunny (for everyone else) | Bunny |

---

## How this changes the v2 plan

### MongoDB decision
- **No change to the v2 plan.** Stay with MongoDB for v1. The v2 plan's existing recommendations stand.
- **Add to technical-debt section**: "Plan a Postgres migration for the financial core (`enrollments`, `financial_transactions`, `payouts`, `certificates`) in v2 post-launch, via the 4-phase dual-write path. Estimated 4–6 weeks."

### CDN decision (replaces §9 of v2 plan)
- **§9.1 Architecture**: Replace S3 + CloudFront with Bunny Stream + Bunny CDN. Backend stores `MediaAsset.BunnyVideoID` instead of `StorageURL`. Upload flow uses Bunny's direct upload URL instead of S3 presigned PUT.
- **§9.2 Transcoding**: Bunny Stream handles this automatically. Remove the "no transcoding for MVP" recommendation — Bunny does it for free (first 5000 min/month).
- **§9.3 Storage metering**: Same — meter via `UsageEvent` for `video_storage_mb_hours` and `video_bandwidth_gb`.
- **§9.4 Access gating**: Replace S3 presigned URLs with Bunny token-signed URLs. Same enrollment check, different URL signing code.
- **§9.5 DRM**: Defer. Bunny's token auth + 10-min TTL + enrollment check is the MVP.
- **§5.3 Caddyfile**: Now Caddy only handles the platform domain + routes to backend. Custom creator domains go directly to Bunny via DNS CNAME (creator's domain → Bunny pullzone). This **simplifies** the custom domain flow significantly — no more Caddy `on_demand_tls` needed for video; only for the SPA HTML if you want it served from the creator's domain.
- **§2.10 Infrastructure**: Add Bunny env vars (`BUNNY_API_KEY`, `BUNNY_LIBRARY_ID`, `BUNNY_STORAGE_ZONE`, `BUNNY_PULLZONE_ID`, `BUNNY_TOKEN_AUTH_KEY`). Remove the AWS S3 IAM setup.
- **§2.1 Models**: `MediaAsset` gets `BunnyVideoID string` instead of `StorageURL`. Keep an optional `OriginalS3Key` field as backup storage of the raw MP4.
- **§13.3 Cost profile at scale**: Update from "~$1,200 S3 storage + ~$2,500 CloudFront bandwidth" to "~$10 Bunny storage + ~$50 Bunny bandwidth" for the same 50TB. Massive savings.

### New env vars to add

```yaml
bunny:
  api_key: "${BUNNY_API_KEY}"
  library_id: "${BUNNY_LIBRARY_ID}"
  storage_zone: "${BUNNY_STORAGE_ZONE}"
  storage_password: "${BUNNY_STORAGE_PASSWORD}"
  pullzone_id: "${BUNNY_PULLZONE_ID}"
  pullzone_hostname: "mycourses.b-cdn.net"
  token_auth_key: "${BUNNY_TOKEN_AUTH_KEY}"   # for signed video URLs
  webhook_secret: "${BUNNY_WEBHOOK_SECRET}"   # for verifying Bunny webhooks
```

### New handler

Add `backend/internal/api/handlers/bunny_webhook.go`:

| Endpoint | Purpose |
|---|---|
| `POST /api/internal/bunny/webhook` | Receives `video.encoded`, `video.deleted` events from Bunny Stream. Verifies HMAC signature. Updates `MediaAsset.Status` accordingly. |

### Simplified custom domain flow (big win)

With Bunny handling video delivery on custom creator domains, the platform backend only needs to serve the SPA HTML + API on custom domains. You have two options:

**Option A (recommended for MVP)**: Platform SPA served from a single platform domain (`app.mycourses.com`). Creator's custom domain (`academy.johndoe.com`) CNAMEs to Bunny. Bunny serves the SPA HTML (cached) + all video assets. Backend API calls go to `api.mycourses.com` (fixed subdomain). This is how Teachable does it — creator's custom domain serves the storefront, but the API is centralized.

**Option B (more complex)**: Caddy `on_demand_tls` for the SPA on creator domains (as originally planned in v2). Use this only if you need the creator's entire school (including non-video pages) to be served from their custom domain with platform-controlled TLS.

Either way, Bunny handles the heavy lifting for video delivery — no per-creator cert management, no Caddy config churn.

---

## Summary

| Question | Answer | Confidence |
|---|---|---|
| MongoDB vs PostgreSQL for the course SaaS | **MongoDB for v1 (stay with lastsaas default). Plan Postgres migration for financial core in v2.** Migration cost (3–4 weeks of zero new features) isn't worth it before product-market fit. | High — wiki confirmed the coupling is moderate-to-high; the cost of migration is real |
| Bunny vs CloudFront/Cloudflare/Mux | **Bunny CDN + Bunny Stream for MVP.** 10–100× cheaper than alternatives for the specific use case of per-creator custom domains + video streaming. Switch to Mux only when you need DRM. | High — Bunny's feature set is documented and matches the requirements; cost math is decisive |
