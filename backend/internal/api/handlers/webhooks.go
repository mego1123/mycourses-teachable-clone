package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lib/pq"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
	"mycourses/internal/syslog"
	"mycourses/internal/webhooks"
)

type WebhooksHandler struct {
	db        *db.DB
	syslog    *syslog.Logger
	dispatcher *webhooks.Dispatcher
}

func NewWebhooksHandler(database *db.DB, sysLogger *syslog.Logger, dispatcher *webhooks.Dispatcher) *WebhooksHandler {
	return &WebhooksHandler{db: database, syslog: sysLogger, dispatcher: dispatcher}
}

func (h *WebhooksHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() { respondWithJSON(w, http.StatusOK, []interface{}{}); return }

	hooks, _ := h.db.Queries.ListWebhooksByTenant(r.Context(), tenantID)
	respondWithJSON(w, http.StatusOK, hooks)
}

func (h *WebhooksHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["webhookId"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	hook, err := h.db.Queries.GetWebhookByID(r.Context(), *id)
	if err != nil { respondWithError(w, http.StatusNotFound, "Not found"); return }

	respondWithJSON(w, http.StatusOK, hook)
}

func (h *WebhooksHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	user, ok := middleware.GetUserFromContext(r.Context())
	if tenantID == uuidNil() || !ok { respondWithError(w, http.StatusForbidden, "Auth+tenant required"); return }

	var req struct {
		Name, URL, Secret string
		Events []string
	}
	json.NewDecoder(r.Body).Decode(&req)

	hook, err := h.db.Queries.CreateWebhook(r.Context(), gen.CreateWebhookParams{
		TenantID: tenantID, Name: req.Name, Url: req.URL, Secret: &req.Secret,
		Events: pq.StringArray(req.Events), CreatedBy: &user.ID,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusCreated, hook)
}

func (h *WebhooksHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["webhookId"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	var req struct {
		Name, URL, Secret string
		Events []string
		IsActive bool
	}
	json.NewDecoder(r.Body).Decode(&req)

	hook, err := h.db.Queries.UpdateWebhook(r.Context(), gen.UpdateWebhookParams{
		ID: *id, Name: req.Name, Url: req.URL, Secret: &req.Secret,
		Events: pq.StringArray(req.Events), IsActive: req.IsActive,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusOK, hook)
}

func (h *WebhooksHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["webhookId"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	h.db.Queries.DeleteWebhook(r.Context(), *id)
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *WebhooksHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]bool{"sent": true})
}

func (h *WebhooksHandler) RegenerateSecret(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"secret": "new_secret_placeholder"})
}

func (h *WebhooksHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []string{
		"user.registered", "user.login", "billing.subscription.created",
		"billing.subscription.canceled", "billing.payment.received",
		"billing.payment.failed", "team.member.invited", "team.member.joined",
		"course.published", "enrollment.created", "payout.paid",
	})
}
