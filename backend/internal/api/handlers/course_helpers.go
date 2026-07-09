package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"mycourses/internal/middleware"
)

// getUserIDFromContext extracts the authenticated user ID from the request context.
// TODO: wire to auth middleware once integrated with Postgres.
func getUserIDFromContext(r *http.Request) *uuid.UUID {
	return nil
}

// getTenantIDFromContext extracts the tenant ID from the Postgres-resolved tenant in context.
func getTenantIDFromContext(r *http.Request) *uuid.UUID {
	tenant := middleware.GetPgTenantFromContext(r.Context())
	if tenant == nil {
		return nil
	}
	return &tenant.ID
}

func uuidNil() uuid.UUID { return uuid.Nil }

func strPtr(s string) *string { return &s }
