package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
	"mycourses/internal/syslog"
)

type BundlesHandler struct{ db *db.DB; syslog *syslog.Logger }

func NewBundlesHandler(database *db.DB, sysLogger *syslog.Logger) *BundlesHandler {
	return &BundlesHandler{db: database, syslog: sysLogger}
}

func (h *BundlesHandler) ListBundles(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() { respondWithJSON(w, http.StatusOK, []interface{}{}); return }

	bundles, _ := h.db.Queries.ListCreditBundlesByTenant(r.Context(), &tenantID)
	respondWithJSON(w, http.StatusOK, bundles)
}

func (h *BundlesHandler) ListBundlesPublic(w http.ResponseWriter, r *http.Request) {
	bundles, _ := h.db.Queries.ListCreditBundlesPublic(r.Context())
	respondWithJSON(w, http.StatusOK, bundles)
}

func (h *BundlesHandler) CreateBundle(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	var req struct {
		Name string `json:"name"`
		Description string `json:"description"`
		Credits int `json:"credits"`
		PriceCents int64 `json:"priceCents"`
		Currency string `json:"currency"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Currency == "" { req.Currency = "usd" }

	bundle, err := h.db.Queries.CreateCreditBundle(r.Context(), gen.CreateCreditBundleParams{
		TenantID: &tenantID, Name: req.Name, Description: req.Description,
		Credits: int32(req.Credits), PriceCents: req.PriceCents, Currency: req.Currency,
		IsActive: true, SortOrder: 0,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusCreated, bundle)
}

func (h *BundlesHandler) UpdateBundle(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	var req struct {
		Name string `json:"name"`
		Description string `json:"description"`
		Credits int `json:"credits"`
		PriceCents int64 `json:"priceCents"`
		IsActive bool `json:"isActive"`
		SortOrder int `json:"sortOrder"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	bundle, err := h.db.Queries.UpdateCreditBundle(r.Context(), gen.UpdateCreditBundleParams{
		ID: *id, Name: req.Name, Description: req.Description,
		Credits: int32(req.Credits), PriceCents: req.PriceCents,
		IsActive: req.IsActive, SortOrder: int32(req.SortOrder),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusOK, bundle)
}

func (h *BundlesHandler) DeleteBundle(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	h.db.Queries.DeleteCreditBundle(r.Context(), *id)
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
