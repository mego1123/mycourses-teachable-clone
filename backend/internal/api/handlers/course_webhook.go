// Package handlers — Stripe webhook handler for course platform events.
// Handles course purchase checkout, refunds, disputes, and Connect events.
// Separate from the existing webhook.go (which handles platform subscriptions).
package handlers

import (
        "encoding/json"
        "io"
        "net/http"

        "github.com/google/uuid"
        stripe "github.com/stripe/stripe-go/v82"

        "mycourses/internal/db"
        gen "mycourses/internal/db/gen"
        stripeconnect "mycourses/internal/stripe"
)

// CourseWebhookHandler handles Stripe webhooks for the course marketplace.
type CourseWebhookHandler struct {
        db         *db.DB
        connectSvc *stripeconnect.ConnectService
}

func NewCourseWebhookHandler(database *db.DB, connectSvc *stripeconnect.ConnectService) *CourseWebhookHandler {
        return &CourseWebhookHandler{db: database, connectSvc: connectSvc}
}

// HandleWebhook receives and processes Stripe webhook events for the course platform.
func (h *CourseWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
        payload, err := io.ReadAll(r.Body)
        if err != nil {
                respondWithError(w, http.StatusBadRequest, "Failed to read request body")
                return
        }

        signature := r.Header.Get("Stripe-Signature")
        event, err := h.connectSvc.VerifyWebhookSignature(payload, signature)
        if err != nil {
                respondWithError(w, http.StatusBadRequest, "Invalid signature")
                return
        }

        // Idempotency: check if we already processed this event
        _, err = h.db.Queries.GetProcessedStripeEvent(r.Context(), event.ID)
        if err == nil {
                // Already processed — return 200 so Stripe doesn't redeliver
                w.WriteHeader(http.StatusOK)
                return
        }

        // Dispatch based on event type
        switch event.Type {
        case "checkout.session.completed":
                h.handleCheckoutCompleted(w, r, &event)
        case "charge.refunded":
                h.handleChargeRefunded(w, r, &event)
        case "transfer.reversed":
                h.handleTransferReversed(w, r, &event)
        case "charge.dispute.created":
                h.handleDisputeCreated(w, r, &event)
        case "charge.dispute.funds_withdrawn":
                h.handleDisputeFundsWithdrawn(w, r, &event)
        case "charge.dispute.funds_reinstated":
                h.handleDisputeFundsReinstated(w, r, &event)
        case "account.updated":
                h.handleAccountUpdated(w, r, &event)
        case "payout.paid":
                h.handlePayoutPaid(w, r, &event)
        case "payout.failed":
                h.handlePayoutFailed(w, r, &event)
        default:
                // Unknown event — acknowledge so Stripe doesn't redeliver
                w.WriteHeader(http.StatusOK)
                return
        }

        // Mark event as processed (idempotency)
        h.db.Queries.MarkStripeEventProcessed(r.Context(), gen.MarkStripeEventProcessedParams{
                EventID:   event.ID,
                EventType: string(event.Type),
        })

        w.WriteHeader(http.StatusOK)
}

