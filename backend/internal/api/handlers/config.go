package handlers

import (
	"net/http"

	"mycourses/internal/configstore"
	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type ConfigHandler struct {
	db     *db.DB
	store  *configstore.Store
	syslog *syslog.Logger
}

func NewConfigHandler(database *db.DB, store *configstore.Store, sysLogger *syslog.Logger) *ConfigHandler {
	return &ConfigHandler{db: database, store: store, syslog: sysLogger}
}

func (h *ConfigHandler) ListConfig(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *ConfigHandler) ResetConfig(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}

func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}

func (h *ConfigHandler) CreateConfig(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented) }
func (h *ConfigHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented) }
