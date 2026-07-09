package handlers

import (
	"net/http"

	stripeservice "mycourses/internal/stripe"
	"mycourses/internal/db"
	"mycourses/internal/events"
	"mycourses/internal/syslog"
)

type WebhookHandler struct {
	db        *db.DB
	stripeSvc *stripeservice.Service
	emitter   events.Emitter
	syslog    *syslog.Logger
	getConfig func(string) string
}

func NewWebhookHandler(stripeSvc *stripeservice.Service, database *db.DB, emitter events.Emitter, sysLogger *syslog.Logger, getConfig func(string) string) *WebhookHandler {
	return &WebhookHandler{stripeSvc: stripeSvc, db: database, emitter: emitter, syslog: sysLogger, getConfig: getConfig}
}

func (h *WebhookHandler) SetTelemetry(svc interface{}) {}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Webhook handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
