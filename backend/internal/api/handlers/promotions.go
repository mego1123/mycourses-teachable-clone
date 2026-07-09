package handlers

import (
	"net/http"

	"mycourses/internal/configstore"
	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type PromotionsHandler struct {
	db     *db.DB
	syslog *syslog.Logger
}

func NewPromotionsHandler(database *db.DB, stripeSvc interface{}, store *configstore.Store) *PromotionsHandler {
	return &PromotionsHandler{db: database}
}

func (h *PromotionsHandler) ListPromotions(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *PromotionsHandler) CreatePromotion(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *PromotionsHandler) UpdatePromotion(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *PromotionsHandler) DeactivatePromotion(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *PromotionsHandler) ListEligibleProducts(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
