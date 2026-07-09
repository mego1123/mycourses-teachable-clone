package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"mycourses/internal/middleware"
)

// getUserIDFromContext extracts the Postgres user UUID from the request context.
// Uses the AuthBridge middleware's context key.
func getUserIDFromContext(r *http.Request) *uuid.UUID {
	userID := middleware.GetPgUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		return nil
	}
	return &userID
}

// getTenantIDFromContext extracts the Postgres tenant UUID from the request context.
// Checks both AuthBridge (from membership) and CustomDomainTenantMiddleware (from domain).
func getTenantIDFromContext(r *http.Request) *uuid.UUID {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuid.Nil {
		return nil
	}
	return &tenantID
}

func uuidNil() uuid.UUID { return uuid.Nil }

func strPtr(s string) *string { return &s }