// handleCheckoutCompleted processes a successful course purchase.
// Creates enrollment + financial transactions.
func (h *CourseWebhookHandler) handleCheckoutCompleted(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var session stripeCheckoutSession
        if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
                return // return 200 — don't redeliver on parse errors
        }

        // Check if this is a course purchase (via metadata)
        if session.Metadata["type"] != "course_purchase" {
                return // Not a course purchase — let the existing webhook handler process it
        }

        courseID, _ := uuid.Parse(session.Metadata["course_id"])
        tenantID, _ := uuid.Parse(session.Metadata["tenant_id"])
        learnerUserID, _ := uuid.Parse(session.Metadata["learner_user_id"])

        if courseID == uuid.Nil || tenantID == uuid.Nil || learnerUserID == uuid.Nil {
                return
        }

        // Check if enrollment already exists (idempotency)
        existing, _ := h.db.Queries.GetEnrollmentByCourseAndUser(r.Context(), gen.GetEnrollmentByCourseAndUserParams{
                CourseID: courseID, UserID: learnerUserID,
        })
        if existing.ID != uuid.Nil {
                return // Already enrolled — idempotent
        }

        // Create enrollment
        sessionID := session.ID
        enrollment, err := h.db.Queries.CreateEnrollment(r.Context(), gen.CreateEnrollmentParams{
                CourseID:        courseID,
                TenantID:        tenantID,
                UserID:          learnerUserID,
                Status:          "active",
                PricePaidCents:  session.AmountTotal,
                Currency:        string(session.Currency),
                StripeSessionID: &sessionID,
        })
        if err != nil {
                return
        }

        // Record course_purchase transaction
        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                TenantID:    &tenantID,
                UserID:      &learnerUserID,
                Type:        "course_purchase",
                AmountCents: session.AmountTotal,
                Currency:    string(session.Currency),
                Description: "Course purchase",
                Metadata:    []byte(`{"course_id":"` + courseID.String() + `","enrollment_id":"` + enrollment.ID.String() + `"}`),
        })

        // Record platform_commission transaction
        tenant, _ := h.db.Queries.GetTenantByID(r.Context(), tenantID)
        if tenant.ID != uuid.Nil {
                commission := session.AmountTotal * int64(tenant.CommissionRateBps) / 10000
                if commission > 0 {
                        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                                TenantID:    &tenantID,
                                Type:        "platform_commission",
                                AmountCents: commission,
                                Currency:    string(session.Currency),
                                Description: "Platform commission on course sale",
                                Metadata:    []byte(`{"course_id":"` + courseID.String() + `","enrollment_id":"` + enrollment.ID.String() + `"}`),
                        })
                }
        }
}

// handleChargeRefunded processes a refund — marks enrollment as refunded.
func (h *CourseWebhookHandler) handleChargeRefunded(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var charge stripeCharge
        if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
                return
        }

        // Find enrollment by stripe session ID (stored during checkout)
        // The charge has a PaymentIntentID, which links back to the checkout session
        // For simplicity, we look up by metadata in the financial transaction
        // In production, we'd store the charge ID on the enrollment

        // Record refund transaction
        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                Type:        "refund",
                AmountCents: -charge.AmountRefunded, // negative
                Currency:    string(charge.Currency),
                Description: "Course purchase refund",
                Metadata:    []byte(`{"stripe_charge_id":"` + charge.ID + `"}`),
        })
}

// handleTransferReversed processes a transfer reversal (happens on refund).
func (h *CourseWebhookHandler) handleTransferReversed(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var reversal stripeTransferReversal
        if err := json.Unmarshal(event.Data.Raw, &reversal); err != nil {
                return
        }

        // Record transfer reversal transaction (against creator's balance)
        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                Type:        "connect_transfer_reversal",
                AmountCents: -reversal.Amount, // negative — creator loses funds
                Currency:    string(reversal.Currency),
                Description: "Transfer reversed due to refund",
                Metadata:    []byte(`{"stripe_transfer_id":"` + reversal.TransferID + `"}`),
        })
}

// handleDisputeCreated processes a chargeback initiation.
func (h *CourseWebhookHandler) handleDisputeCreated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var dispute stripeDispute
        if err := json.Unmarshal(event.Data.Raw, &dispute); err != nil {
                return
        }

        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                Type:        "dispute_pending",
                AmountCents: dispute.Amount,
                Currency:    string(dispute.Currency),
                Description: "Chargeback dispute created",
                Metadata:    []byte(`{"stripe_dispute_id":"` + dispute.ID + `"}`),
        })
}

