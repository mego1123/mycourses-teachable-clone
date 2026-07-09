// Package middleware — Auth bridge: translates MongoDB user context to Postgres user ID.
//
// PROBLEM: The existing lastsaas auth middleware stores *models.User (MongoDB)
// in context. The new course handlers use Postgres and need a uuid.UUID.
//
// SOLUTION: This bridge middleware runs AFTER the existing auth middleware.
// It extracts the user's email from the MongoDB user, looks up (or creates)
// that user in the Postgres users table, and stores the Postgres UUID in context.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/uuid"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

// PgUserIDKey is the context key for the Postgres user UUID.
type PgUserIDKey string

const PgUserIDKeyConst PgUserIDKey = "pg_user_id"

// PgTenantIDKey is the context key for the Postgres tenant UUID (from auth, not domain).
type PgTenantIDKey string

const PgTenantIDKeyConst PgTenantIDKey = "pg_tenant_id"

// AuthBridge middleware translates the MongoDB user in context to a Postgres user ID.
// Must run AFTER the existing RequireAuth middleware.
func AuthBridge(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mongoUser, ok := GetUserFromContext(r.Context())
			if !ok || mongoUser == nil {
				next.ServeHTTP(w, r)
				return
			}

			emailNormalized := normalizeEmail(mongoUser.Email)
			if emailNormalized == "" {
				next.ServeHTTP(w, r)
				return
			}

			pgUser, err := database.Queries.GetUserByEmailNormalized(r.Context(), emailNormalized)
			if err != nil {
				// User doesn't exist in Postgres — sync from MongoDB
				pgUser, err = syncUserToPostgres(r.Context(), database, mongoUser, emailNormalized)
				if err != nil {
					slog.Error("Auth bridge: failed to sync user to Postgres",
						"email", emailNormalized, "error", err)
					next.ServeHTTP(w, r)
					return
				}
				slog.Info("Auth bridge: synced user to Postgres",
					"email", emailNormalized, "pg_user_id", pgUser.ID)
			}

			ctx := context.WithValue(r.Context(), PgUserIDKeyConst, pgUser.ID)

			// Resolve tenant from membership
			tenantID := resolveTenantFromMembership(r.Context(), database, pgUser.ID)
			if tenantID != uuid.Nil {
				ctx = context.WithValue(ctx, PgTenantIDKeyConst, tenantID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetPgUserIDFromContext extracts the Postgres user UUID from context.
func GetPgUserIDFromContext(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(PgUserIDKeyConst).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// GetPgTenantIDFromContext extracts the Postgres tenant UUID from context.
// Checks both AuthBridge (from membership) and CustomDomainTenantMiddleware (from domain).
func GetPgTenantIDFromContext(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(PgTenantIDKeyConst).(uuid.UUID); ok {
		return v
	}
	if tenant := GetPgTenantFromContext(ctx); tenant != nil {
		return tenant.ID
	}
	return uuid.Nil
}

// syncUserToPostgres creates a user in Postgres based on the MongoDB user.
func syncUserToPostgres(ctx context.Context, database *db.DB, mongoUser interface{}, emailNormalized string) (gen.User, error) {
	fields := extractMongoUserFields(mongoUser)

	pgUser, err := database.Queries.CreateUser(ctx, gen.CreateUserParams{
		Email:            fields.Email,
		EmailNormalized:  emailNormalized,
		DisplayName:      fields.DisplayName,
		LocalePreference: "en",
		ThemePreference:  "system",
		EmailVerified:    fields.EmailVerified,
		IsActive:         fields.IsActive,
		IsAdmin:          false,
	})
	if err != nil {
		return gen.User{}, err
	}

	return pgUser, nil
}

// extractMongoUserFields uses reflection to get fields from *models.User
// without importing the models package (avoids circular dependency).
func extractMongoUserFields(user interface{}) struct {
	Email         string
	DisplayName   string
	EmailVerified bool
	IsActive      bool
} {
	result := struct {
		Email         string
		DisplayName   string
		EmailVerified bool
		IsActive      bool
	}{}

	v := reflect.ValueOf(user)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		if email := v.FieldByName("Email"); email.IsValid() {
			result.Email = email.String()
		}
		if name := v.FieldByName("DisplayName"); name.IsValid() {
			result.DisplayName = name.String()
		}
		if verified := v.FieldByName("EmailVerified"); verified.IsValid() {
			result.EmailVerified = verified.Bool()
		}
		if active := v.FieldByName("IsActive"); active.IsValid() {
			result.IsActive = active.Bool()
		}
	}

	// Fallback: if Email is empty, try to get it from the email_normalized
	if result.Email == "" {
		result.Email = "" // Will be set by caller
	}

	return result
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func resolveTenantFromMembership(ctx context.Context, database *db.DB, userID uuid.UUID) uuid.UUID {
	memberships, err := database.Queries.ListMembershipsByUser(ctx, userID)
	if err != nil || len(memberships) == 0 {
		return uuid.Nil
	}

	for _, m := range memberships {
		if m.Role == "owner" || m.Role == "admin" || m.Role == "creator" {
			return m.TenantID
		}
	}

	return memberships[0].TenantID
}
