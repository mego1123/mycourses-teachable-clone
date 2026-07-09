package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/models"
)

// MustConnectTestDB creates a test database connection.
// NOTE: During migration, this uses the Postgres DB layer.
func MustConnectTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()
	// In production, this would use testcontainers or embedded-postgres
	// For now, return nil — tests should use the e2e_test.go pattern instead
	return nil, func() {}
}

func CreateTestUser(t *testing.T, database *db.DB, email, password, displayName string) *models.User {
	t.Helper()
	ctx := context.Background()
	user, err := database.Queries.CreateUser(ctx, gen.CreateUserParams{
		Email: email, EmailNormalized: email, DisplayName: displayName,
		LocalePreference: "en", ThemePreference: "system", IsActive: true,
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	u := db.ToUser(user)
	return &u
}

func CreateTestTenant(t *testing.T, database *db.DB, name string, ownerID uuid.UUID, isRoot bool) *models.Tenant {
	t.Helper()
	ctx := context.Background()
	tenant, err := database.Queries.CreateTenant(ctx, gen.CreateTenantParams{
		Name: name, Slug: name, BillingStatus: "active", StripeConnectStatus: "pending",
		CommissionRateBps: 1000, PayoutSchedule: "weekly", IsRoot: isRoot, IsActive: true,
	})
	if err != nil {
		t.Fatalf("failed to create test tenant: %v", err)
	}
	t2 := db.ToTenant(tenant)
	return &t2
}

func CreateTestMembership(t *testing.T, database *db.DB, userID, tenantID uuid.UUID, role models.MemberRole) *models.TenantMembership {
	t.Helper()
	ctx := context.Background()
	m, err := database.Queries.CreateMembership(ctx, gen.CreateMembershipParams{
		TenantID: tenantID, UserID: userID, Role: string(role),
	})
	if err != nil {
		t.Fatalf("failed to create test membership: %v", err)
	}
	m2 := db.ToMembership(m)
	return &m2
}

func MarkSystemInitialized(t *testing.T, database *db.DB) {
	t.Helper()
	ctx := context.Background()
	database.Pool.Exec(ctx, "INSERT INTO config_vars (key, value, description, category) VALUES ('system.initialized', 'true', 'System initialized', 'system') ON CONFLICT (key) DO NOTHING")
}

func CleanupCollections(t *testing.T, database *db.DB) {
	t.Helper()
	ctx := context.Background()
	tables := []string{"financial_transactions", "course_progress", "enrollments", "lessons", "sections", "courses",
		"tenant_memberships", "tenants", "plans", "users"}
	for _, table := range tables {
		database.Pool.Exec(ctx, "DELETE FROM "+table)
	}
}

func SetConfigDir(t *testing.T) {}
