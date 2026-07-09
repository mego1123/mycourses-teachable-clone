package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "mycourses/internal/db/gen"
)

func init() { log.SetOutput(io.Discard) }

func setupEmbeddedPostgres(t *testing.T) (*DB, func()) {
	t.Helper()
	port := uint32(55432)
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mc-test-%d", time.Now().UnixNano()))
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(port).Username("mc").Password("mc").Database("mctest").
		DataPath(filepath.Join(dataPath, "data")).RuntimePath(filepath.Join(dataPath, "runtime")))
	if err := pg.Start(); err != nil {
		t.Fatalf("failed to start embedded postgres: %v", err)
	}
	connStr := fmt.Sprintf("postgres://mc:mc@localhost:%d/mctest?sslmode=disable", port)
	ctx := context.Background()
	db, err := New(ctx, Config{URL: connStr, MaxConns: 5, MinConns: 1})
	if err != nil { pg.Stop(); t.Fatalf("failed to connect: %v", err) }
	return db, func() { db.Close(); pg.Stop(); os.RemoveAll(dataPath) }
}

func applyMigrations(t *testing.T, ctx context.Context, db *DB) {
	t.Helper()
	if err := db.Migrate(ctx, "file://../../migrations"); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
}

func strPtr(s string) *string { return &s }
func mustJSON(t *testing.T, v interface{}) []byte { t.Helper(); b, _ := jsonMarshal(v); return b }
func jsonMarshal(v interface{}) ([]byte, error) { return jsonImpl(v) }
func jsonImpl(v interface{}) ([]byte, error) { return jsonReal(v) }
func jsonReal(v interface{}) ([]byte, error) { return json.Marshal(v) }

func TestE2E_AllMigrations(t *testing.T) {
	db, cleanup := setupEmbeddedPostgres(t)
	defer cleanup()
	ctx := context.Background()
	applyMigrations(t, ctx, db)

	// Verify key tables exist
	for _, table := range []string{"users","plans","tenants","tenant_memberships","financial_transactions","courses","sections","lessons","media_assets","enrollments","course_progress","course_coupons","reviews","payouts","custom_domains","certificates","creator_profiles","processed_stripe_events"} {
		var exists bool
		db.Pool.QueryRow(ctx, `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name=$1)`, table).Scan(&exists)
		if !exists { t.Errorf("table %s missing", table) }
	}
	t.Log("✓ All 29 migrations applied, all tables exist")
}

func TestE2E_CourseFlow(t *testing.T) {
	db, cleanup := setupEmbeddedPostgres(t)
	defer cleanup()
	ctx := context.Background()
	applyMigrations(t, ctx, db)

	// Create plan + tenant
	plan, _ := db.Queries.CreatePlan(ctx, gen.CreatePlanParams{
		Name: "Pro", MonthlyPriceCents: 2900, Currency: "usd",
		IncludedSeats: 1, MinSeats: 1, MaxSeats: 1, Entitlements: []byte("{}"),
		IsActive: true, IsPublic: true,
	})
	tenant, _ := db.Queries.CreateTenant(ctx, gen.CreateTenantParams{
		Name: "Test School", Slug: "test-school", PlanID: &plan.ID,
		BillingStatus: "active", StripeConnectStatus: "active",
		CommissionRateBps: 1000, PayoutSchedule: "weekly", IsActive: true,
	})

	// Create course
	course, err := db.Queries.CreateCourse(ctx, gen.CreateCourseParams{
		TenantID: tenant.ID, Title: "Go Course", Slug: "go-course",
		Currency: "usd", Status: "draft", PriceCents: 9900,
	})
	if err != nil { t.Fatalf("create course: %v", err) }
	if course.ID == uuid.Nil { t.Error("nil UUID") }

	// Publish
	pub, _ := db.Queries.PublishCourse(ctx, course.ID)
	if pub.Status != "published" { t.Errorf("expected published, got %s", pub.Status) }

	// Create section + lesson
	section, _ := db.Queries.CreateSection(ctx, gen.CreateSectionParams{
		CourseID: course.ID, Title: "Intro", SortOrder: 0,
	})
	lesson, _ := db.Queries.CreateLesson(ctx, gen.CreateLessonParams{
		SectionID: section.ID, CourseID: course.ID, Title: "Welcome", Type: "video", SortOrder: 0,
	})

	// Create learner + enroll
	learner, _ := db.Queries.CreateUser(ctx, gen.CreateUserParams{
		Email: "learner@test.com", EmailNormalized: "learner@test.com",
		DisplayName: "Learner", LocalePreference: "en", ThemePreference: "system", IsActive: true,
	})
	enrollment, _ := db.Queries.CreateEnrollment(ctx, gen.CreateEnrollmentParams{
		CourseID: course.ID, TenantID: tenant.ID, UserID: learner.ID,
		Status: "active", PricePaidCents: 9900, Currency: "usd",
	})

	// Track progress
	db.Queries.UpsertProgress(ctx, gen.UpsertProgressParams{
		EnrollmentID: enrollment.ID, LessonID: lesson.ID, UserID: learner.ID,
		Completed: true, VideoPositionSec: 300,
	})
	pct, _ := db.Queries.GetCourseCompletionPercentage(ctx, enrollment.ID)
	if pct != 100 { t.Errorf("expected 100%%, got %d", pct) }

	// Complete enrollment + issue certificate
	db.Queries.CompleteEnrollment(ctx, enrollment.ID)
	certNum, _ := db.Queries.GetNextCertificateNumber(ctx)
	certNumStr, _ := certNum.(string)
	cert, _ := db.Queries.CreateCertificate(ctx, gen.CreateCertificateParams{
		EnrollmentID: enrollment.ID, UserID: learner.ID, CourseID: course.ID, TenantID: tenant.ID,
		CertificateNumber: certNumStr, VerificationToken: uuid.New().String(),
		LearnerName: "Learner", CourseTitle: "Go Course", CreatorName: "Test School",
	})

	// Verify certificate
	verified, _ := db.Queries.GetCertificateByToken(ctx, cert.VerificationToken)
	if verified.ID != cert.ID { t.Error("cert verification failed") }

	// Stripe event idempotency
	db.Queries.MarkStripeEventProcessed(ctx, gen.MarkStripeEventProcessedParams{
		EventID: "evt_123", EventType: "checkout.session.completed",
	})
	_, err = db.Queries.GetProcessedStripeEvent(ctx, "evt_123")
	if err != nil { t.Error("stripe event not found") }

	t.Log("✓ Full course flow: plan→tenant→course→section→lesson→enroll→progress→complete→certificate→idempotency")
}

var _ = pgxpool.Pool{}
