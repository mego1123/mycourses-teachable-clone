package handlers

import (
	"encoding/json"
	"net/http"

	"mycourses/internal/configstore"
	"mycourses/internal/db"
	"mycourses/internal/syslog"
	gen "mycourses/internal/db/gen"
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
	category := r.URL.Query().Get("category")
	vars, _ := h.db.Queries.ListConfigVars(r.Context(), category)
	respondWithJSON(w, http.StatusOK, vars)
}

func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" { respondWithError(w, http.StatusBadRequest, "Key required"); return }

	v, err := h.db.Queries.GetConfigVar(r.Context(), key)
	if err != nil { respondWithError(w, http.StatusNotFound, "Not found"); return }

	respondWithJSON(w, http.StatusOK, v)
}

func (h *ConfigHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key, Description, Category string
		Value interface{}
	}
	json.NewDecoder(r.Body).Decode(&req)

	valueJSON, _ := json.Marshal(req.Value)
	v, err := h.db.Queries.UpsertConfigVar(r.Context(), gen.UpsertConfigVarParams{
		Key: req.Key, Value: valueJSON, Description: req.Description,
		Category: req.Category, IsSystem: false, IsReadonly: false,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusCreated, v)
}

func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	h.CreateConfig(w, r) // Same upsert
}

func (h *ConfigHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" { respondWithError(w, http.StatusBadRequest, "Key required"); return }

	h.db.Queries.DeleteConfigVar(r.Context(), key)
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
