# Implementation Patterns Addendum (v4.1)
## Companion to the v4 Canonical Plan

> The v4 canonical plan defines WHAT to build. This addendum defines HOW to build it — the exact code patterns to follow so new code matches lastsaas's conventions.
>
> Source: Wiki Q&As #39–#45 (7 sessions on handler structure, test helpers, frontend routing, API client, webhook handler, seeding/bootstrap, entitlements).

---

## 1. Backend Handler Pattern

Every new handler follows this structure:

### Handler struct + constructor
```go
type CourseHandler struct {
    db          *db.DB
    syslog      *syslog.Logger
    eventEmitter events.Emitter
}

func NewCourseHandler(database *db.DB, sysLogger *syslog.Logger, emitter events.Emitter) *CourseHandler {
    return &CourseHandler{db: database, syslog: sysLogger, eventEmitter: emitter}
}
```

### Create method pattern
```go
func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user + tenant from context
    user := middleware.GetUserFromContext(r.Context())
    tenant := middleware.GetTenantFromContext(r.Context())

    // 2. Decode request body
    var req struct { Title string `json:"title"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // 3. Validate
    if req.Title == "" {
        respondWithError(w, http.StatusBadRequest, "Title is required")
        return
    }

    // 4. DB insert (sqlc generated query)
    created, err := h.db.Queries.CreateCourse(r.Context(), gen.CreateCourseParams{...})
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed")
        return
    }

    // 5. Emit event
    h.eventEmitter.Emit(r.Context(), events.Event{Type: "course.created", ...})

    // 6. Respond
    respondWithJSON(w, http.StatusCreated, created)
}
```

### List with pagination pattern
```go
func (h *CourseHandler) List(w http.ResponseWriter, r *http.Request) {
    page, limit := parsePagination(r) // page default 1, limit default 25
    status := r.URL.Query().Get("status")

    courses, err := h.db.Queries.ListCoursesByTenant(r.Context(), gen.ListCoursesByTenantParams{
        TenantID: tenant.ID,
        Column2:  status,  // sqlc names params by position when using $N
        Limit:    int32(limit),
        Offset:   int32((page - 1) * limit),
    })

    respondWithJSON(w, http.StatusOK, map[string]interface{}{
        "courses": courses, "total": total, "page": page, "limit": limit,
    })
}
```

### Error helpers
```go
respondWithError(w, http.StatusBadRequest, "Invalid request body")
apierror.BadRequest(w, r, "Invalid request body")  // structured with code + requestId
```

### Route registration
```go
creator := courseAPI.PathPrefix("/creator").Subrouter()
creator.Use(authMiddleware.RequireAuth)
creator.Use(middleware.CustomDomainTenantMiddleware(pgDB, platformDomain))
creator.HandleFunc("/courses", courseHandler.Create).Methods("POST")
```

---

## 2. Test Pattern

### Test helpers (from `internal/testutil/testutil.go`)
- `MustConnectTestDB(t)` — connect to test Postgres (embedded-postgres)
- `CreateTestUser(t, db, email, password, name)` — create a user
- `CreateTestTenant(t, db, name, ownerID, isRoot)` — create a tenant + owner membership
- `CreateTestMembership(t, db, userID, tenantID, role)` — create a membership
- `CreateTestPlan(t, db, name, monthlyPriceCents, isSystem)` — create a plan
- `CleanupCollections(t, db)` — truncate all tables

### Complete test example
```go
func TestE2E_CourseCRUD(t *testing.T) {
    edb, cleanup := setupEmbeddedPostgres(t)
    defer cleanup()
    ctx := context.Background()
    applyMigrations(t, ctx, edb.db)

    // Create plan + tenant
    plan, _ := edb.db.Queries.CreatePlan(ctx, gen.CreatePlanParams{...})
    tenant, _ := edb.db.Queries.CreateTenant(ctx, gen.CreateTenantParams{...})

    // Create course
    course, err := edb.db.Queries.CreateCourse(ctx, gen.CreateCourseParams{
        TenantID: tenant.ID, Title: "Test", Slug: "test", ...
    })

    // Assert
    if course.ID == uuid.Nil { t.Error("expected non-nil UUID") }
}
```

### Test conventions
- One `_test.go` per handler file
- Test happy path + unauthorized + forbidden + not found + bad request + tenant isolation
- For Stripe: mock via `httptest.NewServer`
- For webhooks: test idempotency by sending same event twice

---

## 3. Frontend Routing Pattern

### Route structure in App.tsx
```tsx
{/* Storefront routes (public) */}
<Route path="/courses" element={<StorefrontHomePage />} />
<Route path="/courses/:slug" element={<CourseDetailPage />} />
<Route path="/checkout/:courseSlug" element={<CheckoutPage />} />
<Route path="/certificates/verify/:token" element={<CertificateVerifyPage />} />

