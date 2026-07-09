package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/google/uuid"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

type CouponHandler struct{ db *db.DB }

func NewCouponHandler(database *db.DB) *CouponHandler { return &CouponHandler{db: database} }

func (h *CouponHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	var req struct{ Code, DiscountType string; DiscountValue int; Currency, CourseID string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }
	if req.Code == "" { respondWithError(w, http.StatusBadRequest, "Code required"); return }
	if req.Currency == "" { req.Currency = "usd" }

	var courseID *uuid.UUID
	if req.CourseID != "" { courseID = parseUUID(req.CourseID) }

	coupon, err := h.db.Queries.CreateCoupon(r.Context(), gen.CreateCouponParams{
		TenantID: *tenantID, Upper: req.Code, DiscountType: req.DiscountType,
		DiscountValue: int32(req.DiscountValue), Currency: req.Currency, CourseID: courseID,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed: "+err.Error()); return }
	respondWithJSON(w, http.StatusCreated, coupon)
}

func (h *CouponHandler) Delete(w http.ResponseWriter, r *http.Request) {
	couponID := parseUUID(mux.Vars(r)["id"])
	if couponID == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }
	if err := h.db.Queries.DeleteCoupon(r.Context(), *couponID); err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *CouponHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }
	page, limit := parsePagination(r)
	coupons, err := h.db.Queries.ListCouponsByTenant(r.Context(), gen.ListCouponsByTenantParams{
		TenantID: *tenantID, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"coupons": coupons, "page": page, "limit": limit})
}

func (h *CouponHandler) Validate(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	var req struct{ Code string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }

	coupon, err := h.db.Queries.ValidateCoupon(r.Context(), gen.ValidateCouponParams{
		TenantID: *tenantID, Upper: req.Code,
	})
	if err != nil { respondWithError(w, http.StatusNotFound, "Invalid or expired coupon"); return }
	respondWithJSON(w, http.StatusOK, coupon)
}
