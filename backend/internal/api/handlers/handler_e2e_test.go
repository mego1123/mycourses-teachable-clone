// Handler E2E tests: start embedded Postgres, apply migrations, start HTTP server,
// and test the course API endpoints end-to-end.
package handlers_test

import (
        "bytes"
        "context"
        "encoding/json"
        "fmt"
        "io"
        "log"
        "net/http"
        "net/http/httptest"
        "os"
        "path/filepath"
        "testing"
        "time"

        "github.com/fergusstrange/embedded-postgres"
        "github.com/gorilla/mux"
        "github.com/jackc/pgx/v5/pgxpool"

        "mycourses/internal/api/handlers"
        "mycourses/internal/db"
        gen "mycourses/internal/db/gen"
        "mycourses/internal/middleware"
)

func init() { log.SetOutput(io.Discard) }

type testEnv struct {
        server *httptest.Server
        pgDB   *db.DB
        cleanup func()
}

func setupHandlerTestEnv(t *testing.T) *testEnv {
        t.Helper()
        port := uint32(55433)
        dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mc-handler-test-%d", time.Now().UnixNano()))
        pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
                Port(port).Username("mc").Password("mc").Database("mctest").
                DataPath(filepath.Join(dataPath, "data")).RuntimePath(filepath.Join(dataPath, "runtime")))
        if err := pg.Start(); err != nil {
                t.Fatalf("failed to start embedded postgres: %v", err)
        }

        connStr := fmt.Sprintf("postgres://mc:mc@localhost:%d/mctest?sslmode=disable", port)
        ctx := context.Background()
        pgDB, err := db.New(ctx, db.Config{URL: connStr, MaxConns: 5, MinConns: 1})
        if err != nil { pg.Stop(); t.Fatalf("failed to connect: %v", err) }

        if err := pgDB.Migrate(ctx, "file://../../../migrations"); err != nil {
                pgDB.Close(); pg.Stop(); t.Fatalf("failed to migrate: %v", err)
        }

        // Set up router with course routes
        router := mux.NewRouter()
        router.Use(middleware.CustomDomainTenantMiddleware(pgDB, "mycourses.test"))

        courseHandler := handlers.NewCourseHandler(pgDB)
        sectionHandler := handlers.NewSectionHandler(pgDB)
        lessonHandler := handlers.NewLessonHandler(pgDB)
        enrollmentHandler := handlers.NewEnrollmentHandler(pgDB)
        progressHandler := handlers.NewProgressHandler(pgDB)
        _ = handlers.NewCouponHandler(pgDB)
        _ = handlers.NewReviewHandler(pgDB)
        _ = handlers.NewCustomDomainHandler(pgDB, nil)
        certificateHandler := handlers.NewCertificateHandler(pgDB)

        api := router.PathPrefix("/api").Subrouter()

        // Storefront routes
        storefront := api.PathPrefix("/storefront").Subrouter()
        storefront.HandleFunc("/courses", courseHandler.ListStorefront).Methods("GET")
        storefront.HandleFunc("/courses/{slug}", courseHandler.GetStorefront).Methods("GET")
        storefront.HandleFunc("/marketplace", courseHandler.ListMarketplace).Methods("GET")

        // Creator routes
        creator := api.PathPrefix("/creator").Subrouter()
        creator.HandleFunc("/courses", courseHandler.Create).Methods("POST")
        creator.HandleFunc("/courses", courseHandler.ListByCreator).Methods("GET")
        creator.HandleFunc("/courses/{id}", courseHandler.Get).Methods("GET")
        creator.HandleFunc("/courses/{id}", courseHandler.Update).Methods("PUT")
        creator.HandleFunc("/courses/{id}", courseHandler.Delete).Methods("DELETE")
        creator.HandleFunc("/courses/{id}/publish", courseHandler.Publish).Methods("POST")
        creator.HandleFunc("/courses/{id}/unpublish", courseHandler.Unpublish).Methods("POST")
        creator.HandleFunc("/courses/{courseId}/sections", sectionHandler.Create).Methods("POST")
        creator.HandleFunc("/courses/{courseId}/sections", sectionHandler.ListByCourse).Methods("GET")
        creator.HandleFunc("/courses/{courseId}/lessons", lessonHandler.ListByCourse).Methods("GET")
        creator.HandleFunc("/sections/{sectionId}/lessons", lessonHandler.Create).Methods("POST")

        // Learner routes
        learner := api.PathPrefix("/learner").Subrouter()
        learner.HandleFunc("/enrollments/{courseId}", enrollmentHandler.CreateEnrollment).Methods("POST")
        learner.HandleFunc("/progress/{lessonId}", progressHandler.Update).Methods("POST")

        // Public certificate verification
        api.HandleFunc("/public/certificates/verify/{token}", certificateHandler.Verify).Methods("GET")

        server := httptest.NewServer(router)

        cleanup := func() {
                server.Close()
                pgDB.Close()
                pg.Stop()
                os.RemoveAll(dataPath)
        }

        return &testEnv{server: server, pgDB: pgDB, cleanup: cleanup}
}