{/* Learner routes (auth required) */}
<Route path="/learn/my-courses" element={<MyCoursesPage />} />
<Route path="/learn/course/:courseId/:lessonId" element={<CoursePlayerPage />} />

{/* Creator studio routes (auth + role) */}
<Route path="/studio" element={<CreatorLayout />}>
  <Route path="dashboard" element={<CreatorDashboardPage />} />
  <Route path="courses" element={<CoursesPage />} />
  <Route path="courses/:id/edit" element={<CourseEditorPage />} />
  <Route path="payouts" element={<PayoutsPage />} />
  <Route path="connect" element={<ConnectOnboardingPage />} />
  ...
</Route>
```

### ProtectedRoute (extend for role checks)
```tsx
<ProtectedRoute requiredRoles={['owner', 'admin', 'creator']}>
  <CreatorLayout />
</ProtectedRoute>
```

### TenantContext population
- **App routes**: JWT claims → API call to `/api/tenant`
- **Storefront routes**: `CustomDomainTenantMiddleware` on backend resolves tenant from Host header → `GET /api/storefront/branding` returns correct branding

---

## 4. API Client Pattern

### Axios instance with interceptors
```typescript
const api = axios.create({ baseURL: '/api' })

// Auth token
export function setAuthToken(token: string | null) { ... }

// Tenant header
export function setTenantHeader(tenantId: string | null) { ... }

// 401 refresh interceptor (with request queuing)
api.interceptors.response.use(res => res, async (error) => {
    if (error.response?.status === 401) {
        // Refresh token, replay queued requests
    }
})
```

### API namespaces (response unwrapped via `.then(r => r.data)`)
```typescript
export const courseApi = {
    listStorefront: () => api.get<{courses: Course[]}>('/storefront/courses').then(r => r.data),
    create: (data) => api.post<Course>('/creator/courses', data).then(r => r.data),
    ...
}
```

---

## 5. Stripe Webhook Handler Pattern

### Event dispatch + idempotency
```go
func (h *CourseWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    // 1. Verify signature
    event, err := h.connectSvc.VerifyWebhookSignature(payload, signature)

    // 2. Idempotency check
    _, err = h.db.Queries.GetProcessedStripeEvent(ctx, event.ID)
    if err == nil { w.WriteHeader(200); return }  // already processed

    // 3. Dispatch
    switch event.Type {
    case "checkout.session.completed":
        h.handleCheckoutCompleted(w, r, &event)
    case "transfer.reversed":
        h.handleTransferReversed(w, r, &event)
    ...
    }

    // 4. Mark processed
    h.db.Queries.MarkStripeEventProcessed(ctx, gen.MarkStripeEventProcessedParams{
        EventID: event.ID, EventType: string(event.Type),
    })
}
```

### Error handling for webhooks
- Return 200 for: successfully processed, already processed (idempotency), invalid payload, unknown event type
- Return 500 only for transient errors (DB connection lost) — Stripe redelivers on 5xx

---

## 6. Seeding + Bootstrap Pattern

### Planstore seed (3 creator plans with entitlements)
```go
plans := []models.Plan{
    {Name: "Creator Free", Entitlements: map[string]EntitlementValue{
        "max_courses": {Type: Numeric, NumericValue: 1},
        "custom_domain_enabled": {Type: Bool, BoolValue: false},
        "commission_rate_bps": {Type: Numeric, NumericValue: 2000},  // 20%
    }},
    {Name: "Creator Pro", Entitlements: ...},
    {Name: "Creator Business", Entitlements: ...},
}
```

### Configstore seed (dynamic config defaults)
```go
{Key: "course.default_commission_rate_bps", Value: 1000}  // 10%
{Key: "course.payout_schedule", Value: "weekly"}
{Key: "custom_domains.enabled", Value: true}
```

---

## 7. Entitlements Pattern

### Checking entitlements in handlers
```go
plan, _ := h.db.Queries.GetPlanByID(ctx, tenant.PlanID)
// entitlements is JSONB — parse and check
if maxCourses, ok := plan.Entitlements["max_courses"]; ok {
    count, _ := h.db.Queries.CountCoursesByTenant(ctx, tenant.ID)
    if count >= maxCourses.NumericValue {
        apierror.Forbidden(w, r, "Plan limit reached")
        return
    }
}
```

---

## Summary — What This Addendum Adds

1. **Handler pattern**: struct + constructor + create + list + error helpers + route registration
2. **Test pattern**: testutil helpers + embedded-postgres + complete test examples
3. **Frontend routing**: App.tsx structure + ProtectedRoute + TenantContext
4. **API client**: axios interceptors + namespace pattern + response unwrapping
5. **Webhook handler**: idempotency + event dispatch + error handling
6. **Seeding**: 3 creator plans + 7 config defaults
7. **Entitlements**: JSONB map + checking in handlers

**With v4 + this addendum, an engineer can start writing code immediately.**
