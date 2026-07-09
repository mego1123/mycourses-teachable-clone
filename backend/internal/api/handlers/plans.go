package handlers

import (
	"net/http"

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
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *PlansHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *PlansHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *PlansHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *PlansHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}

func (h *PlansHandler) ListPlansPublic(w http.ResponseWriter, r *http.Request) { respondWithJSON(w, http.StatusOK, []interface{}{}) }

func (h *PlansHandler) ListEntitlementKeys(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}

func (h *PlansHandler) ArchivePlan(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented) }
func (h *PlansHandler) UnarchivePlan(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented) }
func (h *PlansHandler) AssignPlan(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented) }
