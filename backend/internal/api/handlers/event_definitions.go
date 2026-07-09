package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

type EventDefinitionsHandler struct{ db *db.DB }

func NewEventDefinitionsHandler(database *db.DB, _ interface{}) *EventDefinitionsHandler {
	return &EventDefinitionsHandler{db: database}
}

func (h *EventDefinitionsHandler) ListEventDefinitions(w http.ResponseWriter, r *http.Request) {
	defs, _ := h.db.Queries.ListEventDefinitions(r.Context(), false)
	respondWithJSON(w, http.StatusOK, defs)
}

func (h *EventDefinitionsHandler) CreateEventDefinition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name, Description, Category string
		IsActive bool
	}
	json.NewDecoder(r.Body).Decode(&req)

	def, err := h.db.Queries.CreateEventDefinition(r.Context(), gen.CreateEventDefinitionParams{
		Name: req.Name, Description: req.Description, Category: req.Category,
		IsActive: req.IsActive, IsSystem: false,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusCreated, def)
}

func (h *EventDefinitionsHandler) UpdateEventDefinition(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	var req struct{ Description string; IsActive bool }
	json.NewDecoder(r.Body).Decode(&req)

	def, err := h.db.Queries.UpdateEventDefinition(r.Context(), gen.UpdateEventDefinitionParams{
		ID: *id, Description: req.Description, IsActive: req.IsActive,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusOK, def)
}

func (h *EventDefinitionsHandler) DeleteEventDefinition(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	h.db.Queries.DeleteEventDefinition(r.Context(), *id)
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *EventDefinitionsHandler) GetSankeyData(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
