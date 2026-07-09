package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/events"
	"mycourses/internal/syslog"
)

type APIKeysHandler struct {
	db     *db.DB
	syslog *syslog.Logger
}

func NewAPIKeysHandler(database *db.DB, emitter events.Emitter, sysLogger *syslog.Logger) *APIKeysHandler {
	return &APIKeysHandler{db: database, syslog: sysLogger}
}

func (h *APIKeysHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *APIKeysHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *APIKeysHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