// seedTenantForHandler creates a tenant and returns it, injecting the tenant into
// requests via the middleware context (simulated by setting the Host header).
func seedTenantForHandler(t *testing.T, pgDB *db.DB, slug string) gen.Tenant {
        ctx := context.Background()
        plan, _ := pgDB.Queries.CreatePlan(ctx, gen.CreatePlanParams{
                Name: "Plan-" + slug, MonthlyPriceCents: 2900, Currency: "usd",
                IncludedSeats: 1, MinSeats: 1, MaxSeats: 1, Entitlements: []byte("{}"),
                IsActive: true, IsPublic: true,
        })
        tenant, err := pgDB.Queries.CreateTenant(ctx, gen.CreateTenantParams{
                Name: "Test School", Slug: slug, PlanID: &plan.ID,
                BillingStatus: "active", StripeConnectStatus: "active",
                CommissionRateBps: 1000, PayoutSchedule: "weekly", IsActive: true,
        })
        if err != nil { t.Fatalf("failed to create tenant: %v", err) }
        return tenant
}

// makeStorefrontRequest makes an HTTP request with the subdomain Host header
// so the CustomDomainTenantMiddleware resolves the tenant.
func makeStorefrontRequest(env *testEnv, method, path string, body interface{}) *http.Request {
        var bodyReader io.Reader
        if body != nil {
                b, _ := json.Marshal(body)
                bodyReader = bytes.NewReader(b)
        }
        req, _ := http.NewRequest(method, env.server.URL+path, bodyReader)
        req.Host = "test-school.mycourses.test" // wildcard subdomain
        if body != nil {
                req.Header.Set("Content-Type", "application/json")
        }
        return req
}

func TestHandler_CreateCourse(t *testing.T) {
        env := setupHandlerTestEnv(t)
        defer env.cleanup()
        seedTenantForHandler(t, env.pgDB, "test-school")

        // Create a course
        body := map[string]interface{}{
                "title":      "Go Fundamentals",
                "slug":       "go-fundamentals",
                "priceCents": 9900,
                "currency":   "usd",
        }
        resp, err := env.server.Client().Do(makeStorefrontRequest(env, "POST", "/api/creator/courses", body))
        if err != nil { t.Fatalf("request failed: %v", err) }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusCreated {
                t.Errorf("expected 201, got %d", resp.StatusCode)
        }

        var course gen.Course
        json.NewDecoder(resp.Body).Decode(&course)
        if course.Title != "Go Fundamentals" {
                t.Errorf("expected title 'Go Fundamentals', got %s", course.Title)
        }
        if course.Status != "draft" {
                t.Errorf("expected draft, got %s", course.Status)
        }

        t.Log("✓ Handler: create course via API")
}

func TestHandler_PublishAndStorefront(t *testing.T) {
        env := setupHandlerTestEnv(t)
        defer env.cleanup()
        seedTenantForHandler(t, env.pgDB, "test-school")

        // Create a course
        createBody := map[string]interface{}{
                "title": "Published Course", "slug": "published-course", "priceCents": 4900, "currency": "usd",
        }
        resp, _ := env.server.Client().Do(makeStorefrontRequest(env, "POST", "/api/creator/courses", createBody))
        var course gen.Course
        json.NewDecoder(resp.Body).Decode(&course)
        resp.Body.Close()

        // Publish it
        resp, _ = env.server.Client().Do(makeStorefrontRequest(env, "POST", "/api/creator/courses/"+course.ID.String()+"/publish", nil))
        if resp.StatusCode != http.StatusOK {
                t.Errorf("expected 200 on publish, got %d", resp.StatusCode)
        }
        resp.Body.Close()

        // Fetch from storefront
        resp, _ = env.server.Client().Do(makeStorefrontRequest(env, "GET", "/api/storefront/courses/published-course", nil))
        if resp.StatusCode != http.StatusOK {
                t.Errorf("expected 200 on storefront fetch, got %d", resp.StatusCode)
        }
        var fetched gen.Course
        json.NewDecoder(resp.Body).Decode(&fetched)
        resp.Body.Close()
        if fetched.Title != "Published Course" {
                t.Errorf("expected 'Published Course', got %s", fetched.Title)
        }

        // List storefront courses
        resp, _ = env.server.Client().Do(makeStorefrontRequest(env, "GET", "/api/storefront/courses", nil))
        if resp.StatusCode != http.StatusOK {
                t.Errorf("expected 200 on storefront list, got %d", resp.StatusCode)
        }
        var listResp struct{ Courses []gen.Course }
        json.NewDecoder(resp.Body).Decode(&listResp)
        resp.Body.Close()
        if len(listResp.Courses) != 1 {
                t.Errorf("expected 1 course in storefront, got %d", len(listResp.Courses))
        }

        t.Log("✓ Handler: create → publish → fetch from storefront → list storefront")
}