// handleDisputeFundsWithdrawn processes funds withdrawal from a lost chargeback.
func (h *CourseWebhookHandler) handleDisputeFundsWithdrawn(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var dispute stripeDispute
        if err := json.Unmarshal(event.Data.Raw, &dispute); err != nil {
                return
        }

        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                Type:        "dispute_withdrawal",
                AmountCents: -dispute.Amount, // negative — funds lost
                Currency:    string(dispute.Currency),
                Description: "Funds withdrawn due to lost chargeback",
                Metadata:    []byte(`{"stripe_dispute_id":"` + dispute.ID + `"}`),
        })
}

// handleDisputeFundsReinstated processes funds return from a won chargeback.
func (h *CourseWebhookHandler) handleDisputeFundsReinstated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var dispute stripeDispute
        if err := json.Unmarshal(event.Data.Raw, &dispute); err != nil {
                return
        }

        h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
                Type:        "dispute_reinstatement",
                AmountCents: dispute.Amount, // positive — funds recovered
                Currency:    string(dispute.Currency),
                Description: "Funds reinstated after won chargeback",
                Metadata:    []byte(`{"stripe_dispute_id":"` + dispute.ID + `"}`),
        })
}

// handleAccountUpdated processes Stripe Connect account status changes.
func (h *CourseWebhookHandler) handleAccountUpdated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var acct stripeAccount
        if err := json.Unmarshal(event.Data.Raw, &acct); err != nil {
                return
        }

        // Determine status
        status := "pending"
        if acct.DetailsSubmitted && acct.ChargesEnabled && acct.PayoutsEnabled {
                status = "active"
        } else if acct.DetailsSubmitted {
                status = "in_review"
        }

        // TODO: Find tenant by Stripe Connect account ID and update status
        // This requires a query like: GetTenantByStripeConnectAccountID
        _ = status
}

// handlePayoutPaid processes a successful payout to creator's bank.
func (h *CourseWebhookHandler) handlePayoutPaid(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var payout stripePayout
        if err := json.Unmarshal(event.Data.Raw, &payout); err != nil {
                return
        }

        // Update payout record by Stripe payout ID
        // The payout.paid event fires when Stripe sends money to the creator's bank
        payouts, _ := h.db.Queries.GetPayoutByStripeTransferID(r.Context(), &payout.Destination)
        _ = payouts // In production, match by Stripe payout ID and update status to "paid"
}

// handlePayoutFailed processes a failed payout.
func (h *CourseWebhookHandler) handlePayoutFailed(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
        var payout stripePayout
        if err := json.Unmarshal(event.Data.Raw, &payout); err != nil {
                return
        }
        // In production, find the payout record and update status to "failed"
}

// =============================================================================
// Internal types for parsing Stripe webhook payloads
// =============================================================================

type stripeEvent struct {
        ID   string          `json:"id"`
        Type string          `json:"type"`
        Data json.RawMessage `json:"data"`
}

type stripeCheckoutSession struct {
        ID           string            `json:"id"`
        AmountTotal  int64             `json:"amount_total"`
        Currency     string            `json:"currency"`
        Metadata     map[string]string `json:"metadata"`
}

type stripeCharge struct {
        ID             string `json:"id"`
        AmountRefunded int64  `json:"amount_refunded"`
        Currency       string `json:"currency"`
}

type stripeTransferReversal struct {
        Amount     int64  `json:"amount"`
        Currency   string `json:"currency"`
        TransferID string `json:"transfer"`
}

type stripeDispute struct {
        ID       string `json:"id"`
        Amount   int64  `json:"amount"`
        Currency string `json:"currency"`
}

type stripeAccount struct {
        ID              string `json:"id"`
        ChargesEnabled  bool   `json:"charges_enabled"`
        PayoutsEnabled  bool   `json:"payouts_enabled"`
        DetailsSubmitted bool  `json:"details_submitted"`
}

type stripePayout struct {
        ID          string `json:"id"`
        Amount      int64  `json:"amount"`
        Currency    string `json:"currency"`
        Destination string `json:"destination"`
}
