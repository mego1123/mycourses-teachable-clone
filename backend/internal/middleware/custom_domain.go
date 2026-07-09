// Package middleware — Custom domain tenant resolution for Postgres-based handlers.
// Resolves tenant from wildcard subdomain (john.mycourses.com) or custom domain.
// Uses a separate context key from the existing MongoDB-based tenant middleware
// so both can coexist during the migration period.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

// PgTenantContextKey is the context key for Postgres-resolved tenants.
// Separate from the existing TenantContextKey (which holds *models.Tenant) to avoid type conflicts.
type PgTenantContextKey string

const PgTenantKey PgTenantContextKey = "pg_tenant"

// CustomDomainTenantMiddleware resolves the tenant from the Host header using Postgres queries.
func CustomDomainTenantMiddleware(database *db.DB, platformDomain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			if idx := strings.LastIndex(host, ":"); idx != -1 {
				host = host[:idx]
			}

			var tenant *gen.Tenant

			// Case 1: Wildcard subdomain (john.mycourses.com)
			if strings.HasSuffix(host, "."+platformDomain) {
				slug := strings.TrimSuffix(host, "."+platformDomain)
				if slug != "" && slug != "www" && slug != "api" {
					t, err := database.Queries.GetTenantBySlug(r.Context(), slug)
					if err == nil {
						tenant = &t
					}
				}
			}

			// Case 2: Custom domain (academy.johndoe.com)
			if tenant == nil && host != platformDomain && host != "www."+platformDomain && !strings.HasSuffix(host, "."+platformDomain) {
				t, err := database.Queries.GetTenantByCustomDomain(r.Context(), host)
				if err == nil {
					tenant = &t
				}
			}

			if tenant != nil {
				ctx := context.WithValue(r.Context(), PgTenantKey, tenant)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetPgTenantFromContext extracts the Postgres-resolved tenant from the request context.
func GetPgTenantFromContext(ctx context.Context) *gen.Tenant {
	if v, ok := ctx.Value(PgTenantKey).(*gen.Tenant); ok {
		return v
	}
	return nil
}
