package handlers

import (
	"context"
	"net/http"

	"mycourses/internal/configstore"
	"mycourses/internal/db"
	"mycourses/internal/events"
	"mycourses/internal/syslog"
	"mycourses/internal/telemetry"
	stripeservice "mycourses/internal/stripe"
)

type BillingHandler struct {
	stripeSvc   *stripeservice.Service
	db          *db.DB
	emitter     events.Emitter
	syslog      *syslog.Logger
	store       *configstore.Store
	telemetrySvc *telemetry.Service
}

func NewBillingHandler(stripeSvc *stripeservice.Service, database *db.DB, emitter events.Emitter, sysLogger *syslog.Logger, store *configstore.Store) *BillingHandler {
	return &BillingHandler{stripeSvc: stripeSvc, db: database, emitter: emitter, syslog: sysLogger, store: store}
}

func (h *BillingHandler) SetTelemetry(svc *telemetry.Service) { h.telemetrySvc = svc }

func (h *BillingHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Billing handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) Portal(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) GetInvoicePDF(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) AdminListTransactions(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) AdminGetMetrics(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) AdminCancelSubscription(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) AdminUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BillingHandler) computeLiveMetric(ctx context.Context, metric, dateStr string) int64 { return 0 }
func (h *BillingHandler) computeLiveRevenue(ctx context.Context, dateStr string) int64 { return 0 }
func (h *BillingHandler) computeLiveARR(ctx context.Context) int64 { return 0 }
