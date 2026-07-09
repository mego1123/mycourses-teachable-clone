package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/models"
)

const (
	TenantContextKey     contextKey = "tenant"
	MembershipContextKey contextKey = "membership"
)

type TenantMiddleware struct {
	db *db.DB
}

func NewTenantMiddleware(database *db.DB) *TenantMiddleware {
	return &TenantMiddleware{db: database}
}

func (m *TenantMiddleware) RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If API key auth already populated tenant context, pass through
		if _, ok := GetTenantFromContext(r.Context()); ok {
			if _, ok := GetMembershipFromContext(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}
		}

		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			http.Error(w, `{"error":"X-Tenant-ID header required"}`, http.StatusBadRequest)
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			http.Error(w, `{"error":"Invalid tenant ID"}`, http.StatusBadRequest)
			return
		}

		// Query Postgres for tenant
		pgTenant, err := m.db.Queries.GetTenantByID(r.Context(), tenantID)
		if err != nil {
			http.Error(w, `{"error":"Tenant not found"}`, http.StatusNotFound)
			return
		}

		tenant := db.ToTenant(pgTenant)

		if !tenant.IsActive {
			http.Error(w, `{"error":"Tenant is not active"}`, http.StatusForbidden)
			return
		}

		user, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, `{"error":"Not authenticated"}`, http.StatusUnauthorized)
			return
		}

		// Query Postgres for membership
		pgMembership, err := m.db.Queries.GetMembership(r.Context(), gen.GetMembershipParams{
			TenantID: tenantID,
			UserID:   user.ID,
		})
		if err != nil {
			http.Error(w, `{"error":"Not a member of this tenant"}`, http.StatusForbidden)
			return
		}

		membership := db.ToMembership(pgMembership)

		ctx := context.WithValue(r.Context(), TenantContextKey, &tenant)
		ctx = context.WithValue(ctx, MembershipContextKey, &membership)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetTenantFromContext(ctx context.Context) (*models.Tenant, bool) {
	tenant, ok := ctx.Value(TenantContextKey).(*models.Tenant)
	return tenant, ok
}

func GetMembershipFromContext(ctx context.Context) (*models.TenantMembership, bool) {
	membership, ok := ctx.Value(MembershipContextKey).(*models.TenantMembership)
	return membership, ok
}

// RequireActiveBilling returns middleware that blocks requests when the tenant's
// billing status is not active (and not waived/root).
func RequireActiveBilling() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := GetTenantFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"No tenant context"}`, http.StatusBadRequest)
				return
			}

			if tenant.IsRoot || tenant.BillingWaived {
				next.ServeHTTP(w, r)
				return
			}

			if tenant.BillingStatus == models.BillingStatusActive || tenant.BillingStatus == models.BillingStatusNone {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, `{"error":"subscription_required","code":"BILLING_INACTIVE"}`, http.StatusPaymentRequired)
		})
	}
}

// RequireEntitlement returns middleware that checks whether the tenant's plan
// grants a specific boolean entitlement.
func RequireEntitlement(database *db.DB, feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := GetTenantFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"No tenant context"}`, http.StatusBadRequest)
				return
			}

			if tenant.IsRoot || tenant.BillingWaived {
				next.ServeHTTP(w, r)
				return
			}

			if tenant.PlanID == nil {
				http.Error(w, fmt.Sprintf(`{"error":"Feature '%s' requires an active plan","code":"ENTITLEMENT_REQUIRED"}`, feature), http.StatusForbidden)
				return
			}

			pgPlan, err := database.Queries.GetPlanByID(r.Context(), *tenant.PlanID)
			if err != nil {
				http.Error(w, `{"error":"Plan not found"}`, http.StatusInternalServerError)
				return
			}

			plan := db.ToPlan(pgPlan)

			ent, exists := plan.Entitlements[feature]
			if !exists || (ent.Type == models.EntitlementTypeBool && !ent.BoolValue) {
				http.Error(w, fmt.Sprintf(`{"error":"Feature '%s' is not included in your plan","code":"ENTITLEMENT_REQUIRED"}`, feature), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
