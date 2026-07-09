// Package handlers — Payout handler for creator payout management.
package handlers

import (
        "net/http"
        "time"

        "mycourses/internal/db"
        gen "mycourses/internal/db/gen"
        "mycourses/internal/middleware"
        stripeconnect "mycourses/internal/stripe"
)

// PayoutHandler handles creator payout listing and requests.
type PayoutHandler struct {
        db         *db.DB
        connectSvc *stripeconnect.ConnectService
}

func NewPayoutHandler(database *db.DB, connectSvc *stripeconnect.ConnectService) *PayoutHandler {
        return &PayoutHandler{db: database, connectSvc: connectSvc}
}

// List returns the creator's payout history.
func (h *PayoutHandler) List(w http.ResponseWriter, r *http.Request) {
        tenant := middleware.GetPgTenantFromContext(r.Context())
        if tenant == nil {
                respondWithError(w, http.StatusForbidden, "Tenant context required")
                return
        }

        page, limit := parsePagination(r)
        payouts, err := h.db.Queries.ListPayoutsByTenant(r.Context(), gen.ListPayoutsByTenantParams{
                TenantID: tenant.ID,
                Limit:    int32(limit),
                Offset:   int32((page - 1) * limit),
        })
        if err != nil {
                respondWithError(w, http.StatusInternalServerError, "Failed to list payouts")
                return
        }

        totalPaid, _ := h.db.Queries.SumPayoutsByTenant(r.Context(), tenant.ID)

        respondWithJSON(w, http.StatusOK, map[string]interface{}{
                "payouts":    payouts,
                "totalPaid":  totalPaid,
                "page":       page,
                "limit":      limit,
        })
}

// Request creates a manual payout (transfers available balance to creator's bank).
func (h *PayoutHandler) Request(w http.ResponseWriter, r *http.Request) {
        tenant := middleware.GetPgTenantFromContext(r.Context())
        if tenant == nil {
                respondWithError(w, http.StatusForbidden, "Tenant context required")
                return
        }

        // Check Connect status is active
        if tenant.StripeConnectStatus != "active" || tenant.StripeConnectAccountID == nil {
                respondWithError(w, http.StatusForbidden, "Stripe Connect onboarding not complete")
                return
        }

        // Calculate available balance: total course sales - commission - already paid out
        salesRevenue, _ := h.db.Queries.SumTransactionsByTenantAndType(r.Context(), gen.SumTransactionsByTenantAndTypeParams{
                TenantID:  &tenant.ID,
                Type:      "course_purchase",
                CreatedAt: parseTime("2000-01-01"),
                CreatedAt_2: parseTime("2100-01-01"),
        })

        commissionPaid, _ := h.db.Queries.SumTransactionsByTenantAndType(r.Context(), gen.SumTransactionsByTenantAndTypeParams{
                TenantID:  &tenant.ID,
                Type:      "platform_commission",
                CreatedAt: parseTime("2000-01-01"),
                CreatedAt_2: parseTime("2100-01-01"),
        })

        alreadyPaid, _ := h.db.Queries.SumPayoutsByTenant(r.Context(), tenant.ID)

        // Available = sales - commission - already paid out
        available := salesRevenue - commissionPaid - alreadyPaid
        if available <= 0 {
                respondWithError(w, http.StatusBadRequest, "No available balance to payout")
                return
        }

        // Create Stripe transfer
        transferID, err := h.connectSvc.InitiatePayout(r.Context(), tenant.ID.String(), available, "usd")
        if err != nil {
                respondWithError(w, http.StatusInternalServerError, "Failed to initiate payout: "+err.Error())
                return
        }

        // Record payout in database
        var transferIDPtr *string
        if transferID != "" {
                transferIDPtr = &transferID
        }

        payout, err := h.db.Queries.CreatePayout(r.Context(), gen.CreatePayoutParams{
                TenantID:         tenant.ID,
                StripeTransferID: transferIDPtr,
                AmountCents:      available,
                Currency:         "usd",
                Status:           "pending",
        })
        if err != nil {
                respondWithError(w, http.StatusInternalServerError, "Failed to record payout")
                return
        }

        // Record financial transaction
        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                TenantID:    &tenant.ID,
                Type:        "creator_payout",
                AmountCents: -available, // negative — money out
                Currency:    "usd",
                Description: "Manual payout request",
                Metadata:    []byte(`{"payout_id":"` + payout.ID.String() + `"}`),
        })

        respondWithJSON(w, http.StatusCreated, payout)
}

// Helper — parse time string (simple version)
func parseTime(s string) (t time.Time) {
        t, _ = time.Parse("2006-01-02", s)
        return
}
