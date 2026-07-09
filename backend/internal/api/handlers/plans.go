package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/configstore"
	"mycourses/internal/db"
	"mycourses/internal/syslog"
	stripeservice "mycourses/internal/stripe"
)

type PlansHandler struct {
	db     *db.DB
	syslog *syslog.Logger
	store  *configstore.Store
	stripe *stripeservice.Service
}

func NewPlansHandler(database *db.DB, sysLogger *syslog.Logger, store *configstore.Store, stripeSvc *stripeservice.Service) *PlansHandler {
	return &PlansHandler{db: database, syslog: sysLogger, store: store, stripe: stripeSvc}
}

func (h *PlansHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, _ := h.db.Queries.ListActivePublicPlans(r.Context())
	respondWithJSON(w, http.StatusOK, plans)
}

func (h *PlansHandler) ListPlansPublic(w http.ResponseWriter, r *http.Request) {
	h.ListPlans(w, r)
}

func (h *PlansHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	plan, err := h.db.Queries.GetPlanByID(r.Context(), *id)
	if err != nil { respondWithError(w, http.StatusNotFound, "Plan not found"); return }

	respondWithJSON(w, http.StatusOK, plan)
}

func (h *PlansHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Plan creation via API not yet implemented — use seeding")
}

func (h *PlansHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Not yet implemented")
}

func (h *PlansHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }
	h.db.Pool.Exec(r.Context(), "DELETE FROM plans WHERE id = $1 AND is_system = FALSE", *id)
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *PlansHandler) ArchivePlan(w http.ResponseWriter, r *http.Request) { respondWithError(w, http.StatusNotImplemented, "Not yet") }
func (h *PlansHandler) UnarchivePlan(w http.ResponseWriter, r *http.Request) { respondWithError(w, http.StatusNotImplemented, "Not yet") }
func (h *PlansHandler) AssignPlan(w http.ResponseWriter, r *http.Request) { respondWithError(w, http.StatusNotImplemented, "Not yet") }
func (h *PlansHandler) ListEntitlementKeys(w http.ResponseWriter, r *http.Request) { respondWithJSON(w, http.StatusOK, []interface{}{}) }
