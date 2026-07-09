package handlers

import (
	"net/http"

	"mycourses/internal/db"
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
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *WebhooksHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *WebhooksHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *WebhooksHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *WebhooksHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *WebhooksHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *WebhooksHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *WebhooksHandler) RegenerateSecret(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
