// Package handlers — Stripe Connect handler for creator onboarding and status.
package handlers

import (
	"net/http"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
	stripeconnect "mycourses/internal/stripe"
)

// ConnectHandler handles Stripe Connect onboarding and status.
type ConnectHandler struct {
	db          *db.DB
	connectSvc  *stripeconnect.ConnectService
	frontendURL string
}

func NewConnectHandler(database *db.DB, connectSvc *stripeconnect.ConnectService, frontendURL string) *ConnectHandler {
	return &ConnectHandler{db: database, connectSvc: connectSvc, frontendURL: frontendURL}
}

// Onboard returns a Stripe AccountLink URL for the creator to complete onboarding.
func (h *ConnectHandler) Onboard(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetPgTenantFromContext(r.Context())
	if tenant == nil {
		respondWithError(w, http.StatusForbidden, "Tenant context required")
		return
	}

	returnURL := h.frontendURL + "/studio/connect?status=return"
	refreshURL := h.frontendURL + "/studio/connect?status=refresh"

	url, err := h.connectSvc.OnboardCreator(r.Context(), tenant.ID.String(), returnURL, refreshURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to start onboarding: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"url": url})
}

// Status returns the creator's Stripe Connect account status.
func (h *ConnectHandler) Status(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetPgTenantFromContext(r.Context())
	if tenant == nil {
		respondWithError(w, http.StatusForbidden, "Tenant context required")
		return
	}

	// If no account ID yet, return not-connected status
	if tenant.StripeConnectAccountID == nil || *tenant.StripeConnectAccountID == "" {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"connected":    false,
			"status":       "not_connected",
			"accountId":    "",
			"chargesEnabled": false,
			"payoutsEnabled": false,
		})
		return
	}

	// Query Stripe for live status
	chargesEnabled, payoutsEnabled, detailsSubmitted, err := h.connectSvc.GetAccountStatus(*tenant.StripeConnectAccountID)
	if err != nil {
		// Fall back to stored status
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"connected":       true,
			"status":          tenant.StripeConnectStatus,
			"accountId":       *tenant.StripeConnectAccountID,
			"chargesEnabled":  false,
			"payoutsEnabled":  false,
			"detailsSubmitted": false,
		})
		return
	}

	// Determine status
	status := "pending"
	if detailsSubmitted && chargesEnabled && payoutsEnabled {
		status = "active"
	} else if detailsSubmitted {
		status = "in_review"
	}

	// Update stored status if changed
	if status != tenant.StripeConnectStatus {
		h.db.Queries.UpdateTenantStripeConnect(r.Context(), gen.UpdateTenantStripeConnectParams{
			ID:                     tenant.ID,
			StripeConnectAccountID: tenant.StripeConnectAccountID,
			StripeConnectStatus:    status,
		})
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"connected":       true,
		"status":          status,
		"accountId":       *tenant.StripeConnectAccountID,
		"chargesEnabled":  chargesEnabled,
		"payoutsEnabled":  payoutsEnabled,
		"detailsSubmitted": detailsSubmitted,
	})
}