func TestHandler_SectionAndLesson(t *testing.T) {
        env := setupHandlerTestEnv(t)
        defer env.cleanup()
        seedTenantForHandler(t, env.pgDB, "test-school")

        // Create course
        resp, _ := env.server.Client().Do(makeStorefrontRequest(env, "POST", "/api/creator/courses", map[string]interface{}{
                "title": "Course With Sections", "slug": "sections-course", "currency": "usd",
        }))
        var course gen.Course
        json.NewDecoder(resp.Body).Decode(&course)
        resp.Body.Close()

        // Create section
        resp, _ = env.server.Client().Do(makeStorefrontRequest(env, "POST", "/api/creator/courses/"+course.ID.String()+"/sections", map[string]interface{}{
                "title": "Section 1", "sortOrder": 0,
        }))
        if resp.StatusCode != http.StatusCreated {
                t.Errorf("expected 201 on section create, got %d", resp.StatusCode)
        }
        var section gen.Section
        json.NewDecoder(resp.Body).Decode(&section)
        resp.Body.Close()

        // Create lesson
        resp, _ = env.server.Client().Do(makeStorefrontRequest(env, "POST", "/api/creator/sections/"+section.ID.String()+"/lessons", map[string]interface{}{
                "title": "Lesson 1", "type": "video", "sortOrder": 0,
        }))
        if resp.StatusCode != http.StatusCreated {
                t.Errorf("expected 201 on lesson create, got %d", resp.StatusCode)
        }
        var lesson gen.Lesson
        json.NewDecoder(resp.Body).Decode(&lesson)
        resp.Body.Close()
        if lesson.Title != "Lesson 1" {
                t.Errorf("expected 'Lesson 1', got %s", lesson.Title)
        }

        // List lessons by course
        resp, _ = env.server.Client().Do(makeStorefrontRequest(env, "GET", "/api/creator/courses/"+course.ID.String()+"/lessons", nil))
        var lessons []gen.Lesson
        json.NewDecoder(resp.Body).Decode(&lessons)
        resp.Body.Close()
        if len(lessons) != 1 {
                t.Errorf("expected 1 lesson, got %d", len(lessons))
        }

        t.Log("✓ Handler: create course → create section → create lesson → list lessons")
}

func TestHandler_MarketplaceListing(t *testing.T) {
        env := setupHandlerTestEnv(t)
        defer env.cleanup()

        // Create 2 tenants with published courses
        for i := 0; i < 2; i++ {
                tenant := seedTenantForHandler(t, env.pgDB, fmt.Sprintf("market-school-%d", i))
                course, _ := env.pgDB.Queries.CreateCourse(context.Background(), gen.CreateCourseParams{
                        TenantID: tenant.ID, Title: fmt.Sprintf("Market Course %d", i),
                        Slug: fmt.Sprintf("market-course-%d", i), Currency: "usd", Status: "draft",
                })
                env.pgDB.Queries.PublishCourse(context.Background(), course.ID)
        }

        // List marketplace (no tenant context — platform-level)
        req, _ := http.NewRequest("GET", env.server.URL+"/api/storefront/marketplace", nil)
        req.Host = "mycourses.test" // platform domain, not a subdomain
        resp, err := env.server.Client().Do(req)
        if err != nil { t.Fatalf("request failed: %v", err) }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                t.Errorf("expected 200, got %d", resp.StatusCode)
        }

        var result struct{ Courses []gen.Course }
        json.NewDecoder(resp.Body).Decode(&result)
        if len(result.Courses) < 2 {
                t.Errorf("expected at least 2 marketplace courses, got %d", len(result.Courses))
        }

        t.Log("✓ Handler: marketplace lists published courses across all tenants")
}

// Ensure pgxpool is imported (used indirectly)
var _ = pgxpool.Pool{}
